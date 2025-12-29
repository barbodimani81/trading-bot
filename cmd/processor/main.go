package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/barbodimani81/trading-bot.git/internal/workerpool"
)

func main() {
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
