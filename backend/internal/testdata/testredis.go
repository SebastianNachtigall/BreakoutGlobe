package testdata

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestRedis provides Redis integration testing infrastructure
type TestRedis struct {
	t      TestingT
	client *redis.Client
	dbNum  int
}

// SetupRedis creates an isolated Redis database for integration testing
func SetupRedis(t TestingT) *TestRedis {
	t.Helper()
	
	// Get Redis connection details from environment
	host := getRedisEnvOrDefault("TEST_REDIS_HOST", "localhost")
	port := getRedisEnvOrDefault("TEST_REDIS_PORT", "6379") // Use production Redis port for tests
	password := getRedisEnvOrDefault("TEST_REDIS_PASSWORD", "")
	
	// Generate unique database number for this test
	dbNum := int(time.Now().UnixNano() % 16)
	
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       dbNum,
	})
	
	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Errorf("Failed to connect to Redis at %s:%s: %v", host, port, err)
		t.Errorf("Make sure Redis is running with: docker compose up -d redis")
		return nil
	}
	
	// Clear the database to ensure clean state
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Errorf("Failed to flush Redis database: %v", err)
		return nil
	}
	
	// Register cleanup
	t.Cleanup(func() {
		client.FlushDB(context.Background())
		client.Close()
	})
	
	return &TestRedis{
		t:      t,
		client: client,
		dbNum:  dbNum,
	}
}

// Client returns the Redis client for testing
func (tr *TestRedis) Client() *redis.Client {
	if tr == nil || tr.client == nil {
		panic("TestRedis client is nil - Redis connection failed during setup")
	}
	return tr.client
}

// DBNum returns the Redis database number being used
func (tr *TestRedis) DBNum() int {
	return tr.dbNum
}

// Subscribe creates a subscription to channels
func (tr *TestRedis) Subscribe(channels ...string) *redis.PubSub {
	return tr.client.Subscribe(context.Background(), channels...)
}

// AssertKeyExists asserts that a key exists in Redis
func (tr *TestRedis) AssertKeyExists(key string) {
	tr.t.Helper()
	count, err := tr.client.Exists(context.Background(), key).Result()
	if err != nil || count == 0 {
		tr.t.Errorf("Expected Redis key %s to exist, but it doesn't", key)
	}
}

// AssertKeyNotExists asserts that a key does not exist in Redis
func (tr *TestRedis) AssertKeyNotExists(key string) {
	tr.t.Helper()
	count, err := tr.client.Exists(context.Background(), key).Result()
	if err != nil {
		tr.t.Errorf("Error checking Redis key existence: %v", err)
		return
	}
	if count > 0 {
		tr.t.Errorf("Expected Redis key %s to not exist, but it does", key)
	}
}

// AssertSetContains asserts that a set contains a specific member
func (tr *TestRedis) AssertSetContains(key string, member interface{}) {
	tr.t.Helper()
	isMember, err := tr.client.SIsMember(context.Background(), key, member).Result()
	if err != nil || !isMember {
		tr.t.Errorf("Expected Redis set %s to contain %v, but it doesn't", key, member)
	}
}

// AssertSetNotContains asserts that a set does not contain a specific member
func (tr *TestRedis) AssertSetNotContains(key string, member interface{}) {
	tr.t.Helper()
	isMember, err := tr.client.SIsMember(context.Background(), key, member).Result()
	if err != nil {
		tr.t.Errorf("Error checking Redis set membership: %v", err)
		return
	}
	if isMember {
		tr.t.Errorf("Expected Redis set %s to not contain %v, but it does", key, member)
	}
}

// AssertHashField asserts that a hash field has a specific value
func (tr *TestRedis) AssertHashField(key, field, expectedValue string) {
	tr.t.Helper()
	actualValue, err := tr.client.HGet(context.Background(), key, field).Result()
	if err != nil {
		tr.t.Errorf("Failed to get Redis hash field: %v", err)
		return
	}
	if actualValue != expectedValue {
		tr.t.Errorf("Expected Redis hash %s field %s to have value %q, got %q", key, field, expectedValue, actualValue)
	}
}

// AssertSetSize asserts that a set has a specific size
func (tr *TestRedis) AssertSetSize(key string, expectedSize int) {
	tr.t.Helper()
	actualSize, err := tr.client.SCard(context.Background(), key).Result()
	if err != nil {
		tr.t.Errorf("Failed to get Redis set size: %v", err)
		return
	}
	if int(actualSize) != expectedSize {
		tr.t.Errorf("Expected Redis set %s to have %d members, got %d", key, expectedSize, int(actualSize))
	}
}

// Keys returns all keys matching a pattern
func (tr *TestRedis) Keys(pattern string) []string {
	tr.t.Helper()
	
	keys, err := tr.client.Keys(context.Background(), pattern).Result()
	if err != nil {
		tr.t.Errorf("Failed to get Redis keys with pattern %s: %v", pattern, err)
		return nil
	}
	return keys
}

// KeyCount returns the number of keys matching a pattern
func (tr *TestRedis) KeyCount(pattern string) int {
	return len(tr.Keys(pattern))
}

// SetMembers gets all members of a set
func (tr *TestRedis) SetMembers(key string) []string {
	tr.t.Helper()
	
	members, err := tr.client.SMembers(context.Background(), key).Result()
	if err != nil {
		tr.t.Errorf("Failed to get Redis set members %s: %v", key, err)
		return nil
	}
	return members
}

// getRedisEnvOrDefault returns environment variable value or default
func getRedisEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetAllKeys returns all keys in the Redis database
func (tr *TestRedis) GetAllKeys() []string {
	return tr.Keys("*")
}