package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MockPOIService is a mock implementation of POIServiceInterface
type MockPOIService struct {
	mock.Mock
}

func (m *MockPOIService) CreatePOI(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int) (*models.POI, error) {
	args := m.Called(ctx, mapID, name, description, position, createdBy, maxParticipants)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.POI), args.Error(1)
}

func (m *MockPOIService) CreatePOIWithImage(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int, imageFile *multipart.FileHeader) (*models.POI, error) {
	args := m.Called(ctx, mapID, name, description, position, createdBy, maxParticipants, imageFile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.POI), args.Error(1)
}

func (m *MockPOIService) GetPOI(ctx context.Context, poiID string) (*models.POI, error) {
	args := m.Called(ctx, poiID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.POI), args.Error(1)
}

func (m *MockPOIService) GetPOIsForMap(ctx context.Context, mapID string) ([]*models.POI, error) {
	args := m.Called(ctx, mapID)
	return args.Get(0).([]*models.POI), args.Error(1)
}

func (m *MockPOIService) GetPOIsInBounds(ctx context.Context, mapID string, bounds services.POIBounds) ([]*models.POI, error) {
	args := m.Called(ctx, mapID, bounds)
	return args.Get(0).([]*models.POI), args.Error(1)
}

func (m *MockPOIService) UpdatePOI(ctx context.Context, poiID string, updateData services.POIUpdateData) (*models.POI, error) {
	args := m.Called(ctx, poiID, updateData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.POI), args.Error(1)
}

func (m *MockPOIService) DeletePOI(ctx context.Context, poiID string) error {
	args := m.Called(ctx, poiID)
	return args.Error(0)
}

func (m *MockPOIService) JoinPOI(ctx context.Context, poiID, userID string) error {
	args := m.Called(ctx, poiID, userID)
	return args.Error(0)
}

func (m *MockPOIService) LeavePOI(ctx context.Context, poiID, userID string) error {
	args := m.Called(ctx, poiID, userID)
	return args.Error(0)
}

