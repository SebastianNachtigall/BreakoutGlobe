package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionRepository_Integration demonstrates database integration testing
// These tests use real PostgreSQL database to validate repository behavior

func TestSessionRepository_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Setup test database
	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Create session using builder
	session := testdata.NewDatabaseSession().
		WithUserID("integration-user").
		WithMapID("integration-map").
		WithPosition(40.7128, -74.0060).
		WithActive(true).
		Build()

	// Test repository create
	createdSession, err := repo.Create(ctx, session)
	require.NoError(t, err)
	require.NotNil(t, createdSession)

	// Verify session was persisted correctly
	assert.NotEmpty(t, createdSession.ID)
	assert.WithinDuration(t, time.Now(), createdSession.CreatedAt, time.Second)

	// Verify in database
	var count int64
	err = testDB.DB.Model(&models.Session{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Retrieve and verify data integrity
	var savedSession models.Session
	err = testDB.DB.First(&savedSession).Error
	require.NoError(t, err)
	assert.Equal(t, createdSession.UserID, savedSession.UserID)
	assert.Equal(t, createdSession.MapID, savedSession.MapID)
	assert.Equal(t, createdSession.AvatarPos.Lat, savedSession.AvatarPos.Lat)
	assert.Equal(t, createdSession.AvatarPos.Lng, savedSession.AvatarPos.Lng)
	assert.Equal(t, createdSession.IsActive, savedSession.IsActive)
}

func TestSessionRepository_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data
	session1 := testdata.NewDatabaseSession().WithUserID("user-1").WithMapID("map-1").Build()
	session2 := testdata.NewDatabaseSession().WithUserID("user-2").WithMapID("map-1").Build()
	err := testDB.SeedFixtures(session1, session2)
	require.NoError(t, err)

	// Test successful retrieval
	foundSession, err := repo.GetByID(ctx, session1.ID)
	require.NoError(t, err)
	require.NotNil(t, foundSession)
	assert.Equal(t, session1.ID, foundSession.ID)
	assert.Equal(t, "user-1", foundSession.UserID)

	// Test not found scenario
	notFoundSession, err := repo.GetByID(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, notFoundSession)
}

func TestSessionRepository_GetByUserAndMap_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data for same user across different maps
	session1 := testdata.NewDatabaseSession().WithUserID("user-1").WithMapID("map-1").WithActive(true).Build()
	session2 := testdata.NewDatabaseSession().WithUserID("user-1").WithMapID("map-2").WithActive(true).Build()
	session3 := testdata.NewDatabaseSession().WithUserID("user-2").WithMapID("map-1").WithActive(true).Build()
	err := testDB.SeedFixtures(session1, session2, session3)
	require.NoError(t, err)

	// Test retrieval by user and map ID
	foundSession, err := repo.GetByUserAndMap(ctx, "user-1", "map-1")
	require.NoError(t, err)
	require.NotNil(t, foundSession)
	assert.Equal(t, "user-1", foundSession.UserID)
	assert.Equal(t, "map-1", foundSession.MapID)
	assert.True(t, foundSession.IsActive)

	// Test user with no session in specific map
	noSession, err := repo.GetByUserAndMap(ctx, "user-1", "non-existent-map")
	assert.Error(t, err)
	assert.Nil(t, noSession)
}

func TestSessionRepository_GetActiveByMap_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data with active and inactive sessions
	activeSession1 := testdata.NewDatabaseSession().WithUserID("user-1").WithMapID("map-1").WithActive(true).Build()
	activeSession2 := testdata.NewDatabaseSession().WithUserID("user-2").WithMapID("map-1").WithActive(true).Build()
	inactiveSession := testdata.NewDatabaseSession().WithUserID("user-3").WithMapID("map-1").WithActive(false).Build()
	otherMapSession := testdata.NewDatabaseSession().WithUserID("user-4").WithMapID("map-2").WithActive(true).Build()
	err := testDB.SeedFixtures(activeSession1, activeSession2, inactiveSession, otherMapSession)
	require.NoError(t, err)

	// Test retrieval of active sessions for map-1
	activeSessions, err := repo.GetActiveByMap(ctx, "map-1")
	require.NoError(t, err)
	assert.Len(t, activeSessions, 2)

	// Verify all returned sessions are active and belong to correct map
	for _, session := range activeSessions {
		assert.Equal(t, "map-1", session.MapID)
		assert.True(t, session.IsActive)
	}

	// Test map with no active sessions
	noActiveSessions, err := repo.GetActiveByMap(ctx, "empty-map")
	require.NoError(t, err)
	assert.Empty(t, noActiveSessions)
}

