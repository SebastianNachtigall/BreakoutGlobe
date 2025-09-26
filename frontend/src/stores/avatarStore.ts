import { create } from 'zustand';
import type { AvatarData } from '../components/MapContainer';

interface AvatarState {
  // State
  avatars: Map<string, AvatarData>; // sessionId -> AvatarData
  currentMap: string | null;
  
  // Actions
  addOrUpdateAvatar: (avatar: AvatarData) => void;
  removeAvatar: (sessionId: string) => void;
  updateAvatarPosition: (sessionId: string, position: { lat: number; lng: number }, isMoving: boolean) => void;
  loadInitialUsers: (users: AvatarData[]) => void;
  setCurrentMap: (mapId: string | null) => void;
  clearAllAvatars: () => void;
  
  // Getters
  getOtherUsersAvatars: () => AvatarData[];
  getAvatarsForCurrentMap: () => AvatarData[];
  getAvatarBySessionId: (sessionId: string) => AvatarData | undefined;
}

export const avatarStore = create<AvatarState>((set, get) => ({
  // Initial state
  avatars: new Map(),
  currentMap: null,
  
  // Actions
  addOrUpdateAvatar: (avatar: AvatarData) => {
    set((state) => {
      const newAvatars = new Map(state.avatars);
      
      // Don't add current user to the avatars map
      if (avatar.isCurrentUser) {
        return state;
      }
      
      newAvatars.set(avatar.sessionId, avatar);
      return { avatars: newAvatars };
    });
  },
  
  removeAvatar: (sessionId: string) => {
    set((state) => {
      const newAvatars = new Map(state.avatars);
      newAvatars.delete(sessionId);
      return { avatars: newAvatars };
    });
  },
  
  updateAvatarPosition: (sessionId: string, position: { lat: number; lng: number }, isMoving: boolean) => {
    set((state) => {
      const existingAvatar = state.avatars.get(sessionId);
      if (!existingAvatar) {
        // Avatar doesn't exist, don't create a new one
        return state;
      }
      
      const newAvatars = new Map(state.avatars);
      const updatedAvatar: AvatarData = {
        ...existingAvatar,
        position,
        isMoving
      };
      
      newAvatars.set(sessionId, updatedAvatar);
      return { avatars: newAvatars };
    });
  },
  
  loadInitialUsers: (users: AvatarData[]) => {
    set(() => {
      const newAvatars = new Map<string, AvatarData>();
      
      // Only add other users (not current user)
      users.forEach(user => {
        if (!user.isCurrentUser) {
          newAvatars.set(user.sessionId, user);
        }
      });
      
      return { avatars: newAvatars };
    });
  },
  
  setCurrentMap: (mapId: string | null) => {
    set({ currentMap: mapId });
  },
  
  clearAllAvatars: () => {
    set({ avatars: new Map() });
  },
  
  // Getters
  getOtherUsersAvatars: () => {
    const state = get();
    return Array.from(state.avatars.values());
  },
  
  getAvatarsForCurrentMap: () => {
    const state = get();
    const allAvatars = Array.from(state.avatars.values());
    
    // If no current map is set, return all avatars
    if (!state.currentMap) {
      return allAvatars;
    }
    
    // Filter by current map (if avatars have mapId property)
    return allAvatars.filter(avatar => {
      // For now, we'll assume all avatars are on the same map
      // This can be extended when multi-map support is added
      const avatarMapId = (avatar as any).mapId;
      return !avatarMapId || avatarMapId === state.currentMap;
    });
  },
  
  getAvatarBySessionId: (sessionId: string) => {
    const state = get();
    return state.avatars.get(sessionId);
  }
}));