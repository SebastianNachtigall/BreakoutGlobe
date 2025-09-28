import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { videoCallStore } from '../stores/videoCallStore';
import { avatarStore } from '../stores/avatarStore';

describe('Participant Display Names', () => {
  beforeEach(() => {
    // Reset stores
    videoCallStore.getState().leavePOICall();
    avatarStore.getState().clearAllAvatars();
  });

  it('should use display name from avatar store when adding participants', () => {
    // Setup avatar with display name
    const avatar = {
      sessionId: 'session-123',
      userId: 'user-123',
      displayName: 'John Doe',
      avatarURL: 'https://example.com/avatar.jpg',
      aboutMe: 'Test user',
      position: { lat: 40.7128, lng: -74.0060 },
      isCurrentUser: false,
      isMoving: false,
      role: 'user' as const
    };

    // Add avatar to store
    avatarStore.getState().addOrUpdateAvatar(avatar);

    // Join POI call
    videoCallStore.getState().joinPOICall('poi-123');

    // Add participant using the same userId
    videoCallStore.getState().addGroupCallParticipant('user-123', {
      displayName: 'John Doe',
      avatarUrl: 'https://example.com/avatar.jpg'
    });

    // Verify participant was added with correct display name
    const participants = videoCallStore.getState().groupCallParticipants;
    expect(participants.has('user-123')).toBe(true);
    expect(participants.get('user-123')?.displayName).toBe('John Doe');
  });

  it('should fallback to user ID when no display name is available', () => {
    // Join POI call
    videoCallStore.getState().joinPOICall('poi-123');

    // Add participant without avatar in store (simulating WebRTC offer scenario)
    videoCallStore.getState().addGroupCallParticipant('user-456', {
      displayName: 'User user-456', // Fallback format
      avatarUrl: null
    });

    // Verify participant was added with fallback name
    const participants = videoCallStore.getState().groupCallParticipants;
    expect(participants.has('user-456')).toBe(true);
    expect(participants.get('user-456')?.displayName).toBe('User user-456');
  });

  it('should get avatar by userId correctly', () => {
    const avatar = {
      sessionId: 'session-789',
      userId: 'user-789',
      displayName: 'Jane Smith',
      avatarURL: null,
      aboutMe: null,
      position: { lat: 40.7128, lng: -74.0060 },
      isCurrentUser: false,
      isMoving: false,
      role: 'user' as const
    };

    // Add avatar to store
    avatarStore.getState().addOrUpdateAvatar(avatar);

    // Get avatar by userId
    const retrievedAvatar = avatarStore.getState().getAvatarByUserId('user-789');
    
    expect(retrievedAvatar).toBeDefined();
    expect(retrievedAvatar?.displayName).toBe('Jane Smith');
    expect(retrievedAvatar?.userId).toBe('user-789');
  });

  it('should return undefined when avatar not found by userId', () => {
    const retrievedAvatar = avatarStore.getState().getAvatarByUserId('non-existent-user');
    expect(retrievedAvatar).toBeUndefined();
  });
});