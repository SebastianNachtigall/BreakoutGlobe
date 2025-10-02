/**
 * Test for POI Group Call Duplicate Initialization Bug
 * 
 * This test reproduces the bug where starting a second group call fails
 * because WebRTC services are initialized multiple times without proper cleanup.
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

describe('POI Group Call Duplicate Initialization Bug', () => {
  beforeEach(() => {
    // Reset stores
    videoCallStore.getState().leavePOICall();
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

  it('should handle multiple group call initializations without duplicate WebRTC services', async () => {
    const poiId = 'test-poi-123';

    // First call initialization
    console.log('🧪 Test: Starting first group call');
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Get fresh state after initialization
    let store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(true);
    expect(store.currentPOI).toBe(poiId);
    expect(store.groupWebRTCService).toBeTruthy();
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1);
    
    const firstService = store.groupWebRTCService;
    
    // End first call
    console.log('🧪 Test: Ending first group call');
    videoCallStore.getState().leavePOICall();
    
    // Get fresh state after leaving
    store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(false);
    expect(store.currentPOI).toBe(null);
    expect(store.groupWebRTCService).toBe(null);
    
    // Second call initialization (this should work without issues)
    console.log('🧪 Test: Starting second group call');
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Get fresh state after second initialization
    store = videoCallStore.getState();
    expect(store.isGroupCallActive).toBe(true);
    expect(store.currentPOI).toBe(poiId);
    expect(store.groupWebRTCService).toBeTruthy();
    expect(store.groupWebRTCService).not.toBe(firstService); // Should be a new instance
    expect(mockGetUserMedia).toHaveBeenCalledTimes(2); // Should be called again for new stream
  });

  it('should prevent duplicate initialization when already active', async () => {
    const poiId = 'test-poi-123';

    // First initialization
    videoCallStore.getState().joinPOICall(poiId);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    let store = videoCallStore.getState();
    const firstService = store.groupWebRTCService;
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1);
    
    // Attempt duplicate initialization (should be prevented)
    console.log('🧪 Test: Attempting duplicate initialization');
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Get fresh state and verify no duplicate initialization
    store = videoCallStore.getState();
    expect(store.groupWebRTCService).toBe(firstService);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(1); // Should not be called again
  });

  it('should cleanup properly when switching between POIs', async () => {
    const poiId1 = 'test-poi-1';
    const poiId2 = 'test-poi-2';

    // Join first POI
    videoCallStore.getState().joinPOICall(poiId1);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    let store = videoCallStore.getState();
    const firstService = store.groupWebRTCService;
    expect(store.currentPOI).toBe(poiId1);
    
    // Switch to second POI (should cleanup first and initialize second)
    videoCallStore.getState().leavePOICall();
    videoCallStore.getState().joinPOICall(poiId2);
    await videoCallStore.getState().initializeGroupWebRTC();
    
    // Get fresh state after switching
    store = videoCallStore.getState();
    expect(store.currentPOI).toBe(poiId2);
    expect(store.groupWebRTCService).not.toBe(firstService);
    expect(mockGetUserMedia).toHaveBeenCalledTimes(2);
  });
});