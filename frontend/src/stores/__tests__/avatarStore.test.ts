import { avatarStore } from '../avatarStore';
import type { AvatarData } from '../../components/MapContainer';

// Mock data for testing
const mockAvatar1: AvatarData = {
  sessionId: 'session-1',
  userId: 'user-1',
  displayName: 'Alice',
  avatarURL: 'https://example.com/alice.jpg',
  position: { lat: 40.7128, lng: -74.0060 },
  isCurrentUser: false,
  isMoving: false,
  role: 'user'
};

const mockAvatar2: AvatarData = {
  sessionId: 'session-2',
  userId: 'user-2',
  displayName: 'Bob',
  position: { lat: 40.7589, lng: -73.9851 },
  isCurrentUser: false,
  isMoving: true,
  role: 'admin'
};

const mockCurrentUser: AvatarData = {
  sessionId: 'current-session',
  userId: 'current-user',
  displayName: 'Current User',
  position: { lat: 40.7500, lng: -73.9900 },
  isCurrentUser: true,
  isMoving: false,
  role: 'user'
};

describe('avatarStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    avatarStore.getState().clearAllAvatars();
  });

  describe('Multi-User Avatar Display', () => {
    it('should add other users avatars to the store', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      store.addOrUpdateAvatar(mockAvatar2);
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars).toHaveLength(2);
      expect(avatars[0]).toEqual(mockAvatar1);
      expect(avatars[1]).toEqual(mockAvatar2);
    });

    it('should update existing avatar when user moves', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      
      const updatedAvatar = {
        ...mockAvatar1,
        position: { lat: 41.0000, lng: -75.0000 },
        isMoving: true
      };
      
      store.addOrUpdateAvatar(updatedAvatar);
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars).toHaveLength(1);
      expect(avatars[0].position).toEqual({ lat: 41.0000, lng: -75.0000 });
      expect(avatars[0].isMoving).toBe(true);
    });

    it('should remove avatar when user leaves', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      store.addOrUpdateAvatar(mockAvatar2);
      
      expect(store.getOtherUsersAvatars()).toHaveLength(2);
      
      store.removeAvatar('session-1');
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars).toHaveLength(1);
      expect(avatars[0].sessionId).toBe('session-2');
    });

    it('should not include current user in other users avatars', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockCurrentUser);
      store.addOrUpdateAvatar(mockAvatar1);
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars).toHaveLength(1);
      expect(avatars[0].sessionId).toBe('session-1');
    });
  });

  describe('Avatar Position Sync', () => {
    it('should handle real-time position updates', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      
      // Simulate WebSocket position update
      store.updateAvatarPosition('session-1', { lat: 42.0000, lng: -76.0000 }, true);
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars[0].position).toEqual({ lat: 42.0000, lng: -76.0000 });
      expect(avatars[0].isMoving).toBe(true);
    });

    it('should handle position updates for non-existent avatars gracefully', () => {
      const store = avatarStore.getState();
      
      // Try to update position for non-existent avatar
      store.updateAvatarPosition('non-existent', { lat: 42.0000, lng: -76.0000 }, false);
      
      // Should not crash and should not add new avatar
      expect(store.getOtherUsersAvatars()).toHaveLength(0);
    });
  });

  describe('Initial User State Load', () => {
    it('should load multiple users from initial state', () => {
      const store = avatarStore.getState();
      
      const initialUsers = [mockAvatar1, mockAvatar2];
      store.loadInitialUsers(initialUsers);
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars).toHaveLength(2);
      expect(avatars).toEqual(initialUsers);
    });

    it('should replace existing avatars when loading initial state', () => {
      const store = avatarStore.getState();
      
      // Add some avatars first
      store.addOrUpdateAvatar(mockAvatar1);
      
      // Load initial state with different avatars
      const initialUsers = [mockAvatar2];
      store.loadInitialUsers(initialUsers);
      
      const avatars = store.getOtherUsersAvatars();
      expect(avatars).toHaveLength(1);
      expect(avatars[0].sessionId).toBe('session-2');
    });
  });

  describe('Map User Isolation', () => {
    it('should filter avatars by current map', () => {
      const store = avatarStore.getState();
      
      const avatar1OnMap1 = { ...mockAvatar1, mapId: 'map-1' };
      const avatar2OnMap2 = { ...mockAvatar2, mapId: 'map-2' };
      
      store.addOrUpdateAvatar(avatar1OnMap1);
      store.addOrUpdateAvatar(avatar2OnMap2);
      
      // Set current map to map-1
      store.setCurrentMap('map-1');
      
      const avatars = store.getAvatarsForCurrentMap();
      expect(avatars).toHaveLength(1);
      expect(avatars[0].sessionId).toBe('session-1');
    });

    it('should return all avatars when no current map is set', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      store.addOrUpdateAvatar(mockAvatar2);
      
      const avatars = store.getAvatarsForCurrentMap();
      expect(avatars).toHaveLength(2);
    });
  });

  describe('Store State Management', () => {
    it('should clear all avatars', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      store.addOrUpdateAvatar(mockAvatar2);
      
      expect(store.getOtherUsersAvatars()).toHaveLength(2);
      
      store.clearAllAvatars();
      
      expect(store.getOtherUsersAvatars()).toHaveLength(0);
    });

    it('should get avatar by session ID', () => {
      const store = avatarStore.getState();
      
      store.addOrUpdateAvatar(mockAvatar1);
      
      const avatar = store.getAvatarBySessionId('session-1');
      expect(avatar).toEqual(mockAvatar1);
      
      const nonExistent = store.getAvatarBySessionId('non-existent');
      expect(nonExistent).toBeUndefined();
    });
  });
});