/**
 * Verification test for POI Group Call Second Call Bug Fix
 * 
 * This test verifies that the specific bug reported is fixed:
 * "when two users join a POI the group video call starts and is working fine. 
 * however, if the users start the call for a second time, no video can be established."
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { videoCallStore } from '../stores/videoCallStore';

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

describe('POI Group Call Bug Fix Verification', () => {
  beforeEach(() => {
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

  it('should fix the second call video establishment bug', async () => {
    const poiId = 'test-poi-office';
    
    console.log('🧪 Bug Fix Test: Reproducing the exact scenario from bug report');
    
    // === FIRST CALL (Working) ===
    console.log('📞 First call: Two users join POI, group video call starts');
    
    // User joins POI and group call starts
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Verify first call works
    let state = videoCallStore.getState();
    expect(state.isGroupCallActive).toBe(true);
    expect(state.groupWebRTCService).toBeTruthy();
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1);
    
    console.log('✅ First call established successfully');
    
    // === END FIRST CALL ===
    console.log('📵 First call ends');
    videoCallStore.getState().leavePOICall();
    
    // Verify cleanup
    state = videoCallStore.getState();
    expect(state.isGroupCallActive).toBe(false);
    expect(state.groupWebRTCService).toBe(null);
    
    console.log('✅ First call cleaned up successfully');
    
    // === SECOND CALL (Previously Broken, Now Fixed) ===
    console.log('📞 Second call: Users start call for second time (this was broken before)');
    
    // Users start the call for a second time
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Verify second call works (this would fail before the fix)
    state = videoCallStore.getState();
    expect(state.isGroupCallActive).toBe(true);
    expect(state.groupWebRTCService).toBeTruthy();
    expect(mockGetUserMedia).toHaveBeenCalledTimes(2); // Should be called again for new session
    
    console.log('✅ Second call established successfully - BUG IS FIXED!');
    
    // Cleanup
    videoCallStore.getState().leavePOICall();
  });

  it('should prevent duplicate WebRTC service creation during race conditions', async () => {
    const poiId = 'test-poi-race';
    
    console.log('🧪 Race Condition Test: Preventing duplicate initialization');
    
    // Join POI
    videoCallStore.getState().joinPOICall(poiId);
    
    // Simulate race condition: both App.tsx and WebSocket try to initialize simultaneously
    const initPromise1 = videoCallStore.getState().initializeGroupWebRTC();
    const initPromise2 = videoCallStore.getState().initializeGroupWebRTC();
    
    await Promise.all([initPromise1, initPromise2]);
    
    // Should only have one service and one media request
    const state = videoCallStore.getState();
    expect(state.groupWebRTCService).toBeTruthy();
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1); // Should not be called twice
    
    console.log('✅ Race condition handled correctly - no duplicate services created');
    
    // Cleanup
    videoCallStore.getState().leavePOICall();
  });

  it('should handle multiple POI switches without WebRTC conflicts', async () => {
    console.log('🧪 POI Switching Test: Multiple POI switches');
    
    const poi1 = 'poi-office-1';
    const poi2 = 'poi-office-2';
    
    // First POI
    videoCallStore.getState().joinPOICall(poi1);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    let state = videoCallStore.getState();
    expect(state.currentPOI).toBe(poi1);
    expect(state.groupWebRTCService).toBeTruthy();
    
    // Switch to second POI
    videoCallStore.getState().leavePOICall();
    videoCallStore.getState().joinPOICall(poi2);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    state = videoCallStore.getState();
    expect(state.currentPOI).toBe(poi2);
    expect(state.groupWebRTCService).toBeTruthy();
    
    // Switch back to first POI
    videoCallStore.getState().leavePOICall();
    videoCallStore.getState().joinPOICall(poi1);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    state = videoCallStore.getState();
    expect(state.currentPOI).toBe(poi1);
    expect(state.groupWebRTCService).toBeTruthy();
    
    console.log('✅ POI switching handled correctly');
    
    // Cleanup
    videoCallStore.getState().leavePOICall();
  });
});