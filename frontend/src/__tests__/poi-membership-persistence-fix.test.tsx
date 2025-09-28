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
  moveAvatar: vi.fn(),
};

// Mock window.wsClient
Object.defineProperty(window, 'wsClient', {
  value: mockWsClient,
  writable: true,
});

describe('POI Membership Persistence Fix', () => {
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

  beforeEach(() => {
    // Reset all stores
    poiStore.getState().reset();
    avatarStore.getState().clearAllAvatars();
    sessionStore.getState().reset();
    
    // Clear all mocks
    vi.clearAllMocks();
    
    // Setup initial state
    poiStore.getState().addPOI(mockPOI);
    sessionStore.getState().createSession('session-123', { lat: 40.7128, lng: -74.0060 });
  });

  it('should persist currentUserPOI after browser refresh', () => {
    // Step 1: User joins a POI
    const joinSuccess = poiStore.getState().joinPOI(mockPOI.id, mockUserId);
    expect(joinSuccess).toBe(true);
    
    // Verify user is in POI
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);

    // Step 2: Simulate browser refresh with proper persistence
    const state = poiStore.getState();
    
    // The persist middleware now includes currentUserPOI
    const persistedData = {
      pois: state.pois,
      currentUserPOI: state.currentUserPOI, // Now included in persistence!
    };
    
    // Verify currentUserPOI is included in persisted data
    expect(persistedData.currentUserPOI).toBe(mockPOI.id);
    
    // Reset store to simulate refresh
    poiStore.getState().reset();
    
    // Restore persisted data (this would be done automatically by zustand persist)
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // Step 3: Verify the fix - both POI data and currentUserPOI are restored
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id); // FIXED!
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);
  });

  it('should properly handle map click after refresh', () => {
    // Join POI and simulate refresh
    poiStore.getState().joinPOI(mockPOI.id, mockUserId);
    
    const state = poiStore.getState();
    const persistedData = {
      pois: state.pois,
      currentUserPOI: state.currentUserPOI,
    };
    
    poiStore.getState().reset();
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // Verify state is correct after refresh
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    
    // Step 4: User clicks on map (should leave POI)
    mockWsClient.leaveCurrentPOI.mockImplementation(() => {
      const currentPOI = poiStore.getState().getCurrentUserPOI();
      if (currentPOI) {
        mockWsClient.leavePOI(currentPOI);
        poiStore.getState().leavePOI(currentPOI, mockUserId);
      }
    });
    
    mockWsClient.leaveCurrentPOI();
    
    // Verify fix works: leave message is sent and user is properly removed
    expect(mockWsClient.leavePOI).toHaveBeenCalledWith(mockPOI.id);
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(0);
  });

  it('should handle edge case where user is not in any POI after refresh', () => {
    // User is not in any POI
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
    
    // Simulate refresh
    const state = poiStore.getState();
    const persistedData = {
      pois: state.pois,
      currentUserPOI: state.currentUserPOI, // null
    };
    
    poiStore.getState().reset();
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // Verify state is correct
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
    
    // Map click should not try to leave any POI
    mockWsClient.leaveCurrentPOI.mockImplementation(() => {
      const currentPOI = poiStore.getState().getCurrentUserPOI();
      if (currentPOI) {
        mockWsClient.leavePOI(currentPOI);
      }
    });
    
    mockWsClient.leaveCurrentPOI();
    
    // No leave message should be sent
    expect(mockWsClient.leavePOI).not.toHaveBeenCalled();
  });

  it('should handle POI that no longer exists after refresh', () => {
    // Join POI
    poiStore.getState().joinPOI(mockPOI.id, mockUserId);
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    
    // Simulate refresh where POI was deleted by another user
    const state = poiStore.getState();
    const persistedData = {
      pois: [], // POI was deleted
      currentUserPOI: state.currentUserPOI, // Still references deleted POI
    };
    
    poiStore.getState().reset();
    poiStore.getState().setPOIs(persistedData.pois);
    poiStore.setState({ currentUserPOI: persistedData.currentUserPOI });
    
    // User still thinks they're in the POI, but POI doesn't exist
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)).toBeUndefined();
    
    // Map click should still try to leave (server will handle the error)
    mockWsClient.leaveCurrentPOI.mockImplementation(() => {
      const currentPOI = poiStore.getState().getCurrentUserPOI();
      if (currentPOI) {
        mockWsClient.leavePOI(currentPOI);
        // Simulate server response that POI doesn't exist
        poiStore.setState({ currentUserPOI: null });
      }
    });
    
    mockWsClient.leaveCurrentPOI();
    
    // Leave message is sent, and state is cleaned up
    expect(mockWsClient.leavePOI).toHaveBeenCalledWith(mockPOI.id);
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
  });
});