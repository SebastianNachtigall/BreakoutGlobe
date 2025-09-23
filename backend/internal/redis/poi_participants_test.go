package redis

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type POIParticipantsTestSuite struct {
	suite.Suite
	client       *redis.Client
	participants *POIParticipants
}

func (suite *POIParticipantsTestSuite) SetupSuite() {
	// Skip integration tests in short mode
	if testing.Short() {
		suite.T().Skip("Skipping Redis integration test in short mode")
	}
	
	// Connect to test Redis instance
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use DB 1 for tests to avoid conflicts
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	suite.Require().NoError(err, "Redis connection failed - make sure Redis is running")

	suite.client = client
	suite.participants = NewPOIParticipants(client)
}

func (suite *POIParticipantsTestSuite) SetupTest() {
	// Clean up Redis before each test
	ctx := context.Background()
	suite.client.FlushDB(ctx)
}

func (suite *POIParticipantsTestSuite) TearDownSuite() {
	if suite.client != nil {
		suite.client.Close()
	}
}

func (suite *POIParticipantsTestSuite) TestJoinPOI() {
	ctx := context.Background()
	poiID := "poi-123"
	sessionID := "session-456"

	// Execute
	err := suite.participants.JoinPOI(ctx, poiID, sessionID)

	// Assert
	suite.NoError(err)

	// Verify participant was added
	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(1, count)

	// Verify participant is in the set
	isParticipant, err := suite.participants.IsParticipant(ctx, poiID, sessionID)
	suite.NoError(err)
	suite.True(isParticipant)
}

func (suite *POIParticipantsTestSuite) TestJoinPOI_AlreadyParticipant() {
	ctx := context.Background()
	poiID := "poi-123"
	sessionID := "session-456"

	// Join first time
	err := suite.participants.JoinPOI(ctx, poiID, sessionID)
	suite.Require().NoError(err)

	// Join second time (should be idempotent)
	err = suite.participants.JoinPOI(ctx, poiID, sessionID)

	// Assert - should not error, but count should remain 1
	suite.NoError(err)

	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(1, count)
}

func (suite *POIParticipantsTestSuite) TestLeavePOI() {
	ctx := context.Background()
	poiID := "poi-123"
	sessionID := "session-456"

	// Setup - join first
	err := suite.participants.JoinPOI(ctx, poiID, sessionID)
	suite.Require().NoError(err)

	// Verify joined
	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.Require().NoError(err)
	suite.Equal(1, count)

	// Execute - leave
	err = suite.participants.LeavePOI(ctx, poiID, sessionID)

	// Assert
	suite.NoError(err)

	// Verify participant was removed
	count, err = suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(0, count)

	// Verify participant is not in the set
	isParticipant, err := suite.participants.IsParticipant(ctx, poiID, sessionID)
	suite.NoError(err)
	suite.False(isParticipant)
}

func (suite *POIParticipantsTestSuite) TestLeavePOI_NotParticipant() {
	ctx := context.Background()
	poiID := "poi-123"
	sessionID := "session-456"

	// Execute - leave without joining (should be idempotent)
	err := suite.participants.LeavePOI(ctx, poiID, sessionID)

	// Assert - should not error
	suite.NoError(err)

	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *POIParticipantsTestSuite) TestGetParticipantCount() {
	ctx := context.Background()
	poiID := "poi-123"

	// Initially should be 0
	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(0, count)

	// Add participants
	sessions := []string{"session-1", "session-2", "session-3"}
	for _, sessionID := range sessions {
		err := suite.participants.JoinPOI(ctx, poiID, sessionID)
		suite.Require().NoError(err)
	}

	// Check count
	count, err = suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(3, count)

	// Remove one participant
	err = suite.participants.LeavePOI(ctx, poiID, "session-2")
	suite.Require().NoError(err)

	// Check count again
	count, err = suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(2, count)
}

func (suite *POIParticipantsTestSuite) TestGetParticipants() {
	ctx := context.Background()
	poiID := "poi-123"

	// Initially should be empty
	participants, err := suite.participants.GetParticipants(ctx, poiID)
	suite.NoError(err)
	suite.Empty(participants)

	// Add participants
	sessions := []string{"session-1", "session-2", "session-3"}
	for _, sessionID := range sessions {
		err := suite.participants.JoinPOI(ctx, poiID, sessionID)
		suite.Require().NoError(err)
	}

	// Get participants
	participants, err = suite.participants.GetParticipants(ctx, poiID)
	suite.NoError(err)
	suite.Len(participants, 3)

	// Verify all sessions are present
	for _, sessionID := range sessions {
		suite.Contains(participants, sessionID)
	}
}

