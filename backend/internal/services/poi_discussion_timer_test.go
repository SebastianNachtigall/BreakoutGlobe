package services

import (
	"context"
	"testing"
	"time"

	"breakoutglobe/internal/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// POI Discussion Timer Test Scenario
type poiDiscussionTimerScenario struct {
	t               *testing.T
	mockRepo        *MockPOIRepository
	mockParts       *MockPOIParticipants
	mockPubsub      *MockPubSub
	mockUserService *MockUserService
	service         *POIService
}

func newPOIDiscussionTimerScenario(t *testing.T) *poiDiscussionTimerScenario {
	mockRepo := new(MockPOIRepository)
	mockParts := new(MockPOIParticipants)
	mockPubsub := new(MockPubSub)
	mockUserService := new(MockUserService)
	
	service := NewPOIService(mockRepo, mockParts, mockPubsub, mockUserService)
	
	return &poiDiscussionTimerScenario{
		t:               t,
		mockRepo:        mockRepo,
		mockParts:       mockParts,
		mockPubsub:      mockPubsub,
		mockUserService: mockUserService,
		service:         service,
	}
}

func (s *poiDiscussionTimerScenario) cleanup() {
	s.mockRepo.AssertExpectations(s.t)
	s.mockParts.AssertExpectations(s.t)
	s.mockPubsub.AssertExpectations(s.t)
	s.mockUserService.AssertExpectations(s.t)
}

func TestPOIService_DiscussionTimer_SimplifiedLogic(t *testing.T) {
	scenario := newPOIDiscussionTimerScenario(t)
	defer scenario.cleanup()

	poiID := "test-poi-1"
	user1ID := "user-1"
	user2ID := "user-2"
	
	// Create initial POI with no discussion active
	initialPOI := &models.POI{
		ID:                  poiID,
		MapID:              "map-123",
		Name:               "Test POI",
		Description:        "Test description",
		MaxParticipants:    10,
		IsDiscussionActive: false,
		DiscussionStartTime: nil,
	}

	// Test: Add first user - should not start discussion (need 2+ users)
	scenario.mockRepo.On("GetByID", mock.Anything, poiID).Return(initialPOI, nil).Once()
	scenario.mockParts.On("IsParticipant", mock.Anything, poiID, user1ID).Return(false, nil).Once()
	scenario.mockParts.On("CanJoinPOI", mock.Anything, poiID, 10).Return(true, nil).Once()
	scenario.mockParts.On("JoinPOI", mock.Anything, poiID, user1ID).Return(nil).Once()
	
	// updateDiscussionTimer call with 1 participant
	scenario.mockRepo.On("GetByID", mock.Anything, poiID).Return(initialPOI, nil).Once()
	scenario.mockParts.On("GetParticipantCount", mock.Anything, poiID).Return(1, nil).Once()
	// Should not update POI since discussion should remain inactive (no Update call expected)
	
	scenario.mockParts.On("GetParticipantCount", mock.Anything, poiID).Return(1, nil).Once()
	scenario.mockParts.On("GetParticipants", mock.Anything, poiID).Return([]string{user1ID}, nil).Once()
	scenario.mockUserService.On("GetUser", mock.Anything, user1ID).Return(&models.User{ID: user1ID, DisplayName: "User 1"}, nil).Once()
	scenario.mockPubsub.On("PublishPOIJoinedWithParticipants", mock.Anything, mock.AnythingOfType("redis.POIJoinedEventWithParticipants")).Return(nil).Once()

	err := scenario.service.JoinPOI(context.Background(), poiID, user1ID)
	require.NoError(t, err)

	// Test: Add second user - should start discussion
	poiWith1User := &models.POI{
		ID:                  poiID,
		MapID:              "map-123",
		Name:               "Test POI",
		Description:        "Test description",
		MaxParticipants:    10,
		IsDiscussionActive: false,
		DiscussionStartTime: nil,
	}
	
	scenario.mockRepo.On("GetByID", mock.Anything, poiID).Return(poiWith1User, nil).Once()
	scenario.mockParts.On("IsParticipant", mock.Anything, poiID, user2ID).Return(false, nil).Once()
	scenario.mockParts.On("CanJoinPOI", mock.Anything, poiID, 10).Return(true, nil).Once()
	scenario.mockParts.On("JoinPOI", mock.Anything, poiID, user2ID).Return(nil).Once()
	
	// updateDiscussionTimer call with 2 participants - should start discussion
	scenario.mockRepo.On("GetByID", mock.Anything, poiID).Return(poiWith1User, nil).Once()
	scenario.mockParts.On("GetParticipantCount", mock.Anything, poiID).Return(2, nil).Once()
	
	// Should update POI to start discussion
	scenario.mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(poi *models.POI) bool {
		return poi.ID == poiID && poi.IsDiscussionActive && poi.DiscussionStartTime != nil
	})).Return(nil).Once()
	
	scenario.mockParts.On("GetParticipantCount", mock.Anything, poiID).Return(2, nil).Once()
	scenario.mockParts.On("GetParticipants", mock.Anything, poiID).Return([]string{user1ID, user2ID}, nil).Once()
	scenario.mockUserService.On("GetUser", mock.Anything, user1ID).Return(&models.User{ID: user1ID, DisplayName: "User 1"}, nil).Once()
	scenario.mockUserService.On("GetUser", mock.Anything, user2ID).Return(&models.User{ID: user2ID, DisplayName: "User 2"}, nil).Once()
	scenario.mockPubsub.On("PublishPOIJoinedWithParticipants", mock.Anything, mock.AnythingOfType("redis.POIJoinedEventWithParticipants")).Return(nil).Once()

	beforeJoin := time.Now()
	err = scenario.service.JoinPOI(context.Background(), poiID, user2ID)
	require.NoError(t, err)

	// Test: Remove one user - should stop discussion
	poiWith2Users := &models.POI{
		ID:                  poiID,
		MapID:              "map-123",
		Name:               "Test POI",
		Description:        "Test description",
		MaxParticipants:    10,
		IsDiscussionActive: true,
		DiscussionStartTime: &beforeJoin,
	}
	
	scenario.mockRepo.On("GetByID", mock.Anything, poiID).Return(poiWith2Users, nil).Once()
	scenario.mockParts.On("IsParticipant", mock.Anything, poiID, user2ID).Return(true, nil).Once()
	scenario.mockParts.On("LeavePOI", mock.Anything, poiID, user2ID).Return(nil).Once()
	
	// updateDiscussionTimer call with 1 participant - should stop discussion
	scenario.mockRepo.On("GetByID", mock.Anything, poiID).Return(poiWith2Users, nil).Once()
	scenario.mockParts.On("GetParticipantCount", mock.Anything, poiID).Return(1, nil).Once()
	
	// Should update POI to stop discussion
	scenario.mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(poi *models.POI) bool {
		return poi.ID == poiID && !poi.IsDiscussionActive && poi.DiscussionStartTime == nil
	})).Return(nil).Once()
	
	scenario.mockParts.On("GetParticipantCount", mock.Anything, poiID).Return(1, nil).Once()
	scenario.mockParts.On("GetParticipants", mock.Anything, poiID).Return([]string{user1ID}, nil).Once()
	scenario.mockUserService.On("GetUser", mock.Anything, user1ID).Return(&models.User{ID: user1ID, DisplayName: "User 1"}, nil).Once()
	scenario.mockPubsub.On("PublishPOILeftWithParticipants", mock.Anything, mock.AnythingOfType("redis.POILeftEventWithParticipants")).Return(nil).Once()

	err = scenario.service.LeavePOI(context.Background(), poiID, user2ID)
	require.NoError(t, err)
}