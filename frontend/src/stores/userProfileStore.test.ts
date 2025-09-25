import { describe, it, expect, beforeEach, vi } from 'vitest';
import { userProfileStore } from './userProfileStore';
import type { UserProfile } from '../types/models';

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Test data builders
const createTestProfile = (overrides: Partial<UserProfile> = {}): UserProfile => ({
  id: 'test-user-123',
  displayName: 'Test User',
  email: 'test@example.com',
  avatarURL: 'http://localhost:8080/api/users/avatar/test.png',
  aboutMe: 'Test about me',
  accountType: 'guest',
  role: 'user',
  isActive: true,
  emailVerified: false,
  createdAt: new Date('2024-01-01T00:00:00Z'),
  lastActiveAt: new Date('2024-01-01T12:00:00Z'),
  ...overrides,
});

describe('userProfileStore', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    vi.clearAllMocks();
    // Reset store state
    userProfileStore.getState().clearProfile();
  });

  describe('localStorage sync patterns', () => {
    it('should save profile to localStorage when profile is set', () => {
      // Arrange
      const testProfile = createTestProfile();
      
      // Act
      userProfileStore.getState().setProfile(testProfile);
      
      // Assert - expectLocalStorageSync()
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'breakoutglobe_user_profile',
        expect.stringContaining('"profile":')
      );
      
      // Verify the saved data structure
      const savedData = JSON.parse(localStorageMock.setItem.mock.calls[0][1]);
      expect(savedData).toHaveProperty('profile');
      expect(savedData).toHaveProperty('timestamp');
      expect(savedData.profile.id).toBe(testProfile.id);
      expect(savedData.profile.displayName).toBe(testProfile.displayName);
      // Dates are serialized as strings in localStorage
      expect(savedData.profile.createdAt).toBe(testProfile.createdAt.toISOString());
    });

    it('should load profile from localStorage on store initialization', () => {
      // Arrange
      const testProfile = createTestProfile();
      localStorageMock.getItem.mockReturnValue(JSON.stringify(testProfile));
      
      // Act - Initialize a new store instance
      const { getProfile } = userProfileStore.getState();
      const loadedProfile = getProfile();
      
      // Assert - expectLocalStorageSync()
      expect(localStorageMock.getItem).toHaveBeenCalledWith('breakoutglobe_user_profile');
      expect(loadedProfile).toEqual(testProfile);
    });

    it('should handle corrupted localStorage data gracefully', () => {
      // Arrange
      localStorageMock.getItem.mockReturnValue('invalid-json');
      
      // Act & Assert - should not throw
      expect(() => {
        userProfileStore.getState().loadFromLocalStorage();
      }).not.toThrow();
      
      // Should return null for corrupted data
      expect(userProfileStore.getState().getProfile()).toBeNull();
    });

    it('should remove profile from localStorage when profile is cleared', () => {
      // Arrange
      const testProfile = createTestProfile();
      userProfileStore.getState().setProfile(testProfile);
      
      // Act
      userProfileStore.getState().clearProfile();
      
      // Assert
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('breakoutglobe_user_profile');
    });
  });

  describe('backend synchronization workflows', () => {
    it('should sync profile changes to backend when online', async () => {
      // Arrange
      const testProfile = createTestProfile();
      const mockUpdateProfile = vi.fn().mockResolvedValue(testProfile);
      
      // Act
      await userProfileStore.getState().syncToBackend(testProfile, mockUpdateProfile);
      
      // Assert - expectBackendSynchronization()
      expect(mockUpdateProfile).toHaveBeenCalledWith({
        displayName: testProfile.displayName,
        aboutMe: testProfile.aboutMe,
      });
    });

    it('should queue profile changes for sync when offline', () => {
      // Arrange
      const testProfile = createTestProfile();
      
      // Act
      userProfileStore.getState().queueForSync(testProfile);
      
      // Assert
      const pendingChanges = userProfileStore.getState().getPendingChanges();
      expect(pendingChanges).toHaveLength(1);
      expect(pendingChanges[0]).toHaveProperty('profile');
      expect(pendingChanges[0]).toHaveProperty('timestamp');
      expect(pendingChanges[0]).toHaveProperty('retryCount');
      expect(pendingChanges[0].profile).toEqual(testProfile);
    });

    it('should retry failed sync operations', async () => {
      // Arrange
      const testProfile = createTestProfile();
      const mockUpdateProfile = vi.fn()
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce(testProfile);
      
      // Act
      await userProfileStore.getState().syncWithRetry(testProfile, mockUpdateProfile);
      
      // Assert - expectBackendSynchronization()
      expect(mockUpdateProfile).toHaveBeenCalledTimes(2);
    });
  });

  describe('offline profile access', () => {
    it('should provide profile access when backend is unavailable', () => {
      // Arrange
      const testProfile = createTestProfile();
      localStorageMock.getItem.mockReturnValue(JSON.stringify(testProfile));
      
      // Act
      const profile = userProfileStore.getState().getProfileOffline();
      
      // Assert - expectOfflineProfileAccess()
      expect(profile).toEqual(testProfile);
      expect(localStorageMock.getItem).toHaveBeenCalledWith('breakoutglobe_user_profile');
    });

    it('should indicate when profile is from localStorage cache', () => {
      // Arrange
      const testProfile = createTestProfile();
      localStorageMock.getItem.mockReturnValue(JSON.stringify(testProfile));
      
      // Act
      userProfileStore.getState().loadFromLocalStorage();
      
      // Assert
      expect(userProfileStore.getState().isFromCache()).toBe(true);
    });

    it('should update cache timestamp when profile is accessed', () => {
      // Arrange
      const testProfile = createTestProfile();
      const beforeTime = Date.now();
      
      // Act
      userProfileStore.getState().setProfile(testProfile);
      
      // Assert
      const cacheInfo = userProfileStore.getState().getCacheInfo();
      expect(cacheInfo.lastUpdated).toBeGreaterThanOrEqual(beforeTime);
    });

    it('should expire cached profiles after configured time', () => {
      // Arrange
      const testProfile = createTestProfile();
      const expiredTime = Date.now() - (25 * 60 * 60 * 1000); // 25 hours ago
      localStorageMock.getItem.mockReturnValue(JSON.stringify({
        profile: testProfile,
        timestamp: expiredTime,
      }));
      
      // Act
      const profile = userProfileStore.getState().getProfileOffline();
      
      // Assert
      expect(profile).toBeNull(); // Should be expired
    });
  });

  describe('profile store state management', () => {
    it('should maintain profile state across store operations', () => {
      // Arrange
      const testProfile = createTestProfile();
      
      // Act
      userProfileStore.getState().setProfile(testProfile);
      const retrievedProfile = userProfileStore.getState().getProfile();
      
      // Assert
      expect(retrievedProfile).toEqual(testProfile);
    });

    it('should notify subscribers when profile changes', () => {
      // Arrange
      const testProfile = createTestProfile();
      const mockSubscriber = vi.fn();
      const unsubscribe = userProfileStore.subscribe(mockSubscriber);
      
      // Act
      userProfileStore.getState().setProfile(testProfile);
      
      // Assert
      expect(mockSubscriber).toHaveBeenCalled();
      
      // Cleanup
      unsubscribe();
    });

    it('should handle multiple concurrent profile updates', async () => {
      // Arrange
      const profile1 = createTestProfile({ displayName: 'User 1' });
      const profile2 = createTestProfile({ displayName: 'User 2' });
      
      // Act
      const promises = [
        Promise.resolve(userProfileStore.getState().setProfile(profile1)),
        Promise.resolve(userProfileStore.getState().setProfile(profile2)),
      ];
      await Promise.all(promises);
      
      // Assert - Last update should win
      const finalProfile = userProfileStore.getState().getProfile();
      expect(finalProfile?.displayName).toBe('User 2');
    });
  });
});