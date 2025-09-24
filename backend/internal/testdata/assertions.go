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
	Cleanup(func())
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

// UserAssertion provides fluent assertions for User models
type UserAssertion struct {
	t    TestingT
	user *models.User
}

// AssertUser provides fluent assertions for User models
func AssertUser(t TestingT, user *models.User) *UserAssertion {
	t.Helper()
	return &UserAssertion{t: t, user: user}
}

// HasID asserts the user has the expected ID
func (a *UserAssertion) HasID(expectedID string) *UserAssertion {
	a.t.Helper()
	if a.user.ID != expectedID {
		a.t.Errorf("Expected user ID %s, got %s", expectedID, a.user.ID)
	}
	return a
}

// HasEmail asserts the user has the expected email
func (a *UserAssertion) HasEmail(expectedEmail string) *UserAssertion {
	a.t.Helper()
	if a.user.Email != expectedEmail {
		a.t.Errorf("Expected user email %s, got %s", expectedEmail, a.user.Email)
	}
	return a
}

// HasDisplayName asserts the user has the expected display name
func (a *UserAssertion) HasDisplayName(expectedDisplayName string) *UserAssertion {
	a.t.Helper()
	if a.user.DisplayName != expectedDisplayName {
		a.t.Errorf("Expected user display name %s, got %s", expectedDisplayName, a.user.DisplayName)
	}
	return a
}

// HasAvatarURL asserts the user has the expected avatar URL
func (a *UserAssertion) HasAvatarURL(expectedAvatarURL string) *UserAssertion {
	a.t.Helper()
	if a.user.AvatarURL != expectedAvatarURL {
		a.t.Errorf("Expected user avatar URL %s, got %s", expectedAvatarURL, a.user.AvatarURL)
	}
	return a
}

// HasAboutMe asserts the user has the expected about me text
func (a *UserAssertion) HasAboutMe(expectedAboutMe string) *UserAssertion {
	a.t.Helper()
	if a.user.AboutMe != expectedAboutMe {
		a.t.Errorf("Expected user about me %s, got %s", expectedAboutMe, a.user.AboutMe)
	}
	return a
}

// HasAccountType asserts the user has the expected account type
func (a *UserAssertion) HasAccountType(expectedAccountType models.AccountType) *UserAssertion {
	a.t.Helper()
	if a.user.AccountType != expectedAccountType {
		a.t.Errorf("Expected user account type %s, got %s", expectedAccountType, a.user.AccountType)
	}
	return a
}

// HasRole asserts the user has the expected role
func (a *UserAssertion) HasRole(expectedRole models.UserRole) *UserAssertion {
	a.t.Helper()
	if a.user.Role != expectedRole {
		a.t.Errorf("Expected user role %s, got %s", expectedRole, a.user.Role)
	}
	return a
}

// HasPassword asserts the user has a password set
func (a *UserAssertion) HasPassword() *UserAssertion {
	a.t.Helper()
	if !a.user.HasPassword() {
		a.t.Errorf("Expected user to have a password, but password hash is empty")
	}
	return a
}

// HasNoPassword asserts the user has no password set
func (a *UserAssertion) HasNoPassword() *UserAssertion {
	a.t.Helper()
	if a.user.HasPassword() {
		a.t.Errorf("Expected user to have no password, but password hash is set")
	}
	return a
}

// IsActive asserts the user is active
func (a *UserAssertion) IsActive() *UserAssertion {
	a.t.Helper()
	if !a.user.IsActive {
		a.t.Errorf("Expected user to be active, but user is inactive")
	}
	return a
}

// IsInactive asserts the user is inactive
func (a *UserAssertion) IsInactive() *UserAssertion {
	a.t.Helper()
	if a.user.IsActive {
		a.t.Errorf("Expected user to be inactive, but user is active")
	}
	return a
}

// IsGuest asserts the user is a guest account
func (a *UserAssertion) IsGuest() *UserAssertion {
	a.t.Helper()
	if !a.user.IsGuest() {
		a.t.Errorf("Expected user to be a guest account")
	}
	return a
}

// IsFull asserts the user is a full account
func (a *UserAssertion) IsFull() *UserAssertion {
	a.t.Helper()
	if !a.user.IsFull() {
		a.t.Errorf("Expected user to be a full account")
	}
	return a
}

// IsAdmin asserts the user has admin privileges
func (a *UserAssertion) IsAdmin() *UserAssertion {
	a.t.Helper()
	if !a.user.IsAdmin() {
		a.t.Errorf("Expected user to have admin privileges")
	}
	return a
}

