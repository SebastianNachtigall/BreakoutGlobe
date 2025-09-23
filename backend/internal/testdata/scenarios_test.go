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
		scenario := NewPOITestScenario()

		assert.NotNil(t, scenario)
		assert.NotEmpty(t, scenario.userID)
		assert.NotEmpty(t, scenario.mapID)
		assert.NotNil(t, scenario.mockSetup)
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()

		scenario := NewPOITestScenario().
			WithUser(userID).
			WithMap(mapID)

		assert.Equal(t, userID, scenario.userID)
		assert.Equal(t, mapID, scenario.mapID)
	})

	t.Run("ExpectRateLimitSuccess sets up rate limiter mock", func(t *testing.T) {
		scenario := NewPOITestScenario().
			ExpectRateLimitSuccess()

		// Verify the expectation was set up (this will be validated when we execute)
		assert.NotNil(t, scenario.mockSetup.RateLimiter)
	})

	t.Run("ExpectCreationSuccess sets up POI service mock", func(t *testing.T) {
		expectedPOI := NewPOI().Build()
		scenario := NewPOITestScenario().
			ExpectCreationSuccess(expectedPOI)

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

		scenario := NewPOITestScenario().
			WithUser(userID).
			WithMap(mapID).
			ExpectRateLimitSuccess().
			ExpectCreationSuccess(expectedPOI)

		request := CreatePOIRequest{
			MapID:           mapID.String(),
			Name:            "Coffee Shop",
			Description:     "Great coffee place",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       userID.String(),
			MaxParticipants: 10,
		}

		response := scenario.CreatePOI(t, request)

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

		scenario := NewPOITestScenario().
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

		recorder := scenario.CreatePOIExpectingError(t, request)

		// Verify rate limit error response
		AssertHTTPStatus(t, recorder, http.StatusTooManyRequests)
		AssertErrorResponse(t, recorder, "RATE_LIMIT_EXCEEDED")

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI creation with validation error", func(t *testing.T) {
		scenario := NewPOITestScenario()

		request := CreatePOIRequest{
			MapID:           "", // Invalid: empty map ID
			Name:            "Coffee Shop",
			Description:     "Great coffee place",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       scenario.userID.String(),
			MaxParticipants: 10,
		}

		recorder := scenario.CreatePOIExpectingError(t, request)

		// Verify validation error response
		AssertHTTPStatus(t, recorder, http.StatusBadRequest)
		AssertErrorResponse(t, recorder, "INVALID_REQUEST")
	})
}

func TestPOITestScenario_JoinPOI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful POI join", func(t *testing.T) {
		userID := GenerateUUID()
		poiID := "poi-123"

		scenario := NewPOITestScenario().
			WithUser(userID).
			ExpectJoinRateLimitSuccessWithHeaders().
			ExpectJoinSuccess()

		request := JoinPOIRequest{
			UserID: userID.String(),
		}

		recorder := scenario.JoinPOI(t, poiID, request)

		// Verify successful join
		AssertHTTPStatus(t, recorder, http.StatusOK)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI join with capacity exceeded", func(t *testing.T) {
		userID := GenerateUUID()
		poiID := "poi-123"

		scenario := NewPOITestScenario().
			WithUser(userID).
			ExpectJoinRateLimitSuccess().
			ExpectCapacityExceeded()

		request := JoinPOIRequest{
			UserID: userID.String(),
		}

		recorder := scenario.JoinPOI(t, poiID, request)

		// Verify capacity exceeded error
		AssertHTTPStatus(t, recorder, http.StatusConflict)
		AssertErrorResponse(t, recorder, "CAPACITY_EXCEEDED")

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI join with POI not found", func(t *testing.T) {
		userID := GenerateUUID()
		poiID := "non-existent-poi"

		scenario := NewPOITestScenario().
			WithUser(userID).
			ExpectJoinRateLimitSuccess().
			ExpectNotFound()

		request := JoinPOIRequest{
			UserID: userID.String(),
		}

		recorder := scenario.JoinPOI(t, poiID, request)

		// Verify internal error (since we're returning a generic error)
		AssertHTTPStatus(t, recorder, http.StatusInternalServerError)
		AssertErrorResponse(t, recorder, "INTERNAL_ERROR")

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

		scenario := NewPOITestScenario().
			ExpectGetSuccess(expectedPOI)

		response := scenario.GetPOI(t, poiID)

		// Verify response
		assert.Equal(t, expectedPOI.ID, response.ID)
		assert.Equal(t, expectedPOI.Name, response.Name)

		scenario.mockSetup.AssertExpectations(t)
	})

	t.Run("POI not found", func(t *testing.T) {
		poiID := "non-existent-poi"

		scenario := NewPOITestScenario().
			ExpectGetNotFound()

		recorder := scenario.GetPOIExpectingError(t, poiID)

		// Verify not found error
		AssertHTTPStatus(t, recorder, http.StatusNotFound)
		AssertErrorResponse(t, recorder, "POI_NOT_FOUND")

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