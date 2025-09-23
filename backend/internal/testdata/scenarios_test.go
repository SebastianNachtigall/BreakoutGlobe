package testdata

import (
	"net/http"
	"testing"

	"breakoutglobe/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPOITestScenario(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates scenario with defaults", func(t *testing.T) {
		scenario := NewPOITestScenario(t)

		assert.NotNil(t, scenario)
		assert.NotEmpty(t, scenario.userID)
		assert.NotEmpty(t, scenario.mapID)
		assert.NotNil(t, scenario.mockSetup)
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewPOITestScenario(t).
			WithUser(userID).
			WithMap(mapID)

		assert.Equal(t, userID, scenario.userID)
		assert.Equal(t, mapID, scenario.mapID)
	})

	t.Run("ExpectRateLimitSuccess sets up rate limiter mock", func(t *testing.T) {
		scenario := NewPOITestScenario(t).
			ExpectRateLimitSuccess()

		// Verify the expectation was set up (this will be validated when we execute)
		assert.NotNil(t, scenario.mockSetup.RateLimiter)
	})

	t.Run("ExpectCreationSuccess sets up POI service mock", func(t *testing.T) {
		scenario := NewPOITestScenario(t).
			ExpectCreationSuccess()

		// Verify the expectation was set up
		assert.NotNil(t, scenario.mockSetup.POIService)
	})
}

func TestPOITestScenario_CreatePOI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful POI creation", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()
		expectedPOI := NewPOI().
			WithID("poi-123").
			WithName("Coffee Shop").
			WithCreator(userID).
			WithMap(mapID).
			Build()

		scenario := NewPOITestScenario(t).
			WithUser(userID).
			WithMap(mapID).
			ExpectRateLimitSuccess().
			ExpectCreationSuccess()

		request := CreatePOIRequest{
			MapID:           mapID.String(),
			Name:            "Coffee Shop",
			Description:     "Great coffee place",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       userID.String(),
			MaxParticipants: 10,
		}

		response := scenario.CreatePOI(request)

		// Verify response
		assert.Equal(t, expectedPOI.ID, response.ID)
		assert.Equal(t, expectedPOI.Name, response.Name)
		assert.Equal(t, expectedPOI.Description, response.Description)
		assert.Equal(t, expectedPOI.Position.Lat, response.Position.Lat)
		assert.Equal(t, expectedPOI.Position.Lng, response.Position.Lng)

		// Verify all mocks were called correctly
		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI creation with rate limit exceeded", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewPOITestScenario(t).
			WithUser(userID).
			WithMap(mapID).
			ExpectRateLimitExceeded()

		request := CreatePOIRequest{
			MapID:           mapID.String(),
			Name:            "Coffee Shop",
			Description:     "Great coffee place",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       userID.String(),
			MaxParticipants: 10,
		}

		errorResponse := scenario.CreatePOIExpectError(request)

		// Verify rate limit error response
		assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorResponse.Code)
		assert.Contains(t, errorResponse.Message, "rate limit")

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI creation with validation error", func(t *testing.T) {
		scenario := NewPOITestScenario(t)

		request := CreatePOIRequest{
			MapID:           "", // Invalid: empty map ID
			Name:            "Coffee Shop",
			Description:     "Great coffee place",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       scenario.userID.String(),
			MaxParticipants: 10,
		}

		errorResponse := scenario.CreatePOIExpectError(request)

		// Verify validation error response
		assert.Equal(t, "INVALID_REQUEST", errorResponse.Code)
		assert.Contains(t, errorResponse.Message, "validation")
	})
}

