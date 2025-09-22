package testdata

import (
	"context"
	"errors"
	"testing"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMockSetup(t *testing.T) {
	t.Run("creates all mock services", func(t *testing.T) {
		setup := NewMockSetup()

		assert.NotNil(t, setup.POIService)
		assert.NotNil(t, setup.SessionService)
		assert.NotNil(t, setup.RateLimiter)
	})

	t.Run("provides access to underlying mocks", func(t *testing.T) {
		setup := NewMockSetup()

		assert.NotNil(t, setup.POIService.Mock())
		assert.NotNil(t, setup.SessionService.Mock())
		assert.NotNil(t, setup.RateLimiter.Mock())
	})

	t.Run("can verify all expectations", func(t *testing.T) {
		setup := NewMockSetup()

		// This should not panic when no expectations are set
		setup.AssertExpectations(t)
	})
}

func TestMockPOIServiceBuilder(t *testing.T) {
	t.Run("ExpectCreatePOI with success", func(t *testing.T) {
		setup := NewMockSetup()
		expectedPOI := NewPOI().WithName("Test POI").Build()

		// Setup expectation
		setup.POIService.ExpectCreatePOI().
			WithMapID("map-123").
			WithCreatedBy("user-123").
			Returns(expectedPOI)

		// Execute the call
		result, err := setup.POIService.Mock().CreatePOI(
			context.Background(),
			"map-123",
			"Test POI",
			"Description",
			models.LatLng{Lat: 40.7128, Lng: -74.0060},
			"user-123",
			10,
		)

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, expectedPOI, result)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectCreatePOI with error", func(t *testing.T) {
		setup := NewMockSetup()
		expectedError := errors.New("creation failed")

		// Setup expectation
		setup.POIService.ExpectCreatePOI().
			WithMapID("map-123").
			WithCreatedBy("user-123").
			ReturnsError(expectedError)

		// Execute the call
		result, err := setup.POIService.Mock().CreatePOI(
			context.Background(),
			"map-123",
			"Test POI",
			"Description",
			models.LatLng{Lat: 40.7128, Lng: -74.0060},
			"user-123",
			10,
		)

		// Verify
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectGetPOI with success", func(t *testing.T) {
		setup := NewMockSetup()
		expectedPOI := NewPOI().WithID("poi-123").Build()

		// Setup expectation
		setup.POIService.ExpectGetPOI("poi-123").Returns(expectedPOI)

		// Execute the call
		result, err := setup.POIService.Mock().GetPOI(context.Background(), "poi-123")

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, expectedPOI, result)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectGetPOI with not found error", func(t *testing.T) {
		setup := NewMockSetup()

		// Setup expectation
		setup.POIService.ExpectGetPOI("poi-123").ReturnsNotFound()

		// Execute the call
		result, err := setup.POIService.Mock().GetPOI(context.Background(), "poi-123")

		// Verify
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "POI not found")
		setup.AssertExpectations(t)
	})
}

func TestMockRateLimiterBuilder(t *testing.T) {
	t.Run("ExpectCheckRateLimit with success", func(t *testing.T) {
		setup := NewMockSetup()

		// Setup expectation
		setup.RateLimiter.ExpectCheckRateLimit().
			WithUserID("user-123").
			WithAction(services.ActionCreatePOI).
			Returns()

		// Execute the call (using gin context)
		gin.SetMode(gin.TestMode)
		ctx, _ := gin.CreateTestContext(nil)
		err := setup.RateLimiter.Mock().CheckRateLimit(ctx, "user-123", services.ActionCreatePOI)

		// Verify
		assert.NoError(t, err)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectCheckRateLimit with rate limit exceeded", func(t *testing.T) {
		setup := NewMockSetup()

		// Setup expectation
		setup.RateLimiter.ExpectCheckRateLimit().
			WithUserID("user-123").
			WithAction(services.ActionCreatePOI).
			ReturnsRateLimitExceeded()

		// Execute the call
		gin.SetMode(gin.TestMode)
		ctx, _ := gin.CreateTestContext(nil)
		err := setup.RateLimiter.Mock().CheckRateLimit(ctx, "user-123", services.ActionCreatePOI)

		// Verify
		assert.IsType(t, &services.RateLimitError{}, err)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectGetRateLimitHeaders with success", func(t *testing.T) {
		setup := NewMockSetup()
		expectedHeaders := map[string]string{
			"X-RateLimit-Limit":     "5",
			"X-RateLimit-Remaining": "4",
		}

		// Setup expectation
		setup.RateLimiter.ExpectGetRateLimitHeaders().
			WithUserID("user-123").
			WithAction(services.ActionCreatePOI).
			Returns(expectedHeaders)

		// Execute the call
		gin.SetMode(gin.TestMode)
		ctx, _ := gin.CreateTestContext(nil)
		headers, err := setup.RateLimiter.Mock().GetRateLimitHeaders(ctx, "user-123", services.ActionCreatePOI)

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, expectedHeaders, headers)
		setup.AssertExpectations(t)
	})
}

func TestMockSessionServiceBuilder(t *testing.T) {
	t.Run("ExpectCreateSession with success", func(t *testing.T) {
		setup := NewMockSetup()
		expectedSession := NewSession().WithID("session-123").Build()

		// Setup expectation
		setup.SessionService.ExpectCreateSession().
			WithUserID("user-123").
			WithMapID("map-123").
			Returns(expectedSession)

		// Execute the call
		result, err := setup.SessionService.Mock().CreateSession(
			context.Background(),
			"user-123",
			"map-123",
			models.LatLng{Lat: 40.7128, Lng: -74.0060},
		)

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, expectedSession, result)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectGetSession with success", func(t *testing.T) {
		setup := NewMockSetup()
		expectedSession := NewSession().WithID("session-123").Build()

		// Setup expectation
		setup.SessionService.ExpectGetSession("session-123").Returns(expectedSession)

		// Execute the call
		result, err := setup.SessionService.Mock().GetSession(context.Background(), "session-123")

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, expectedSession, result)
		setup.AssertExpectations(t)
	})

	t.Run("ExpectGetSession with not found error", func(t *testing.T) {
		setup := NewMockSetup()

		// Setup expectation
		setup.SessionService.ExpectGetSession("session-123").ReturnsNotFound()

		// Execute the call
		result, err := setup.SessionService.Mock().GetSession(context.Background(), "session-123")

		// Verify
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session not found")
		setup.AssertExpectations(t)
	})
}