func (suite *POIParticipantsTestSuite) TestIsParticipant() {
	ctx := context.Background()
	poiID := "poi-123"
	sessionID := "session-456"
	otherSessionID := "session-789"

	// Initially should not be participant
	isParticipant, err := suite.participants.IsParticipant(ctx, poiID, sessionID)
	suite.NoError(err)
	suite.False(isParticipant)

	// Join POI
	err = suite.participants.JoinPOI(ctx, poiID, sessionID)
	suite.Require().NoError(err)

	// Should now be participant
	isParticipant, err = suite.participants.IsParticipant(ctx, poiID, sessionID)
	suite.NoError(err)
	suite.True(isParticipant)

	// Other session should not be participant
	isParticipant, err = suite.participants.IsParticipant(ctx, poiID, otherSessionID)
	suite.NoError(err)
	suite.False(isParticipant)
}

func (suite *POIParticipantsTestSuite) TestCanJoinPOI() {
	ctx := context.Background()
	poiID := "poi-123"
	maxParticipants := 3

	// Initially should be able to join
	canJoin, err := suite.participants.CanJoinPOI(ctx, poiID, maxParticipants)
	suite.NoError(err)
	suite.True(canJoin)

	// Add participants up to limit
	for i := 0; i < maxParticipants; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		err := suite.participants.JoinPOI(ctx, poiID, sessionID)
		suite.Require().NoError(err)
	}

	// Should now be at capacity
	canJoin, err = suite.participants.CanJoinPOI(ctx, poiID, maxParticipants)
	suite.NoError(err)
	suite.False(canJoin)

	// Remove one participant
	err = suite.participants.LeavePOI(ctx, poiID, "session-1")
	suite.Require().NoError(err)

	// Should be able to join again
	canJoin, err = suite.participants.CanJoinPOI(ctx, poiID, maxParticipants)
	suite.NoError(err)
	suite.True(canJoin)
}

func (suite *POIParticipantsTestSuite) TestJoinPOIWithCapacityCheck() {
	ctx := context.Background()
	poiID := "poi-123"
	maxParticipants := 2

	// Join up to capacity
	err := suite.participants.JoinPOIWithCapacityCheck(ctx, poiID, "session-1", maxParticipants)
	suite.NoError(err)

	err = suite.participants.JoinPOIWithCapacityCheck(ctx, poiID, "session-2", maxParticipants)
	suite.NoError(err)

	// Try to join when at capacity
	err = suite.participants.JoinPOIWithCapacityCheck(ctx, poiID, "session-3", maxParticipants)
	suite.Error(err)
	suite.Contains(err.Error(), "POI is at capacity")

	// Verify count is still 2
	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(2, count)

	// Verify session-3 is not a participant
	isParticipant, err := suite.participants.IsParticipant(ctx, poiID, "session-3")
	suite.NoError(err)
	suite.False(isParticipant)
}

func (suite *POIParticipantsTestSuite) TestRemoveAllParticipants() {
	ctx := context.Background()
	poiID := "poi-123"

	// Add multiple participants
	sessions := []string{"session-1", "session-2", "session-3", "session-4"}
	for _, sessionID := range sessions {
		err := suite.participants.JoinPOI(ctx, poiID, sessionID)
		suite.Require().NoError(err)
	}

	// Verify participants exist
	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.Require().NoError(err)
	suite.Equal(4, count)

	// Execute - remove all participants
	removedCount, err := suite.participants.RemoveAllParticipants(ctx, poiID)

	// Assert
	suite.NoError(err)
	suite.Equal(4, removedCount)

	// Verify no participants remain
	count, err = suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(0, count)

	participants, err := suite.participants.GetParticipants(ctx, poiID)
	suite.NoError(err)
	suite.Empty(participants)
}

func (suite *POIParticipantsTestSuite) TestRemoveAllParticipants_EmptyPOI() {
	ctx := context.Background()
	poiID := "poi-123"

	// Execute on empty POI
	removedCount, err := suite.participants.RemoveAllParticipants(ctx, poiID)

	// Assert - should not error
	suite.NoError(err)
	suite.Equal(0, removedCount)
}

