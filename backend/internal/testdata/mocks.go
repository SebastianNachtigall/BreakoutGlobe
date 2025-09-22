package testdata

import (
	"context"
	"fmt"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"github.com/stretchr/testify/mock"
)

// MockSetup provides a centralized way to set up all mocks with fluent API
type MockSetup struct {
	POIService     *MockPOIServiceBuilder
	SessionService *MockSessionServiceBuilder
	RateLimiter    *MockRateLimiterBuilder
}

// NewMockSetup creates a new mock setup with all services initialized
func NewMockSetup() *MockSetup {
	return &MockSetup{
		POIService:     NewMockPOIServiceBuilder(),
		SessionService: NewMockSessionServiceBuilder(),
		RateLimiter:    NewMockRateLimiterBuilder(),
	}
}

// AssertExpectations verifies all mock expectations
func (m *MockSetup) AssertExpectations(t mock.TestingT) {
	m.POIService.Mock().AssertExpectations(t)
	m.SessionService.Mock().AssertExpectations(t)
	m.RateLimiter.Mock().AssertExpectations(t)
}

// MockPOIServiceBuilder provides a fluent API for setting up POI service mocks
type MockPOIServiceBuilder struct {
	mock *MockPOIService
}

// NewMockPOIServiceBuilder creates a new POI service mock builder
func NewMockPOIServiceBuilder() *MockPOIServiceBuilder {
	return &MockPOIServiceBuilder{
		mock: &MockPOIService{},
	}
}

// Mock returns the underlying mock for direct access if needed
func (b *MockPOIServiceBuilder) Mock() *MockPOIService {
	return b.mock
}

// ExpectCreatePOI starts building an expectation for CreatePOI
func (b *MockPOIServiceBuilder) ExpectCreatePOI() *POICreateExpectation {
	return &POICreateExpectation{
		mock: b.mock,
	}
}

// ExpectGetPOI starts building an expectation for GetPOI
func (b *MockPOIServiceBuilder) ExpectGetPOI(poiID string) *POIGetExpectation {
	return &POIGetExpectation{
		mock:  b.mock,
		poiID: poiID,
	}
}

// POICreateExpectation builds expectations for POI creation
type POICreateExpectation struct {
	mock      *MockPOIService
	mapID     string
	createdBy string
}

// WithMapID sets the expected map ID
func (e *POICreateExpectation) WithMapID(mapID string) *POICreateExpectation {
	e.mapID = mapID
	return e
}

// WithCreatedBy sets the expected creator ID
func (e *POICreateExpectation) WithCreatedBy(createdBy string) *POICreateExpectation {
	e.createdBy = createdBy
	return e
}

// Returns sets up the mock to return the specified POI
func (e *POICreateExpectation) Returns(poi *models.POI) {
	e.mock.On("CreatePOI",
		mock.Anything, // Handle any context type automatically
		e.mapID,
		mock.AnythingOfType("string"), // name
		mock.AnythingOfType("string"), // description
		mock.AnythingOfType("models.LatLng"),
		e.createdBy,
		mock.AnythingOfType("int"), // maxParticipants
	).Return(poi, nil)
}

// ReturnsError sets up the mock to return an error
func (e *POICreateExpectation) ReturnsError(err error) {
	e.mock.On("CreatePOI",
		mock.Anything,
		e.mapID,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("models.LatLng"),
		e.createdBy,
		mock.AnythingOfType("int"),
	).Return(nil, err)
}

// POIGetExpectation builds expectations for POI retrieval
type POIGetExpectation struct {
	mock  *MockPOIService
	poiID string
}

// Returns sets up the mock to return the specified POI
func (e *POIGetExpectation) Returns(poi *models.POI) {
	e.mock.On("GetPOI",
		mock.Anything,
		e.poiID,
	).Return(poi, nil)
}

// ReturnsNotFound sets up the mock to return a not found error
func (e *POIGetExpectation) ReturnsNotFound() {
	e.mock.On("GetPOI",
		mock.Anything,
		e.poiID,
	).Return(nil, fmt.Errorf("POI not found: %s", e.poiID))
}

// ReturnsError sets up the mock to return a custom error
func (e *POIGetExpectation) ReturnsError(err error) {
	e.mock.On("GetPOI",
		mock.Anything,
		e.poiID,
	).Return(nil, err)
}

// MockRateLimiterBuilder provides a fluent API for setting up rate limiter mocks
type MockRateLimiterBuilder struct {
	mock *MockRateLimiter
}