// IsSuperAdmin asserts the user is a super admin
func (a *UserAssertion) IsSuperAdmin() *UserAssertion {
	a.t.Helper()
	if !a.user.IsSuperAdmin() {
		a.t.Errorf("Expected user to be a super admin")
	}
	return a
}

// HasCreatedAt asserts the user has a creation timestamp
func (a *UserAssertion) HasCreatedAt() *UserAssertion {
	a.t.Helper()
	if a.user.CreatedAt.IsZero() {
		a.t.Errorf("Expected user to have a creation timestamp")
	}
	return a
}

// HasUpdatedAt asserts the user has an update timestamp
func (a *UserAssertion) HasUpdatedAt() *UserAssertion {
	a.t.Helper()
	if a.user.UpdatedAt.IsZero() {
		a.t.Errorf("Expected user to have an update timestamp")
	}
	return a
}

// UserResponseInterface defines the interface for User response types
type UserResponseInterface interface {
	GetID() string
	GetDisplayName() string
	GetAccountType() string
	GetRole() string
	GetIsActive() bool
}

// UserResponseAssertion provides fluent assertions for User response types
type UserResponseAssertion struct {
	t        TestingT
	response UserResponseInterface
}

// AssertUserResponse provides fluent assertions for User response types
func AssertUserResponse(t TestingT, response UserResponseInterface) *UserResponseAssertion {
	t.Helper()
	return &UserResponseAssertion{t: t, response: response}
}

// HasID asserts the response has the expected ID
func (a *UserResponseAssertion) HasID(expectedID string) *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetID() != expectedID {
		a.t.Errorf("Expected user response ID %s, got %s", expectedID, a.response.GetID())
	}
	return a
}

// HasDisplayName asserts the response has the expected display name
func (a *UserResponseAssertion) HasDisplayName(expectedDisplayName string) *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetDisplayName() != expectedDisplayName {
		a.t.Errorf("Expected user response display name %s, got %s", expectedDisplayName, a.response.GetDisplayName())
	}
	return a
}

// HasAccountType asserts the response has the expected account type
func (a *UserResponseAssertion) HasAccountType(expectedAccountType string) *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetAccountType() != expectedAccountType {
		a.t.Errorf("Expected user response account type %s, got %s", expectedAccountType, a.response.GetAccountType())
	}
	return a
}

// HasRole asserts the response has the expected role
func (a *UserResponseAssertion) HasRole(expectedRole string) *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetRole() != expectedRole {
		a.t.Errorf("Expected user response role %s, got %s", expectedRole, a.response.GetRole())
	}
	return a
}

// IsActive asserts the response indicates the user is active
func (a *UserResponseAssertion) IsActive() *UserResponseAssertion {
	a.t.Helper()
	if !a.response.GetIsActive() {
		a.t.Errorf("Expected user response to indicate active user, but got inactive")
	}
	return a
}

// IsInactive asserts the response indicates the user is inactive
func (a *UserResponseAssertion) IsInactive() *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetIsActive() {
		a.t.Errorf("Expected user response to indicate inactive user, but got active")
	}
	return a
}

// IsGuest asserts the response indicates a guest account
func (a *UserResponseAssertion) IsGuest() *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetAccountType() != "guest" {
		a.t.Errorf("Expected user response to indicate guest account, got %s", a.response.GetAccountType())
	}
	return a
}

// IsFull asserts the response indicates a full account
func (a *UserResponseAssertion) IsFull() *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetAccountType() != "full" {
		a.t.Errorf("Expected user response to indicate full account, got %s", a.response.GetAccountType())
	}
	return a
}

// HasUserRole asserts the response indicates a user role
func (a *UserResponseAssertion) HasUserRole() *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetRole() != "user" {
		a.t.Errorf("Expected user response to indicate user role, got %s", a.response.GetRole())
	}
	return a
}

// HasAdminRole asserts the response indicates an admin role
func (a *UserResponseAssertion) HasAdminRole() *UserResponseAssertion {
	a.t.Helper()
	if a.response.GetRole() != "admin" {
		a.t.Errorf("Expected user response to indicate admin role, got %s", a.response.GetRole())
	}
	return a
}

// HasValidID asserts the response has a valid UUID as ID
func (a *UserResponseAssertion) HasValidID() *UserResponseAssertion {
	a.t.Helper()
	AssertValidUUID(a.t, a.response.GetID())
	return a
}