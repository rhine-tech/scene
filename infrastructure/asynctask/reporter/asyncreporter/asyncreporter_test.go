package asyncreporter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

type fakeReporter struct {
	lock    sync.Mutex
	events  []asynctask.TaskEvent
	err     error
	block   chan struct{}
	entered chan struct{}
}

func newFakeReporter() *fakeReporter {
	return &fakeReporter{}
}

func (r *fakeReporter) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskReporter", "fake")
}

func (r *fakeReporter) ReportTaskEvent(_ context.Context, event asynctask.TaskEvent) error {
	if r.block != nil {
		if r.entered != nil {
			select {
			case r.entered <- struct{}{}:
			default:
			}
		}
		<-r.block
	}
	r.lock.Lock()
	r.events = append(r.events, event)
	r.lock.Unlock()
	return r.err
}

func (r *fakeReporter) waitCount(t *testing.T, count int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		r.lock.Lock()
		current := len(r.events)
		r.lock.Unlock()
		if current >= count {
			return
		}
		time.Sleep(time.Millisecond)
	}
	r.lock.Lock()
	current := len(r.events)
	r.lock.Unlock()
	t.Fatalf("timeout waiting events, got %d want %d", current, count)
}

func (r *fakeReporter) snapshot() []asynctask.TaskEvent {
	r.lock.Lock()
	defer r.lock.Unlock()
	events := make([]asynctask.TaskEvent, len(r.events))
	copy(events, r.events)
	return events
}

func TestAsyncReporterReportsInBackground(t *testing.T) {
	target := newFakeReporter()
	reporter := New(target, Config{BufferSize: 4, Workers: 1, ReportTimeout: time.Second})

	err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-1"})
	if err != nil {
		t.Fatalf("report task event failed: %v", err)
	}
	target.waitCount(t, 1)
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	events := target.snapshot()
	if events[0].TaskID != "task-1" || events[0].UpdatedAt.IsZero() {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestAsyncReporterReportDoesNotBlockWhenTargetBlocked(t *testing.T) {
	target := newFakeReporter()
	target.block = make(chan struct{})
	target.entered = make(chan struct{}, 1)
	reporter := New(target, Config{BufferSize: 1, Workers: 1, OnFull: DropNewest})

	reportFast(t, reporter, "task-1")
	<-target.entered
	reportFast(t, reporter, "task-2")
	reportFast(t, reporter, "task-3")

	if reporter.Dropped() != 1 {
		t.Fatalf("unexpected dropped count: %d", reporter.Dropped())
	}
	close(target.block)
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestAsyncReporterDropNewest(t *testing.T) {
	target := newFakeReporter()
	target.block = make(chan struct{})
	target.entered = make(chan struct{}, 1)
	reporter := New(target, Config{BufferSize: 1, Workers: 1, OnFull: DropNewest})

	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-1"}); err != nil {
		t.Fatalf("report first event failed: %v", err)
	}
	<-target.entered
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-2"}); err != nil {
		t.Fatalf("report second event failed: %v", err)
	}
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-3"}); err != nil {
		t.Fatalf("report third event failed: %v", err)
	}
	if reporter.Dropped() != 1 {
		t.Fatalf("unexpected dropped count: %d", reporter.Dropped())
	}
	close(target.block)
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	events := target.snapshot()
	if len(events) != 2 || events[0].TaskID != "task-1" || events[1].TaskID != "task-2" {
		t.Fatalf("unexpected retained events: %+v", events)
	}
}

func reportFast(t *testing.T, reporter *AsyncReporter, taskID string) {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: taskID})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("report task event failed: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("ReportTaskEvent blocked for task %s", taskID)
	}
}

func TestAsyncReporterDropOldest(t *testing.T) {
	target := newFakeReporter()
	target.block = make(chan struct{})
	target.entered = make(chan struct{}, 1)
	reporter := New(target, Config{BufferSize: 1, Workers: 1, OnFull: DropOldest})

	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-1"}); err != nil {
		t.Fatalf("report first event failed: %v", err)
	}
	<-target.entered
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-2"}); err != nil {
		t.Fatalf("report second event failed: %v", err)
	}
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-3"}); err != nil {
		t.Fatalf("report third event failed: %v", err)
	}
	if reporter.Dropped() != 1 {
		t.Fatalf("unexpected dropped count: %d", reporter.Dropped())
	}
	close(target.block)
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	events := target.snapshot()
	if len(events) != 2 || events[0].TaskID != "task-1" || events[1].TaskID != "task-3" {
		t.Fatalf("unexpected retained events: %+v", events)
	}
}

func TestAsyncReporterBlockRespectsContext(t *testing.T) {
	target := newFakeReporter()
	target.block = make(chan struct{})
	target.entered = make(chan struct{}, 1)
	reporter := New(target, Config{BufferSize: 1, Workers: 1, OnFull: Block})

	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-1"}); err != nil {
		t.Fatalf("report first event failed: %v", err)
	}
	<-target.entered
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-2"}); err != nil {
		t.Fatalf("report second event failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err := reporter.ReportTaskEvent(ctx, asynctask.TaskEvent{TaskID: "task-3"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected block error: %v", err)
	}
	close(target.block)
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestAsyncReporterOnError(t *testing.T) {
	target := newFakeReporter()
	target.err = errors.New("report failed")
	errCh := make(chan error, 1)
	reporter := New(target, Config{
		BufferSize: 1,
		Workers:    1,
		OnError: func(err error, event asynctask.TaskEvent) {
			errCh <- err
		},
	})

	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: "task-1"}); err != nil {
		t.Fatalf("report task event failed: %v", err)
	}
	select {
	case err := <-errCh:
		if !errors.Is(err, target.err) {
			t.Fatalf("unexpected on error value: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting on error")
	}
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}

func TestAsyncReporterFlush(t *testing.T) {
	target := newFakeReporter()
	reporter := New(target, Config{BufferSize: 4, Workers: 1})
	for _, taskID := range []string{"task-1", "task-2"} {
		if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{TaskID: taskID}); err != nil {
			t.Fatalf("report task event failed: %v", err)
		}
	}
	if err := reporter.Flush(context.Background()); err != nil {
		t.Fatalf("flush failed: %v", err)
	}
	target.waitCount(t, 2)
	if err := reporter.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}
