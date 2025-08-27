package longrunning

import (
	"context"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRunAndCancel verifies that a task can be started and then successfully cancelled.
func TestRunAndCancel(t *testing.T) {
	dispatcher := NewLongRunningTaskDispatcher()

	// Use an atomic counter to track if the task is running
	var runningCounter int32
	// Use a WaitGroup to signal that the task has entered its main loop
	var wg sync.WaitGroup
	wg.Add(1)

	taskFunc := func(ctx context.Context) error {
		// Signal that the task has started its execution loop
		if atomic.CompareAndSwapInt32(&runningCounter, 0, 1) {
			wg.Done()
		}
		// Block until the context is cancelled
		<-ctx.Done()
		atomic.StoreInt32(&runningCounter, 0)
		return nil
	}

	// Run the task
	task, err := dispatcher.Run(taskFunc)
	require.NoError(t, err)
	require.NotEmpty(t, task.Identifier())

	// Wait for the task to confirm it's running, with a timeout
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// Task started successfully
		require.Equal(t, int32(1), atomic.LoadInt32(&runningCounter), "Task should be running")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for task to start")
	}

	// Cancel the task
	err = dispatcher.Cancel(task.Identifier())
	require.NoError(t, err)

	// Give the task goroutine a moment to exit and clean up
	time.Sleep(100 * time.Millisecond)

	// Verify the task is no longer running
	require.Equal(t, int32(0), atomic.LoadInt32(&runningCounter), "Task should have stopped")

	// Verify that trying to cancel again returns an error because the task is gone
	err = dispatcher.Cancel(task.Identifier())
	require.ErrorIs(t, err, asynctask.ErrTaskNotFound, "Cancelled task should not be found again")
}

// TestTaskRestart verifies that a task function restarts after it completes.
func TestTaskRestart(t *testing.T) {
	dispatcher := NewLongRunningTaskDispatcher()

	// Use an atomic counter to track how many times the task has run
	var executionCount int32

	taskFunc := func(ctx context.Context) error {
		atomic.AddInt32(&executionCount, 1)
		// The function returns immediately, and the dispatcher should restart it after a delay.
		return nil
	}

	task, err := dispatcher.Run(taskFunc)
	require.NoError(t, err)

	// Use t.Cleanup to ensure the task is cancelled after the test, preventing leaks.
	t.Cleanup(func() {
		_ = dispatcher.Cancel(task.Identifier())
	})

	// Wait for the task to execute multiple times.
	// Since there's a 1-second sleep between executions, waiting for 3 executions
	// should take a bit over 2 seconds. We'll wait for up to 3 seconds.
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&executionCount) >= 3
	}, 3*time.Second, 100*time.Millisecond, "Task should have run at least 3 times")
}

// TestCancelNonExistentTask verifies that cancelling a non-existent task returns the correct error.
func TestCancelNonExistentTask(t *testing.T) {
	dispatcher := NewLongRunningTaskDispatcher()
	err := dispatcher.Cancel("some-id-that-does-not-exist")
	require.ErrorIs(t, err, asynctask.ErrTaskNotFound)
}

// TestPanicRecovery verifies that the dispatcher recovers from a panic in the task function
// and restarts the task.
func TestPanicRecovery(t *testing.T) {
	dispatcher := NewLongRunningTaskDispatcher()

	var executionCount int32

	taskFunc := func(ctx context.Context) error {
		count := atomic.AddInt32(&executionCount, 1)
		// Panic on the first run
		if count == 1 {
			panic("simulating a panic")
		}
		// Succeed on subsequent runs
		return nil
	}

	task, err := dispatcher.Run(taskFunc)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = dispatcher.Cancel(task.Identifier())
	})

	// Check that the task runs again after panicking.
	// It should panic once, sleep for 1s, then run again successfully.
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&executionCount) >= 2
	}, 3*time.Second, 100*time.Millisecond, "Task should restart and run a second time after panicking")
}
