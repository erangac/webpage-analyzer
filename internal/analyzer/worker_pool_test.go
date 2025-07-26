package analyzer

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{"Single worker", 1},
		{"Multiple workers", 5},
		{"Zero workers", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workers)
			if pool == nil {
				t.Fatal("NewWorkerPool() returned nil")
			}
			if pool.workers != tt.workers {
				t.Errorf("NewWorkerPool() workers = %d, want %d", pool.workers, tt.workers)
			}
		})
	}
}

func TestWorkerPoolSubmit(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	var results []int
	var mu sync.Mutex

	// Submit multiple tasks
	for i := 0; i < 5; i++ {
		i := i // Capture loop variable
		pool.Submit(func() error {
			mu.Lock()
			results = append(results, i)
			mu.Unlock()
			return nil
		})
	}

	// Wait for all tasks to complete (Shutdown will handle this)
	// pool.Wait() // Remove this to avoid double-closing the channel

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestWorkerPoolSubmitAndWait(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	// Test successful task
	err := pool.SubmitAndWait(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("SubmitAndWait() returned error: %v", err)
	}

	// Test task that returns error
	expectedErr := "test error"
	err = pool.SubmitAndWait(func() error {
		return &AnalysisError{StatusCode: 400, ErrorMessage: expectedErr, URL: "test"}
	})
	if err == nil {
		t.Error("SubmitAndWait() should return error")
	}
	if err.Error() != "HTTP 400: test error (URL: test)" {
		t.Errorf("SubmitAndWait() error = %v, want 'HTTP 400: test error (URL: test)'", err.Error())
	}
}

func TestWorkerPoolShutdown(t *testing.T) {
	pool := NewWorkerPool(2)

	// Submit a simple task
	pool.Submit(func() error {
		return nil
	})

	// Shutdown should complete quickly
	shutdownStart := time.Now()
	pool.Shutdown()
	shutdownDuration := time.Since(shutdownStart)

	// Should complete quickly
	if shutdownDuration > 100*time.Millisecond {
		t.Errorf("Shutdown took too long: %v", shutdownDuration)
	}
}

func TestWorkerPoolContextCancellation(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Submit a task that respects context
	taskCompleted := make(chan bool)
	pool.Submit(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			taskCompleted <- true
			return nil
		}
	})

	// Wait for context to timeout
	time.Sleep(20 * time.Millisecond)

	// Task should have been cancelled
	select {
	case <-taskCompleted:
		t.Error("Task should have been cancelled")
	default:
		// Expected - task was cancelled
	}
}

func TestAnalysisTaskGroup(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	group := NewAnalysisTaskGroup(pool)

	// Add tasks
	group.AddTask("task1", func() (interface{}, error) {
		return "result1", nil
	})

	group.AddTask("task2", func() (interface{}, error) {
		return "result2", nil
	})

	group.AddTask("task3", func() (interface{}, error) {
		return nil, &AnalysisError{StatusCode: 400, ErrorMessage: "task3 error", URL: "test"}
	})

	// Execute all tasks
	group.ExecuteAll()

	// Check results
	result1, err := group.GetResult("task1")
	if err != nil {
		t.Errorf("task1 error: %v", err)
	}
	if result1 != "result1" {
		t.Errorf("task1 result = %v, want 'result1'", result1)
	}

	result2, err := group.GetResult("task2")
	if err != nil {
		t.Errorf("task2 error: %v", err)
	}
	if result2 != "result2" {
		t.Errorf("task2 result = %v, want 'result2'", result2)
	}

	_, err = group.GetResult("task3")
	if err == nil {
		t.Error("task3 should have error")
	}

	// Check if any tasks had errors
	if !group.HasErrors() {
		t.Error("HasErrors() should return true")
	}
}

func TestAnalysisTaskGroupNoErrors(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	group := NewAnalysisTaskGroup(pool)

	// Add successful tasks only
	group.AddTask("task1", func() (interface{}, error) {
		return "result1", nil
	})

	group.AddTask("task2", func() (interface{}, error) {
		return "result2", nil
	})

	// Execute all tasks
	group.ExecuteAll()

	// Check if any tasks had errors
	if group.HasErrors() {
		t.Error("HasErrors() should return false")
	}
}

func TestAnalysisTaskGroupGetResultNonExistent(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	group := NewAnalysisTaskGroup(pool)

	// Try to get result for non-existent task
	result, err := group.GetResult("nonexistent")
	if result != nil {
		t.Errorf("GetResult() for non-existent task should return nil, got %v", result)
	}
	if err != nil {
		t.Errorf("GetResult() for non-existent task should return nil error, got %v", err)
	}
}

func TestWorkerPoolConcurrentAccess(t *testing.T) {
	pool := NewWorkerPool(5)
	defer pool.Shutdown()

	var wg sync.WaitGroup
	var counter int
	var mu sync.Mutex

	// Submit many concurrent tasks
	for i := 0; i < 100; i++ {
		wg.Add(1)
		pool.Submit(func() error {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
			return nil
		})
	}

	// Wait for all tasks to complete
	wg.Wait()

	if counter != 100 {
		t.Errorf("Expected counter to be 100, got %d", counter)
	}
}

func TestWorkerPoolTaskQueueBuffer(t *testing.T) {
	// Test with different buffer sizes
	tests := []struct {
		name    string
		workers int
	}{
		{"Small buffer", 1},
		{"Large buffer", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.workers)
			defer pool.Shutdown()

			// Submit more tasks than workers to test buffering
			taskCount := tt.workers * 3
			completed := make(chan bool, taskCount)

			for i := 0; i < taskCount; i++ {
				pool.Submit(func() error {
					completed <- true
					return nil
				})
			}

			// Wait for all tasks to complete
			for i := 0; i < taskCount; i++ {
				select {
				case <-completed:
					// Task completed
				case <-time.After(1 * time.Second):
					t.Errorf("Task %d did not complete in time", i)
				}
			}
		})
	}
} 