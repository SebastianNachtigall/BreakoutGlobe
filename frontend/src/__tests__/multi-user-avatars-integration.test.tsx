import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import App from '../App';
import { avatarStore } from '../stores/avatarStore';
import type { AvatarData } from '../components/MapContainer';

// Mock the API calls
global.fetch = vi.fn();

// Mock WebSocket
class MockWebSocket {
  onopen: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  
  send = vi.fn();
  close = vi.fn();
  
  simulateConnection() {
    if (this.onopen) {
      this.onopen({} as Event);
    }
  }
}

(global as any).WebSocket = MockWebSocket;

// Mock MapLibre GL
vi.mock('maplibre-gl', () => ({
  Map: vi.fn(() => ({
    addControl: vi.fn(),
    on: vi.fn(),
    remove: vi.fn(),
  })),
  NavigationControl: vi.fn(),
  ScaleControl: vi.fn(),
  Marker: vi.fn(() => ({
    setLngLat: vi.fn().mockReturnThis(),
    addTo: vi.fn().mockReturnThis(),
    remove: vi.fn(),
    getElement: vi.fn(() => document.createElement('div')),
    getLngLat: vi.fn(() => ({ lng: 0, lat: 0 })),
  })),
}));

describe('Multi-User Avatars Integration', () => {
  beforeEach(() => {
    // Reset avatar store
    avatarStore.getState().clearAllAvatars();
    
    // Mock successful API responses
    (global.fetch as any).mockImplementation((url: string) => {
      if (url.includes('/api/users/profile')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            id: 'user-123',
            displayName: 'Test User',
            accountType: 'guest',
            role: 'user',
            createdAt: new Date().toISOString(),
          }),
        });
      }
      
      if (url.includes('/api/sessions')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            sessionId: 'session-123',
            position: { lat: 40.7128, lng: -74.0060 },
          }),
        });
      }
      
      if (url.includes('/api/pois')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ pois: [] }),
        });
      }
      
      return Promise.reject(new Error('Unknown URL'));
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should include other users avatars in the avatar store', () => {
    // Add some other users to the avatar store
    const otherUser1: AvatarData = {
      sessionId: 'other-session-1',
      userId: 'other-user-1',
      displayName: 'Alice',
      avatarURL: 'https://example.com/alice.jpg',
      position: { lat: 40.7589, lng: -73.9851 },
      isCurrentUser: false,
      isMoving: false,
      role: 'user'
    };
    
    const otherUser2: AvatarData = {
      sessionId: 'other-session-2',
      userId: 'other-user-2',
      displayName: 'Bob',
      position: { lat: 40.7300, lng: -73.9900 },
      isCurrentUser: false,
      isMoving: true,
      role: 'admin'
    };
    
    avatarStore.getState().addOrUpdateAvatar(otherUser1);
    avatarStore.getState().addOrUpdateAvatar(otherUser2);
    
    // Verify that the avatar store has the other users
    const avatars = avatarStore.getState().getOtherUsersAvatars();
    expect(avatars).toHaveLength(2);
    expect(avatars[0].displayName).toBe('Alice');
    expect(avatars[1].displayName).toBe('Bob');
  });

  it('should update avatar positions when receiving WebSocket events', () => {
    // Add a user to the store
    const otherUser: AvatarData = {
      sessionId: 'moving-user',
      userId: 'moving-user-id',
      displayName: 'Moving User',
      position: { lat: 40.0, lng: -74.0 },
      isCurrentUser: false,
      isMoving: false,
      role: 'user'
    };
    
    avatarStore.getState().addOrUpdateAvatar(otherUser);
    
    // Simulate position update
    avatarStore.getState().updateAvatarPosition(
      'moving-user',
      { lat: 41.0, lng: -75.0 },
      true
    );
    
    // Verify position was updated
    const updatedAvatar = avatarStore.getState().getAvatarBySessionId('moving-user');
    expect(updatedAvatar?.position).toEqual({ lat: 41.0, lng: -75.0 });
    expect(updatedAvatar?.isMoving).toBe(true);
  });

  it('should handle user join and leave events', () => {
    // Initially no other users
    expect(avatarStore.getState().getOtherUsersAvatars()).toHaveLength(0);
    
    // Simulate user joining
    const newUser: AvatarData = {
      sessionId: 'new-user',
      userId: 'new-user-id',
      displayName: 'New User',
      position: { lat: 42.0, lng: -76.0 },
      isCurrentUser: false,
      isMoving: false,
      role: 'user'
    };
    
    avatarStore.getState().addOrUpdateAvatar(newUser);
    expect(avatarStore.getState().getOtherUsersAvatars()).toHaveLength(1);
    
    // Simulate user leaving
    avatarStore.getState().removeAvatar('new-user');
    expect(avatarStore.getState().getOtherUsersAvatars()).toHaveLength(0);
  });

  it('should load initial users when joining a map', () => {
    const initialUsers: AvatarData[] = [
      {
        sessionId: 'existing-user-1',
        userId: 'existing-user-1-id',
        displayName: 'Existing User 1',
        position: { lat: 40.0, lng: -74.0 },
        isCurrentUser: false,
        isMoving: false,
        role: 'user'
      },
      {
        sessionId: 'existing-user-2',
        userId: 'existing-user-2-id',
        displayName: 'Existing User 2',
        position: { lat: 41.0, lng: -75.0 },
        isCurrentUser: false,
        isMoving: false,
        role: 'admin'
      }
    ];
    
    avatarStore.getState().loadInitialUsers(initialUsers);
    
    const avatars = avatarStore.getState().getOtherUsersAvatars();
    expect(avatars).toHaveLength(2);
    expect(avatars[0].displayName).toBe('Existing User 1');
    expect(avatars[1].displayName).toBe('Existing User 2');
  });
});