import { create } from 'zustand';
import type { UserProfile } from '../types/models';

// Cache configuration
const CACHE_KEY = 'breakoutglobe_user_profile';
const CACHE_EXPIRY_HOURS = 24;

interface CacheInfo {
  lastUpdated: number;
  isFromCache: boolean;
}

interface PendingChange {
  profile: UserProfile;
  timestamp: number;
  retryCount: number;
}

interface UserProfileState {
  profile: UserProfile | null;
  cacheInfo: CacheInfo;
  pendingChanges: PendingChange[];
  
  // Core profile management
  setProfile: (profile: UserProfile) => void;
  getProfile: () => UserProfile | null;
  clearProfile: () => void;
  
  // localStorage sync
  loadFromLocalStorage: () => void;
  saveToLocalStorage: (profile: UserProfile) => void;
  
  // Backend synchronization
  syncToBackend: (profile: UserProfile, updateFn: (updates: Partial<UserProfile>) => Promise<UserProfile>) => Promise<void>;
  queueForSync: (profile: UserProfile) => void;
  syncWithRetry: (profile: UserProfile, updateFn: (updates: Partial<UserProfile>) => Promise<UserProfile>) => Promise<void>;
  
  // Offline access
  getProfileOffline: () => UserProfile | null;
  isFromCache: () => boolean;
  getCacheInfo: () => CacheInfo;
  getPendingChanges: () => PendingChange[];
}

export const userProfileStore = create<UserProfileState>((set, get) => ({
  profile: null,
  cacheInfo: {
    lastUpdated: 0,
    isFromCache: false,
  },
  pendingChanges: [],

  setProfile: (profile: UserProfile) => {
    console.log('🏪 UserProfileStore: setProfile called with:', {
      id: profile.id,
      displayName: profile.displayName,
      aboutMe: profile.aboutMe,
      aboutMeType: typeof profile.aboutMe,
      avatarURL: profile.avatarURL,
    });
    
    set({
      profile,
      cacheInfo: {
        lastUpdated: Date.now(),
        isFromCache: false,
      },
    });
    
    // Save to localStorage
    get().saveToLocalStorage(profile);
    
    console.log('✅ UserProfileStore: Profile set and saved to localStorage');
  },

  getProfile: () => {
    const state = get();
    if (state.profile) {
      return state.profile;
    }
    
    // Try to load from localStorage if no profile in memory
    state.loadFromLocalStorage();
    return get().profile;
  },

  clearProfile: () => {
    set({
      profile: null,
      cacheInfo: {
        lastUpdated: 0,
        isFromCache: false,
      },
    });
    
    // Remove from localStorage
    try {
      localStorage.removeItem(CACHE_KEY);
    } catch (error) {
      console.warn('Failed to remove profile from localStorage:', error);
    }
  },

  loadFromLocalStorage: () => {
    try {
      const cached = localStorage.getItem(CACHE_KEY);
      if (!cached) return;

      const data = JSON.parse(cached);
      
      // Handle both old format (direct profile) and new format (with timestamp)
      let profile: UserProfile;
      let timestamp: number;
      
      if (data.profile && data.timestamp) {
        // New format with timestamp
        profile = data.profile;
        timestamp = data.timestamp;
      } else {
        // Old format (direct profile)
        profile = data;
        timestamp = Date.now(); // Assume current time for old format
      }

      // Check if cache is expired
      const isExpired = Date.now() - timestamp > (CACHE_EXPIRY_HOURS * 60 * 60 * 1000);
      if (isExpired) {
        localStorage.removeItem(CACHE_KEY);
        return;
      }

      // Transform dates from strings back to Date objects
      const transformedProfile: UserProfile = {
        ...profile,
        createdAt: new Date(profile.createdAt),
        lastActiveAt: profile.lastActiveAt ? new Date(profile.lastActiveAt) : undefined,
      };

      set({
        profile: transformedProfile,
        cacheInfo: {
          lastUpdated: timestamp,
          isFromCache: true,
        },
      });
    } catch (error) {
      console.warn('Failed to load profile from localStorage:', error);
      // Clear corrupted data
      try {
        localStorage.removeItem(CACHE_KEY);
      } catch (clearError) {
        console.warn('Failed to clear corrupted localStorage data:', clearError);
      }
    }
  },

  saveToLocalStorage: (profile: UserProfile) => {
    try {
      const cacheData = {
        profile,
        timestamp: Date.now(),
      };
      localStorage.setItem(CACHE_KEY, JSON.stringify(cacheData));
    } catch (error) {
      console.warn('Failed to save profile to localStorage:', error);
    }
  },

  syncToBackend: async (profile: UserProfile, updateFn: (updates: Partial<UserProfile>) => Promise<UserProfile>) => {
    try {
      const updates = {
        displayName: profile.displayName,
        aboutMe: profile.aboutMe,
      };
      
      const updatedProfile = await updateFn(updates);
      
      // Update local state with backend response
      set({
        profile: updatedProfile,
        cacheInfo: {
          lastUpdated: Date.now(),
          isFromCache: false,
        },
      });
      
      // Save updated profile to localStorage
      get().saveToLocalStorage(updatedProfile);
    } catch (error) {
      console.error('Failed to sync profile to backend:', error);
      throw error;
    }
  },

  queueForSync: (profile: UserProfile) => {
    const pendingChange: PendingChange = {
      profile,
      timestamp: Date.now(),
      retryCount: 0,
    };
    
    set((state) => ({
      pendingChanges: [...state.pendingChanges, pendingChange],
    }));
  },

  syncWithRetry: async (profile: UserProfile, updateFn: (updates: Partial<UserProfile>) => Promise<UserProfile>) => {
    const maxRetries = 3;
    let lastError: Error | null = null;
    
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        await get().syncToBackend(profile, updateFn);
        return; // Success, exit retry loop
      } catch (error) {
        lastError = error as Error;
        
        // Wait before retry (exponential backoff)
        if (attempt < maxRetries - 1) {
          const delay = Math.pow(2, attempt) * 1000; // 1s, 2s, 4s
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }
    
    // All retries failed, queue for later sync
    get().queueForSync(profile);
    throw lastError;
  },

  getProfileOffline: () => {
    const state = get();
    if (state.profile) {
      return state.profile;
    }
    
    // Load from localStorage
    state.loadFromLocalStorage();
    return get().profile;
  },

  isFromCache: () => {
    return get().cacheInfo.isFromCache;
  },

  getCacheInfo: () => {
    return get().cacheInfo;
  },

  getPendingChanges: () => {
    return get().pendingChanges;
  },
}));

// Initialize store by loading from localStorage on startup
userProfileStore.getState().loadFromLocalStorage();