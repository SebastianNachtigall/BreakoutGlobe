package repository

import (
	"context"
	"testing"
	"time"

	"breakoutglobe/internal/database"
	"breakoutglobe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	repo SessionRepository
}

func (suite *SessionRepositoryTestSuite) SetupSuite() {
	// Set up test database
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := database.Initialize(testURL)
	require.NoError(suite.T(), err)
	
	suite.db = db
	suite.repo = NewSessionRepository(db)
}

func (suite *SessionRepositoryTestSuite) TearDownSuite() {
	database.CloseConnection(suite.db)
}

func (suite *SessionRepositoryTestSuite) SetupTest() {
	// Clean up sessions table before each test
	suite.db.Exec("DELETE FROM sessions")
	suite.db.Exec("DELETE FROM maps")
}

func (suite *SessionRepositoryTestSuite) TestCreate() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create test session
	session, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	// Test Create
	created, err := suite.repo.Create(ctx, session)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), created)
	assert.Equal(suite.T(), session.UserID, created.UserID)
	assert.Equal(suite.T(), session.MapID, created.MapID)
	assert.Equal(suite.T(), session.AvatarPos, created.AvatarPos)
	assert.True(suite.T(), created.IsActive)
	assert.NotEmpty(suite.T(), created.ID)
}

func (suite *SessionRepositoryTestSuite) TestCreate_DuplicateUserInMap() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create first session
	session1, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	_, err = suite.repo.Create(ctx, session1)
	require.NoError(suite.T(), err)
	
	// Try to create second session for same user in same map
	session2, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 41.0, Lng: -75.0})
	require.NoError(suite.T(), err)
	
	_, err = suite.repo.Create(ctx, session2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user already has an active session in this map")
}

func (suite *SessionRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create and save session
	session, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	created, err := suite.repo.Create(ctx, session)
	require.NoError(suite.T(), err)
	
	// Test GetByID
	found, err := suite.repo.GetByID(ctx, created.ID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), created.ID, found.ID)
	assert.Equal(suite.T(), created.UserID, found.UserID)
	assert.Equal(suite.T(), created.MapID, found.MapID)
	assert.Equal(suite.T(), created.AvatarPos, found.AvatarPos)
}

func (suite *SessionRepositoryTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()
	
	found, err := suite.repo.GetByID(ctx, "non-existent-id")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *SessionRepositoryTestSuite) TestGetByUserAndMap() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create and save session
	session, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	created, err := suite.repo.Create(ctx, session)
	require.NoError(suite.T(), err)
	
	// Test GetByUserAndMap
	found, err := suite.repo.GetByUserAndMap(ctx, "user-123", "map-123")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), created.ID, found.ID)
	assert.Equal(suite.T(), "user-123", found.UserID)
	assert.Equal(suite.T(), "map-123", found.MapID)
}

func (suite *SessionRepositoryTestSuite) TestGetByUserAndMap_NotFound() {
	ctx := context.Background()
	
	found, err := suite.repo.GetByUserAndMap(ctx, "non-existent-user", "non-existent-map")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *SessionRepositoryTestSuite) TestGetActiveByMap() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create multiple sessions
	session1, err := models.NewSession("user-1", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	session2, err := models.NewSession("user-2", "map-123", models.LatLng{Lat: 41.0, Lng: -75.0})
	require.NoError(suite.T(), err)
	
	session3, err := models.NewSession("user-3", "map-123", models.LatLng{Lat: 42.0, Lng: -76.0})
	require.NoError(suite.T(), err)
	
	_, err = suite.repo.Create(ctx, session1)
	require.NoError(suite.T(), err)
	
	_, err = suite.repo.Create(ctx, session2)
	require.NoError(suite.T(), err)
	
	created3, err := suite.repo.Create(ctx, session3)
	require.NoError(suite.T(), err)
	
	// Deactivate session3 after creation
	created3.IsActive = false
	_, err = suite.repo.Update(ctx, created3)
	require.NoError(suite.T(), err)
	
	// Test GetActiveByMap
	sessions, err := suite.repo.GetActiveByMap(ctx, "map-123")
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2) // Only active sessions
	
	userIDs := make([]string, len(sessions))
	for i, s := range sessions {
		userIDs[i] = s.UserID
		assert.True(suite.T(), s.IsActive)
		assert.Equal(suite.T(), "map-123", s.MapID)
	}
	assert.Contains(suite.T(), userIDs, "user-1")
	assert.Contains(suite.T(), userIDs, "user-2")
	assert.NotContains(suite.T(), userIDs, "user-3")
}

