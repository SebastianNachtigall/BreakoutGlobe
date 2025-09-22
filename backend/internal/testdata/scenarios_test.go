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