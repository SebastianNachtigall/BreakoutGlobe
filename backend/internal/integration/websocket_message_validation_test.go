package integration

import (
	"testing"
	"time"

	"breakoutglobe/internal/testdata"
	"breakoutglobe/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketMessageTypeValidation tests WebSocket message type validation and format assertions
func TestWebSocketMessageTypeValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket message validation integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	t.Run("WelcomeMessageHandling", func(t *testing.T) {
		// Create client and properly consume welcome message
		client := testWS.CreateClient("session-welcome-test", "user-welcome-test", "map-welcome-test")
		require.NotNil(t, client)

		// Wait for connection
		time.Sleep(50 * time.Millisecond)

		// First message should be welcome message
		msg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should receive welcome message")
		assert.Equal(t, "welcome", msg.Type, "First message should be welcome")

		// Validate welcome message data structure
		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok, "Welcome message should have valid data structure")
		assert.Equal(t, "session-welcome-test", data["sessionId"])
		assert.Equal(t, "user-welcome-test", data["userId"])
		assert.Equal(t, "map-welcome-test", data["mapId"])
	})

	t.Run("MessageTypeAssertion", func(t *testing.T) {
		// Create client and consume welcome message first
		client := testWS.CreateClient("session-type-test", "user-type-test", "map-type-test")
		require.NotNil(t, client)

		// Wait for connection and consume welcome message
		time.Sleep(50 * time.Millisecond)
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should consume welcome message")

		// Now test specific message type broadcasting
		testMessage := websocket.Message{
			Type: "poi_created",
			Data: map[string]interface{}{
				"poiId":      "test-poi-123",
				"name":       "Test POI",
				"position":   map[string]float64{"lat": 40.7128, "lng": -74.0060},
				"createdBy":  "user-type-test",
				"mapId":      "map-type-test",
			},
			Timestamp: time.Now(),
		}

		testWS.BroadcastToMap("map-type-test", testMessage)

		// Should receive the poi_created message
		receivedMsg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should receive poi_created message")
		assert.Equal(t, "poi_created", receivedMsg.Type, "Should receive correct message type")

		// Validate message data structure
		data, ok := receivedMsg.Data.(map[string]interface{})
		require.True(t, ok, "Message should have valid data structure")
		assert.Equal(t, "test-poi-123", data["poiId"])
		assert.Equal(t, "Test POI", data["name"])
	})

	t.Run("MessageDataValidation", func(t *testing.T) {
		// Create client and consume welcome message
		client := testWS.CreateClient("session-data-test", "user-data-test", "map-data-test")
		require.NotNil(t, client)

		time.Sleep(50 * time.Millisecond)
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should consume welcome message")

		// Test avatar movement message with proper data types
		movementMessage := websocket.Message{
			Type: "avatar_movement",
			Data: map[string]interface{}{
				"sessionId": "session-data-test",
				"userId":    "user-data-test",
				"position": map[string]float64{
					"lat": 40.7589,
					"lng": -73.9851,
				},
				"mapId": "map-data-test",
			},
			Timestamp: time.Now(),
		}

		testWS.BroadcastToMap("map-data-test", movementMessage)

		// Receive and validate message
		receivedMsg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should receive avatar_movement message")
		assert.Equal(t, "avatar_movement", receivedMsg.Type)

		// Validate data types are preserved correctly
		data, ok := receivedMsg.Data.(map[string]interface{})
		require.True(t, ok, "Message data should be map")

		position, ok := data["position"].(map[string]float64)
		require.True(t, ok, "Position should be map[string]float64")
		assert.Equal(t, 40.7589, position["lat"])
		assert.Equal(t, -73.9851, position["lng"])

		assert.Equal(t, "session-data-test", data["sessionId"])
		assert.Equal(t, "user-data-test", data["userId"])
		assert.Equal(t, "map-data-test", data["mapId"])
	})

	t.Run("InvalidMessageTypeHandling", func(t *testing.T) {
		// Create client and consume welcome message
		client := testWS.CreateClient("session-invalid-test", "user-invalid-test", "map-invalid-test")
		require.NotNil(t, client)

		time.Sleep(50 * time.Millisecond)
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should consume welcome message")

		// Send message with invalid/unknown type
		invalidMessage := websocket.Message{
			Type: "unknown_message_type",
			Data: map[string]interface{}{
				"test": "data",
			},
			Timestamp: time.Now(),
		}

		testWS.BroadcastToMap("map-invalid-test", invalidMessage)

		// Should still receive the message (broadcasting doesn't validate types)
		receivedMsg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should receive message even with unknown type")
		assert.Equal(t, "unknown_message_type", receivedMsg.Type)
	})

	t.Run("MessageSequenceValidation", func(t *testing.T) {
		// Create client and consume welcome message
		client := testWS.CreateClient("session-sequence-test", "user-sequence-test", "map-sequence-test")
		require.NotNil(t, client)

		time.Sleep(50 * time.Millisecond)
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should consume welcome message")

		// Send sequence of different message types
		messageTypes := []string{"poi_created", "poi_joined", "avatar_movement", "poi_left"}
		
		for i, msgType := range messageTypes {
			testMessage := websocket.Message{
				Type: msgType,
				Data: map[string]interface{}{
					"sequence": i,
					"type":     msgType,
				},
				Timestamp: time.Now(),
			}

			testWS.BroadcastToMap("map-sequence-test", testMessage)
			time.Sleep(10 * time.Millisecond) // Small delay between messages
		}

		// Receive and validate all messages in sequence
		for i, expectedType := range messageTypes {
			receivedMsg, err := client.ReceiveMessage(200 * time.Millisecond)
			require.NoError(t, err, "Should receive message %d", i)
			assert.Equal(t, expectedType, receivedMsg.Type, "Message %d should have correct type", i)

			data, ok := receivedMsg.Data.(map[string]interface{})
			require.True(t, ok, "Message %d should have valid data", i)
			
			// JSON unmarshaling converts numbers to float64, but we sent int
			// Check both int and float64 types
			var sequence float64
			if seqInt, ok := data["sequence"].(int); ok {
				sequence = float64(seqInt)
			} else if seqFloat, ok := data["sequence"].(float64); ok {
				sequence = seqFloat
			} else {
				t.Errorf("Sequence should be a number, got %T: %v", data["sequence"], data["sequence"])
				continue
			}
			assert.Equal(t, float64(i), sequence, "Message %d should have correct sequence", i)
		}
	})
}