func TestPOITestScenario_JoinPOI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful POI join", func(t *testing.T) {
		userID := GenerateUUID()
		poiID := "poi-123"

		scenario := NewPOITestScenario(t).
			WithUser(userID).
			ExpectJoinRateLimitSuccessWithHeaders().
			ExpectJoinSuccess()

		response := scenario.JoinPOI(poiID, userID.String())

		// Verify successful join
		assert.True(t, response.Success)
		assert.Equal(t, poiID, response.POIID)
		assert.Equal(t, userID.String(), response.UserID)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI join with capacity exceeded", func(t *testing.T) {
		userID := GenerateUUID()
		poiID := "poi-123"

		scenario := NewPOITestScenario(t).
			WithUser(userID).
			ExpectJoinRateLimitSuccess().
			ExpectCapacityExceeded()

		errorResponse := scenario.JoinPOIExpectError(poiID, userID.String())

		// Verify capacity exceeded error
		assert.Equal(t, "CAPACITY_EXCEEDED", errorResponse.Code)
		assert.Contains(t, errorResponse.Message, "capacity")

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI join with POI not found", func(t *testing.T) {
		userID := GenerateUUID()
		poiID := "non-existent-poi"

		scenario := NewPOITestScenario(t).
			WithUser(userID).
			ExpectJoinRateLimitSuccess().
			ExpectNotFound()

		errorResponse := scenario.JoinPOIExpectError(poiID, userID.String())

		// Verify not found error
		assert.Equal(t, "POI_NOT_FOUND", errorResponse.Code)
		assert.Contains(t, errorResponse.Message, "not found")

		scenario.mockSetup.AssertExpectations(t)
	})
}

func TestPOITestScenario_GetPOI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful POI retrieval", func(t *testing.T) {
		poiID := "poi-123"
		expectedPOI := NewPOI().
			WithID(poiID).
			WithName("Coffee Shop").
			Build()

		scenario := NewPOITestScenario(t).
			ExpectGetSuccess(expectedPOI)

		response := scenario.GetPOI(poiID)

		// Verify response
		assert.Equal(t, expectedPOI.ID, response.POI.ID)
		assert.Equal(t, expectedPOI.Name, response.POI.Name)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI not found", func(t *testing.T) {
		poiID := "non-existent-poi"

		scenario := NewPOITestScenario(t).
			ExpectGetNotFound()

		errorResponse := scenario.GetPOIExpectError(poiID)

		// Verify not found error
		assert.Equal(t, "POI_NOT_FOUND", errorResponse.Code)
		assert.Contains(t, errorResponse.Message, "not found")

		scenario.mockSetup.AssertExpectations(t)
	})
}

// Request/Response types are now defined in scenarios.go

func TestSessionTestScenario(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates scenario with defaults", func(t *testing.T) {
		scenario := NewSessionTestScenario()

		assert.NotNil(t, scenario)
		assert.NotEmpty(t, scenario.userID)
		assert.NotEmpty(t, scenario.mapID)
		assert.NotNil(t, scenario.mockSetup)
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewSessionTestScenario().
			WithUser(userID).
			WithMap(mapID)

		assert.Equal(t, userID, scenario.userID)
		assert.Equal(t, mapID, scenario.mapID)
	})

	t.Run("ExpectRateLimitSuccess sets up rate limiter mock", func(t *testing.T) {
		scenario := NewSessionTestScenario().
			ExpectRateLimitSuccess()

		assert.NotNil(t, scenario.mockSetup.RateLimiter)
	})

	t.Run("ExpectCreationSuccess sets up session service mock", func(t *testing.T) {
		expectedSession := NewSession().Build()
		scenario := NewSessionTestScenario().
			ExpectCreationSuccess(expectedSession)

		assert.NotNil(t, scenario.mockSetup.SessionService)
	})
}

func TestSessionTestScenario_CreateSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful session creation", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()
		expectedSession := NewSession().
			WithID("session-123").
			WithUser(userID).
			WithMap(mapID).
			WithPosition(models.LatLng{Lat: 37.7749, Lng: -122.4194}).
			Build()

		scenario := NewSessionTestScenario().
			WithUser(userID).
			WithMap(mapID).
			ExpectRateLimitSuccess().
			ExpectCreationSuccess(expectedSession)

		request := CreateSessionRequest{
			UserID:         userID.String(),
			MapID:          mapID.String(),
			AvatarPosition: models.LatLng{Lat: 37.7749, Lng: -122.4194},
		}

		response := scenario.CreateSession(t, request)

		// Verify response
		assert.Equal(t, expectedSession.ID, response.SessionID)
		assert.Equal(t, expectedSession.UserID, response.UserID)
		assert.Equal(t, expectedSession.MapID, response.MapID)
		assert.Equal(t, expectedSession.AvatarPos.Lat, response.AvatarPosition.Lat)
		assert.Equal(t, expectedSession.AvatarPos.Lng, response.AvatarPosition.Lng)

		// Verify all mocks were called correctly
		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("session creation with rate limit exceeded", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewSessionTestScenario().
			WithUser(userID).
			WithMap(mapID).
			ExpectRateLimitExceeded()

		request := CreateSessionRequest{
			UserID:         userID.String(),
			MapID:          mapID.String(),
			AvatarPosition: models.LatLng{Lat: 37.7749, Lng: -122.4194},
		}

		recorder := scenario.CreateSessionExpectingError(t, request)

		// Verify rate limit error response
		AssertHTTPStatus(t, recorder, http.StatusTooManyRequests)
		AssertErrorResponse(t, recorder, "RATE_LIMIT_EXCEEDED")

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("session creation with validation error", func(t *testing.T) {
		scenario := NewSessionTestScenario()

		request := CreateSessionRequest{
			UserID:         "", // Invalid: empty user ID
			MapID:          scenario.mapID.String(),
			AvatarPosition: models.LatLng{Lat: 37.7749, Lng: -122.4194},
		}

		recorder := scenario.CreateSessionExpectingError(t, request)

		// Verify validation error response
		AssertHTTPStatus(t, recorder, http.StatusBadRequest)
		AssertErrorResponse(t, recorder, "INVALID_REQUEST")
	})
}

