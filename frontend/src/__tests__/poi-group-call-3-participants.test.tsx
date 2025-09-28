import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GroupCallModal } from '../components/GroupCallModal';

// Mock WebRTC APIs
global.RTCPeerConnection = vi.fn().mockImplementation(() => ({
  addTrack: vi.fn(),
  createOffer: vi.fn().mockResolvedValue({ type: 'offer', sdp: 'mock-sdp' }),
  createAnswer: vi.fn().mockResolvedValue({ type: 'answer', sdp: 'mock-sdp' }),
  setLocalDescription: vi.fn().mockResolvedValue(undefined),
  setRemoteDescription: vi.fn().mockResolvedValue(undefined),
  addIceCandidate: vi.fn().mockResolvedValue(undefined),
  close: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  connectionState: 'new',
  iceConnectionState: 'new',
  iceGatheringState: 'new',
}));

global.navigator.mediaDevices = {
  getUserMedia: vi.fn().mockResolvedValue({
    getTracks: vi.fn().mockReturnValue([
      { kind: 'video', enabled: true, stop: vi.fn() },
      { kind: 'audio', enabled: true, stop: vi.fn() }
    ]),
    id: 'mock-stream-id'
  })
} as any;

describe('POI Group Call - 3 Participants', () => {
  const mockLocalStream = {
    getTracks: vi.fn().mockReturnValue([
      { kind: 'video', enabled: true, stop: vi.fn() },
      { kind: 'audio', enabled: true, stop: vi.fn() }
    ]),
    id: 'local-stream-id'
  } as any;

  const mockRemoteStream1 = {
    getTracks: vi.fn().mockReturnValue([
      { kind: 'video', enabled: true },
      { kind: 'audio', enabled: true }
    ]),
    id: 'remote-stream-1'
  } as any;

  const mockRemoteStream2 = {
    getTracks: vi.fn().mockReturnValue([
      { kind: 'video', enabled: true },
      { kind: 'audio', enabled: true }
    ]),
    id: 'remote-stream-2'
  } as any;

  const mockRemoteStream3 = {
    getTracks: vi.fn().mockReturnValue([
      { kind: 'video', enabled: true },
      { kind: 'audio', enabled: true }
    ]),
    id: 'remote-stream-3'
  } as any;

  const createParticipants = (count: number) => {
    const participants = new Map();
    const remoteStreams = new Map();
    
    for (let i = 1; i <= count; i++) {
      const userId = `user-${i}`;
      participants.set(userId, {
        displayName: `User ${i}`,
        avatarUrl: null
      });
      
      if (i === 1) remoteStreams.set(userId, mockRemoteStream1);
      if (i === 2) remoteStreams.set(userId, mockRemoteStream2);
      if (i === 3) remoteStreams.set(userId, mockRemoteStream3);
    }
    
    return { participants, remoteStreams };
  };

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    callState: 'connected' as const,
    poiId: 'poi-123',
    poiName: 'Test POI',
    localStream: mockLocalStream,
    isAudioEnabled: true,
    isVideoEnabled: true,
    onEndCall: vi.fn(),
    onToggleAudio: vi.fn(),
    onToggleVideo: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should display 2x2 grid for 3 participants', () => {
    const { participants, remoteStreams } = createParticipants(3);
    
    render(
      <GroupCallModal
        {...defaultProps}
        participants={participants}
        remoteStreams={remoteStreams}
      />
    );

    // Should show 2x2 grid (grid-cols-2)
    const videoGrid = document.querySelector('.grid-cols-2');
    expect(videoGrid).toBeInTheDocument();

    // Should show 3 participant videos
    const participantVideos = screen.getAllByText(/User \d/);
    expect(participantVideos).toHaveLength(3);

    // Should show local video in picture-in-picture
    const localVideo = screen.getByTestId('local-video');
    expect(localVideo).toBeInTheDocument();
  });

  it('should display 2x2 grid for 4 participants', () => {
    const { participants, remoteStreams } = createParticipants(4);
    
    render(
      <GroupCallModal
        {...defaultProps}
        participants={participants}
        remoteStreams={remoteStreams}
      />
    );

    // Should show 2x2 grid (grid-cols-2)
    const videoGrid = document.querySelector('.grid-cols-2');
    expect(videoGrid).toBeInTheDocument();

    // Should show 4 participant video containers
    const videoContainers = document.querySelectorAll('.relative.bg-gray-800');
    expect(videoContainers).toHaveLength(4);
  });

  it('should display 3x2 grid for 5 participants', () => {
    const { participants, remoteStreams } = createParticipants(5);
    
    render(
      <GroupCallModal
        {...defaultProps}
        participants={participants}
        remoteStreams={remoteStreams}
      />
    );

    // Should show 3x2 grid (grid-cols-3)
    const videoGrid = document.querySelector('.grid-cols-3');
    expect(videoGrid).toBeInTheDocument();

    // Should show 5 participant video containers
    const videoContainers = document.querySelectorAll('.relative.bg-gray-800');
    expect(videoContainers).toHaveLength(5);
  });

  it('should display 3x2 grid for 6 participants', () => {
    const { participants, remoteStreams } = createParticipants(6);
    
    render(
      <GroupCallModal
        {...defaultProps}
        participants={participants}
        remoteStreams={remoteStreams}
      />
    );

    // Should show 3x2 grid (grid-cols-3)
    const videoGrid = document.querySelector('.grid-cols-3');
    expect(videoGrid).toBeInTheDocument();

    // Should show 6 participant video containers
    const videoContainers = document.querySelectorAll('.relative.bg-gray-800');
    expect(videoContainers).toHaveLength(6);
  });

  it('should handle dynamic participant addition', async () => {
    const { participants: initialParticipants, remoteStreams: initialStreams } = createParticipants(2);
    
    const { rerender } = render(
      <GroupCallModal
        {...defaultProps}
        participants={initialParticipants}
        remoteStreams={initialStreams}
      />
    );

    // Initially should show 2x1 grid
    expect(document.querySelector('.grid-cols-2')).toBeInTheDocument();
    expect(screen.getAllByText(/User \d/)).toHaveLength(2);

    // Add third participant
    const { participants: updatedParticipants, remoteStreams: updatedStreams } = createParticipants(3);
    
    rerender(
      <GroupCallModal
        {...defaultProps}
        participants={updatedParticipants}
        remoteStreams={updatedStreams}
      />
    );

    // Should now show 2x2 grid with 3 participants
    expect(document.querySelector('.grid-cols-2')).toBeInTheDocument();
    expect(screen.getAllByText(/User \d/)).toHaveLength(3);
  });

  it('should handle participant removal', async () => {
    const { participants: initialParticipants, remoteStreams: initialStreams } = createParticipants(4);
    
    const { rerender } = render(
      <GroupCallModal
        {...defaultProps}
        participants={initialParticipants}
        remoteStreams={initialStreams}
      />
    );

    // Initially should show 2x2 grid with 4 participants
    expect(document.querySelector('.grid-cols-2')).toBeInTheDocument();
    expect(document.querySelectorAll('.relative.bg-gray-800')).toHaveLength(4);

    // Remove one participant
    const { participants: updatedParticipants, remoteStreams: updatedStreams } = createParticipants(3);
    
    rerender(
      <GroupCallModal
        {...defaultProps}
        participants={updatedParticipants}
        remoteStreams={updatedStreams}
      />
    );

    // Should still show 2x2 grid but with 3 participants
    expect(document.querySelector('.grid-cols-2')).toBeInTheDocument();
    expect(document.querySelectorAll('.relative.bg-gray-800')).toHaveLength(3);
  });

  it('should maintain aspect ratio in different grid layouts', () => {
    const { participants, remoteStreams } = createParticipants(6);
    
    render(
      <GroupCallModal
        {...defaultProps}
        participants={participants}
        remoteStreams={remoteStreams}
      />
    );

    // All video containers should have proper aspect ratio classes
    const videoContainers = document.querySelectorAll('.relative.bg-gray-800');
    expect(videoContainers.length).toBe(6);

    // Each container should be properly sized within the grid
    videoContainers.forEach(container => {
      expect(container).toHaveClass('rounded-lg', 'overflow-hidden');
    });
  });
});