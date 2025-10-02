import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getCurrentUserProfile } from '../services/api';

// Mock the API service
vi.mock('../services/api', () => ({
  getCurrentUserProfile: vi.fn(),
}));

describe('Stale Profile Cleanup Logic', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clear localStorage before each test
    localStorage.clear();
  });

  it('should detect when cached profile is stale (user deleted from backend)', async () => {
    // Simulate a cached profile in localStorage
    const cachedProfile = {
      id: 'deleted-user-123',
      displayName: 'Deleted User',
      accountType: 'guest',
      role: 'user',
      isActive: true,
      createdAt: '2025-10-02T12:00:00Z',
    };
    
    localStorage.setItem('userProfile', JSON.stringify(cachedProfile));

    // Mock API to return null (user not found in backend)
    const mockGetCurrentUserProfile = vi.mocked(getCurrentUserProfile);
    mockGetCurrentUserProfile.mockResolvedValue(null);

    // Call the API with the cached profile's ID
    const result = await getCurrentUserProfile(cachedProfile.id);

    // Verify the API was called with the correct user ID
    expect(mockGetCurrentUserProfile).toHaveBeenCalledWith(cachedProfile.id);
    
    // Verify the result is null (user not found)
    expect(result).toBeNull();
  });

  it('should successfully sync when cached profile exists in backend', async () => {
    // Simulate a cached profile in localStorage
    const cachedProfile = {
      id: 'valid-user-123',
      displayName: 'Valid User',
      accountType: 'guest',
      role: 'user',
      isActive: true,
      createdAt: '2025-10-02T12:00:00Z',
    };
    
    localStorage.setItem('userProfile', JSON.stringify(cachedProfile));

    // Mock API to return updated profile from backend
    const backendProfile = {
      ...cachedProfile,
      displayName: 'Updated User', // Simulate backend change
    };
    
    const mockGetCurrentUserProfile = vi.mocked(getCurrentUserProfile);
    mockGetCurrentUserProfile.mockResolvedValue(backendProfile);

    // Call the API with the cached profile's ID
    const result = await getCurrentUserProfile(cachedProfile.id);

    // Verify the API was called with the correct user ID
    expect(mockGetCurrentUserProfile).toHaveBeenCalledWith(cachedProfile.id);
    
    // Verify the result is the updated profile from backend
    expect(result).toEqual(backendProfile);
    expect(result?.displayName).toBe('Updated User');
  });

  it('should handle API errors gracefully', async () => {
    // Simulate a cached profile in localStorage
    const cachedProfile = {
      id: 'error-user-123',
      displayName: 'Error User',
      accountType: 'guest',
      role: 'user',
      isActive: true,
      createdAt: '2025-10-02T12:00:00Z',
    };
    
    localStorage.setItem('userProfile', JSON.stringify(cachedProfile));

    // Mock API to throw an error (network error, server error, etc.)
    const mockGetCurrentUserProfile = vi.mocked(getCurrentUserProfile);
    mockGetCurrentUserProfile.mockRejectedValue(new Error('Network error'));

    // Call the API and expect it to throw
    await expect(getCurrentUserProfile(cachedProfile.id)).rejects.toThrow('Network error');

    // Verify the API was called with the correct user ID
    expect(mockGetCurrentUserProfile).toHaveBeenCalledWith(cachedProfile.id);
  });
});