func TestSessionTestScenario_GetSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful session retrieval", func(t *testing.T) {
		sessionID := "session-123"
		expectedSession := NewSession().
			WithID(sessionID).
			WithUser(GenerateUUID()).
			WithMap(GenerateUUID()).
			Build()

		scenario := NewSessionTestScenario().
			ExpectGetSuccess(expectedSession)

		response := scenario.GetSession(t, sessionID)

		// Verify response
		assert.Equal(t, expectedSession.ID, response.SessionID)
		assert.Equal(t, expectedSession.UserID, response.UserID)
		assert.Equal(t, expectedSession.MapID, response.MapID)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		sessionID := "non-existent-session"

		scenario := NewSessionTestScenario().
			ExpectGetNotFound()

		recorder := scenario.GetSessionExpectingError(t, sessionID)

		// Verify not found error
		AssertHTTPStatus(t, recorder, http.StatusNotFound)
		AssertErrorResponse(t, recorder, "SESSION_NOT_FOUND")

		scenario.mockSetup.AssertExpectations(t)
	})
}

func TestSessionTestScenario_UpdateAvatarPosition(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful avatar position update", func(t *testing.T) {
		sessionID := "session-123"
		userID := GenerateUUID()

		scenario := NewSessionTestScenario().
			WithUser(userID).
			ExpectUpdateRateLimitSuccess().
			ExpectUpdatePositionSuccess()

		request := UpdateAvatarPositionRequest{
			Position: models.LatLng{Lat: 40.7589, Lng: -73.9851}, // New York
		}

		recorder := scenario.UpdateAvatarPosition(t, sessionID, request)

		// Verify successful update
		AssertHTTPStatus(t, recorder, http.StatusOK)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("avatar position update with session not found", func(t *testing.T) {
		sessionID := "non-existent-session"
		userID := GenerateUUID()

		scenario := NewSessionTestScenario().
			WithUser(userID).
			ExpectUpdatePositionNotFound()

		request := UpdateAvatarPositionRequest{
			Position: models.LatLng{Lat: 40.7589, Lng: -73.9851},
		}

		recorder := scenario.UpdateAvatarPosition(t, sessionID, request)

		// Verify not found error
		AssertHTTPStatus(t, recorder, http.StatusNotFound)
		AssertErrorResponse(t, recorder, "SESSION_NOT_FOUND")

		scenario.mockSetup.AssertExpectations(t)
	})
}

func TestSessionTestScenario_SessionHeartbeat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful session heartbeat", func(t *testing.T) {
		sessionID := "session-123"

		scenario := NewSessionTestScenario().
			ExpectHeartbeatSuccess()

		recorder := scenario.SessionHeartbeat(t, sessionID)

		// Verify successful heartbeat
		AssertHTTPStatus(t, recorder, http.StatusOK)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("session heartbeat with session not found", func(t *testing.T) {
		sessionID := "non-existent-session"

		scenario := NewSessionTestScenario().
			ExpectHeartbeatNotFound()

		recorder := scenario.SessionHeartbeat(t, sessionID)

		// Verify not found error
		AssertHTTPStatus(t, recorder, http.StatusNotFound)
		AssertErrorResponse(t, recorder, "SESSION_NOT_FOUND")

		scenario.mockSetup.AssertExpectations(t)
	})
}

