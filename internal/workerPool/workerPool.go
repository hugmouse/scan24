package workerPool

import (
	"fmt"
	"github.com/hugmouse/scan24/internal/job"
	"sync"
)

type WorkerStats struct {
	WorkerID            int
	JobsProcessed       int
	ErrorsEncountered   int
	JobsProcessedByType map[string]int
}

// WorkerPool manages the creation, distribution, and monitoring of workers
type WorkerPool struct {
	jobQueue       chan job.Job
	results        chan job.Result
	workerStats    chan WorkerStats // Channel to receive stats from workers upon completion
	workerCount    int
	busyWorkers    int
	freeWorkers    int
	mu             sync.Mutex // Mutex for worker counters
	wg             sync.WaitGroup
	allWorkerStats []WorkerStats
	statsMu        sync.Mutex // Mutex for stat counters
}

// NewWorkerPool creates and initializes a new WorkerPool
func NewWorkerPool(workers, queueSize int) *WorkerPool {
	return &WorkerPool{
		jobQueue:    make(chan job.Job, queueSize),
		results:     make(chan job.Result, queueSize),
		workerStats: make(chan WorkerStats, workers),
		workerCount: workers,
		busyWorkers: 0,
		freeWorkers: workers,
	}
}

// worker represents an individual worker in the pool
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	workerStats := WorkerStats{WorkerID: id}
	workerStats.JobsProcessedByType = make(map[string]int)

	for _job := range wp.jobQueue {

		wp.mu.Lock()
		wp.busyWorkers++
		wp.freeWorkers--
		wp.mu.Unlock()

		fmt.Printf("Worker %d starting %s job %s\n", id, _job.GetType(), _job.GetID())

		jobResult := _job.Execute()
		jobResult.WorkerID = id

		wp.results <- jobResult //

		workerStats.JobsProcessed++
		// workerStats.JobsProcessedByType[job.GetType()]++

		if jobResult.Error != nil {
			workerStats.ErrorsEncountered++
		}

		fmt.Printf("Worker %d finished %s job %s (Result: %#+v, Error: %#+v)\n",
			id, _job.GetType(), _job.GetID(), jobResult.Data, jobResult.Error)

		wp.mu.Lock()
		wp.busyWorkers--
		wp.freeWorkers++
		wp.mu.Unlock()
	}

	wp.workerStats <- workerStats
}

// StartWorkers starts the specified number of worker goroutines and a goroutine to collect results
func (wp *WorkerPool) StartWorkers() {
	for i := 1; i <= wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	go func() {
		wp.wg.Wait()
		close(wp.workerStats)
	}()
}

// SubmitJob sends a job to the job queue
func (wp *WorkerPool) SubmitJob(job job.Job) {
	wp.jobQueue <- job
}

// Results returns the channel where job results are sent
func (wp *WorkerPool) Results() <-chan job.Result {
	return wp.results
}

// Close closes the job queue, signaling workers to shut down after processing
// all remaining jobs, and then collects all worker statistics
func (wp *WorkerPool) Close() {
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.results)

	wp.statsMu.Lock()
	defer wp.statsMu.Unlock()
	for stats := range wp.workerStats {
		wp.allWorkerStats = append(wp.allWorkerStats, stats)
	}
}

// GetStatus returns the current status of the worker pool
func (wp *WorkerPool) GetStatus() (total, busy, free int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.workerCount, wp.busyWorkers, wp.freeWorkers
}

// GetAllWorkerStats returns the collected statistics from all workers after the pool is closed
func (wp *WorkerPool) GetAllWorkerStats() []WorkerStats {
	wp.statsMu.Lock()
	defer wp.statsMu.Unlock()
	return wp.allWorkerStats
}
