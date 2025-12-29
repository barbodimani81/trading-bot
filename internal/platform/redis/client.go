package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // Default for local docker
		DB:           0,  // Default DB
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10, // Max number of simultaneous connections
	})

	// 1. Verify connection immediately (Ping)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	fmt.Println("Redis connected successfully at: ", addr)
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) SetPrice(ctx context.Context, symbol string, price string) error {
	key := fmt.Sprintf("price:%s", symbol)
	return r.Client.Set(ctx, key, price, 10*time.Minute).Err()
}

func (r *RedisClient) GetPrice(ctx context.Context, symbol string) (string, error) {
	key := fmt.Sprintf("price:%s", symbol)
	return r.Client.Get(ctx, key).Result()
}
