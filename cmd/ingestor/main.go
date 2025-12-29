package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/barbodimani81/trading-bot.git/internal/platform/kafka"
	"github.com/barbodimani81/trading-bot.git/internal/platform/redis"
	"github.com/barbodimani81/trading-bot.git/internal/workerpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/adshao/go-binance/v2"
)

var (
    messagesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "ingestor_messages_total",
        Help: "Total number of price updates received from Binance",
    })
)

func main() {
	prometheus.MustRegister(messagesProcessed)

	go func() {
        http.Handle("/metrics", promhttp.Handler())
        http.ListenAndServe(":2112", nil)
    }()
	
	rdb, err := redis.NewRedisClient("localhost:6379")
	if err != nil {
		log.Fatalf("Redis Init Failed: %v", err)
	}
	wp := workerpool.NewWorkerPool(5)
	wp.Redis = rdb
	wp.Start(context.Background())

	producer, err := kafka.NewProducer([]string{"localhost:9092"})
	if err != nil {
		log.Fatalf("Kafka Init Failed: %v", err)
	}
	defer producer.Close()

	wsHandler := func(event *binance.WsMarketStatEvent) {
		price := event.LastPrice
		symbol := event.Symbol

		messagesProcessed.Inc()

		err := kafka.PushMessage(producer, "market_data", symbol, price)
		if err != nil {
			log.Printf("Kafka Push Error: %v", err)
		}

		err = rdb.SetPrice(context.Background(), symbol, price)
		if err != nil {
			log.Printf("Redis Cache Error: %v", err)
		}
	}

	errHandler := func(err error) {
		log.Printf("Binance WS Error: %v", err)
	}

	doneC, stopC, err := binance.WsMarketStatServe("BTCUSDT", wsHandler, errHandler)
if err != nil {
    log.Fatal(err)
}

	fmt.Println("Ingestor is LIVE. Streaming data to Kafka & Redis...")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	stopC <- struct{}{}
	<-doneC
	fmt.Println("\nShutting down gracefully...")
}