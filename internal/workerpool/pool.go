package workerpool

import (
	"context"
	"fmt"
	"sync"
)

type Job struct {
	Symbol string
	Price  string
}

type Workerpool struct {
	NumWorkers int
	JobQueue   chan Job
	wg sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *Workerpool {
	return &Workerpool {
		NumWorkers: numWorkers,
		JobQueue:   make(chan Job, 100),
	}
}

func (wp *Workerpool) Start(ctx context.Context) {
	for i := 0; i <= wp.NumWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *Workerpool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	fmt.Printf("Worker %d started and waiting for jobs...\n", id)

	for {
		select {
		case job, ok := <-wp.JobQueue:
			if !ok {
				return // Channel closed, worker shuts down
			}
			// This is where your Trading Strategy logic would go!
			wp.processTrade(id, job)
		case <-ctx.Done():
			return // Context cancelled, shut down
		}
	}
}

func (wp *Workerpool) processTrade(workerID int, job Job) {
	fmt.Printf("[Worker %d] Analyzing %s: Price is %s\n", workerID, job.Symbol, job.Price)
}

func (wp *Workerpool) Wait() {
	close(wp.JobQueue)
	wp.wg.Wait()
}