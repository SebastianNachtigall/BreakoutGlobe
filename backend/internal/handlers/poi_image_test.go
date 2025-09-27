package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// POI Image Creation Test Scenario
type poiImageScenario struct {
	t               *testing.T
	mockPOIService  *MockPOIService
	mockRateLimiter *MockRateLimiter
	handler         *POIHandler
	router          *gin.Engine
	userID          string
	mapID           string
}

func newPOIImageScenario(t *testing.T) *poiImageScenario {
	gin.SetMode(gin.TestMode)
	
	mockPOIService := new(MockPOIService)
	mockRateLimiter := new(MockRateLimiter)
	handler := NewPOIHandler(mockPOIService, mockRateLimiter)
	
	router := gin.New()
	handler.RegisterRoutes(router)
	
	return &poiImageScenario{
		t:               t,
		mockPOIService:  mockPOIService,
		mockRateLimiter: mockRateLimiter,
		handler:         handler,
		router:          router,
		userID:          "user-123",
		mapID:           "map-123",
	}
}

func (s *poiImageScenario) expectRateLimitSuccess() *poiImageScenario {
	s.mockRateLimiter.On("CheckRateLimit", mock.Anything, s.userID, services.ActionCreatePOI).Return(nil)
	s.mockRateLimiter.On("GetRateLimitHeaders", mock.Anything, s.userID, services.ActionCreatePOI).Return(map[string]string{
		"X-RateLimit-Limit":     "5",
		"X-RateLimit-Remaining": "4",
	}, nil)
	return s
}

func (s *poiImageScenario) expectPOICreationWithImage() *poiImageScenario {
	expectedPOI := &models.POI{
		ID:              "poi-789",
		MapID:           s.mapID,
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       s.userID,
		MaxParticipants: 15,
		ImageURL:        "https://example.com/uploads/poi-789.jpg",
		CreatedAt:       time.Now(),
	}
	
	s.mockPOIService.On("CreatePOIWithImage", mock.Anything, s.mapID, "Coffee Shop", "Great place to meet", 
		models.LatLng{Lat: 40.7128, Lng: -74.0060}, s.userID, 15, mock.AnythingOfType("*multipart.FileHeader")).Return(expectedPOI, nil)
	return s
}

func (s *poiImageScenario) expectPOICreationWithoutImage() *poiImageScenario {
	expectedPOI := &models.POI{
		ID:              "poi-789",
		MapID:           s.mapID,
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       s.userID,
		MaxParticipants: 15,
		ImageURL:        "", // No image
		CreatedAt:       time.Now(),
	}
	
	s.mockPOIService.On("CreatePOI", mock.Anything, s.mapID, "Coffee Shop", "Great place to meet", 
		models.LatLng{Lat: 40.7128, Lng: -74.0060}, s.userID, 15).Return(expectedPOI, nil)
	return s
}

func (s *poiImageScenario) createPOIWithMultipartForm(includeImage bool) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	
	// Add form fields
	writer.WriteField("mapId", s.mapID)
	writer.WriteField("name", "Coffee Shop")
	writer.WriteField("description", "Great place to meet")
	writer.WriteField("position.lat", "40.7128")
	writer.WriteField("position.lng", "-74.0060")
	writer.WriteField("createdBy", s.userID)
	writer.WriteField("maxParticipants", "15")
	
	// Add image file if requested
	if includeImage {
		part, err := writer.CreateFormFile("image", "coffee-shop.jpg")
		assert.NoError(s.t, err)
		
		// Write fake image data
		_, err = io.WriteString(part, "fake-image-data")
		assert.NoError(s.t, err)
	}
	
	writer.Close()
	
	req := httptest.NewRequest(http.MethodPost, "/api/pois", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	
	s.router.ServeHTTP(recorder, req)
	return recorder
}

func (s *poiImageScenario) cleanup() {
	s.mockPOIService.AssertExpectations(s.t)
	s.mockRateLimiter.AssertExpectations(s.t)
}

func TestCreatePOI_WithImage_Success(t *testing.T) {
	scenario := newPOIImageScenario(t)
	defer scenario.cleanup()

	scenario.expectRateLimitSuccess().
		expectPOICreationWithImage()

	recorder := scenario.createPOIWithMultipartForm(true)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	
	var response CreatePOIResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Coffee Shop", response.Name)
	assert.Equal(t, "https://example.com/uploads/poi-789.jpg", response.ImageURL)
}

func TestCreatePOI_WithoutImage_Success(t *testing.T) {
	scenario := newPOIImageScenario(t)
	defer scenario.cleanup()

	scenario.expectRateLimitSuccess().
		expectPOICreationWithoutImage()

	recorder := scenario.createPOIWithMultipartForm(false)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	
	var response CreatePOIResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Coffee Shop", response.Name)
	assert.Empty(t, response.ImageURL)
}

func TestCreatePOI_JSONRequest_StillWorks(t *testing.T) {
	// Test that existing JSON-based POI creation still works
	scenario := newPOIImageScenario(t)
	defer scenario.cleanup()

	scenario.expectRateLimitSuccess().
		expectPOICreationWithoutImage()

	// Create JSON request (existing functionality)
	request := CreatePOIRequest{
		MapID:           scenario.mapID,
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       scenario.userID,
		MaxParticipants: 15,
	}
	
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	scenario.router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	
	var response CreatePOIResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Coffee Shop", response.Name)
	assert.Empty(t, response.ImageURL)
}