func TestSessionRepository_UpdateAvatarPosition_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed initial session
	originalSession := testdata.NewDatabaseSession().
		WithUserID("user-1").
		WithMapID("map-1").
		WithPosition(40.7128, -74.0060).
		Build()
	err := testDB.SeedFixtures(originalSession)
	require.NoError(t, err)

	// Update avatar position
	newPosition := models.LatLng{Lat: 41.0, Lng: -75.0}
	updatedSession, err := repo.UpdateAvatarPosition(ctx, originalSession.ID, newPosition)
	require.NoError(t, err)
	require.NotNil(t, updatedSession)

	// Verify update in database
	retrievedSession, err := repo.GetByID(ctx, originalSession.ID)
	require.NoError(t, err)
	assert.Equal(t, newPosition.Lat, retrievedSession.AvatarPos.Lat)
	assert.Equal(t, newPosition.Lng, retrievedSession.AvatarPos.Lng)

	// Verify other fields unchanged
	assert.Equal(t, originalSession.UserID, retrievedSession.UserID)
	assert.Equal(t, originalSession.MapID, retrievedSession.MapID)
	assert.Equal(t, originalSession.IsActive, retrievedSession.IsActive)
}

func TestSessionRepository_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed active session
	session := testdata.NewDatabaseSession().
		WithUserID("user-1").
		WithMapID("map-1").
		WithActive(true).
		Build()
	err := testDB.SeedFixtures(session)
	require.NoError(t, err)

	// Update session
	session.IsActive = false
	session.AvatarPos = models.LatLng{Lat: 41.0, Lng: -75.0}
	updatedSession, err := repo.Update(ctx, session)
	require.NoError(t, err)
	require.NotNil(t, updatedSession)

	// Verify session is updated
	retrievedSession, err := repo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	assert.False(t, retrievedSession.IsActive)
	assert.Equal(t, 41.0, retrievedSession.AvatarPos.Lat)
	assert.Equal(t, -75.0, retrievedSession.AvatarPos.Lng)
}

