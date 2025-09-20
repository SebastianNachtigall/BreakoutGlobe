package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

// NewClient creates a new Redis client with the given configuration
func NewClient(config Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
}

// TestConnection tests the Redis connection
func TestConnection(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return nil
}

// CloseClient closes the Redis client connection
func CloseClient(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}