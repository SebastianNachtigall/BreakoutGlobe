package services

import (
	"context"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(session *models.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(id string) (*models.Session, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByUserAndMap(userID, mapID string) (*models.Session, error) {
	args := m.Called(userID, mapID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(session *models.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockSessionRepository) UpdateAvatarPosition(sessionID string, position models.LatLng) error {
	args := m.Called(sessionID, position)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSessionRepository) GetActiveByMap(mapID string) ([]*models.Session, error) {
	args := m.Called(mapID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) ExpireOldSessions(timeout time.Duration) error {
	args := m.Called(timeout)
	return args.Error(0)
}

// MockSessionPresence is a mock implementation of SessionPresence
type MockSessionPresence struct {
	mock.Mock
}

func (m *MockSessionPresence) SetSessionPresence(ctx context.Context, sessionID string, data redis.SessionPresenceData, ttl time.Duration) error {
	args := m.Called(ctx, sessionID, data, ttl)
	return args.Error(0)
}

func (m *MockSessionPresence) GetSessionPresence(ctx context.Context, sessionID string) (*redis.SessionPresenceData, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*redis.SessionPresenceData), args.Error(1)
}

func (m *MockSessionPresence) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	args := m.Called(ctx, sessionID, position)
	return args.Error(0)
}

func (m *MockSessionPresence) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionPresence) SessionHeartbeat(ctx context.Context, sessionID string, ttl time.Duration) error {
	args := m.Called(ctx, sessionID, ttl)
	return args.Error(0)
}

func (m *MockSessionPresence) RemoveSessionPresence(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionPresence) GetActiveSessionsForMap(ctx context.Context, mapID string) ([]*redis.SessionPresenceData, error) {
	args := m.Called(ctx, mapID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*redis.SessionPresenceData), args.Error(1)
}

func (m *MockSessionPresence) SetCurrentPOI(ctx context.Context, sessionID, poiID string) error {
	args := m.Called(ctx, sessionID, poiID)
	return args.Error(0)
}

func (m *MockSessionPresence) CleanupExpiredSessions(ctx context.Context, maxAge time.Duration) (int, error) {
	args := m.Called(ctx, maxAge)
	return args.Int(0), args.Error(1)
}

// MockPubSub is a mock implementation of PubSub
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) PublishAvatarMovement(ctx context.Context, event redis.AvatarMovementEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOICreated(ctx context.Context, event redis.POICreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOIUpdated(ctx context.Context, event redis.POIUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOIJoined(ctx context.Context, event redis.POIJoinedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOILeft(ctx context.Context, event redis.POILeftEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// SessionServiceTestSuite tests the SessionService
type SessionServiceTestSuite struct {
	suite.Suite
	service         *SessionService
	mockRepo        *MockSessionRepository
	mockPresence    *MockSessionPresence
	mockPubSub      *MockPubSub
}

func (suite *SessionServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockSessionRepository)
	suite.mockPresence = new(MockSessionPresence)
	suite.mockPubSub = new(MockPubSub)
	
	suite.service = NewSessionService(suite.mockRepo, suite.mockPresence, suite.mockPubSub)
}

func (suite *SessionServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockPresence.AssertExpectations(suite.T())
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestCreateSession() {
	ctx := context.Background()
	userID := "user-123"
	mapID := "map-456"
	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}

	// Mock repository to return no existing session
	suite.mockRepo.On("GetByUserAndMap", userID, mapID).Return(nil, gorm.ErrRecordNotFound)
	
	// Mock repository create
	suite.mockRepo.On("Create", mock.MatchedBy(func(session *models.Session) bool {
		return session.UserID == userID && 
			   session.MapID == mapID && 
			   session.AvatarPos.Lat == position.Lat &&
			   session.AvatarPos.Lng == position.Lng &&
			   session.IsActive &&
			   session.ID != ""
	})).Return(nil)

	// Mock presence creation
	suite.mockPresence.On("SetSessionPresence", ctx, mock.AnythingOfType("string"), mock.MatchedBy(func(data redis.SessionPresenceData) bool {
		return data.UserID == userID && 
			   data.MapID == mapID &&
			   data.AvatarPosition.Lat == position.Lat &&
			   data.AvatarPosition.Lng == position.Lng
	}), 30*time.Minute).Return(nil)

	// Execute
	session, err := suite.service.CreateSession(ctx, userID, mapID, position)

	// Assert
	suite.NoError(err)
	suite.NotNil(session)
	suite.Equal(userID, session.UserID)
	suite.Equal(mapID, session.MapID)
	suite.Equal(position.Lat, session.AvatarPos.Lat)
	suite.Equal(position.Lng, session.AvatarPos.Lng)
	suite.True(session.IsActive)
	suite.NotEmpty(session.ID)
}

func (suite *SessionServiceTestSuite) TestCreateSession_UserAlreadyInMap() {
	ctx := context.Background()
	userID := "user-123"
	mapID := "map-456"
	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}

	existingSession := &models.Session{
		ID:         "existing-session-id",
		UserID:     userID,
		MapID:      mapID,
		AvatarPos:  models.LatLng{Lat: 41.0, Lng: -75.0},
		IsActive:   true,
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		LastActive: time.Now().Add(-10 * time.Minute),
	}

	// Mock repository to return existing session
	suite.mockRepo.On("GetByUserAndMap", userID, mapID).Return(existingSession, nil)

	// Execute
	session, err := suite.service.CreateSession(ctx, userID, mapID, position)

	// Assert
	suite.Error(err)
	suite.Nil(session)
	suite.Contains(err.Error(), "user already has an active session")
}

func (suite *SessionServiceTestSuite) TestGetSession() {
	ctx := context.Background()
	sessionID := "session-123"

	expectedSession := &models.Session{
		ID:         sessionID,
		UserID:     "user-456",
		MapID:      "map-789",
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		IsActive:   true,
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		LastActive: time.Now().Add(-5 * time.Minute),
	}

	// Mock repository
	suite.mockRepo.On("GetByID", sessionID).Return(expectedSession, nil)

	// Execute
	session, err := suite.service.GetSession(ctx, sessionID)

	// Assert
	suite.NoError(err)
	suite.NotNil(session)
	suite.Equal(expectedSession.ID, session.ID)
	suite.Equal(expectedSession.UserID, session.UserID)
	suite.Equal(expectedSession.MapID, session.MapID)
}

func (suite *SessionServiceTestSuite) TestUpdateAvatarPosition() {
	ctx := context.Background()
	sessionID := "session-123"
	newPosition := models.LatLng{Lat: 41.0, Lng: -75.0}

	existingSession := &models.Session{
		ID:         sessionID,
		UserID:     "user-456",
		MapID:      "map-789",
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		IsActive:   true,
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		LastActive: time.Now().Add(-5 * time.Minute),
	}

	// Mock repository calls
	suite.mockRepo.On("GetByID", sessionID).Return(existingSession, nil)
	suite.mockRepo.On("UpdateAvatarPosition", sessionID, newPosition).Return(nil)

	// Mock presence update
	suite.mockPresence.On("UpdateAvatarPosition", ctx, sessionID, newPosition).Return(nil)

	// Mock pub/sub event
	suite.mockPubSub.On("PublishAvatarMovement", ctx, mock.MatchedBy(func(event redis.AvatarMovementEvent) bool {
		return event.SessionID == sessionID &&
			   event.UserID == existingSession.UserID &&
			   event.MapID == existingSession.MapID &&
			   event.Position.Lat == newPosition.Lat &&
			   event.Position.Lng == newPosition.Lng
	})).Return(nil)

	// Execute
	err := suite.service.UpdateAvatarPosition(ctx, sessionID, newPosition)

	// Assert
	suite.NoError(err)
}

func (suite *SessionServiceTestSuite) TestSessionHeartbeat() {
	ctx := context.Background()
	sessionID := "session-123"

	existingSession := &models.Session{
		ID:         sessionID,
		UserID:     "user-456",
		MapID:      "map-789",
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		IsActive:   true,
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		LastActive: time.Now().Add(-5 * time.Minute),
	}

	// Mock repository calls
	suite.mockRepo.On("GetByID", sessionID).Return(existingSession, nil)
	suite.mockRepo.On("Update", mock.AnythingOfType("*models.Session")).Return(nil)

	// Mock presence heartbeat
	suite.mockPresence.On("SessionHeartbeat", ctx, sessionID, 30*time.Minute).Return(nil)

	// Execute
	err := suite.service.SessionHeartbeat(ctx, sessionID)

	// Assert
	suite.NoError(err)
}

func (suite *SessionServiceTestSuite) TestEndSession() {
	ctx := context.Background()
	sessionID := "session-123"

	existingSession := &models.Session{
		ID:         sessionID,
		UserID:     "user-456",
		MapID:      "map-789",
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		IsActive:   true,
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		LastActive: time.Now().Add(-5 * time.Minute),
	}

	// Mock repository calls
	suite.mockRepo.On("GetByID", sessionID).Return(existingSession, nil)
	suite.mockRepo.On("Update", mock.AnythingOfType("*models.Session")).Return(nil)

	// Mock presence removal
	suite.mockPresence.On("RemoveSessionPresence", ctx, sessionID).Return(nil)

	// Execute
	err := suite.service.EndSession(ctx, sessionID)

	// Assert
	suite.NoError(err)
}

func (suite *SessionServiceTestSuite) TestGetActiveSessionsForMap() {
	ctx := context.Background()
	mapID := "map-123"

	expectedSessions := []*models.Session{
		{
			ID:         "session-1",
			UserID:     "user-1",
			MapID:      mapID,
			AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
			IsActive:   true,
		},
		{
			ID:         "session-2",
			UserID:     "user-2",
			MapID:      mapID,
			AvatarPos:  models.LatLng{Lat: 41.0, Lng: -75.0},
			IsActive:   true,
		},
	}

	// Mock repository
	suite.mockRepo.On("GetActiveByMap", mapID).Return(expectedSessions, nil)

	// Execute
	sessions, err := suite.service.GetActiveSessionsForMap(ctx, mapID)

	// Assert
	suite.NoError(err)
	suite.Len(sessions, 2)
	suite.Equal(expectedSessions[0].ID, sessions[0].ID)
	suite.Equal(expectedSessions[1].ID, sessions[1].ID)
}

func (suite *SessionServiceTestSuite) TestCleanupExpiredSessions() {
	ctx := context.Background()

	// Mock repository cleanup
	suite.mockRepo.On("ExpireOldSessions", 30*time.Minute).Return(nil)

	// Mock presence cleanup
	suite.mockPresence.On("CleanupExpiredSessions", ctx, 30*time.Minute).Return(5, nil)

	// Execute
	cleanedCount, err := suite.service.CleanupExpiredSessions(ctx)

	// Assert
	suite.NoError(err)
	suite.Equal(5, cleanedCount)
}

func TestSessionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SessionServiceTestSuite))
}