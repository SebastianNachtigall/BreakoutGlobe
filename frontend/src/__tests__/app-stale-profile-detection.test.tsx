import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { userProfileStore } from '../stores/userProfileStore';
import { sessionStore } from '../stores/sessionStore';
import { getCurrentUserProfile } from '../services/api';

// Mock the API service
vi.mock('../services/api', () => ({
  getCurrentUserProfile: vi.fn(),
}));

describe('App Stale Profile Detection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clear stores before each test
    userProfileStore.getState().clearProfile();
    sessionStore.getState().reset();
    // Clear localStorage after clearing stores
    localStorage.clear();
  });

  afterEach(() => {
    // Clean up after each test
    localStorage.clear();
    userProfileStore.getState().clearProfile();
    sessionStore.getState().reset();
  });

  it('should clear stale profile when backend returns null', async () => {
    // Setup: Create a cached profile
    const cachedProfile = {
      id: 'stale-user-123',
      displayName: 'Stale User',
      accountType: 'guest' as const,
      role: 'user' as const,
      isActive: true,
      createdAt: '2025-10-02T12:00:00Z',
    };

    // Set profile in store (this automatically saves to localStorage)
    userProfileStore.getState().setProfile(cachedProfile);
    localStorage.setItem('sessionId', 'fake-session-123');

    // Verify initial state
    expect(userProfileStore.getState().getProfileOffline()).toEqual(cachedProfile);
    expect(localStorage.getItem('sessionId')).toBe('fake-session-123');

    // Mock API to return null (user not found - was deleted)
    const mockGetCurrentUserProfile = vi.mocked(getCurrentUserProfile);
    mockGetCurrentUserProfile.mockResolvedValue(null);

    // Simulate the App.tsx sync logic
    const profile = userProfileStore.getState().getProfileOffline();
    if (profile) {
      try {
        const backendProfile = await getCurrentUserProfile(profile.id);
        if (backendProfile && backendProfile.id === profile.id) {
          // Profile exists and matches - would update
          userProfileStore.getState().setProfile(backendProfile);
        } else if (backendProfile === null) {
          // Backend returned null - user was deleted (e.g., via "nuke users")
          console.warn('⚠️ Cached profile not found in backend - clearing stale data');
          // Clear stale localStorage data
          localStorage.removeItem('userProfile');
          localStorage.removeItem('sessionId');
          userProfileStore.getState().clearProfile();
          sessionStore.getState().reset();
        }
      } catch (syncError) {
        // Would continue with cached profile
      }
    }

    // Verify cleanup happened
    expect(mockGetCurrentUserProfile).toHaveBeenCalledWith(cachedProfile.id);
    expect(userProfileStore.getState().getProfileOffline()).toBeNull();
    expect(localStorage.getItem('userProfile')).toBeNull();
    expect(localStorage.getItem('sessionId')).toBeNull();
  });

  it('should keep profile when backend sync succeeds', async () => {
    // Setup: Create a cached profile
    const cachedProfile = {
      id: 'valid-user-123',
      displayName: 'Valid User',
      accountType: 'guest' as const,
      role: 'user' as const,
      isActive: true,
      createdAt: '2025-10-02T12:00:00Z',
    };

    // Set profile in store
    userProfileStore.getState().setProfile(cachedProfile);

    // Mock API to return updated profile from backend
    const backendProfile = {
      ...cachedProfile,
      displayName: 'Updated Valid User',
    };
    
    const mockGetCurrentUserProfile = vi.mocked(getCurrentUserProfile);
    mockGetCurrentUserProfile.mockResolvedValue(backendProfile);

    // Simulate the App.tsx sync logic
    const profile = userProfileStore.getState().getProfileOffline();
    if (profile) {
      try {
        const backendProfile = await getCurrentUserProfile(profile.id);
        if (backendProfile && backendProfile.id === profile.id) {
          // Profile exists and matches - update with backend data
          userProfileStore.getState().setProfile(backendProfile);
        } else if (backendProfile === null) {
          // Would clear stale data
          localStorage.removeItem('userProfile');
          localStorage.removeItem('sessionId');
          userProfileStore.getState().clearProfile();
          sessionStore.getState().reset();
        }
      } catch (syncError) {
        // Would continue with cached profile
      }
    }

    // Verify profile was updated, not cleared
    expect(mockGetCurrentUserProfile).toHaveBeenCalledWith(cachedProfile.id);
    expect(userProfileStore.getState().getProfileOffline()?.displayName).toBe('Updated Valid User');
  });

  it('should keep cached profile when backend sync fails with error', async () => {
    // Setup: Create a cached profile
    const cachedProfile = {
      id: 'error-user-123',
      displayName: 'Error User',
      accountType: 'guest' as const,
      role: 'user' as const,
      isActive: true,
      createdAt: '2025-10-02T12:00:00Z',
    };

    // Set profile in store
    userProfileStore.getState().setProfile(cachedProfile);

    // Mock API to throw an error (network error, server error, etc.)
    const mockGetCurrentUserProfile = vi.mocked(getCurrentUserProfile);
    mockGetCurrentUserProfile.mockRejectedValue(new Error('Network error'));

    // Simulate the App.tsx sync logic
    const profile = userProfileStore.getState().getProfileOffline();
    if (profile) {
      try {
        const backendProfile = await getCurrentUserProfile(profile.id);
        if (backendProfile && backendProfile.id === profile.id) {
          // Would update profile
          userProfileStore.getState().setProfile(backendProfile);
        } else if (backendProfile === null) {
          // Would clear stale data
          localStorage.removeItem('userProfile');
          localStorage.removeItem('sessionId');
          userProfileStore.getState().clearProfile();
          sessionStore.getState().reset();
        }
      } catch (syncError) {
        // Continue with cached profile on error
        console.info('Backend sync failed, using cached profile');
      }
    }

    // Verify profile was kept (not cleared) on error
    expect(mockGetCurrentUserProfile).toHaveBeenCalledWith(cachedProfile.id);
    expect(userProfileStore.getState().getProfileOffline()).toEqual(cachedProfile);
  });
});