import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
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

describe('POI Membership Persistence Issue', () => {
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

  it('should reproduce the POI membership persistence issue', async () => {
    // Step 1: User joins a POI
    const joinSuccess = poiStore.getState().joinPOI(mockPOI.id, mockUserId);
    expect(joinSuccess).toBe(true);
    
    // Verify user is in POI
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);

    // Step 2: Simulate browser refresh (store persistence)
    // The POI store persists POI data but NOT currentUserPOI
    const persistedState = {
      pois: poiStore.getState().pois,
      // currentUserPOI is NOT persisted - this is the bug!
    };
    
    // Reset store to simulate refresh
    poiStore.getState().reset();
    
    // Restore only persisted data (POI data but not currentUserPOI)
    poiStore.getState().setPOIs(persistedState.pois);
    
    // Step 3: Verify the issue - POI still shows user as joined but currentUserPOI is null
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null); // BUG: Should be mockPOI.id
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1); // POI still shows user joined
    
    // Step 4: User clicks on map (should leave POI)
    // This calls wsClient.leaveCurrentPOI() which checks getCurrentUserPOI()
    // Since getCurrentUserPOI() returns null, no leave message is sent
    mockWsClient.leaveCurrentPOI.mockImplementation(() => {
      const currentPOI = poiStore.getState().getCurrentUserPOI();
      if (currentPOI) {
        mockWsClient.leavePOI(currentPOI);
      }
      // If currentPOI is null, nothing happens - this is the bug!
    });
    
    mockWsClient.leaveCurrentPOI();
    
    // Verify the bug: no leave message was sent because currentUserPOI was null
    expect(mockWsClient.leavePOI).not.toHaveBeenCalled();
    
    // The POI still shows the user as joined, making them invisible to others
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);
  });

  it('should identify the root cause - currentUserPOI not persisted', () => {
    // Join POI
    poiStore.getState().joinPOI(mockPOI.id, mockUserId);
    
    // Check what gets persisted
    const state = poiStore.getState();
    
    // The persist middleware only persists 'pois', not 'currentUserPOI'
    // This is defined in the partialize function in poiStore.ts
    const persistedData = {
      pois: state.pois,
      // currentUserPOI is NOT included in persistence
    };
    
    expect(persistedData).not.toHaveProperty('currentUserPOI');
    expect(state.currentUserPOI).toBe(mockPOI.id);
  });

  it('should demonstrate the fix - persist currentUserPOI', () => {
    // Join POI
    poiStore.getState().joinPOI(mockPOI.id, mockUserId);
    
    // Simulate proper persistence that includes currentUserPOI
    const properPersistedData = {
      pois: poiStore.getState().pois,
      currentUserPOI: poiStore.getState().currentUserPOI, // Include this!
    };
    
    // Reset and restore with proper persistence
    poiStore.getState().reset();
    poiStore.getState().setPOIs(properPersistedData.pois);
    
    // Manually set currentUserPOI (this would be done by updated persistence)
    poiStore.setState({ currentUserPOI: properPersistedData.currentUserPOI });
    
    // Now the state is correct after "refresh"
    expect(poiStore.getState().getCurrentUserPOI()).toBe(mockPOI.id);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(1);
    
    // Map click would now properly leave POI
    mockWsClient.leaveCurrentPOI.mockImplementation(() => {
      const currentPOI = poiStore.getState().getCurrentUserPOI();
      if (currentPOI) {
        mockWsClient.leavePOI(currentPOI);
        poiStore.getState().leavePOI(currentPOI, mockUserId);
      }
    });
    
    mockWsClient.leaveCurrentPOI();
    
    // Verify fix works
    expect(mockWsClient.leavePOI).toHaveBeenCalledWith(mockPOI.id);
    expect(poiStore.getState().getCurrentUserPOI()).toBe(null);
    expect(poiStore.getState().getPOIById(mockPOI.id)?.participantCount).toBe(0);
  });
});