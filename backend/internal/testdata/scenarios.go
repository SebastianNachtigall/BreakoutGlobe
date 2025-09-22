package testdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"breakoutglobe/internal/handlers"
	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// POITestScenario provides a fluent API for testing POI-related functionality
type POITestScenario struct {
	mockSetup *MockSetup
	userID    uuid.UUID
	mapID     uuid.UUID
	handler   *handlers.POIHandler
	router    *gin.Engine
}

// NewPOITestScenario creates a new POI test scenario with sensible defaults
func NewPOITestScenario() *POITestScenario {
	mockSetup := NewMockSetup()
	
	scenario := &POITestScenario{
		mockSetup: mockSetup,
		userID:    GenerateUUID(),
		mapID:     GenerateUUID(),
	}
	
	// Create handler with mocks
	scenario.handler = handlers.NewPOIHandler(
		mockSetup.POIService.Mock(),
		mockSetup.RateLimiter.Mock(),
	)
	
	// Setup router
	gin.SetMode(gin.TestMode)
	scenario.router = gin.New()
	scenario.handler.RegisterRoutes(scenario.router)
	
	return scenario
}

// WithUser sets a custom user ID for the scenario
func (s *POITestScenario) WithUser(userID uuid.UUID) *POITestScenario {
	s.userID = userID
	return s
}

// WithMap sets a custom map ID for the scenario
func (s *POITestScenario) WithMap(mapID uuid.UUID) *POITestScenario {
	s.mapID = mapID
	return s
}

// ExpectRateLimitSuccess sets up the rate limiter to allow the request for CREATE action
func (s *POITestScenario) ExpectRateLimitSuccess() *POITestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreatePOI).
		Returns()
	
	s.mockSetup.RateLimiter.ExpectGetRateLimitHeaders().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreatePOI).
		Returns(map[string]string{
			"X-RateLimit-Limit":     "5",
			"X-RateLimit-Remaining": "4",
		})
	
	return s
}

// ExpectJoinRateLimitSuccess sets up the rate limiter to allow JOIN requests
func (s *POITestScenario) ExpectJoinRateLimitSuccess() *POITestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionJoinPOI).
		Returns()
	
	return s
}

// ExpectJoinRateLimitSuccessWithHeaders sets up rate limiter for successful join (with headers)
func (s *POITestScenario) ExpectJoinRateLimitSuccessWithHeaders() *POITestScenario {
	s.ExpectJoinRateLimitSuccess()
	
	s.mockSetup.RateLimiter.ExpectGetRateLimitHeaders().
		WithUserID(s.userID.String()).
		WithAction(services.ActionJoinPOI).
		Returns(map[string]string{
			"X-RateLimit-Limit":     "20",
			"X-RateLimit-Remaining": "19",
		})
	
	return s
}

// ExpectRateLimitExceeded sets up the rate limiter to reject the request
func (s *POITestScenario) ExpectRateLimitExceeded() *POITestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreatePOI).
		ReturnsRateLimitExceeded()
	
	return s
}

// ExpectCreationSuccess sets up the POI service to successfully create a POI
func (s *POITestScenario) ExpectCreationSuccess(expectedPOI *models.POI) *POITestScenario {
	s.mockSetup.POIService.ExpectCreatePOI().
		WithMapID(s.mapID.String()).
		WithCreatedBy(s.userID.String()).
		Returns(expectedPOI)
	
	return s
}

// ExpectJoinSuccess sets up the POI service to successfully join a POI
func (s *POITestScenario) ExpectJoinSuccess() *POITestScenario {
	// Mock JoinPOI method
	s.mockSetup.POIService.Mock().On("JoinPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"), // poiID
		s.userID.String(),
	).Return(nil)
	
	return s
}

// ExpectCapacityExceeded sets up the POI service to return capacity exceeded error
func (s *POITestScenario) ExpectCapacityExceeded() *POITestScenario {
	s.mockSetup.POIService.Mock().On("JoinPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"),
		s.userID.String(),
	).Return(fmt.Errorf("POI capacity exceeded"))
	
	return s
}

// ExpectNotFound sets up the POI service to return not found error
func (s *POITestScenario) ExpectNotFound() *POITestScenario {
	s.mockSetup.POIService.Mock().On("JoinPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"),
		s.userID.String(),
	).Return(fmt.Errorf("POI not found"))
	
	return s
}

// ExpectGetSuccess sets up the POI service to successfully retrieve a POI
func (s *POITestScenario) ExpectGetSuccess(expectedPOI *models.POI) *POITestScenario {
	s.mockSetup.POIService.ExpectGetPOI(expectedPOI.ID).Returns(expectedPOI)
	return s
}

// ExpectGetNotFound sets up the POI service to return not found for GET
func (s *POITestScenario) ExpectGetNotFound() *POITestScenario {
	s.mockSetup.POIService.Mock().On("GetPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"),
	).Return(nil, gorm.ErrRecordNotFound)
	
	return s
}

// CreatePOI executes a POI creation request and returns the response
func (s *POITestScenario) CreatePOI(t TestingT, request CreatePOIRequest) *CreatePOIResponse {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	// Verify success status
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			http.StatusCreated, recorder.Code, recorder.Body.String())
		return nil
	}
	
	// Parse response
	var response CreatePOIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// CreatePOIExpectingError executes a POI creation request expecting an error
func (s *POITestScenario) CreatePOIExpectingError(t TestingT, request CreatePOIRequest) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// JoinPOI executes a POI join request
func (s *POITestScenario) JoinPOI(t TestingT, poiID string, request JoinPOIRequest) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	url := fmt.Sprintf("/api/pois/%s/join", poiID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// GetPOI executes a POI retrieval request and returns the response
func (s *POITestScenario) GetPOI(t TestingT, poiID string) *GetPOIResponse {
	t.Helper()
	
	url := fmt.Sprintf("/api/pois/%s", poiID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	// Verify success status
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			http.StatusOK, recorder.Code, recorder.Body.String())
		return nil
	}
	
	// Parse response
	var response GetPOIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// GetPOIExpectingError executes a POI retrieval request expecting an error
func (s *POITestScenario) GetPOIExpectingError(t TestingT, poiID string) *httptest.ResponseRecorder {
	t.Helper()
	
	url := fmt.Sprintf("/api/pois/%s", poiID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// Request/Response types for POI scenarios

// CreatePOIRequest represents a POI creation request
type CreatePOIRequest struct {
	MapID           string        `json:"mapId"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Position        models.LatLng `json:"position"`
	CreatedBy       string        `json:"createdBy"`
	MaxParticipants int           `json:"maxParticipants"`
}

// CreatePOIResponse is defined in assertions.go

// JoinPOIRequest represents a POI join request
type JoinPOIRequest struct {
	UserID string `json:"userId"`
}

// GetPOIResponse represents a POI retrieval response
type GetPOIResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Position    models.LatLng `json:"position"`
	CreatedBy   string        `json:"createdBy"`
}