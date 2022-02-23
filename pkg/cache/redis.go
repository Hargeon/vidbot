package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	timeOut = time.Second
)

type Redis struct {
	client *redis.Client
}

// NewService creates redis cache
func NewService(client *redis.Client) *Redis {
	return &Redis{client: client}
}

// Read from redis cache
func (r *Redis) Read(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	return r.client.Get(ctx, key).Result()
}

// Write to redis cache
func (r *Redis) Write(key string, value interface{}, exp time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	return r.client.Set(ctx, key, value, exp).Err()
}

// Delete key in redis
func (r *Redis) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	return r.client.Del(ctx, key).Err()
}
