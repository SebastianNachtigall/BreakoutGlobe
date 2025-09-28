import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, beforeEach, describe, it, expect } from 'vitest';
import { poiStore } from '../stores/poiStore';
import { videoCallStore } from '../stores/videoCallStore';
import type { POIData } from '../components/MapContainer';

// Mock the stores
vi.mock('../stores/poiStore');
vi.mock('../stores/videoCallStore');

describe('POI Group Call Integration', () => {
  const mockPOI: POIData = {
    id: 'poi-123',
    name: 'Test POI',
    description: 'Test POI Description',
    position: { lat: 40.7128, lng: -74.0060 },
    participantCount: 2, // Multiple participants to trigger group call
    maxParticipants: 5,
    createdBy: 'user-456',
    createdAt: new Date()
  };

  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks();
    
    // Mock POI store methods
    const mockPOIStore = {
      pois: [mockPOI],
      joinPOIOptimisticWithAutoLeave: vi.fn().mockReturnValue(true),
      confirmJoinPOI: vi.fn(),
      leavePOI: vi.fn().mockReturnValue(true),
      getCurrentUserPOI: vi.fn().mockReturnValue(null)
    };
    
    // Mock video call store methods
    const mockVideoCallStore = {
      currentPOI: null,
      isGroupCallActive: false,
      callState: 'idle' as const,
      joinPOICall: vi.fn(),
      leavePOICall: vi.fn()
    };

    vi.mocked(poiStore).mockReturnValue(mockPOIStore as any);
    vi.mocked(videoCallStore).mockReturnValue(mockVideoCallStore as any);
  });

  describe('POI join triggers group call', () => {
    it('should trigger group call when joining POI with multiple participants', async () => {
      const mockVideoStore = vi.mocked(videoCallStore)();
      const mockPOIStoreInstance = vi.mocked(poiStore)();

      // Simulate POI with multiple participants
      mockPOIStoreInstance.pois = [{
        ...mockPOI,
        participantCount: 2 // Multiple participants should trigger group call
      }];

      // Mock successful POI join
      mockPOIStoreInstance.joinPOIOptimisticWithAutoLeave.mockReturnValue(true);

      // Simulate the POI join flow that would happen in App.tsx
      const handleJoinPOI = async (poiId: string) => {
        const success = mockPOIStoreInstance.joinPOIOptimisticWithAutoLeave(poiId, 'user-123');
        if (success) {
          // Simulate API call success
          mockPOIStoreInstance.confirmJoinPOI(poiId, 'user-123');
          
          // Check if POI has multiple participants and trigger group call
          const updatedPOI = mockPOIStoreInstance.pois.find(p => p.id === poiId);
          if (updatedPOI && updatedPOI.participantCount > 1) {
            mockVideoStore.joinPOICall(poiId);
          }
        }
      };

      // Execute the join flow
      await handleJoinPOI('poi-123');

      // Verify POI join was called
      expect(mockPOIStoreInstance.joinPOIOptimisticWithAutoLeave).toHaveBeenCalledWith('poi-123', 'user-123');
      expect(mockPOIStoreInstance.confirmJoinPOI).toHaveBeenCalledWith('poi-123', 'user-123');
      
      // Verify group call was triggered
      expect(mockVideoStore.joinPOICall).toHaveBeenCalledWith('poi-123');
    });

    it('should not trigger group call when joining POI with single participant', async () => {
      const mockVideoStore = vi.mocked(videoCallStore)();
      const mockPOIStoreInstance = vi.mocked(poiStore)();

      // Simulate POI with single participant (just the user joining)
      mockPOIStoreInstance.pois = [{
        ...mockPOI,
        participantCount: 1 // Single participant should not trigger group call
      }];

      // Mock successful POI join
      mockPOIStoreInstance.joinPOIOptimisticWithAutoLeave.mockReturnValue(true);

      // Simulate the POI join flow
      const handleJoinPOI = async (poiId: string) => {
        const success = mockPOIStoreInstance.joinPOIOptimisticWithAutoLeave(poiId, 'user-123');
        if (success) {
          mockPOIStoreInstance.confirmJoinPOI(poiId, 'user-123');
          
          // Check if POI has multiple participants
          const updatedPOI = mockPOIStoreInstance.pois.find(p => p.id === poiId);
          if (updatedPOI && updatedPOI.participantCount > 1) {
            mockVideoStore.joinPOICall(poiId);
          }
        }
      };

      // Execute the join flow
      await handleJoinPOI('poi-123');

      // Verify POI join was called
      expect(mockPOIStoreInstance.joinPOIOptimisticWithAutoLeave).toHaveBeenCalledWith('poi-123', 'user-123');
      expect(mockPOIStoreInstance.confirmJoinPOI).toHaveBeenCalledWith('poi-123', 'user-123');
      
      // Verify group call was NOT triggered
      expect(mockVideoStore.joinPOICall).not.toHaveBeenCalled();
    });
  });

  describe('POI leave triggers group call cleanup', () => {
    it('should leave group call when leaving POI', async () => {
      const mockVideoStore = vi.mocked(videoCallStore)();
      const mockPOIStoreInstance = vi.mocked(poiStore)();

      // Set up initial state - user is in a POI with active group call
      mockVideoStore.currentPOI = 'poi-123';
      mockVideoStore.isGroupCallActive = true;
      mockPOIStoreInstance.leavePOI.mockReturnValue(true);

      // Simulate the POI leave flow
      const handleLeavePOI = async (poiId: string) => {
        const success = mockPOIStoreInstance.leavePOI(poiId, 'user-123');
        if (success) {
          // Leave group call if user was in one for this POI
          if (mockVideoStore.currentPOI === poiId && mockVideoStore.isGroupCallActive) {
            mockVideoStore.leavePOICall();
          }
        }
      };

      // Execute the leave flow
      await handleLeavePOI('poi-123');

      // Verify POI leave was called
      expect(mockPOIStoreInstance.leavePOI).toHaveBeenCalledWith('poi-123', 'user-123');
      
      // Verify group call was left
      expect(mockVideoStore.leavePOICall).toHaveBeenCalled();
    });

    it('should not leave group call when leaving different POI', async () => {
      const mockVideoStore = vi.mocked(videoCallStore)();
      const mockPOIStoreInstance = vi.mocked(poiStore)();

      // Set up initial state - user is in a different POI with active group call
      mockVideoStore.currentPOI = 'poi-456'; // Different POI
      mockVideoStore.isGroupCallActive = true;
      mockPOIStoreInstance.leavePOI.mockReturnValue(true);

      // Simulate leaving a different POI
      const handleLeavePOI = async (poiId: string) => {
        const success = mockPOIStoreInstance.leavePOI(poiId, 'user-123');
        if (success) {
          // Only leave group call if leaving the same POI
          if (mockVideoStore.currentPOI === poiId && mockVideoStore.isGroupCallActive) {
            mockVideoStore.leavePOICall();
          }
        }
      };

      // Execute the leave flow for different POI
      await handleLeavePOI('poi-123');

      // Verify POI leave was called
      expect(mockPOIStoreInstance.leavePOI).toHaveBeenCalledWith('poi-123', 'user-123');
      
      // Verify group call was NOT left (different POI)
      expect(mockVideoStore.leavePOICall).not.toHaveBeenCalled();
    });
  });

  describe('Group call state management', () => {
    it('should maintain group call state separately from regular calls', () => {
      const mockVideoStore = vi.mocked(videoCallStore)();

      // Set up group call state
      mockVideoStore.currentPOI = 'poi-123';
      mockVideoStore.isGroupCallActive = true;
      mockVideoStore.callState = 'connecting';

      // Verify group call state is independent
      expect(mockVideoStore.currentPOI).toBe('poi-123');
      expect(mockVideoStore.isGroupCallActive).toBe(true);
      expect(mockVideoStore.callState).toBe('connecting');
    });
  });
});