// NewMockRateLimiterBuilder creates a new rate limiter mock builder
func NewMockRateLimiterBuilder() *MockRateLimiterBuilder {
	return &MockRateLimiterBuilder{
		mock: &MockRateLimiter{},
	}
}

// Mock returns the underlying mock for direct access if needed
func (b *MockRateLimiterBuilder) Mock() *MockRateLimiter {
	return b.mock
}

// ExpectCheckRateLimit starts building an expectation for CheckRateLimit
func (b *MockRateLimiterBuilder) ExpectCheckRateLimit() *RateLimitCheckExpectation {
	return &RateLimitCheckExpectation{
		mock: b.mock,
	}
}

// ExpectGetRateLimitHeaders starts building an expectation for GetRateLimitHeaders
func (b *MockRateLimiterBuilder) ExpectGetRateLimitHeaders() *RateLimitHeadersExpectation {
	return &RateLimitHeadersExpectation{
		mock: b.mock,
	}
}

// RateLimitCheckExpectation builds expectations for rate limit checking
type RateLimitCheckExpectation struct {
	mock   *MockRateLimiter
	userID string
	action services.ActionType
}

// WithUserID sets the expected user ID
func (e *RateLimitCheckExpectation) WithUserID(userID string) *RateLimitCheckExpectation {
	e.userID = userID
	return e
}

// WithAction sets the expected action
func (e *RateLimitCheckExpectation) WithAction(action services.ActionType) *RateLimitCheckExpectation {
	e.action = action
	return e
}

// Returns sets up the mock to return success (no error)
func (e *RateLimitCheckExpectation) Returns() {
	e.mock.On("CheckRateLimit",
		mock.Anything, // Handle any context type automatically
		e.userID,
		e.action,
	).Return(nil)
}

// ReturnsRateLimitExceeded sets up the mock to return a rate limit exceeded error
func (e *RateLimitCheckExpectation) ReturnsRateLimitExceeded() {
	rateLimitError := services.NewRateLimitError(
		e.userID,
		e.action,
		services.RateLimit{Requests: 5, Window: time.Minute},
		time.Now().Add(time.Minute),
	)
	e.mock.On("CheckRateLimit",
		mock.Anything,
		e.userID,
		e.action,
	).Return(rateLimitError)
}

// ReturnsError sets up the mock to return a custom error
func (e *RateLimitCheckExpectation) ReturnsError(err error) {
	e.mock.On("CheckRateLimit",
		mock.Anything,
		e.userID,
		e.action,
	).Return(err)
}

// RateLimitHeadersExpectation builds expectations for rate limit headers
type RateLimitHeadersExpectation struct {
	mock   *MockRateLimiter
	userID string
	action services.ActionType
}

// WithUserID sets the expected user ID
func (e *RateLimitHeadersExpectation) WithUserID(userID string) *RateLimitHeadersExpectation {
	e.userID = userID
	return e
}

// WithAction sets the expected action
func (e *RateLimitHeadersExpectation) WithAction(action services.ActionType) *RateLimitHeadersExpectation {
	e.action = action
	return e
}

// Returns sets up the mock to return the specified headers
func (e *RateLimitHeadersExpectation) Returns(headers map[string]string) {
	e.mock.On("GetRateLimitHeaders",
		mock.Anything,
		e.userID,
		e.action,
	).Return(headers, nil)
}

// ReturnsError sets up the mock to return an error
func (e *RateLimitHeadersExpectation) ReturnsError(err error) {
	e.mock.On("GetRateLimitHeaders",
		mock.Anything,
		e.userID,
		e.action,
	).Return(nil, err)
}

// MockSessionServiceBuilder provides a fluent API for setting up session service mocks
type MockSessionServiceBuilder struct {
	mock *MockSessionService
}

// NewMockSessionServiceBuilder creates a new session service mock builder
func NewMockSessionServiceBuilder() *MockSessionServiceBuilder {
	return &MockSessionServiceBuilder{
		mock: &MockSessionService{},
	}
}

// Mock returns the underlying mock for direct access if needed
func (b *MockSessionServiceBuilder) Mock() *MockSessionService {
	return b.mock
}

// ExpectCreateSession starts building an expectation for CreateSession
func (b *MockSessionServiceBuilder) ExpectCreateSession() *SessionCreateExpectation {
	return &SessionCreateExpectation{
		mock: b.mock,
	}
}

