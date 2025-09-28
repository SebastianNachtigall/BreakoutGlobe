import { videoCallStore } from '../videoCallStore';
import { vi } from 'vitest';

// Mock RTCPeerConnection
const mockPeerConnection = {
  addTrack: vi.fn(),
  createOffer: vi.fn(),
  createAnswer: vi.fn(),
  setLocalDescription: vi.fn(),
  setRemoteDescription: vi.fn(),
  addIceCandidate: vi.fn(),
  close: vi.fn(),
  connectionState: 'new' as RTCPeerConnectionState,
  iceConnectionState: 'new' as RTCIceConnectionState,
  iceGatheringState: 'new' as RTCIceGatheringState,
  onicecandidate: null,
  ontrack: null,
  onconnectionstatechange: null,
  oniceconnectionstatechange: null,
  onicegatheringstatechange: null
};

// Mock MediaStream
const mockMediaStream = {
  getTracks: vi.fn().mockReturnValue([
    { kind: 'video', enabled: true, stop: vi.fn() },
    { kind: 'audio', enabled: true, stop: vi.fn() }
  ])
} as unknown as MediaStream;

// Mock getUserMedia
Object.defineProperty(global.navigator, 'mediaDevices', {
  value: {
    getUserMedia: vi.fn().mockResolvedValue(mockMediaStream)
  },
  writable: true
});

// Mock RTCPeerConnection constructor
global.RTCPeerConnection = vi.fn().mockImplementation(() => mockPeerConnection);

