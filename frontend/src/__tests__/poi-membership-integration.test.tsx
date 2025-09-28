import { describe, it, expect, beforeEach, vi } from 'vitest';
import { poiStore } from '../stores/poiStore';
import { avatarStore } from '../stores/avatarStore';
import { sessionStore } from '../stores/sessionStore';
import type { POIData } from '../components/MapContainer';

// Mock WebSocket client
const mockWsClient = {
  isConnected: vi.fn(() => true),
  joinPOI: vi.fn(),
  leavePOI: vi.fn(),
  leaveCurrentPOI: vi.fn(),
  joinPOIWithAutoLeave: vi.fn(),
  send: vi.fn(),
};

describe('POI Membership Integration Test', () => {
  
  const mockPOI: POIData = {
    id: 'poi-1',
    name: 'Test POI',
    description: 'A test POI',
    position: { lat: 40.7128, lng: -74.0060 },
    participantCount: 0,
    maxParticipants: 5,
    participants: [],
    createdBy: 'user-1',
    createdAt: new Date(),
  };

  const mockUserId = 'user-123';
  const mockSessionId = 'session-123';

  beforeEach(() => {
    // Reset all stores
    poiStore.getState().reset();
    avatarStore.getState().clearAllAvatars();
    sessionStore.getState().reset();
    
    // Clear all mocks
    vi.clearAllMocks();
    
    // Setup initial state
    poiStore.getState().addPOI(mockPOI);
    sessionStore.getState().createSession(mockSessionId, { lat: 40.7128, lng: -74.0060 });
    
    // Setup mock WebSocket client behavior
    mockWsClient.leaveCurrentPOI.mockImplementation(() => {
      const currentPOI = poiStore.getState().getCurrentUserPOI();
      if (currentPOI) {
        mockWsClient.leavePOI(currentPOI);
        poiStore.getState().leavePOI(currentPOI, mockSessionId);
      }
    });
    
    mockWsClient.joinPOI.mockImplementation((poiId: string) => {
      poiStore.getState().joinPOI(poiId, mockSessionId);
    });
    
    mockWsClient.joinPOIWithAutoLeave.mockImplementation((poiId: string) => {
      poiStore.getState().joinPOIWithAutoLeave(poiId, mockSessionId);
    });
  });

  it('should handle complete POI join/leave flow with persistence', () => {
    // Step 1: User joins POI
    mockWsClient.joinPOI(mockPOI.id);
    
    // Verify optimistic update
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);
    
    // Verify WebSocket method was called
    expect(mockWsClient.joinPOI).toHaveBeenCalledWith(mockPOI.id);
    
    // Step 2: Simulate browser refresh (persistence test)
    const state = poiStore.getState();
    const persistedData = {
      pois: state.pois,
      currentUserPOI: state.currentUserPOI, // This should be persisted now
    };
    
    // Reset store to simulate refresh
    poiStore.getState().reset();
    
    // Restore persisted data
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // Verify state is correctly restored
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);
    
    // Step 3: User clicks on map (should leave POI)
    mockWsClient.leaveCurrentPOI();
    
    // Verify leave methods were called
    expect(mockWsClient.leaveCurrentPOI).toHaveBeenCalled();
    expect(mockWsClient.leavePOI).toHaveBeenCalledWith(mockPOI.id);
    
    // Verify optimistic update
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(0);
  });

  it('should handle edge case where user tries to leave non-existent POI after refresh', () => {
    // Step 1: User joins POI
    mockWsClient.joinPOI(mockPOI.id);
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    
    // Step 2: Simulate refresh where POI was deleted
    const state = poiStore.getState();
    const persistedData = {
      pois: [], // POI was deleted by another user
      currentUserPOI: state.currentUserPOI, // Still references deleted POI
    };
    
    poiStore.getState().reset();
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // User still thinks they're in the POI
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)).toBeUndefined();
    
    // Step 3: User clicks on map
    mockWsClient.leaveCurrentPOI();
    
    // Leave message should still be sent (server will handle the error)
    expect(mockWsClient.leaveCurrentPOI).toHaveBeenCalled();
    expect(mockWsClient.leavePOI).toHaveBeenCalledWith(mockPOI.id);
    
    // The store should be updated even if POI doesn't exist locally
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
  });

  it('should handle auto-leave when joining another POI after refresh', () => {
    // Create second POI
    const secondPOI: POIData = {
      ...mockPOI,
      id: 'poi-2',
      name: 'Second POI'
    };
    poiStore.getState().addPOI(secondPOI);
    
    // Step 1: User joins first POI
    mockWsClient.joinPOI(mockPOI.id);
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    
    // Step 2: Simulate refresh
    const state = poiStore.getState();
    const persistedData = {
      pois: state.pois,
      currentUserPOI: state.currentUserPOI,
    };
    
    poiStore.getState().reset();
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // Verify state after refresh
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    
    // Step 3: User joins second POI (should auto-leave first)
    mockWsClient.joinPOIWithAutoLeave(secondPOI.id);
    
    // Should call the auto-leave method
    expect(mockWsClient.joinPOIWithAutoLeave).toHaveBeenCalledWith(secondPOI.id);
    
    // Verify optimistic updates
    expect(poiStore.getState().getCurrentUserPOI()).toBe(secondPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(0);
    expect(poiStore.getState().getPOIById(secondPOI.id)?.participantCount).toBe(1);
  });
});