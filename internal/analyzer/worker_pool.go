package analyzer

import (
	"context"
	"log"
	"sync"
)

// TaskFunc represents a unit of work to be executed by a worker.
// It should return an error if the task fails, or nil if successful.
type TaskFunc func() error

// Task represents a unit of work to be executed.
type Task = TaskFunc

// WorkerPool manages a pool of workers for concurrent task execution.
type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, workers*2), // Buffer for better performance.
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start workers.
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker is the main worker goroutine that processes tasks.
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				return // Channel closed, exit worker.
			}
			if err := task(); err != nil {
				// Log error but continue processing other tasks.
				// In a production system, you might want to handle errors differently.
				log.Printf("Worker task failed: %v", err)
			}
		case <-wp.ctx.Done():
			return // Context cancelled, exit worker.
		}
	}
}

// Submit adds a task to the worker pool.
func (wp *WorkerPool) Submit(task Task) {
	select {
	case wp.taskQueue <- task:
		// Task submitted successfully.
	case <-wp.ctx.Done():
		// Pool is shutting down.
	}
}

// SubmitAndWait submits a task and waits for it to complete.
func (wp *WorkerPool) SubmitAndWait(task Task) error {
	resultChan := make(chan error, 1)

	wp.Submit(func() error {
		err := task()
		resultChan <- err
		return err
	})

	return <-resultChan
}

// Wait waits for all submitted tasks to complete.
func (wp *WorkerPool) Wait() {
	close(wp.taskQueue) // Signal workers to stop accepting new tasks.
	wp.wg.Wait()        // Wait for all workers to finish.
}

// Shutdown gracefully shuts down the worker pool.
func (wp *WorkerPool) Shutdown() {
	wp.cancel() // Cancel context to stop workers.
	close(wp.taskQueue) // Close task queue to signal workers to stop
	wp.wg.Wait() // Wait for all workers to finish.
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

// NewAnalysisTaskGroup creates a new task group for analysis.
func NewAnalysisTaskGroup(pool *WorkerPool) *AnalysisTaskGroup {
	return &AnalysisTaskGroup{
		tasks: make([]*AnalysisTask, 0),
		pool:  pool,
	}
}

// AddTask adds a task to the group.
func (atg *AnalysisTaskGroup) AddTask(name string, task func() (interface{}, error)) {
	analysisTask := &AnalysisTask{
		Name: name,
		Task: task,
	}
	atg.tasks = append(atg.tasks, analysisTask)
}

// ExecuteAll runs all tasks in parallel and waits for completion.
func (atg *AnalysisTaskGroup) ExecuteAll() {
	var wg sync.WaitGroup

	for _, task := range atg.tasks {
		wg.Add(1)
		atg.pool.Submit(func() error {
			defer wg.Done()
			result, err := task.Task()
			task.Result = result
			task.Error = err
			if err != nil {
				log.Printf("Analysis task '%s' failed: %v", task.Name, err)
			}
			return err
		})
	}

	wg.Wait()
}

// GetResult retrieves the result of a specific task.
func (atg *AnalysisTaskGroup) GetResult(taskName string) (interface{}, error) {
	for _, task := range atg.tasks {
		if task.Name == taskName {
			return task.Result, task.Error
		}
	}
	return nil, nil
}

// HasErrors checks if any tasks had errors.
func (atg *AnalysisTaskGroup) HasErrors() bool {
	for _, task := range atg.tasks {
		if task.Error != nil {
			return true
		}
	}
	return false
}