describe('VideoCallStore - Group Call Functionality', () => {
  beforeEach(() => {
    // Reset store state before each test
    const store = videoCallStore.getState();
    store.clearCall();
    store.leavePOICall();
  });

  describe('joinPOICall', () => {
    it('should set group call state when joining POI call', () => {
      const poiId = 'poi-123';

      videoCallStore.getState().joinPOICall(poiId);

      const state = videoCallStore.getState();
      expect(state.currentPOI).toBe(poiId);
      expect(state.isGroupCallActive).toBe(true);
      expect(state.callState).toBe('connecting');
    });

    it('should handle joining different POI calls', () => {
      const firstPOI = 'poi-123';
      const secondPOI = 'poi-456';

      // Join first POI
      videoCallStore.getState().joinPOICall(firstPOI);
      let state = videoCallStore.getState();
      expect(state.currentPOI).toBe(firstPOI);
      expect(state.isGroupCallActive).toBe(true);

      // Join second POI (should switch)
      videoCallStore.getState().joinPOICall(secondPOI);
      state = videoCallStore.getState();
      expect(state.currentPOI).toBe(secondPOI);
      expect(state.isGroupCallActive).toBe(true);
    });
  });

  describe('leavePOICall', () => {
    it('should clear group call state when leaving POI call', () => {
      const poiId = 'poi-123';

      // First join a POI call
      videoCallStore.getState().joinPOICall(poiId);
      let state = videoCallStore.getState();
      expect(state.currentPOI).toBe(poiId);
      expect(state.isGroupCallActive).toBe(true);

      // Then leave
      videoCallStore.getState().leavePOICall();
      state = videoCallStore.getState();
      expect(state.currentPOI).toBe(null);
      expect(state.isGroupCallActive).toBe(false);
      expect(state.callState).toBe('idle');
    });

    it('should clean up WebRTC resources when leaving POI call', () => {
      const poiId = 'poi-123';

      // Mock WebRTC service
      const mockWebRTCService = {
        cleanup: vi.fn()
      };

      // Mock MediaStream for test environment
      const mockMediaStream = {} as MediaStream;

      // Join POI call and set mock WebRTC service
      videoCallStore.getState().joinPOICall(poiId);
      videoCallStore.getState().setLocalStream(mockMediaStream);
      videoCallStore.getState().setRemoteStream(mockMediaStream);
      // Manually set webrtcService for test
      videoCallStore.setState({ webrtcService: mockWebRTCService as any });

      // Leave POI call
      videoCallStore.getState().leavePOICall();

      // Verify cleanup
      const state = videoCallStore.getState();
      expect(mockWebRTCService.cleanup).toHaveBeenCalled();
      expect(state.localStream).toBe(null);
      expect(state.remoteStream).toBe(null);
      expect(state.webrtcService).toBe(null);
    });
  });

  describe('group call state management', () => {
    it('should maintain separate state from regular video calls', () => {
      const poiId = 'poi-123';

      // Join POI call
      videoCallStore.getState().joinPOICall(poiId);
      let state = videoCallStore.getState();
      expect(state.currentPOI).toBe(poiId);
      expect(state.isGroupCallActive).toBe(true);
      expect(state.currentCall).toBe(null); // No regular call

      // Leave POI call
      videoCallStore.getState().leavePOICall();
      state = videoCallStore.getState();
      expect(state.currentPOI).toBe(null);
      expect(state.isGroupCallActive).toBe(false);
      expect(state.callState).toBe('idle');
    });

    it('should reset audio/video settings when leaving POI call', () => {
      const poiId = 'poi-123';

      // Join POI call
      videoCallStore.getState().joinPOICall(poiId);
      
      // Simulate changed audio/video settings
      videoCallStore.setState({ 
        isAudioEnabled: false, 
        isVideoEnabled: false 
      });

      // Leave POI call
      videoCallStore.getState().leavePOICall();

      // Settings should be reset to defaults
      const state = videoCallStore.getState();
      expect(state.isAudioEnabled).toBe(true);
      expect(state.isVideoEnabled).toBe(true);
    });
  });

  describe('dual peer WebRTC support', () => {
    it('should track second participant in POI', () => {
      const poiId = 'poi-123';
      const participant = {
        userId: 'user-456',
        displayName: 'Test User',
        avatarURL: 'https://example.com/avatar.jpg'
      };

      // Join POI call
      videoCallStore.getState().joinPOICall(poiId);
      
      // Add second participant
      videoCallStore.getState().addGroupCallParticipant(participant.userId, participant);

      const state = videoCallStore.getState();
      expect(state.groupCallParticipants.has(participant.userId)).toBe(true);
      expect(state.groupCallParticipants.get(participant.userId)).toEqual(participant);
    });

    it('should remove participant from POI call', () => {
      const poiId = 'poi-123';
      const participant = {
        userId: 'user-456',
        displayName: 'Test User',
        avatarURL: 'https://example.com/avatar.jpg'
      };

      // Join POI call and add participant
      videoCallStore.getState().joinPOICall(poiId);
      videoCallStore.getState().addGroupCallParticipant(participant.userId, participant);
      
      // Verify participant is added
      let state = videoCallStore.getState();
      expect(state.groupCallParticipants.has(participant.userId)).toBe(true);

      // Remove participant
      videoCallStore.getState().removeGroupCallParticipant(participant.userId);

      state = videoCallStore.getState();
      expect(state.groupCallParticipants.has(participant.userId)).toBe(false);
    });

    it('should track multiple remote streams for group call', () => {
      const poiId = 'poi-123';
      const userId1 = 'user-456';
      const userId2 = 'user-789';
      const mockStream1 = {} as MediaStream;
      const mockStream2 = {} as MediaStream;

      // Join POI call
      videoCallStore.getState().joinPOICall(poiId);
      
      // Add remote streams for different users
      videoCallStore.getState().setRemoteStreamForUser(userId1, mockStream1);
      videoCallStore.getState().setRemoteStreamForUser(userId2, mockStream2);

      const state = videoCallStore.getState();
      expect(state.remoteStreams.has(userId1)).toBe(true);
      expect(state.remoteStreams.get(userId1)).toBe(mockStream1);
      expect(state.remoteStreams.has(userId2)).toBe(true);
      expect(state.remoteStreams.get(userId2)).toBe(mockStream2);
    });

    it('should clear all participants and streams when leaving POI call', () => {
      const poiId = 'poi-123';
      const participant = {
        userId: 'user-456',
        displayName: 'Test User'
      };
      const mockStream = {} as MediaStream;

      // Join POI call and add participant/stream
      videoCallStore.getState().joinPOICall(poiId);
      videoCallStore.getState().addGroupCallParticipant(participant.userId, participant);
      videoCallStore.getState().setRemoteStreamForUser(participant.userId, mockStream);

      // Verify state is set
      let state = videoCallStore.getState();
      expect(state.groupCallParticipants.size).toBe(1);
      expect(state.remoteStreams.size).toBe(1);

      // Leave POI call
      videoCallStore.getState().leavePOICall();

      // Verify all cleared
      state = videoCallStore.getState();
      expect(state.groupCallParticipants.size).toBe(0);
      expect(state.remoteStreams.size).toBe(0);
    });

    it('should initialize GroupWebRTCService for group calls', async () => {
      const poiId = 'poi-123';

      // Join POI call
      videoCallStore.getState().joinPOICall(poiId);
      
      // Initialize group WebRTC service
      await videoCallStore.getState().initializeGroupWebRTC();

      const state = videoCallStore.getState();
      expect(state.groupWebRTCService).toBeDefined();
      expect(state.groupWebRTCService).not.toBe(null);
    });

    it('should add peer to group WebRTC service', async () => {
      const poiId = 'poi-123';
      const userId = 'user-456';
      const participant = {
        userId,
        displayName: 'Test User'
      };

      // Join POI call and initialize WebRTC
      videoCallStore.getState().joinPOICall(poiId);
      await videoCallStore.getState().initializeGroupWebRTC();
      
      // Add participant and peer
      videoCallStore.getState().addGroupCallParticipant(userId, participant);
      await videoCallStore.getState().addPeerToGroupCall(userId);

      const state = videoCallStore.getState();
      expect(state.groupWebRTCService?.peerConnections.has(userId)).toBe(true);
    });

    it('should remove peer from group WebRTC service', async () => {
      const poiId = 'poi-123';
      const userId = 'user-456';
      const participant = {
        userId,
        displayName: 'Test User'
      };

      // Join POI call and initialize WebRTC
      videoCallStore.getState().joinPOICall(poiId);
      await videoCallStore.getState().initializeGroupWebRTC();
      
      // Add participant and peer
      videoCallStore.getState().addGroupCallParticipant(userId, participant);
      await videoCallStore.getState().addPeerToGroupCall(userId);
      
      // Verify peer is added
      let state = videoCallStore.getState();
      expect(state.groupWebRTCService?.peerConnections.has(userId)).toBe(true);

      // Remove peer
      videoCallStore.getState().removePeerFromGroupCall(userId);

      state = videoCallStore.getState();
      expect(state.groupWebRTCService?.peerConnections.has(userId)).toBe(false);
    });
  });
});