func (suite *POIParticipantsTestSuite) TestRemoveParticipantFromAllPOIs() {
	ctx := context.Background()
	sessionID := "session-123"

	// Add session to multiple POIs
	pois := []string{"poi-1", "poi-2", "poi-3"}
	for _, poiID := range pois {
		err := suite.participants.JoinPOI(ctx, poiID, sessionID)
		suite.Require().NoError(err)
	}

	// Add other participants to verify they remain
	for _, poiID := range pois {
		err := suite.participants.JoinPOI(ctx, poiID, "other-session")
		suite.Require().NoError(err)
	}

	// Verify session is in all POIs
	for _, poiID := range pois {
		isParticipant, err := suite.participants.IsParticipant(ctx, poiID, sessionID)
		suite.Require().NoError(err)
		suite.True(isParticipant)

		count, err := suite.participants.GetParticipantCount(ctx, poiID)
		suite.Require().NoError(err)
		suite.Equal(2, count)
	}

	// Execute - remove session from all POIs
	removedCount, err := suite.participants.RemoveParticipantFromAllPOIs(ctx, sessionID)

	// Assert
	suite.NoError(err)
	suite.Equal(3, removedCount) // Removed from 3 POIs

	// Verify session is no longer in any POI
	for _, poiID := range pois {
		isParticipant, err := suite.participants.IsParticipant(ctx, poiID, sessionID)
		suite.NoError(err)
		suite.False(isParticipant)

		// Verify other participants remain
		count, err := suite.participants.GetParticipantCount(ctx, poiID)
		suite.NoError(err)
		suite.Equal(1, count)
	}
}

func (suite *POIParticipantsTestSuite) TestGetPOIsForParticipant() {
	ctx := context.Background()
	sessionID := "session-123"

	// Initially should have no POIs
	pois, err := suite.participants.GetPOIsForParticipant(ctx, sessionID)
	suite.NoError(err)
	suite.Empty(pois)

	// Join multiple POIs
	poiIDs := []string{"poi-1", "poi-2", "poi-3"}
	for _, poiID := range poiIDs {
		err := suite.participants.JoinPOI(ctx, poiID, sessionID)
		suite.Require().NoError(err)
	}

	// Get POIs for participant
	pois, err = suite.participants.GetPOIsForParticipant(ctx, sessionID)
	suite.NoError(err)
	suite.Len(pois, 3)

	// Verify all POIs are present
	for _, poiID := range poiIDs {
		suite.Contains(pois, poiID)
	}

	// Leave one POI
	err = suite.participants.LeavePOI(ctx, "poi-2", sessionID)
	suite.Require().NoError(err)

	// Should now have 2 POIs
	pois, err = suite.participants.GetPOIsForParticipant(ctx, sessionID)
	suite.NoError(err)
	suite.Len(pois, 2)
	suite.Contains(pois, "poi-1")
	suite.Contains(pois, "poi-3")
	suite.NotContains(pois, "poi-2")
}

func (suite *POIParticipantsTestSuite) TestConcurrentJoinLeave() {
	ctx := context.Background()
	poiID := "poi-123"
	maxParticipants := 5

	// Test concurrent operations
	const numGoroutines = 10
	const numOperations = 20

	// Channel to collect errors
	errChan := make(chan error, numGoroutines*numOperations)
	doneChan := make(chan bool, numGoroutines)

	// Start multiple goroutines doing join/leave operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { doneChan <- true }()

			for j := 0; j < numOperations; j++ {
				sessionID := fmt.Sprintf("session-%d-%d", goroutineID, j)

				// Try to join
				err := suite.participants.JoinPOIWithCapacityCheck(ctx, poiID, sessionID, maxParticipants)
				if err != nil {
					errChan <- err
				}

				// Immediately leave
				err = suite.participants.LeavePOI(ctx, poiID, sessionID)
				if err != nil {
					errChan <- err
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-doneChan
	}
	close(errChan)

	// Check for capacity errors (expected) vs unexpected errors
	capacityErrors := 0
	unexpectedErrors := 0
	for err := range errChan {
		if err != nil {
			if suite.Contains(err.Error(), "POI is at capacity") {
				capacityErrors++
			} else {
				unexpectedErrors++
				suite.T().Logf("Unexpected error: %v", err)
			}
		}
	}

	// Should have no unexpected errors
	suite.Equal(0, unexpectedErrors)

	// Final count should be 0 (all sessions left)
	count, err := suite.participants.GetParticipantCount(ctx, poiID)
	suite.NoError(err)
	suite.Equal(0, count)
}

func TestPOIParticipantsTestSuite(t *testing.T) {
	suite.Run(t, new(POIParticipantsTestSuite))
}