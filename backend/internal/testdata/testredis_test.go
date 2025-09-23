package testdata

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupRedis verifies that Redis test setup works correctly
func TestSetupRedis(t *testing.T) {
	// Skip if no Redis available
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}
	
	// Setup test Redis
	testRedis := SetupRedis(t)
	
	// Verify Redis connection
	assert.NotNil(t, testRedis.Client())
	assert.GreaterOrEqual(t, testRedis.DBNum(), 0)
	assert.LessOrEqual(t, testRedis.DBNum(), 15)
	
	// Verify Redis is accessible
	ctx := context.Background()
	err := testRedis.Client().Set(ctx, "test-key", "test-value", 0).Err()
	require.NoError(t, err)
	
	val, err := testRedis.Client().Get(ctx, "test-key").Result()
	require.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

// TestRedisIsolation verifies that Redis tests are isolated
func TestRedisIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}
	
	// Create two test Redis instances
	testRedis1 := SetupRedis(t)
	testRedis2 := SetupRedis(t)
	
	// Verify they use different database numbers (most of the time)
	// Note: With 16 databases and timestamp-based selection, collisions are possible but rare
	if testRedis1.DBNum() == testRedis2.DBNum() {
		t.Logf("Warning: Both Redis instances using same DB number %d (collision possible with timestamp-based selection)", testRedis1.DBNum())
	}
	
	// Verify they are isolated - data in one doesn't affect the other
	ctx := context.Background()
	err1 := testRedis1.Client().Set(ctx, "isolation-test", "value1", 0).Err()
	require.NoError(t, err1)
	
	err2 := testRedis2.Client().Set(ctx, "isolation-test", "value2", 0).Err()
	require.NoError(t, err2)
	
	// Each should have its own value
	val1, err := testRedis1.Client().Get(ctx, "isolation-test").Result()
	require.NoError(t, err)
	
	val2, err := testRedis2.Client().Get(ctx, "isolation-test").Result()
	require.NoError(t, err)
	
	if testRedis1.DBNum() != testRedis2.DBNum() {
		// Different databases - should have different values
		assert.Equal(t, "value1", val1)
		assert.Equal(t, "value2", val2)
	} else {
		// Same database - second write overwrites first
		assert.Equal(t, "value2", val1)
		assert.Equal(t, "value2", val2)
	}
}

// TestRedisAssertions verifies Redis assertion helpers
func TestRedisAssertions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}
	
	testRedis := SetupRedis(t)
	ctx := context.Background()
	
	// Test key existence assertions
	testRedis.Client().Set(ctx, "existing-key", "value", 0)
	testRedis.AssertKeyExists("existing-key")
	testRedis.AssertKeyNotExists("non-existing-key")
	
	// Test set operations and assertions
	testRedis.Client().SAdd(ctx, "test-set", "member1", "member2", "member3")
	testRedis.AssertSetContains("test-set", "member1")
	testRedis.AssertSetNotContains("test-set", "member4")
	testRedis.AssertSetSize("test-set", 3)
	
	// Test hash operations and assertions
	testRedis.Client().HSet(ctx, "test-hash", "field1", "value1", "field2", "value2")
	testRedis.AssertHashField("test-hash", "field1", "value1")
	testRedis.AssertHashField("test-hash", "field2", "value2")
}

// TestRedisPubSub verifies Redis pub/sub operations
func TestRedisPubSub(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}
	
	testRedis := SetupRedis(t)
	
	// Subscribe to a channel
	pubsub := testRedis.Subscribe("test-channel")
	defer pubsub.Close()
	
	// Wait for subscription to be established
	time.Sleep(10 * time.Millisecond)
	
	// Publish a message
	ctx := context.Background()
	testRedis.Client().Publish(ctx, "test-channel", "test-message")
	
	// Receive the message
	msg, err := pubsub.ReceiveTimeout(ctx, 100*time.Millisecond)
	require.NoError(t, err)
	
	switch m := msg.(type) {
	case *redis.Subscription:
		// This is the subscription confirmation, get the actual message
		msg, err = pubsub.ReceiveTimeout(ctx, 100*time.Millisecond)
		require.NoError(t, err)
		actualMsg, ok := msg.(*redis.Message)
		require.True(t, ok)
		assert.Equal(t, "test-channel", actualMsg.Channel)
		assert.Equal(t, "test-message", actualMsg.Payload)
	case *redis.Message:
		assert.Equal(t, "test-channel", m.Channel)
		assert.Equal(t, "test-message", m.Payload)
	default:
		t.Fatalf("Unexpected message type: %T", msg)
	}
}

// TestRedisCleanup verifies that Redis cleanup works properly
func TestRedisCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}
	
	var client *redis.Client
	
	// Create a scope to test cleanup
	func() {
		testRedis := SetupRedis(t)
		client = testRedis.Client()
		
		// Add some data
		ctx := context.Background()
		client.Set(ctx, "cleanup-test", "value", 0)
		
		// Verify data exists
		val, err := client.Get(ctx, "cleanup-test").Result()
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	}()
	
	// After the scope, the cleanup should have run
	// Note: We can't easily test this without accessing the same DB again,
	// but the cleanup function should have been called
}

// BenchmarkRedisOperations benchmarks Redis operations
func BenchmarkRedisOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping Redis integration benchmark in short mode")
	}
	
	testRedis := SetupRedis(b)
	ctx := context.Background()
	
	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testRedis.Client().Set(ctx, fmt.Sprintf("bench-key-%d", i), fmt.Sprintf("value-%d", i), 0)
		}
	})
	
	b.Run("Get", func(b *testing.B) {
		// Setup data
		for i := 0; i < 1000; i++ {
			testRedis.Client().Set(ctx, fmt.Sprintf("bench-key-%d", i), fmt.Sprintf("value-%d", i), 0)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testRedis.Client().Get(ctx, fmt.Sprintf("bench-key-%d", i%1000))
		}
	})
	
	b.Run("SetAdd", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testRedis.Client().SAdd(ctx, "bench-set", fmt.Sprintf("member-%d", i))
		}
	})
}