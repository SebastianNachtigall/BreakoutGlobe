package websocket

import (
	"context"
	"testing"
	"time"

	"breakoutglobe/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_POIJoin_Success(t *testing.T) {
	// Setup mocks
	mockPOIService := new(MockPOIService)
	mockSessionService := new(MockSessionService)
	mockRateLimiter := new(MockRateLimiter)

	// Create handler
	handler := NewHandler(mockSessionService, mockRateLimiter, nil, mockPOIService)

	// Create test client
	client := &Client{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     "map-789",
		Send:      make(chan Message, 10),
	}

	// Setup expectations
	mockSessionService.On("GetSession", mock.Anything, "session-123").Return(nil, nil) // Session not needed for this test
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionJoinPOI).Return(nil)
	mockPOIService.On("JoinPOI", mock.Anything, "poi-123", "user-456").Return(nil)

	// Create test message
	msg := Message{
		Type: "poi_join",
		Data: map[string]interface{}{
			"poiId": "poi-123",
		},
		Timestamp: time.Now(),
	}

	// Execute
	handler.handlePOIJoin(context.Background(), client, msg)

	// Verify acknowledgment was sent
	select {
	case ackMsg := <-client.Send:
		assert.Equal(t, "poi_join_ack", ackMsg.Type)
		data, ok := ackMsg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "session-123", data["sessionId"])
		assert.Equal(t, "poi-123", data["poiId"])
		assert.Equal(t, true, data["success"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected acknowledgment message not received")
	}

	// Verify mocks were called
	mockRateLimiter.AssertExpectations(t)
	mockPOIService.AssertExpectations(t)
}

func TestHandler_POILeave_Success(t *testing.T) {
	// Setup mocks
	mockPOIService := new(MockPOIService)
	mockSessionService := new(MockSessionService)
	mockRateLimiter := new(MockRateLimiter)

	// Create handler
	handler := NewHandler(mockSessionService, mockRateLimiter, nil, mockPOIService)

	// Create test client
	client := &Client{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     "map-789",
		Send:      make(chan Message, 10),
	}

	// Setup expectations
	mockSessionService.On("GetSession", mock.Anything, "session-123").Return(nil, nil) // Session not needed for this test
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionLeavePOI).Return(nil)
	mockPOIService.On("LeavePOI", mock.Anything, "poi-123", "user-456").Return(nil)

	// Create test message
	msg := Message{
		Type: "poi_leave",
		Data: map[string]interface{}{
			"poiId": "poi-123",
		},
		Timestamp: time.Now(),
	}

	// Execute
	handler.handlePOILeave(context.Background(), client, msg)

	// Verify acknowledgment was sent
	select {
	case ackMsg := <-client.Send:
		assert.Equal(t, "poi_leave_ack", ackMsg.Type)
		data, ok := ackMsg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "session-123", data["sessionId"])
		assert.Equal(t, "poi-123", data["poiId"])
		assert.Equal(t, true, data["success"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected acknowledgment message not received")
	}

	// Verify mocks were called
	mockRateLimiter.AssertExpectations(t)
	mockPOIService.AssertExpectations(t)
}

func TestHandler_POIJoin_ServiceError(t *testing.T) {
	// Setup mocks
	mockPOIService := new(MockPOIService)
	mockSessionService := new(MockSessionService)
	mockRateLimiter := new(MockRateLimiter)

	// Create handler
	handler := NewHandler(mockSessionService, mockRateLimiter, nil, mockPOIService)

	// Create test client
	client := &Client{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     "map-789",
		Send:      make(chan Message, 10),
	}

	// Setup expectations - POI service returns error
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionJoinPOI).Return(nil)
	mockPOIService.On("JoinPOI", mock.Anything, "poi-123", "user-456").Return(assert.AnError)

	// Create test message
	msg := Message{
		Type: "poi_join",
		Data: map[string]interface{}{
			"poiId": "poi-123",
		},
		Timestamp: time.Now(),
	}

	// Execute
	handler.handlePOIJoin(context.Background(), client, msg)

	// Verify error message was sent
	select {
	case errorMsg := <-client.Send:
		assert.Equal(t, "error", errorMsg.Type)
		data, ok := errorMsg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "Failed to join POI")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error message not received")
	}

	// Verify mocks were called
	mockRateLimiter.AssertExpectations(t)
	mockPOIService.AssertExpectations(t)
}