func TestSessionRepository_Delete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data
	session1 := testdata.NewDatabaseSession().WithUserID("user-1").WithMapID("map-1").Build()
	session2 := testdata.NewDatabaseSession().WithUserID("user-2").WithMapID("map-1").Build()
	err := testDB.SeedFixtures(session1, session2)
	require.NoError(t, err)

	// Verify initial state
	var count int64
	err = testDB.DB.Model(&models.Session{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Delete session
	err = repo.Delete(ctx, session1.ID)
	require.NoError(t, err)

	// Verify deletion
	err = testDB.DB.Model(&models.Session{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify correct session was deleted
	remainingSession, err := repo.GetByID(ctx, session2.ID)
	require.NoError(t, err)
	assert.Equal(t, "user-2", remainingSession.UserID)

	// Verify deleted session is not found
	deletedSession, err := repo.GetByID(ctx, session1.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedSession)
}

func TestSessionRepository_ExpireOldSessions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed sessions with different ages
	now := time.Now()
	oldActiveSession := testdata.NewDatabaseSession().
		WithUserID("user-1").
		WithMapID("map-1").
		WithActive(true).
		Build()
	// Manually set last_active to simulate old session
	oldActiveSession.LastActive = now.Add(-2 * time.Hour)

	recentActiveSession := testdata.NewDatabaseSession().
		WithUserID("user-2").
		WithMapID("map-1").
		WithActive(true).
		Build()
	recentActiveSession.LastActive = now.Add(-30 * time.Minute)

	inactiveSession := testdata.NewDatabaseSession().
		WithUserID("user-3").
		WithMapID("map-1").
		WithActive(false).
		Build()

	err := testDB.SeedFixtures(oldActiveSession, recentActiveSession, inactiveSession)
	require.NoError(t, err)

	// Update last_active timestamps in database
	err = testDB.DB.Model(&models.Session{}).Where("id = ?", oldActiveSession.ID).
		Update("last_active", oldActiveSession.LastActive).Error
	require.NoError(t, err)
	err = testDB.DB.Model(&models.Session{}).Where("id = ?", recentActiveSession.ID).
		Update("last_active", recentActiveSession.LastActive).Error
	require.NoError(t, err)

	// Expire sessions older than 1 hour
	expiredCount, err := repo.ExpireOldSessions(ctx, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(1), expiredCount) // Only old active session should be expired

	// Verify old active session was expired
	expiredSession, err := repo.GetByID(ctx, oldActiveSession.ID)
	require.NoError(t, err)
	assert.False(t, expiredSession.IsActive)

	// Verify recent active session remains active
	activeSession, err := repo.GetByID(ctx, recentActiveSession.ID)
	require.NoError(t, err)
	assert.True(t, activeSession.IsActive)

	// Verify inactive session remains unchanged
	unchangedSession, err := repo.GetByID(ctx, inactiveSession.ID)
	require.NoError(t, err)
	assert.False(t, unchangedSession.IsActive)
}

func TestSessionRepository_ConcurrentAccess_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Test concurrent session creation
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			session := testdata.NewDatabaseSession().
				WithUserID(fmt.Sprintf("user-%d", index)).
				WithMapID("concurrent-map").
				WithPosition(40.0+float64(index)*0.001, -74.0+float64(index)*0.001).
				WithActive(true).
				Build()

			_, err := repo.Create(ctx, session)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// Verify all sessions were created
	var count int64
	err := testDB.DB.Model(&models.Session{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(numGoroutines), count)

	// Verify all sessions have unique IDs and users
	sessions, err := repo.GetActiveByMap(ctx, "concurrent-map")
	require.NoError(t, err)
	assert.Len(t, sessions, numGoroutines)

	ids := make(map[string]bool)
	users := make(map[string]bool)
	for _, session := range sessions {
		assert.False(t, ids[session.ID], "Duplicate session ID found: %s", session.ID)
		assert.False(t, users[session.UserID], "Duplicate user ID found: %s", session.UserID)
		ids[session.ID] = true
		users[session.UserID] = true
	}
}

// Benchmark tests for performance validation
func BenchmarkSessionRepository_Create_Integration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database integration benchmark in short mode")
	}

	testDB := testdata.Setup(b)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := testdata.NewDatabaseSession().
			WithUserID(fmt.Sprintf("benchmark-user-%d", i)).
			WithMapID("benchmark-map").
			WithPosition(40.0+float64(i)*0.0001, -74.0+float64(i)*0.0001).
			WithActive(true).
			Build()

		if _, err := repo.Create(ctx, session); err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}
	}
}

func BenchmarkSessionRepository_GetActiveByMapID_Integration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database integration benchmark in short mode")
	}

	testDB := testdata.Setup(b)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data
	for i := 0; i < 100; i++ {
		session := testdata.NewDatabaseSession().
			WithUserID(fmt.Sprintf("benchmark-user-%d", i)).
			WithMapID("benchmark-map").
			WithActive(true).
			Build()
		err := testDB.SeedFixtures(session)
		if err != nil {
			b.Fatalf("Failed to seed session: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions, err := repo.GetActiveByMap(ctx, "benchmark-map")
		if err != nil {
			b.Fatalf("Failed to get sessions: %v", err)
		}
		if len(sessions) != 100 {
			b.Fatalf("Expected 100 sessions, got %d", len(sessions))
		}
	}
}

func BenchmarkSessionRepository_UpdateAvatarPosition_Integration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database integration benchmark in short mode")
	}

	testDB := testdata.Setup(b)
	repo := NewSessionRepository(testDB.DB)
	ctx := context.Background()

	// Seed a session for updating
	session := testdata.NewDatabaseSession().
		WithUserID("benchmark-user").
		WithMapID("benchmark-map").
		WithActive(true).
		Build()
	err := testDB.SeedFixtures(session)
	if err != nil {
		b.Fatalf("Failed to seed session: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newPos := models.LatLng{
			Lat: 40.0 + float64(i)*0.0001,
			Lng: -74.0 + float64(i)*0.0001,
		}
		if _, err := repo.UpdateAvatarPosition(ctx, session.ID, newPos); err != nil {
			b.Fatalf("Failed to update avatar position: %v", err)
		}
	}
}