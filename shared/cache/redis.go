package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

func Init() error {
	var addr string

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		addr = redisAddr
	} else {
		env := os.Getenv("ENV")
		if env == "" {
			env = "development"
		}

		switch env {
		case "production":
			addr = "redis-prod:6379"
		case "staging":
			addr = "redis-staging:6379"
		default:
			addr = "redis:6379"
		}
	}

	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	return nil
}

func Get(key string, dest interface{}) error {
	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, data, ttl).Err()
}

func Delete(key string) error {
	return Client.Del(ctx, key).Err()
}

func Exists(key string) (bool, error) {
	result, err := Client.Exists(ctx, key).Result()
	return result > 0, err
}

func GetOrSet(key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	err := Get(key, dest)
	if err == nil {
		return nil
	}

	value, err := fn()
	if err != nil {
		return err
	}

	if err := Set(key, value, ttl); err != nil {
		fmt.Printf("failed to cache value: %v\n", err)
	}

	data, _ := json.Marshal(value)
	return json.Unmarshal(data, dest)
}

func InvalidatePattern(pattern string) error {
	iter := Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := Client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return Client.Ping(ctx).Err()
}
