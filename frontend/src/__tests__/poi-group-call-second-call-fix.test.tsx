/**
 * Integration test for POI Group Call Second Call Fix
 * 
 * This test verifies that the bug where second group calls fail is fixed.
 * It simulates the exact scenario described in the bug report.
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { videoCallStore } from '../stores/videoCallStore';
import { poiStore } from '../stores/poiStore';

// Mock WebRTC APIs
const mockGetUserMedia = vi.fn();
const mockRTCPeerConnection = vi.fn();

Object.defineProperty(global.navigator, 'mediaDevices', {
  value: {
    getUserMedia: mockGetUserMedia,
  },
  writable: true,
});

Object.defineProperty(global, 'RTCPeerConnection', {
  value: mockRTCPeerConnection,
  writable: true,
});

describe('POI Group Call Second Call Fix - Integration Test', () => {
  beforeEach(() => {
    // Reset stores (but don't call leavePOICall as it interferes with tests)
    poiStore.getState().reset();
    
    // Reset mocks
    vi.clearAllMocks();
    
    // Setup mock implementations
    mockGetUserMedia.mockResolvedValue({
      id: 'mock-stream-id',
      getTracks: () => [
        { kind: 'video', enabled: true, stop: vi.fn() },
        { kind: 'audio', enabled: true, stop: vi.fn() }
      ]
    });
    
    mockRTCPeerConnection.mockImplementation(() => ({
      addTrack: vi.fn(),
      createOffer: vi.fn().mockResolvedValue({ type: 'offer', sdp: 'mock-sdp' }),
      setLocalDescription: vi.fn().mockResolvedValue(undefined),
      close: vi.fn(),
      onicecandidate: null,
      ontrack: null,
      onconnectionstatechange: null,
      oniceconnectionstatechange: null,
      onicegatheringstatechange: null,
    }));
  });

  afterEach(() => {
    // Cleanup after each test
    videoCallStore.getState().leavePOICall();
  });

  it('should successfully handle multiple group call sessions without video failure', async () => {
    const poiId = 'test-poi-office';
    
    console.log('🧪 Integration Test: Simulating real-world scenario');
    
    // === FIRST GROUP CALL SESSION ===
    console.log('📞 Starting first group call session');
    
    // User joins POI and starts group call
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Add another participant (simulating second user joining)
    videoCallStore.getState().addGroupCallParticipant('user-sebastian', {
      userId: 'user-sebastian',
      displayName: 'Sebastian Nachtigall',
      avatarURL: 'https://example.com/sebastian.jpg'
    });
    
    // Ensure we have a fresh state reference after async operations
    let store = videoCallStore.getState();
    if (store.groupWebRTCService) {
      await videoCallStore.getState().addPeerToGroupCall('user-sebastian');
    }
    
    // Verify first call is working
    store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(true);
    expect(store.currentPOI).toBe(poiId);
    expect(store.groupWebRTCService).toBeTruthy();
    expect(store.groupCallParticipants.size).toBe(1);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1);
    
    console.log('✅ First group call session established successfully');
    
    // === END FIRST CALL ===
    console.log('📵 Ending first group call session');
    videoCallStore.getState().leavePOICall();
    
    // Verify cleanup
    store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(false);
    expect(store.currentPOI).toBe(null);
    expect(store.groupWebRTCService).toBe(null);
    expect(store.groupCallParticipants.size).toBe(0);
    
    console.log('✅ First group call session cleaned up successfully');
    
    // === SECOND GROUP CALL SESSION (This is where the bug occurred) ===
    console.log('📞 Starting second group call session (critical test)');
    
    // User joins POI again and starts new group call
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Add participant again (simulating users rejoining)
    videoCallStore.getState().addGroupCallParticipant('user-sebastian', {
      userId: 'user-sebastian',
      displayName: 'Sebastian Nachtigall',
      avatarURL: 'https://example.com/sebastian.jpg'
    });
    
    // Ensure we have a fresh state reference after async operations
    store = videoCallStore.getState();
    if (store.groupWebRTCService) {
      await videoCallStore.getState().addPeerToGroupCall('user-sebastian');
    }
    
    // Verify second call is working (this would fail before the fix)
    store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(true);
    expect(store.currentPOI).toBe(poiId);
    expect(store.groupWebRTCService).toBeTruthy();
    expect(store.groupCallParticipants.size).toBe(1);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(2); // Should be called again for new session
    
    console.log('✅ Second group call session established successfully');
    
    // === VERIFY NO DUPLICATE SERVICES ===
    console.log('🔍 Verifying no duplicate WebRTC services');
    
    // Attempt to initialize again (should be prevented)
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Should still have only one service and no additional media requests
    store = videoCallStore.getState();
    expect(store.groupWebRTCService).toBeTruthy();
    expect(mockGetUserMedia).toHaveBeenCalledTimes(2); // Should not increase
    
    console.log('✅ Duplicate initialization prevention working correctly');
    
    // === FINAL CLEANUP ===
    console.log('🧹 Final cleanup');
    videoCallStore.getState().leavePOICall();
    
    store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(false);
    expect(store.currentPOI).toBe(null);
    expect(store.groupWebRTCService).toBe(null);
    
    console.log('✅ Integration test completed successfully - Bug is fixed!');
  });

  it('should handle rapid POI switching without WebRTC conflicts', async () => {
    const poiId1 = 'poi-office-1';
    const poiId2 = 'poi-office-2';
    
    console.log('🧪 Testing rapid POI switching');
    
    // Join first POI
    videoCallStore.getState().joinPOICall(poiId1);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    let store = videoCallStore.getState();
    expect(store.currentPOI).toBe(poiId1);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1);
    
    // Rapidly switch to second POI (simulating user moving between POIs)
    videoCallStore.getState().leavePOICall();
    videoCallStore.getState().joinPOICall(poiId2);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    store = videoCallStore.getState();
    expect(store.currentPOI).toBe(poiId2);
    expect(store.isGroupCallActive).toBe(true);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(2);
    
    // Switch back to first POI
    videoCallStore.getState().leavePOICall();
    videoCallStore.getState().joinPOICall(poiId1);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    store = videoCallStore.getState();
    expect(store.currentPOI).toBe(poiId1);
    expect(store.isGroupCallActive).toBe(true);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(3);
    
    console.log('✅ Rapid POI switching handled correctly');
  });

  it('should prevent WebSocket duplicate initialization race condition', async () => {
    const poiId = 'poi-race-test';
    
    console.log('🧪 Testing WebSocket race condition prevention');
    
    // Simulate the scenario where both App.tsx and WebSocket client try to initialize
    videoCallStore.getState().joinPOICall(poiId);
    
    // Start both initializations simultaneously (race condition)
    const promise1 = videoCallStore.getState().initializeGroupWebRTC();
    const promise2 = videoCallStore.getState().initializeGroupWebRTC();
    
    await Promise.all([promise1, promise2]);
    
    // Should only have one service and one media request
    const store = videoCallStore.getState();
    expect(store.groupWebRTCService).toBeTruthy();
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1); // Should not be called twice
    
    console.log('✅ Race condition prevention working correctly');
  });
});