// ExpectGetSession starts building an expectation for GetSession
func (b *MockSessionServiceBuilder) ExpectGetSession(sessionID string) *SessionGetExpectation {
	return &SessionGetExpectation{
		mock:      b.mock,
		sessionID: sessionID,
	}
}

// SessionCreateExpectation builds expectations for session creation
type SessionCreateExpectation struct {
	mock   *MockSessionService
	userID string
	mapID  string
}

// WithUserID sets the expected user ID
func (e *SessionCreateExpectation) WithUserID(userID string) *SessionCreateExpectation {
	e.userID = userID
	return e
}

// WithMapID sets the expected map ID
func (e *SessionCreateExpectation) WithMapID(mapID string) *SessionCreateExpectation {
	e.mapID = mapID
	return e
}

// Returns sets up the mock to return the specified session
func (e *SessionCreateExpectation) Returns(session *models.Session) {
	e.mock.On("CreateSession",
		mock.Anything,
		e.userID,
		e.mapID,
		mock.AnythingOfType("models.LatLng"),
	).Return(session, nil)
}

// ReturnsError sets up the mock to return an error
func (e *SessionCreateExpectation) ReturnsError(err error) {
	e.mock.On("CreateSession",
		mock.Anything,
		e.userID,
		e.mapID,
		mock.AnythingOfType("models.LatLng"),
	).Return(nil, err)
}

// SessionGetExpectation builds expectations for session retrieval
type SessionGetExpectation struct {
	mock      *MockSessionService
	sessionID string
}

// Returns sets up the mock to return the specified session
func (e *SessionGetExpectation) Returns(session *models.Session) {
	e.mock.On("GetSession",
		mock.Anything,
		e.sessionID,
	).Return(session, nil)
}

// ReturnsNotFound sets up the mock to return a not found error
func (e *SessionGetExpectation) ReturnsNotFound() {
	e.mock.On("GetSession",
		mock.Anything,
		e.sessionID,
	).Return(nil, fmt.Errorf("session not found"))
}

// ReturnsError sets up the mock to return a custom error
func (e *SessionGetExpectation) ReturnsError(err error) {
	e.mock.On("GetSession",
		mock.Anything,
		e.sessionID,
	).Return(nil, err)
}

// Actual mock implementations (these would normally be in separate files)

// MockPOIService is a mock implementation of POI service
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

// MockRateLimiter is a mock implementation of rate limiter
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) IsAllowed(ctx context.Context, userID string, action services.ActionType) (bool, error) {
	args := m.Called(ctx, userID, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockRateLimiter) GetRemainingRequests(ctx context.Context, userID string, action services.ActionType) (int, error) {
	args := m.Called(ctx, userID, action)
	return args.Int(0), args.Error(1)
}

func (m *MockRateLimiter) GetWindowResetTime(ctx context.Context, userID string, action services.ActionType) (time.Time, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockRateLimiter) SetCustomLimit(userID string, action services.ActionType, limit services.RateLimit) {
	m.Called(userID, action, limit)
}

func (m *MockRateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRateLimiter) GetUserStats(ctx context.Context, userID string) (*services.UserRateLimitStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserRateLimitStats), args.Error(1)
}

func (m *MockRateLimiter) CheckRateLimit(ctx context.Context, userID string, action services.ActionType) error {
	args := m.Called(ctx, userID, action)
	return args.Error(0)
}

func (m *MockRateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action services.ActionType) (map[string]string, error) {
	args := m.Called(ctx, userID, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

// MockSessionService is a mock implementation of session service
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID, mapID string, initialPos models.LatLng) (*models.Session, error) {
	args := m.Called(ctx, userID, mapID, initialPos)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) GetSessionByUserAndMap(ctx context.Context, userID, mapID string) (*models.Session, error) {
	args := m.Called(ctx, userID, mapID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) UpdateAvatarPosition(ctx context.Context, sessionID string, newPos models.LatLng) error {
	args := m.Called(ctx, sessionID, newPos)
	return args.Error(0)
}

func (m *MockSessionService) SessionHeartbeat(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionService) EndSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionService) GetActiveSessions(ctx context.Context, mapID string) ([]*models.Session, error) {
	args := m.Called(ctx, mapID)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionService) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) (int, error) {
	args := m.Called(ctx, timeout)
	return args.Int(0), args.Error(1)
}