func (suite *SessionRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create and save session
	session, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	created, err := suite.repo.Create(ctx, session)
	require.NoError(suite.T(), err)
	
	// Update session
	newPos := models.LatLng{Lat: 41.0, Lng: -75.0}
	err = created.UpdateAvatarPosition(newPos)
	require.NoError(suite.T(), err)
	
	// Test Update
	updated, err := suite.repo.Update(ctx, created)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updated)
	assert.Equal(suite.T(), created.ID, updated.ID)
	assert.Equal(suite.T(), newPos, updated.AvatarPos)
	assert.True(suite.T(), updated.LastActive.After(created.CreatedAt))
}

func (suite *SessionRepositoryTestSuite) TestUpdateAvatarPosition() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create and save session
	session, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	created, err := suite.repo.Create(ctx, session)
	require.NoError(suite.T(), err)
	
	// Test UpdateAvatarPosition
	newPos := models.LatLng{Lat: 41.0, Lng: -75.0}
	updated, err := suite.repo.UpdateAvatarPosition(ctx, created.ID, newPos)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updated)
	assert.Equal(suite.T(), created.ID, updated.ID)
	assert.Equal(suite.T(), newPos, updated.AvatarPos)
	assert.True(suite.T(), updated.LastActive.After(created.LastActive))
}

func (suite *SessionRepositoryTestSuite) TestDelete() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create and save session
	session, err := models.NewSession("user-123", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	
	created, err := suite.repo.Create(ctx, session)
	require.NoError(suite.T(), err)
	
	// Test Delete
	err = suite.repo.Delete(ctx, created.ID)
	
	assert.NoError(suite.T(), err)
	
	// Verify session is deleted
	_, err = suite.repo.GetByID(ctx, created.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "session not found")
}

func (suite *SessionRepositoryTestSuite) TestExpireOldSessions() {
	ctx := context.Background()
	
	// Create a test map first
	testMap := &models.Map{
		ID:        "map-123",
		Name:      "Test Map",
		CreatedBy: "facilitator-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := suite.db.Create(testMap).Error
	require.NoError(suite.T(), err)
	
	// Create sessions with different last active times
	oldSession, err := models.NewSession("user-old", "map-123", models.LatLng{Lat: 40.7128, Lng: -74.0060})
	require.NoError(suite.T(), err)
	oldSession.LastActive = time.Now().Add(-45 * time.Minute) // 45 minutes ago
	
	recentSession, err := models.NewSession("user-recent", "map-123", models.LatLng{Lat: 41.0, Lng: -75.0})
	require.NoError(suite.T(), err)
	recentSession.LastActive = time.Now().Add(-15 * time.Minute) // 15 minutes ago
	
	_, err = suite.repo.Create(ctx, oldSession)
	require.NoError(suite.T(), err)
	
	_, err = suite.repo.Create(ctx, recentSession)
	require.NoError(suite.T(), err)
	
	// Test ExpireOldSessions with 30 minute timeout
	count, err := suite.repo.ExpireOldSessions(ctx, 30*time.Minute)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count) // Only old session should be expired
	
	// Verify old session is inactive
	oldFound, err := suite.repo.GetByID(ctx, oldSession.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), oldFound.IsActive)
	
	// Verify recent session is still active
	recentFound, err := suite.repo.GetByID(ctx, recentSession.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), recentFound.IsActive)
}

func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}