func (m *MockPOIService) GetPOIParticipants(ctx context.Context, poiID string) ([]string, error) {
	args := m.Called(ctx, poiID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPOIService) GetPOIParticipantCount(ctx context.Context, poiID string) (int, error) {
	args := m.Called(ctx, poiID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockPOIService) GetUserPOIs(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPOIService) ValidatePOI(ctx context.Context, poiID string) (*models.POI, error) {
	args := m.Called(ctx, poiID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.POI), args.Error(1)
}

// POIHandlerTestSuite contains the test suite for POIHandler
type POIHandlerTestSuite struct {
	suite.Suite
	mockPOIService  *MockPOIService
	mockRateLimiter *MockRateLimiter
	handler         *POIHandler
	router          *gin.Engine
}

func (suite *POIHandlerTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	suite.mockPOIService = new(MockPOIService)
	suite.mockRateLimiter = new(MockRateLimiter)
	suite.handler = NewPOIHandler(suite.mockPOIService, suite.mockRateLimiter)
	
	// Setup router
	suite.router = gin.New()
	suite.handler.RegisterRoutes(suite.router)
}

func (suite *POIHandlerTestSuite) TearDownTest() {
	suite.mockPOIService.AssertExpectations(suite.T())
	suite.mockRateLimiter.AssertExpectations(suite.T())
}

func (suite *POIHandlerTestSuite) TestGetPOIs() {
	mapID := "map-123"
	expectedPOIs := []*models.POI{
		{
			ID:          "poi-1",
			MapID:       mapID,
			Name:        "Coffee Shop",
			Description: "Great place to meet",
			Position:    models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:   "user-1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "poi-2",
			MapID:       mapID,
			Name:        "Park Bench",
			Description: "Quiet spot in the park",
			Position:    models.LatLng{Lat: 40.7589, Lng: -73.9851},
			CreatedBy:   "user-2",
			CreatedAt:   time.Now(),
		},
	}
	
	// Mock expectations
	suite.mockPOIService.On("GetPOIsForMap", mock.AnythingOfType("*gin.Context"), mapID).Return(expectedPOIs, nil)
	
	// Mock participant information for each POI
	for _, poi := range expectedPOIs {
		suite.mockPOIService.On("GetPOIParticipantCount", mock.AnythingOfType("*gin.Context"), poi.ID).Return(2, nil)
		suite.mockPOIService.On("GetPOIParticipants", mock.AnythingOfType("*gin.Context"), poi.ID).Return([]string{"session-1", "session-2"}, nil)
	}
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/pois?mapId="+mapID, nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response GetPOIsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(len(expectedPOIs), len(response.POIs))
	suite.Equal(expectedPOIs[0].ID, response.POIs[0].ID)
	suite.Equal(expectedPOIs[0].Name, response.POIs[0].Name)
	suite.Equal(expectedPOIs[1].ID, response.POIs[1].ID)
	suite.Equal(expectedPOIs[1].Name, response.POIs[1].Name)
	
	// Verify participant information is included
	suite.Equal(2, response.POIs[0].ParticipantCount)
	suite.Equal(2, len(response.POIs[0].Participants))
	suite.Equal("session-1", response.POIs[0].Participants[0].ID)
	suite.Equal("User-session-1", response.POIs[0].Participants[0].Name)
	suite.Equal("session-2", response.POIs[0].Participants[1].ID)
	suite.Equal("User-session-2", response.POIs[0].Participants[1].Name)
}

func (suite *POIHandlerTestSuite) TestGetPOIsWithBounds() {
	mapID := "map-123"
	bounds := services.POIBounds{
		MinLat: 40.7000,
		MaxLat: 40.8000,
		MinLng: -74.1000,
		MaxLng: -73.9000,
	}
	
	expectedPOIs := []*models.POI{
		{
			ID:          "poi-1",
			MapID:       mapID,
			Name:        "Coffee Shop",
			Description: "Great place to meet",
			Position:    models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:   "user-1",
			CreatedAt:   time.Now(),
		},
	}
	
	// Mock expectations
	suite.mockPOIService.On("GetPOIsInBounds", mock.AnythingOfType("*gin.Context"), mapID, bounds).Return(expectedPOIs, nil)
	
	// Mock participant information for each POI
	for _, poi := range expectedPOIs {
		suite.mockPOIService.On("GetPOIParticipantCount", mock.AnythingOfType("*gin.Context"), poi.ID).Return(1, nil)
		suite.mockPOIService.On("GetPOIParticipants", mock.AnythingOfType("*gin.Context"), poi.ID).Return([]string{"session-1"}, nil)
	}
	
	// Create request with bounds query parameters
	req := httptest.NewRequest(http.MethodGet, "/api/pois?mapId="+mapID+"&minLat=40.7000&maxLat=40.8000&minLng=-74.1000&maxLng=-73.9000", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response GetPOIsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(len(expectedPOIs), len(response.POIs))
	suite.Equal(expectedPOIs[0].ID, response.POIs[0].ID)
}

func (suite *POIHandlerTestSuite) TestGetPOIs_MissingMapID() {
	// Create request without mapId parameter
	req := httptest.NewRequest(http.MethodGet, "/api/pois", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INVALID_REQUEST", response.Code)
	suite.Contains(response.Message, "mapId is required")
}

func (suite *POIHandlerTestSuite) TestCreatePOI() {
	reqBody := CreatePOIRequest{
		MapID:           "map-123",
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 15,
	}
	
	expectedPOI := &models.POI{
		ID:              "poi-789",
		MapID:           reqBody.MapID,
		Name:            reqBody.Name,
		Description:     reqBody.Description,
		Position:        reqBody.Position,
		CreatedBy:       reqBody.CreatedBy,
		MaxParticipants: reqBody.MaxParticipants,
		CreatedAt:       time.Now(),
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.CreatedBy, services.ActionCreatePOI).Return(nil)
	suite.mockPOIService.On("CreatePOI", mock.AnythingOfType("*gin.Context"), reqBody.MapID, reqBody.Name, reqBody.Description, reqBody.Position, reqBody.CreatedBy, reqBody.MaxParticipants).Return(expectedPOI, nil)
	suite.mockRateLimiter.On("GetRateLimitHeaders", mock.AnythingOfType("*gin.Context"), reqBody.CreatedBy, services.ActionCreatePOI).Return(map[string]string{
		"X-RateLimit-Limit":     "5",
		"X-RateLimit-Remaining": "4",
	}, nil)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusCreated, w.Code)
	
	var response CreatePOIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(expectedPOI.ID, response.ID)
	suite.Equal(expectedPOI.Name, response.Name)
	suite.Equal(expectedPOI.Position.Lat, response.Position.Lat)
	suite.Equal(expectedPOI.Position.Lng, response.Position.Lng)
	
	// Check rate limit headers
	suite.Equal("5", w.Header().Get("X-RateLimit-Limit"))
	suite.Equal("4", w.Header().Get("X-RateLimit-Remaining"))
}

func (suite *POIHandlerTestSuite) TestCreatePOI_RateLimited() {
	reqBody := CreatePOIRequest{
		MapID:           "map-123",
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 15,
	}
	
	// Mock rate limit exceeded
	rateLimitErr := &services.RateLimitError{
		UserID:     reqBody.CreatedBy,
		Action:     services.ActionCreatePOI,
		Limit:      5,
		Window:     time.Hour,
		RetryAfter: time.Hour,
	}
	
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.CreatedBy, services.ActionCreatePOI).Return(rateLimitErr)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusTooManyRequests, w.Code)
	suite.Equal("3600", w.Header().Get("Retry-After"))
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("RATE_LIMIT_EXCEEDED", response.Code)
}

func (suite *POIHandlerTestSuite) TestCreatePOI_InvalidJSON() {
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INVALID_REQUEST", response.Code)
}

func (suite *POIHandlerTestSuite) TestCreatePOI_ValidationError() {
	reqBody := CreatePOIRequest{
		MapID:           "", // Invalid: empty map ID
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 15,
	}
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INVALID_REQUEST", response.Code)
}

func (suite *POIHandlerTestSuite) TestCreatePOI_DuplicateLocation() {
	reqBody := CreatePOIRequest{
		MapID:           "map-123",
		Name:            "Coffee Shop",
		Description:     "Great place to meet",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 15,
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.CreatedBy, services.ActionCreatePOI).Return(nil)
	suite.mockPOIService.On("CreatePOI", mock.AnythingOfType("*gin.Context"), reqBody.MapID, reqBody.Name, reqBody.Description, reqBody.Position, reqBody.CreatedBy, reqBody.MaxParticipants).Return((*models.POI)(nil), errors.New("duplicate POI location"))
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusConflict, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("DUPLICATE_LOCATION", response.Code)
}

func (suite *POIHandlerTestSuite) TestJoinPOI() {
	poiID := "poi-123"
	reqBody := JoinPOIRequest{
		UserID: "user-456",
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionJoinPOI).Return(nil)
	suite.mockPOIService.On("JoinPOI", mock.AnythingOfType("*gin.Context"), poiID, reqBody.UserID).Return(nil)
	suite.mockRateLimiter.On("GetRateLimitHeaders", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionJoinPOI).Return(map[string]string{
		"X-RateLimit-Limit":     "20",
		"X-RateLimit-Remaining": "19",
	}, nil)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois/"+poiID+"/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response JoinPOIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.Equal(poiID, response.POIID)
	suite.Equal(reqBody.UserID, response.UserID)
	
	// Check rate limit headers
	suite.Equal("20", w.Header().Get("X-RateLimit-Limit"))
	suite.Equal("19", w.Header().Get("X-RateLimit-Remaining"))
}

func (suite *POIHandlerTestSuite) TestJoinPOI_CapacityExceeded() {
	poiID := "poi-123"
	reqBody := JoinPOIRequest{
		UserID: "user-456",
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionJoinPOI).Return(nil)
	suite.mockPOIService.On("JoinPOI", mock.AnythingOfType("*gin.Context"), poiID, reqBody.UserID).Return(errors.New("POI capacity exceeded"))
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois/"+poiID+"/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusConflict, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("CAPACITY_EXCEEDED", response.Code)
}

func (suite *POIHandlerTestSuite) TestJoinPOI_POINotFound() {
	poiID := "non-existent-poi"
	reqBody := JoinPOIRequest{
		UserID: "user-456",
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionJoinPOI).Return(nil)
	suite.mockPOIService.On("JoinPOI", mock.AnythingOfType("*gin.Context"), poiID, reqBody.UserID).Return(gorm.ErrRecordNotFound)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois/"+poiID+"/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("POI_NOT_FOUND", response.Code)
}

func (suite *POIHandlerTestSuite) TestLeavePOI() {
	poiID := "poi-123"
	reqBody := LeavePOIRequest{
		UserID: "user-456",
	}
	
	// Mock expectations
	suite.mockPOIService.On("LeavePOI", mock.AnythingOfType("*gin.Context"), poiID, reqBody.UserID).Return(nil)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois/"+poiID+"/leave", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response LeavePOIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	suite.Equal(poiID, response.POIID)
	suite.Equal(reqBody.UserID, response.UserID)
}

func (suite *POIHandlerTestSuite) TestLeavePOI_POINotFound() {
	poiID := "non-existent-poi"
	reqBody := LeavePOIRequest{
		UserID: "user-456",
	}
	
	// Mock expectations
	suite.mockPOIService.On("LeavePOI", mock.AnythingOfType("*gin.Context"), poiID, reqBody.UserID).Return(gorm.ErrRecordNotFound)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/pois/"+poiID+"/leave", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("POI_NOT_FOUND", response.Code)
}

func TestPOIHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(POIHandlerTestSuite))
}