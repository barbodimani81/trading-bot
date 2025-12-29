
# Go Crypto Trading Bot

A high-performance, event-driven crypto trading system built with **Golang**, **Kafka**, and **Redis**.

This project demonstrates a production-grade microservices architecture designed to ingest real-time market data from Binance, process trading strategies in parallel using worker pools, and expose low-latency data via an API.


### Core Components

1. **Ingestor Service (`cmd/ingestor`)**:
* Connects to Binance WebSockets.
* Acts as a "Gateway", normalizing data and pushing it to Kafka (Storage) and Redis (Hot State).
* **Tech:** `goroutines`, `sarama` (Kafka), `go-redis`.


2. **Processor Service (`cmd/processor`)**:
* Consumes the `market_data` topic from Kafka.
* Uses a **Worker Pool Pattern** to process thousands of price updates in parallel without blocking.
* Executes a "Moving Average" strategy using Redis as a time-series buffer.
* **Tech:** `sync.WaitGroup`, `channels`, `context`.


3. **API Service (`cmd/api`)**:
* Provides a zero-latency HTTP endpoint for frontends/dashboards.
* Reads directly from Redis Cache (avoiding slow database queries).



## 🛠 Tech Stack

* **Language:** Golang (1.21+)
* **Messaging:** Apache Kafka & Zookeeper
* **Cache/State:** Redis (Alpine)
* **Monitoring:** Prometheus & Grafana
* **Containerization:** Docker & Docker Compose
* **External API:** Binance WebSocket API

## ⚡️ Getting Started

### Prerequisites

* Docker & Docker Compose
* Go 1.21+ installed locally
* `make` (optional, for easy commands)

### 1. Start Infrastructure

Spin up Kafka, Zookeeper, Redis, Prometheus, and Grafana in the background.

```bash
make infra
# OR
docker-compose up -d

```

### 2. Run Microservices

You can run all services simultaneously using the Makefile:

```bash
make run-all

```

*Alternatively, run them in separate terminals:*

```bash
go run cmd/ingestor/main.go
go run cmd/processor/main.go
go run cmd/api/main.go

```

## 📊 Monitoring & Observability

The system is fully instrumented with Prometheus metrics to track performance.

### **1. Prometheus (Metrics Collection)**

* **URL:** `http://localhost:9090`
* **Key Metrics:**
* `ingestor_messages_total`: Total price updates received.
* `worker_processing_duration_seconds`: P99 Latency of trading logic.
* `strategy_buy_signals_total`: Number of "BUY" signals triggered.



### **2. Grafana (Visualization)**

* **URL:** `http://localhost:3001`
* **Login:** `admin` / `admin`
* **Setup:** Add Prometheus (`http://prometheus:9090`) as a data source and import the metrics above.

### **3. Kafka UI**

* **URL:** `http://localhost:8080`
* View live topics, partitions, and message offsets.


```text
.
├── cmd/
│   ├── api/           # HTTP API Entrypoint
│   ├── ingestor/      # Binance -> Kafka Producer
│   └── processor/     # Kafka Consumer -> Strategy Logic
├── internal/
│   ├── kafka/         # Shared Kafka Configuration
│   ├── redis/         # Shared Redis Client & Helpers
│   └── workerpool/    # Concurrency Engine & Strategy
├── docker-compose.yml # Infrastructure definition
├── Makefile           # Shortcut commands
└── prometheus.yml     # Monitoring config

```
