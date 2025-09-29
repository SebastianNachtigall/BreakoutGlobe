import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { poiStore } from './poiStore';
import type { POIData } from '../components/MapContainer';

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

describe('poiStore', () => {
  const mockPOI: POIData = {
    id: 'poi-1',
    name: 'Test Meeting Room',
    description: 'A test POI for meetings',
    position: { lat: 40.7128, lng: -74.0060 },
    participantCount: 3,
    maxParticipants: 10,
    createdBy: 'user-123',
    createdAt: new Date(),
    discussionStartTime: null,
    isDiscussionActive: false,
    discussionDuration: 0
  };

  beforeEach(() => {
    // Reset store state
    poiStore.getState().reset();
    vi.clearAllMocks();
  });

  describe('POI Management', () => {
    it('should initialize with empty POI list', () => {
      const state = poiStore.getState();
      
      expect(state.pois).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should add a new POI', () => {
      poiStore.getState().addPOI(mockPOI);
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(1);
      expect(state.pois[0]).toEqual(mockPOI);
    });
    
    it('should initialize with no current user POI', () => {
      const state = poiStore.getState();
      expect(state.currentUserPOI).toBeNull();
    });

    it('should update an existing POI', () => {
      poiStore.getState().addPOI(mockPOI);
      
      const updatedPOI = { ...mockPOI, name: 'Updated Meeting Room', participantCount: 5 };
      poiStore.getState().updatePOI(mockPOI.id, updatedPOI);
      
      const state = poiStore.getState();
      expect(state.pois[0].name).toBe('Updated Meeting Room');
      expect(state.pois[0].participantCount).toBe(5);
    });

    it('should remove a POI', () => {
      poiStore.getState().addPOI(mockPOI);
      poiStore.getState().removePOI(mockPOI.id);
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(0);
    });

    it('should get POI by ID', () => {
      poiStore.getState().addPOI(mockPOI);
      
      const foundPOI = poiStore.getState().getPOIById(mockPOI.id);
      expect(foundPOI).toEqual(mockPOI);
      
      const notFoundPOI = poiStore.getState().getPOIById('non-existent');
      expect(notFoundPOI).toBeUndefined();
    });

    it('should set multiple POIs', () => {
      const pois = [
        mockPOI,
        { ...mockPOI, id: 'poi-2', name: 'Second POI' },
        { ...mockPOI, id: 'poi-3', name: 'Third POI' }
      ];
      
      poiStore.getState().setPOIs(pois);
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(3);
      expect(state.pois).toEqual(pois);
    });
  });

  describe('Participant Management', () => {
    beforeEach(() => {
      poiStore.getState().addPOI(mockPOI);
    });

    it('should join a POI', () => {
      const userId = 'user-456';
      poiStore.getState().joinPOI(mockPOI.id, userId);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(4); // Was 3, now 4
    });

    it('should leave a POI', () => {
      const userId = 'user-456';
      poiStore.getState().joinPOI(mockPOI.id, userId);
      poiStore.getState().leavePOI(mockPOI.id, userId);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(3); // Back to original
    });

    it('should not join POI if at capacity', () => {
      const fullPOI = { ...mockPOI, participantCount: 10, maxParticipants: 10 };
      poiStore.getState().updatePOI(mockPOI.id, fullPOI);
      
      const userId = 'user-456';
      const result = poiStore.getState().joinPOI(mockPOI.id, userId);
      
      expect(result).toBe(false);
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(10); // Unchanged
    });

    it('should not leave POI if participant count is already 0', () => {
      const emptyPOI = { ...mockPOI, participantCount: 0 };
      poiStore.getState().updatePOI(mockPOI.id, emptyPOI);
      
      const userId = 'user-456';
      const result = poiStore.getState().leavePOI(mockPOI.id, userId);
      
      expect(result).toBe(false);
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(0); // Unchanged
    });
  });

  describe('Auto-Leave Functionality', () => {
    const mockPOI2: POIData = {
      id: 'poi-2',
      name: 'Coffee Corner',
      description: 'A place for coffee breaks',
      position: { lat: 40.7589, lng: -73.9851 },
      participantCount: 1,
      maxParticipants: 5,
      createdBy: 'user-789',
      createdAt: new Date()
    };

    beforeEach(() => {
      poiStore.getState().addPOI(mockPOI);
      poiStore.getState().addPOI(mockPOI2);
    });

    it('should track current user POI when joining', () => {
      const userId = 'user-456';
      poiStore.getState().joinPOI(mockPOI.id, userId);
      
      const state = poiStore.getState();
      expect(state.currentUserPOI).toBe(mockPOI.id);
    });

    it('should clear current user POI when leaving', () => {
      const userId = 'user-456';
      poiStore.getState().joinPOI(mockPOI.id, userId);
      poiStore.getState().leavePOI(mockPOI.id, userId);
      
      const state = poiStore.getState();
      expect(state.currentUserPOI).toBeNull();
    });

    it('should auto-leave current POI when joining a different POI', () => {
      const userId = 'user-456';
      
      // Join first POI
      poiStore.getState().joinPOI(mockPOI.id, userId);
      let state = poiStore.getState();
      expect(state.currentUserPOI).toBe(mockPOI.id);
      expect(state.pois.find(p => p.id === mockPOI.id)?.participantCount).toBe(4);
      
      // Join second POI with auto-leave
      poiStore.getState().joinPOIWithAutoLeave(mockPOI2.id, userId);
      state = poiStore.getState();
      
      // Should be in second POI
      expect(state.currentUserPOI).toBe(mockPOI2.id);
      expect(state.pois.find(p => p.id === mockPOI2.id)?.participantCount).toBe(2);
      
      // Should have left first POI
      expect(state.pois.find(p => p.id === mockPOI.id)?.participantCount).toBe(3);
    });

    it('should not auto-leave when joining the same POI', () => {
      const userId = 'user-456';
      
      // Join POI
      poiStore.getState().joinPOI(mockPOI.id, userId);
      let state = poiStore.getState();
      expect(state.pois.find(p => p.id === mockPOI.id)?.participantCount).toBe(4);
      
      // Join same POI with auto-leave (should not double-join)
      poiStore.getState().joinPOIWithAutoLeave(mockPOI.id, userId);
      state = poiStore.getState();
      
      // Should still be in the POI with same count
      expect(state.currentUserPOI).toBe(mockPOI.id);
      expect(state.pois.find(p => p.id === mockPOI.id)?.participantCount).toBe(5); // Joined again
    });

    it('should leave current POI when calling leaveCurrentPOI', () => {
      const userId = 'user-456';
      
      // Join POI
      poiStore.getState().joinPOI(mockPOI.id, userId);
      let state = poiStore.getState();
      expect(state.currentUserPOI).toBe(mockPOI.id);
      expect(state.pois.find(p => p.id === mockPOI.id)?.participantCount).toBe(4);
      
      // Leave current POI
      const result = poiStore.getState().leaveCurrentPOI(userId);
      state = poiStore.getState();
      
      expect(result).toBe(true);
      expect(state.currentUserPOI).toBeNull();
      expect(state.pois.find(p => p.id === mockPOI.id)?.participantCount).toBe(3);
    });

    it('should return false when trying to leave current POI if not in any POI', () => {
      const userId = 'user-456';
      
      const result = poiStore.getState().leaveCurrentPOI(userId);
      const state = poiStore.getState();
      
      expect(result).toBe(false);
      expect(state.currentUserPOI).toBeNull();
    });

    it('should return current user POI', () => {
      const userId = 'user-456';
      
      expect(poiStore.getState().getCurrentUserPOI()).toBeNull();
      
      poiStore.getState().joinPOI(mockPOI.id, userId);
      expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    });
  });

  describe('Loading and Error States', () => {
    it('should set loading state', () => {
      poiStore.getState().setLoading(true);
      expect(poiStore.getState().isLoading).toBe(true);
      
      poiStore.getState().setLoading(false);
      expect(poiStore.getState().isLoading).toBe(false);
    });

    it('should set error state', () => {
      const error = 'Failed to load POIs';
      poiStore.getState().setError(error);
      expect(poiStore.getState().error).toBe(error);
      
      poiStore.getState().setError(null);
      expect(poiStore.getState().error).toBeNull();
    });

    it('should clear error when adding POI', () => {
      poiStore.getState().setError('Previous error');
      poiStore.getState().addPOI(mockPOI);
      
      expect(poiStore.getState().error).toBeNull();
    });
  });

  describe('Optimistic Updates', () => {
    beforeEach(() => {
      poiStore.getState().addPOI(mockPOI);
    });

    it('should perform optimistic POI creation', () => {
      const newPOI = { ...mockPOI, id: 'poi-new', name: 'New POI' };
      
      poiStore.getState().createPOIOptimistic(newPOI);
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(2);
      expect(state.pois.find(p => p.id === 'poi-new')).toEqual(newPOI);
    });

    it('should rollback optimistic POI creation on server rejection', () => {
      const newPOI = { ...mockPOI, id: 'poi-new', name: 'New POI' };
      
      poiStore.getState().createPOIOptimistic(newPOI);
      poiStore.getState().rollbackPOICreation('poi-new');
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(1);
      expect(state.pois.find(p => p.id === 'poi-new')).toBeUndefined();
    });

    it('should perform optimistic join operation', () => {
      const userId = 'user-456';
      
      poiStore.getState().joinPOIOptimistic(mockPOI.id, userId);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(4);
    });

    it('should rollback optimistic join on server rejection', () => {
      const userId = 'user-456';
      
      poiStore.getState().joinPOIOptimistic(mockPOI.id, userId);
      poiStore.getState().rollbackJoinPOI(mockPOI.id, userId);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(3); // Back to original
    });

    it('should confirm optimistic updates on server acceptance', () => {
      const userId = 'user-456';
      
      poiStore.getState().joinPOIOptimistic(mockPOI.id, userId);
      poiStore.getState().confirmJoinPOI(mockPOI.id, userId);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(4);
    });
  });

  describe('Persistence', () => {
    it('should have persistence configuration', () => {
      // Test that the store is configured with persistence
      expect(poiStore.persist).toBeDefined();
      expect(poiStore.persist.getOptions().name).toBe('breakout-globe-pois');
    });

    it('should include correct fields in persistence', () => {
      poiStore.getState().addPOI(mockPOI);
      
      // Test that partialize function includes the right fields
      const options = poiStore.persist.getOptions();
      const state = poiStore.getState();
      const persistedState = options.partialize(state);
      
      expect(persistedState).toHaveProperty('pois');
      expect(persistedState.pois).toHaveLength(1);
      expect(persistedState.pois[0]).toEqual(mockPOI);
    });

    it('should handle store rehydration', () => {
      // Test that the store can be rehydrated (basic functionality test)
      poiStore.getState().addPOI(mockPOI);
      poiStore.getState().reset();
      
      // After reset, state should be back to initial
      const state = poiStore.getState();
      expect(state.pois).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe('Real-time Updates', () => {
    it('should handle real-time POI updates', () => {
      poiStore.getState().addPOI(mockPOI);
      
      const updatedPOI = { ...mockPOI, participantCount: 7 };
      poiStore.getState().handleRealtimeUpdate(updatedPOI);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.participantCount).toBe(7);
    });

    it('should add new POI from real-time update', () => {
      const newPOI = { ...mockPOI, id: 'poi-realtime', name: 'Realtime POI' };
      
      poiStore.getState().handleRealtimeUpdate(newPOI);
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(1);
      expect(state.pois[0]).toEqual(newPOI);
    });

    it('should remove POI from real-time deletion', () => {
      poiStore.getState().addPOI(mockPOI);
      poiStore.getState().handleRealtimeDelete(mockPOI.id);
      
      const state = poiStore.getState();
      expect(state.pois).toHaveLength(0);
    });
  });

  describe('Discussion Timer', () => {
    beforeEach(() => {
      poiStore.getState().addPOI(mockPOI);
    });

    it('should handle discussion timer data from backend', () => {
      // Backend provides discussion timer data
      const discussionStartTime = new Date();
      poiStore.getState().updatePOI(mockPOI.id, { 
        discussionStartTime,
        isDiscussionActive: true,
        discussionDuration: 120
      });
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.discussionStartTime).toBe(discussionStartTime);
      expect(poi?.isDiscussionActive).toBe(true);
      expect(poi?.discussionDuration).toBe(120);
    });

    it('should handle discussion timer state updates from backend', () => {
      // Backend updates discussion state
      poiStore.getState().updatePOI(mockPOI.id, { 
        discussionStartTime: null,
        isDiscussionActive: false,
        discussionDuration: 0
      });
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.discussionStartTime).toBeNull();
      expect(poi?.isDiscussionActive).toBe(false);
      expect(poi?.discussionDuration).toBe(0);
    });

    it('should update discussion duration', () => {
      const newDuration = 180; // 3 minutes
      
      poiStore.getState().updateDiscussionTimer(mockPOI.id, newDuration);
      
      const state = poiStore.getState();
      const poi = state.pois.find(p => p.id === mockPOI.id);
      expect(poi?.discussionDuration).toBe(newDuration);
    });

    it('should get discussion timer state', () => {
      // Set up active discussion
      const startTime = new Date();
      poiStore.getState().updatePOI(mockPOI.id, { 
        participantCount: 3,
        discussionStartTime: startTime,
        isDiscussionActive: true,
        discussionDuration: 90
      });
      
      const timerState = poiStore.getState().getDiscussionTimerState(mockPOI.id);
      
      expect(timerState).toEqual({
        isActive: true,
        duration: 90,
        startTime: startTime
      });
    });

    it('should return null for discussion timer state of non-existent POI', () => {
      const timerState = poiStore.getState().getDiscussionTimerState('non-existent');
      expect(timerState).toBeNull();
    });
  });

  describe('Store Reset', () => {
    it('should reset store to initial state', () => {
      poiStore.getState().addPOI(mockPOI);
      poiStore.getState().setLoading(true);
      poiStore.getState().setError('Test error');
      
      poiStore.getState().reset();
      
      const state = poiStore.getState();
      expect(state.pois).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });
  });
});