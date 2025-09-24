package testdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"breakoutglobe/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPOIResponseAssertion(t *testing.T) {
	t.Run("asserts POI response fields correctly", func(t *testing.T) {
		expectedPOI := NewPOI().
			WithID("poi-123").
			WithName("Coffee Shop").
			WithDescription("Great coffee").
			WithPosition(models.LatLng{Lat: 40.7128, Lng: -74.0060}).
			Build()

		response := CreatePOIResponse{
			ID:          expectedPOI.ID,
			Name:        expectedPOI.Name,
			Description: expectedPOI.Description,
			Position:    expectedPOI.Position,
		}

		// This should not panic or fail
		AssertPOIResponse(t, &response, expectedPOI)
	})

	t.Run("ignores timestamps in comparison", func(t *testing.T) {
		expectedPOI := NewPOI().
			WithID("poi-123").
			WithName("Coffee Shop").
			Build()

		response := CreatePOIResponse{
			ID:          expectedPOI.ID,
			Name:        expectedPOI.Name,
			Description: expectedPOI.Description,
			Position:    expectedPOI.Position,
			// CreatedAt is not compared - should be ignored
		}

		// This should pass even though CreatedAt differs
		AssertPOIResponse(t, &response, expectedPOI)
	})

	t.Run("fails when core fields differ", func(t *testing.T) {
		expectedPOI := NewPOI().
			WithID("poi-123").
			WithName("Coffee Shop").
			Build()

		response := CreatePOIResponse{
			ID:          "different-id", // This should cause failure
			Name:        expectedPOI.Name,
			Description: expectedPOI.Description,
			Position:    expectedPOI.Position,
		}

		// Create a mock testing.T to capture the failure
		mockT := &MockTestingT{}
		AssertPOIResponse(mockT, &response, expectedPOI)

		// Verify that the assertion failed
		assert.True(t, mockT.Failed, "Expected assertion to fail when IDs differ")
	})
}

func TestSessionResponseAssertion(t *testing.T) {
	t.Run("asserts session response fields correctly", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()
		
		expectedSession := NewSession().
			WithID("session-123").
			WithUser(userID).
			WithMap(mapID).
			WithPosition(models.LatLng{Lat: 37.7749, Lng: -122.4194}).
			Build()

		response := SessionResponse{
			ID:            expectedSession.ID,
			UserID:        expectedSession.UserID,
			MapID:         expectedSession.MapID,
			AvatarPos:     expectedSession.AvatarPos,
			IsActive:      expectedSession.IsActive,
		}

		// This should not panic or fail
		AssertSessionResponse(t, &response, expectedSession)
	})

	t.Run("ignores timestamps in comparison", func(t *testing.T) {
		expectedSession := NewSession().
			WithID("session-123").
			Build()

		response := SessionResponse{
			ID:            expectedSession.ID,
			UserID:        expectedSession.UserID,
			MapID:         expectedSession.MapID,
			AvatarPos:     expectedSession.AvatarPos,
			IsActive:      expectedSession.IsActive,
			// CreatedAt and LastActive are not compared
		}

		// This should pass even though timestamps differ
		AssertSessionResponse(t, &response, expectedSession)
	})
}

func TestHTTPStatusAssertion(t *testing.T) {
	t.Run("passes when status matches", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusCreated)

		// This should not panic or fail
		AssertHTTPStatus(t, recorder, http.StatusCreated)
	})

	t.Run("fails with detailed message when status differs", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteHeader(http.StatusBadRequest)
		recorder.WriteString(`{"error": "validation failed"}`)

		// Create a mock testing.T to capture the failure
		mockT := &MockTestingT{}
		AssertHTTPStatus(mockT, recorder, http.StatusCreated)

		// Verify that the assertion failed with detailed message
		assert.True(t, mockT.Failed, "Expected assertion to fail when status differs")
		assert.Contains(t, mockT.ErrorMessage, "Expected status 201, got 400")
		assert.Contains(t, mockT.ErrorMessage, "validation failed")
	})
}

func TestErrorResponseAssertion(t *testing.T) {
	t.Run("asserts error response correctly", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errorResponse := ErrorResponse{
			Code:    "POI_NOT_FOUND",
			Message: "POI not found",
		}
		
		responseBytes, _ := json.Marshal(errorResponse)
		recorder.WriteString(string(responseBytes))

		// This should not panic or fail
		AssertErrorResponse(t, recorder, "POI_NOT_FOUND")
	})

	t.Run("fails when error code differs", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errorResponse := ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input",
		}
		
		responseBytes, _ := json.Marshal(errorResponse)
		recorder.WriteString(string(responseBytes))

		// Create a mock testing.T to capture the failure
		mockT := &MockTestingT{}
		AssertErrorResponse(mockT, recorder, "POI_NOT_FOUND")

		// Verify that the assertion failed
		assert.True(t, mockT.Failed, "Expected assertion to fail when error codes differ")
	})

	t.Run("fails when response is not valid JSON", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		recorder.WriteString("invalid json")

		// Create a mock testing.T to capture the failure
		mockT := &MockTestingT{}
		AssertErrorResponse(mockT, recorder, "POI_NOT_FOUND")

		// Verify that the assertion failed
		assert.True(t, mockT.Failed, "Expected assertion to fail when JSON is invalid")
	})
}

