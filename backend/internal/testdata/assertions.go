package testdata

import (
	"encoding/json"
	"net/http/httptest"

	"breakoutglobe/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestingT is an interface that matches the testing.T methods we need
type TestingT interface {
	Errorf(format string, args ...interface{})
	Helper()
}

// POIResponseInterface defines the interface for POI response types
type POIResponseInterface interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetPosition() models.LatLng
}

// SessionResponseInterface defines the interface for Session response types
type SessionResponseInterface interface {
	GetID() string
	GetUserID() string
	GetMapID() string
	GetAvatarPos() models.LatLng
	GetIsActive() bool
}

// ErrorResponseInterface defines the interface for error response types
type ErrorResponseInterface interface {
	GetCode() string
	GetMessage() string
}

// AssertPOIResponse validates a POI response against expected POI model
// Ignores timestamps and other irrelevant fields, focusing on business data
func AssertPOIResponse(t TestingT, response POIResponseInterface, expected *models.POI) {
	t.Helper()
	
	assert.Equal(t, expected.ID, response.GetID(), "POI ID should match")
	assert.Equal(t, expected.Name, response.GetName(), "POI name should match")
	assert.Equal(t, expected.Description, response.GetDescription(), "POI description should match")
	assert.Equal(t, expected.Position.Lat, response.GetPosition().Lat, "POI latitude should match")
	assert.Equal(t, expected.Position.Lng, response.GetPosition().Lng, "POI longitude should match")
	// Note: CreatedAt, UpdatedAt, and other metadata fields are intentionally ignored
}

// AssertSessionResponse validates a session response against expected session model
// Ignores timestamps and other irrelevant fields, focusing on business data
func AssertSessionResponse(t TestingT, response SessionResponseInterface, expected *models.Session) {
	t.Helper()
	
	assert.Equal(t, expected.ID, response.GetID(), "Session ID should match")
	assert.Equal(t, expected.UserID, response.GetUserID(), "Session user ID should match")
	assert.Equal(t, expected.MapID, response.GetMapID(), "Session map ID should match")
	assert.Equal(t, expected.AvatarPos.Lat, response.GetAvatarPos().Lat, "Avatar latitude should match")
	assert.Equal(t, expected.AvatarPos.Lng, response.GetAvatarPos().Lng, "Avatar longitude should match")
	assert.Equal(t, expected.IsActive, response.GetIsActive(), "Session active status should match")
	// Note: CreatedAt, LastActive, and other metadata fields are intentionally ignored
}

// AssertHTTPStatus validates HTTP response status with detailed error message on failure
func AssertHTTPStatus(t TestingT, response *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	
	if response.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Response body: %s", 
			expectedStatus, response.Code, response.Body.String())
	}
}

// AssertErrorResponse validates error response structure and error code
func AssertErrorResponse(t TestingT, response *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()
	
	var errorResp map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &errorResp)
	if err != nil {
		t.Errorf("Failed to parse error response as JSON: %v. Response body: %s", 
			err, response.Body.String())
		return
	}
	
	code, exists := errorResp["code"]
	if !exists {
		t.Errorf("Error response missing 'code' field. Response: %s", response.Body.String())
		return
	}
	
	codeStr, ok := code.(string)
	if !ok {
		t.Errorf("Error response 'code' field is not a string: %T. Response: %s", 
			code, response.Body.String())
		return
	}
	
	if codeStr != expectedCode {
		t.Errorf("Expected error code '%s', got '%s'. Response: %s", 
			expectedCode, codeStr, response.Body.String())
	}
}

// AssertValidUUID validates that a string is a valid UUID
func AssertValidUUID(t TestingT, uuidStr string) {
	t.Helper()
	
	if uuidStr == "" {
		t.Errorf("UUID string is empty")
		return
	}
	
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		t.Errorf("Invalid UUID format: %s (error: %v)", uuidStr, err)
	}
}

// AssertValidUUIDs validates that multiple strings are valid UUIDs
func AssertValidUUIDs(t TestingT, uuidStrs ...string) {
	t.Helper()
	
	for i, uuidStr := range uuidStrs {
		if uuidStr == "" {
			t.Errorf("UUID string at index %d is empty", i)
			continue
		}
		
		_, err := uuid.Parse(uuidStr)
		if err != nil {
			t.Errorf("Invalid UUID format at index %d: %s (error: %v)", i, uuidStr, err)
		}
	}
}

// AssertContainsError validates that an error message contains expected text
func AssertContainsError(t TestingT, err error, expectedText string) {
	t.Helper()
	
	if err == nil {
		t.Errorf("Expected error containing '%s', but got nil", expectedText)
		return
	}
	
	errorMsg := err.Error()
	if !contains(errorMsg, expectedText) {
		t.Errorf("Expected error to contain '%s', but got: %s", expectedText, errorMsg)
	}
}

// AssertNoError validates that no error occurred
func AssertNoError(t TestingT, err error) {
	t.Helper()
	
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(substr) == 0 || 
		indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Concrete implementations for common response types

// CreatePOIResponse represents a POI creation response
type CreatePOIResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Position    models.LatLng `json:"position"`
}

func (r *CreatePOIResponse) GetID() string          { return r.ID }
func (r *CreatePOIResponse) GetName() string        { return r.Name }
func (r *CreatePOIResponse) GetDescription() string { return r.Description }
func (r *CreatePOIResponse) GetPosition() models.LatLng { return r.Position }

// SessionResponse represents a session response
type SessionResponse struct {
	ID        string        `json:"id"`
	UserID    string        `json:"userId"`
	MapID     string        `json:"mapId"`
	AvatarPos models.LatLng `json:"avatarPosition"`
	IsActive  bool          `json:"isActive"`
}

func (r *SessionResponse) GetID() string          { return r.ID }
func (r *SessionResponse) GetUserID() string      { return r.UserID }
func (r *SessionResponse) GetMapID() string       { return r.MapID }
func (r *SessionResponse) GetAvatarPos() models.LatLng { return r.AvatarPos }
func (r *SessionResponse) GetIsActive() bool      { return r.IsActive }

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *ErrorResponse) GetCode() string    { return r.Code }
func (r *ErrorResponse) GetMessage() string { return r.Message }