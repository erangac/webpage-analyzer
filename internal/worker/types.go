package worker

import "fmt"

// TaskFunc represents a unit of work to be executed by a worker.
// It should return an error if the task fails, or nil if successful.
type TaskFunc func() error

// Task represents a unit of work to be executed.
type Task = TaskFunc

// WorkerPoolManager defines the interface for worker pool operations.
type WorkerPoolManager interface {
	Submit(task Task)
	SubmitAndWait(task Task) error
	Wait()
	Shutdown()
}

// AnalysisTask represents a specific analysis task with result.
type AnalysisTask struct {
	Name   string
	Task   func() (interface{}, error)
	Result interface{}
	Error  error
}

// AnalysisTaskGroup manages a group of related analysis tasks.
type AnalysisTaskGroup struct {
	tasks []*AnalysisTask
	pool  *WorkerPool
}

// AnalysisError represents an error during analysis (for testing purposes).
type AnalysisError struct {
	StatusCode   int
	ErrorMessage string
	URL          string
}

// Error implements the error interface.
func (e *AnalysisError) Error() string {
	return fmt.Sprintf("HTTP %d: %s (URL: %s)", e.StatusCode, e.ErrorMessage, e.URL)
}