func TestUUIDAssertion(t *testing.T) {
	t.Run("passes when UUID is valid", func(t *testing.T) {
		validUUID := GenerateUUID().String()

		// This should not panic or fail
		AssertValidUUID(t, validUUID)
	})

	t.Run("fails when UUID is invalid", func(t *testing.T) {
		invalidUUID := "not-a-uuid"

		// Create a mock testing.T to capture the failure
		mockT := &MockTestingT{}
		AssertValidUUID(mockT, invalidUUID)

		// Verify that the assertion failed
		assert.True(t, mockT.Failed, "Expected assertion to fail for invalid UUID")
	})

	t.Run("fails when UUID is empty", func(t *testing.T) {
		emptyUUID := ""

		// Create a mock testing.T to capture the failure
		mockT := &MockTestingT{}
		AssertValidUUID(mockT, emptyUUID)

		// Verify that the assertion failed
		assert.True(t, mockT.Failed, "Expected assertion to fail for empty UUID")
	})
}

// Mock types for testing - using the ones from assertions.go

// MockTestingT implements a subset of testing.T for capturing test failures
type MockTestingT struct {
	Failed       bool
	ErrorMessage string
}

func (m *MockTestingT) Errorf(format string, args ...interface{}) {
	m.Failed = true
	m.ErrorMessage = strings.TrimSpace(fmt.Sprintf(format, args...))
}

func (m *MockTestingT) Helper() {
	// No-op for mock
}

func (m *MockTestingT) Cleanup(func()) {
	// No-op for mock
}
func TestUserAssertion(t *testing.T) {
	t.Run("HasEmail assertion", func(t *testing.T) {
		user := NewUser().
			WithEmail("test@example.com").
			Build()

		// This should pass
		AssertUser(t, user).HasEmail("test@example.com")
	})

	t.Run("HasDisplayName assertion", func(t *testing.T) {
		user := NewUser().
			WithDisplayName("John Doe").
			Build()

		// This should pass
		AssertUser(t, user).HasDisplayName("John Doe")
	})

	t.Run("HasRole assertion", func(t *testing.T) {
		user := NewUser().
			WithRole(models.UserRoleAdmin).
			Build()

		// This should pass
		AssertUser(t, user).HasRole(models.UserRoleAdmin)
	})

	t.Run("HasAccountType assertion", func(t *testing.T) {
		user := NewUser().
			WithAccountType(models.AccountTypeGuest).
			Build()

		// This should pass
		AssertUser(t, user).HasAccountType(models.AccountTypeGuest)
	})

	t.Run("IsActive assertion", func(t *testing.T) {
		user := NewUser().
			WithActive(true).
			Build()

		// This should pass
		AssertUser(t, user).IsActive()
	})

	t.Run("IsGuest assertion", func(t *testing.T) {
		user := NewUser().
			AsGuest().
			Build()

		// This should pass
		AssertUser(t, user).IsGuest()
	})

	t.Run("IsFull assertion", func(t *testing.T) {
		user := NewUser().
			WithAccountType(models.AccountTypeFull).
			Build()

		// This should pass
		AssertUser(t, user).IsFull()
	})

	t.Run("IsAdmin assertion", func(t *testing.T) {
		user := NewUser().
			AsAdmin().
			Build()

		// This should pass
		AssertUser(t, user).IsAdmin()
	})

	t.Run("IsSuperAdmin assertion", func(t *testing.T) {
		user := NewUser().
			AsSuperAdmin().
			Build()

		// This should pass
		AssertUser(t, user).IsSuperAdmin()
	})

	t.Run("HasPassword assertion", func(t *testing.T) {
		user := NewUser().
			WithPasswordHash("hashed-password").
			Build()

		// This should pass
		AssertUser(t, user).HasPassword()
	})

	t.Run("HasNoPassword assertion", func(t *testing.T) {
		user := NewUser().
			Build() // No password hash set

		// This should pass
		AssertUser(t, user).HasNoPassword()
	})

	t.Run("chained assertions", func(t *testing.T) {
		user := NewUser().
			WithEmail("admin@example.com").
			WithDisplayName("Admin User").
			AsAdmin().
			WithPasswordHash("secure-hash").
			Build()

		// Test fluent chaining
		AssertUser(t, user).
			HasEmail("admin@example.com").
			HasDisplayName("Admin User").
			HasRole(models.UserRoleAdmin).
			IsActive().
			IsFull().
			IsAdmin().
			HasPassword()
	})
}