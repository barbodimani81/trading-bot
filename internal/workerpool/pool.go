package workerpool

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/barbodimani81/trading-bot.git/internal/platform/redis"
)

type Job struct {
	Symbol string
	Price  string
}

type Workerpool struct {
	NumWorkers int
	JobQueue   chan Job
	wg sync.WaitGroup
	Redis      *redis.RedisClient
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
			wp.processTrade(id, job, wp.Redis)
		case <-ctx.Done():
			return // Context cancelled, shut down
		}
	}
}

func (wp *Workerpool) processTrade(workerID int, job Job, rdb *redis.RedisClient) {
	ctx := context.Background()
	key := fmt.Sprintf("history:%s", job.Symbol)

	// 1. Add current price to the Redis List (Push to the head)
	rdb.Client.LPush(ctx, key, job.Price)

	// 2. Keep only the latest 10 prices (Trim the tail)
	rdb.Client.LTrim(ctx, key, 0, 9)

	// 3. Get the whole list to calculate the average
	history, _ := rdb.Client.LRange(ctx, key, 0, -1).Result()
	
	if len(history) < 10 {
		fmt.Printf("[Worker %d] %s: Collecting data... (%d/10)\n", workerID, job.Symbol, len(history))
		return
	}

	// 4. Calculate Average
	var sum float64
	for _, pStr := range history {
		p, _ := strconv.ParseFloat(pStr, 64)
		sum += p
	}
	avg := sum / float64(len(history))
	current, _ := strconv.ParseFloat(job.Price, 64)

	// 5. Strategy: Buy if price is 1% below 10-period average
	threshold := avg * 0.99
	if current < threshold {
		fmt.Printf("ALERT [Worker %d]: %s is DIPPING! Price: %f | Avg: %f\n", workerID, job.Symbol, current, avg)
	} else {
		fmt.Printf("stable [Worker %d]: %s at %f (Avg: %f)\n", workerID, job.Symbol, current, avg)
	}
}

func (wp *Workerpool) Wait() {
	close(wp.JobQueue)
	wp.wg.Wait()
}