package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/barbodimani81/trading-bot.git/internal/platform/redis"
	"github.com/barbodimani81/trading-bot.git/internal/workerpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Println("📊 Processor Metrics running on :2113")
        if err := http.ListenAndServe(":2113", nil); err != nil {
             log.Fatal("Metrics server failed:", err)
        }
    }()
	
	rdb, err := redis.NewRedisClient("localhost:6379")
    if err != nil {
        log.Fatalf("Processor Redis Init Failed: %v", err)
    }
	
	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("market_data", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal(err)
	}

	wp := workerpool.NewWorkerPool(5)
    wp.Redis = rdb
	ctx, cancel := context.WithCancel(context.Background())
	wp.Start(ctx)

	go func() {
		for msg := range partitionConsumer.Messages() {
			wp.JobQueue <- workerpool.Job{
				Symbol: string(msg.Key),
				Price:  string(msg.Value),
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	
	cancel()
	wp.Wait()
	log.Println("Processor shut down clean.")
}
