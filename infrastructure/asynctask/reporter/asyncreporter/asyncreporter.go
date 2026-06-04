package asyncreporter

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

const (
	defaultBufferSize      = 4096
	defaultWorkers         = 1
	defaultReportTimeout   = time.Second
	defaultShutdownTimeout = 3 * time.Second
)

type DropPolicy string

const (
	// DropNewest drops the incoming event when the async buffer is full.
	DropNewest DropPolicy = "drop_newest"

	// DropOldest removes one buffered event and enqueues the incoming event when the async buffer is full.
	DropOldest DropPolicy = "drop_oldest"

	// Block waits until there is buffer capacity or the caller context is canceled.
	Block DropPolicy = "block"
)

type Config struct {
	BufferSize      int
	Workers         int
	ReportTimeout   time.Duration
	ShutdownTimeout time.Duration
	OnFull          DropPolicy
	OnError         func(error, asynctask.TaskEvent)
	OnDrop          func(asynctask.TaskEvent)
}

type AsyncReporter struct {
	target asynctask.TaskReporter
	config Config

	events chan asynctask.TaskEvent
	done   chan struct{}

	startOnce sync.Once
	stopOnce  sync.Once
	workers   sync.WaitGroup

	accepted atomic.Int64
	dropped  atomic.Int64
	pending  atomic.Int64
	stopped  atomic.Bool
}

func New(target asynctask.TaskReporter, config ...Config) *AsyncReporter {
	cfg := Config{}
	if len(config) > 0 {
		cfg = config[0]
	}
	cfg = normalizeConfig(cfg)
	return &AsyncReporter{
		target: target,
		config: cfg,
		events: make(chan asynctask.TaskEvent, cfg.BufferSize),
		done:   make(chan struct{}),
	}
}

func (r *AsyncReporter) ImplName() scene.ImplName {
	if r == nil || r.target == nil {
		return asynctask.Lens.ImplName("TaskReporter", "async")
	}
	return asynctask.Lens.ImplName("TaskReporter", "async."+r.target.ImplName().Identifier())
}

func (r *AsyncReporter) ReportTaskEvent(ctx context.Context, event asynctask.TaskEvent) error {
	if r == nil || r.target == nil {
		return nil
	}
	taskID := strings.TrimSpace(event.TaskID)
	if taskID == "" {
		return asynctask.ErrInvalidQueueTask
	}
	event.TaskID = taskID
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = time.Now()
	}
	if r.stopped.Load() {
		r.drop(event)
		return nil
	}

	r.Start()
	switch r.config.OnFull {
	case Block:
		if ctx == nil {
			ctx = context.Background()
		}
		return r.enqueueBlock(ctx, event)
	case DropOldest:
		if !r.enqueue(event) {
			r.dropOldest()
			if !r.enqueue(event) {
				r.drop(event)
			}
		}
	default:
		if !r.enqueue(event) {
			r.drop(event)
		}
	}
	return nil
}

func (r *AsyncReporter) Start() {
	if r == nil || r.target == nil {
		return
	}
	if r.stopped.Load() {
		return
	}
	r.startOnce.Do(func() {
		for i := 0; i < r.config.Workers; i++ {
			r.workers.Add(1)
			go r.worker()
		}
	})
}

func (r *AsyncReporter) Stop(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if r.config.ShutdownTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.config.ShutdownTimeout)
		defer cancel()
	}

	r.stopOnce.Do(func() {
		r.stopped.Store(true)
		close(r.done)
	})

	done := make(chan struct{})
	go func() {
		r.workers.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *AsyncReporter) Flush(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		if r.pending.Load() == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func (r *AsyncReporter) Accepted() int64 {
	if r == nil {
		return 0
	}
	return r.accepted.Load()
}

func (r *AsyncReporter) Dropped() int64 {
	if r == nil {
		return 0
	}
	return r.dropped.Load()
}

func (r *AsyncReporter) worker() {
	defer r.workers.Done()
	for {
		select {
		case event := <-r.events:
			r.report(event)
		case <-r.done:
			r.drain()
			return
		}
	}
}

func (r *AsyncReporter) drain() {
	for {
		select {
		case event := <-r.events:
			r.report(event)
		default:
			return
		}
	}
}

func (r *AsyncReporter) report(event asynctask.TaskEvent) {
	defer r.pending.Add(-1)
	ctx := context.Background()
	if r.config.ReportTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.config.ReportTimeout)
		defer cancel()
	}
	if err := r.target.ReportTaskEvent(ctx, event); err != nil && r.config.OnError != nil {
		r.config.OnError(err, event)
	}
}

func (r *AsyncReporter) dropOldest() {
	select {
	case event := <-r.events:
		r.pending.Add(-1)
		r.drop(event)
	default:
	}
}

func (r *AsyncReporter) enqueue(event asynctask.TaskEvent) bool {
	r.pending.Add(1)
	select {
	case r.events <- event:
		r.accepted.Add(1)
		return true
	default:
		r.pending.Add(-1)
		return false
	}
}

func (r *AsyncReporter) enqueueBlock(ctx context.Context, event asynctask.TaskEvent) error {
	r.pending.Add(1)
	select {
	case r.events <- event:
		r.accepted.Add(1)
		return nil
	case <-ctx.Done():
		r.pending.Add(-1)
		return ctx.Err()
	case <-r.done:
		r.pending.Add(-1)
		r.drop(event)
		return nil
	}
}

func (r *AsyncReporter) drop(event asynctask.TaskEvent) {
	r.dropped.Add(1)
	if r.config.OnDrop != nil {
		r.config.OnDrop(event)
	}
}

func normalizeConfig(config Config) Config {
	if config.BufferSize <= 0 {
		config.BufferSize = defaultBufferSize
	}
	if config.Workers <= 0 {
		config.Workers = defaultWorkers
	}
	if config.ReportTimeout == 0 {
		config.ReportTimeout = defaultReportTimeout
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = defaultShutdownTimeout
	}
	if config.OnFull == "" {
		config.OnFull = DropNewest
	}
	return config
}
