package analyzer

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NotNil(t, pool, "NewWorkerPool() should not return nil")
			assert.Equal(t, tt.workers, pool.workers, "Worker count should match")
		})
	}
}

func TestWorkerPoolSubmitAndWait(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	// Test successful task
	err := pool.SubmitAndWait(func() error {
		return nil
	})
	assert.NoError(t, err, "SubmitAndWait() should not return error for successful task")

	// Test task that returns error
	expectedErr := "test error"
	err = pool.SubmitAndWait(func() error {
		return &AnalysisError{StatusCode: 400, ErrorMessage: expectedErr, URL: "test"}
	})
	require.Error(t, err, "SubmitAndWait() should return error for failed task")
	assert.Equal(t, "HTTP 400: test error (URL: test)", err.Error(), "Error message should match expected")
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
	assert.Less(t, shutdownDuration, 100*time.Millisecond, "Shutdown should complete quickly")
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
	assert.NoError(t, err, "task1 should not have error")
	assert.Equal(t, "result1", result1, "task1 result should match")

	result2, err := group.GetResult("task2")
	assert.NoError(t, err, "task2 should not have error")
	assert.Equal(t, "result2", result2, "task2 result should match")

	_, err = group.GetResult("task3")
	assert.Error(t, err, "task3 should have error")

	// Check if any tasks had errors
	assert.True(t, group.HasErrors(), "HasErrors() should return true when tasks have errors")
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
	assert.False(t, group.HasErrors(), "HasErrors() should return false when no tasks have errors")
}

func TestAnalysisTaskGroupGetResultNonExistent(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Shutdown()

	group := NewAnalysisTaskGroup(pool)

	// Try to get result for non-existent task
	result, err := group.GetResult("nonexistent")
	assert.Nil(t, result, "GetResult() for non-existent task should return nil result")
	assert.Nil(t, err, "GetResult() for non-existent task should return nil error")
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

	assert.Equal(t, 100, counter, "Counter should be 100 after all tasks complete")
}