// Request/Response types for Session scenarios - using the ones from scenarios.go

func TestWebSocketTestScenario(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates scenario with defaults", func(t *testing.T) {
		scenario := NewWebSocketTestScenario()

		assert.NotNil(t, scenario)
		assert.NotEmpty(t, scenario.sessionID)
		assert.NotEmpty(t, scenario.userID)
		assert.NotEmpty(t, scenario.mapID)
		assert.NotNil(t, scenario.mockSetup)
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		sessionID := "custom-session"
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			WithUser(userID).
			WithMap(mapID)

		assert.Equal(t, sessionID, scenario.sessionID)
		assert.Equal(t, userID, scenario.userID)
		assert.Equal(t, mapID, scenario.mapID)
	})

	t.Run("ExpectConnectionSuccess sets up session service mock", func(t *testing.T) {
		scenario := NewWebSocketTestScenario().
			ExpectConnectionSuccess()

		assert.NotNil(t, scenario.mockSetup.SessionService)
	})
}

func TestWebSocketTestScenario_Connection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful WebSocket connection", func(t *testing.T) {
		sessionID := "session-123"
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			WithUser(userID).
			WithMap(mapID).
			ExpectConnectionSuccess()

		conn, welcomeMsg := scenario.Connect(t)
		defer conn.Close()

		// Verify welcome message
		assert.Equal(t, "welcome", welcomeMsg.Type)
		if data, ok := welcomeMsg.Data["sessionId"]; ok {
			assert.Equal(t, sessionID, data)
		}

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("WebSocket connection with invalid session", func(t *testing.T) {
		sessionID := "invalid-session"

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			ExpectConnectionFailure()

		err := scenario.ConnectExpectingError(t)

		// Verify connection failed
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad handshake")

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("WebSocket connection without authorization", func(t *testing.T) {
		scenario := NewWebSocketTestScenario()

		err := scenario.ConnectWithoutAuth(t)

		// Verify connection failed
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad handshake")
	})
}

func TestWebSocketTestScenario_Heartbeat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful heartbeat message", func(t *testing.T) {
		sessionID := "session-123"
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			WithUser(userID).
			WithMap(mapID).
			ExpectConnectionSuccess().
			ExpectHeartbeatSuccess()

		conn, _ := scenario.Connect(t)
		defer conn.Close()

		// Send heartbeat and verify no error response
		scenario.SendHeartbeat(t, conn)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("heartbeat with session error", func(t *testing.T) {
		sessionID := "session-123"
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			WithUser(userID).
			WithMap(mapID).
			ExpectConnectionSuccess().
			ExpectHeartbeatError()

		conn, _ := scenario.Connect(t)
		defer conn.Close()

		// Send heartbeat and expect error response
		errorMsg := scenario.SendHeartbeatExpectingError(t, conn)
		assert.Equal(t, "error", errorMsg.Type)

		scenario.mockSetup.AssertExpectations(t)
	})
}

func TestWebSocketTestScenario_AvatarMovement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful avatar movement", func(t *testing.T) {
		sessionID := "session-123"
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			WithUser(userID).
			WithMap(mapID).
			ExpectConnectionSuccess().
			ExpectAvatarMoveSuccess()

		conn, _ := scenario.Connect(t)
		defer conn.Close()

		// Send avatar movement
		newPosition := models.LatLng{Lat: 40.7589, Lng: -73.9851}
		scenario.SendAvatarMove(t, conn, newPosition)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("avatar movement with rate limit exceeded", func(t *testing.T) {
		sessionID := "session-123"
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewWebSocketTestScenario().
			WithSession(sessionID).
			WithUser(userID).
			WithMap(mapID).
			ExpectConnectionSuccess().
			ExpectAvatarMoveRateLimited()

		conn, _ := scenario.Connect(t)
		defer conn.Close()

		// Send avatar movement and expect rate limit error
		newPosition := models.LatLng{Lat: 40.7589, Lng: -73.9851}
		errorMsg := scenario.SendAvatarMoveExpectingError(t, conn, newPosition)
		assert.Equal(t, "error", errorMsg.Type)

		scenario.mockSetup.AssertExpectations(t)
	})
}

// Note: Message broadcast testing is complex and requires multiple clients
// For now, we focus on basic WebSocket functionality