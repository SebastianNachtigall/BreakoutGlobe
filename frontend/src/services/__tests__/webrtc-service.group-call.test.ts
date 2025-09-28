import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GroupWebRTCService } from '../webrtc-service';

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

describe('GroupWebRTCService', () => {
  let service: GroupWebRTCService;

  beforeEach(() => {
    vi.clearAllMocks();
    service = new GroupWebRTCService();
  });

  describe('multiple peer connection management', () => {
    it('should add peer connection for new user', async () => {
      const userId = 'user-123';
      
      await service.addPeer(userId);
      
      expect(service.peerConnections.has(userId)).toBe(true);
      expect(RTCPeerConnection).toHaveBeenCalledTimes(2); // Initial + new peer
    });

    it('should remove peer connection for user', async () => {
      const userId = 'user-123';
      
      // Add peer first
      await service.addPeer(userId);
      expect(service.peerConnections.has(userId)).toBe(true);
      
      // Remove peer
      service.removePeer(userId);
      
      expect(service.peerConnections.has(userId)).toBe(false);
      expect(service.remoteStreams.has(userId)).toBe(false);
    });

    it('should create offer for specific peer', async () => {
      const userId = 'user-123';
      const mockOffer = { type: 'offer', sdp: 'mock-offer-sdp' } as RTCSessionDescriptionInit;
      
      // Mock createOffer for the peer
      mockPeerConnection.createOffer.mockResolvedValue(mockOffer);
      mockPeerConnection.setLocalDescription.mockResolvedValue(undefined);
      
      await service.addPeer(userId);
      const offer = await service.createOfferForPeer(userId);
      
      expect(offer).toEqual(mockOffer);
      expect(mockPeerConnection.createOffer).toHaveBeenCalled();
      expect(mockPeerConnection.setLocalDescription).toHaveBeenCalledWith(mockOffer);
    });

    it('should handle answer from specific peer', async () => {
      const userId = 'user-123';
      const mockAnswer = { type: 'answer', sdp: 'mock-answer-sdp' } as RTCSessionDescriptionInit;
      
      mockPeerConnection.setRemoteDescription.mockResolvedValue(undefined);
      
      await service.addPeer(userId);
      await service.handleAnswerFromPeer(userId, mockAnswer);
      
      expect(mockPeerConnection.setRemoteDescription).toHaveBeenCalledWith(mockAnswer);
    });

    it('should track remote streams for multiple users', async () => {
      const userId1 = 'user-123';
      const userId2 = 'user-456';
      const mockStream1 = {} as MediaStream;
      const mockStream2 = {} as MediaStream;
      
      await service.addPeer(userId1);
      await service.addPeer(userId2);
      
      // Simulate receiving remote streams
      service.remoteStreams.set(userId1, mockStream1);
      service.remoteStreams.set(userId2, mockStream2);
      
      expect(service.remoteStreams.get(userId1)).toBe(mockStream1);
      expect(service.remoteStreams.get(userId2)).toBe(mockStream2);
      expect(service.remoteStreams.size).toBe(2);
    });
  });

  describe('error handling', () => {
    it('should throw error when creating offer for non-existent peer', async () => {
      const userId = 'non-existent-user';
      
      await expect(service.createOfferForPeer(userId)).rejects.toThrow('Peer connection not found for user: non-existent-user');
    });

    it('should throw error when handling answer for non-existent peer', async () => {
      const userId = 'non-existent-user';
      const mockAnswer = { type: 'answer', sdp: 'mock-answer-sdp' } as RTCSessionDescriptionInit;
      
      await expect(service.handleAnswerFromPeer(userId, mockAnswer)).rejects.toThrow('Peer connection not found for user: non-existent-user');
    });

    it('should handle peer connection failures gracefully', async () => {
      const userId = 'user-123';
      const onError = vi.fn();
      
      service.setCallbacks({ onError });
      await service.addPeer(userId);
      
      // Simulate connection failure
      const peerConnection = service.peerConnections.get(userId);
      if (peerConnection && peerConnection.onconnectionstatechange) {
        // Simulate failed state
        Object.defineProperty(peerConnection, 'connectionState', { value: 'failed' });
        peerConnection.onconnectionstatechange(new Event('connectionstatechange'));
      }
      
      expect(onError).toHaveBeenCalledWith(expect.any(Error));
    });
  });

  describe('cleanup', () => {
    it('should clean up all peer connections', async () => {
      const userId1 = 'user-123';
      const userId2 = 'user-456';
      
      await service.addPeer(userId1);
      await service.addPeer(userId2);
      
      expect(service.peerConnections.size).toBe(2);
      
      service.cleanup();
      
      expect(service.peerConnections.size).toBe(0);
      expect(service.remoteStreams.size).toBe(0);
      expect(mockPeerConnection.close).toHaveBeenCalled();
    });
  });
});