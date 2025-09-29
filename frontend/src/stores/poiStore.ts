import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { POIData } from '../components/MapContainer';

export interface POIState {
  // POI data
  pois: POIData[];
  isLoading: boolean;
  error: string | null;

  // User participation tracking
  currentUserPOI: string | null;

  // Optimistic update tracking
  optimisticOperations: Map<string, 'create' | 'join' | 'leave'>;

  // Actions
  addPOI: (poi: POIData) => void;
  updatePOI: (id: string, updates: Partial<POIData>) => void;
  removePOI: (id: string) => void;
  getPOIById: (id: string) => POIData | undefined;
  setPOIs: (pois: POIData[]) => void;

  // Discussion timer is now handled directly in POIDetailsPanel component

  // Participant management
  joinPOI: (poiId: string, userId: string) => boolean;
  leavePOI: (poiId: string, userId: string) => boolean;

  // Auto-leave functionality
  joinPOIWithAutoLeave: (poiId: string, userId: string) => boolean;
  leaveCurrentPOI: (userId: string) => boolean;
  getCurrentUserPOI: () => string | null;

  // Loading and error states
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;

  // Optimistic updates
  createPOIOptimistic: (poi: POIData) => void;
  rollbackPOICreation: (poiId: string) => void;
  joinPOIOptimistic: (poiId: string, userId: string) => boolean;
  joinPOIOptimisticWithAutoLeave: (poiId: string, userId: string) => boolean;
  rollbackJoinPOI: (poiId: string, userId: string) => void;
  confirmJoinPOI: (poiId: string, userId: string) => void;

  // Real-time updates
  handleRealtimeUpdate: (poi: POIData) => void;
  handleRealtimeDelete: (poiId: string) => void;
  updatePOIParticipantCount: (poiId: string, count: number) => void;
  updatePOIParticipants: (poiId: string, count: number, participants: any[]) => void;

  // Discussion timer methods
  updateDiscussionTimer: (poiId: string, duration: number) => void;
  getDiscussionTimerState: (poiId: string) => { isActive: boolean; duration: number; startTime: Date | null } | null;

  // Store management
  reset: () => void;
}

const initialState = {
  pois: [],
  isLoading: false,
  error: null,
  currentUserPOI: null,
  optimisticOperations: new Map<string, 'create' | 'join' | 'leave'>(),
};

export const poiStore = create<POIState>()(
  persist(
    (set, get) => ({
      ...initialState,

      addPOI: (poi: POIData) => {
        const poiWithTimer = {
          ...poi,
          discussionStartTime: poi.discussionStartTime || null,
          isDiscussionActive: poi.isDiscussionActive || false
        };

        set((state) => ({
          pois: [...state.pois, poiWithTimer],
          error: null,
        }));
      },

      updatePOI: (id: string, updates: Partial<POIData>) => {
        set((state) => ({
          pois: state.pois.map(poi =>
            poi.id === id ? { ...poi, ...updates } : poi
          ),
        }));
      },

      removePOI: (id: string) => {
        set((state) => ({
          pois: state.pois.filter(poi => poi.id !== id),
        }));
      },

      getPOIById: (id: string) => {
        return get().pois.find(poi => poi.id === id);
      },

      setPOIs: (pois: POIData[]) => {
        set({ pois, error: null });
      },

      joinPOI: (poiId: string, userId: string) => {
        const state = get();
        const poi = state.pois.find(p => p.id === poiId);

        if (!poi || poi.participantCount >= poi.maxParticipants) {
          return false;
        }

        const newParticipantCount = poi.participantCount + 1;

        set((state) => ({
          pois: state.pois.map(p =>
            p.id === poiId
              ? {
                ...p,
                participantCount: newParticipantCount
              }
              : p
          ),
          currentUserPOI: poiId,
        }));

        return true;
      },

      leavePOI: (poiId: string, userId: string) => {
        const state = get();
        const poi = state.pois.find(p => p.id === poiId);

        // Always clear currentUserPOI if user is trying to leave this POI
        // This handles edge cases where POI was deleted but user still thinks they're in it
        const shouldClearCurrentPOI = state.currentUserPOI === poiId;

        if (!poi || poi.participantCount <= 0) {
          // POI doesn't exist or has no participants, but still clear currentUserPOI if needed
          if (shouldClearCurrentPOI) {
            set((state) => ({
              ...state,
              currentUserPOI: null,
            }));
          }
          return false;
        }

        const newParticipantCount = Math.max(0, poi.participantCount - 1);

        set((state) => ({
          pois: state.pois.map(p =>
            p.id === poiId
              ? {
                ...p,
                participantCount: newParticipantCount
              }
              : p
          ),
          currentUserPOI: shouldClearCurrentPOI ? null : state.currentUserPOI,
        }));

        return true;
      },

      // Auto-leave functionality
      joinPOIWithAutoLeave: (poiId: string, userId: string) => {
        const state = get();

        // First, leave current POI if user is in one
        if (state.currentUserPOI && state.currentUserPOI !== poiId) {
          get().leavePOI(state.currentUserPOI, userId);
        }

        // Then join the new POI
        const success = get().joinPOI(poiId, userId);
        if (success) {
          set({ currentUserPOI: poiId });
        }

        return success;
      },

      leaveCurrentPOI: (userId: string) => {
        const state = get();

        if (!state.currentUserPOI) {
          return false;
        }

        const success = get().leavePOI(state.currentUserPOI, userId);
        if (success) {
          set({ currentUserPOI: null });
        }

        return success;
      },

      getCurrentUserPOI: () => {
        return get().currentUserPOI;
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      setError: (error: string | null) => {
        set({ error });
      },

      createPOIOptimistic: (poi: POIData) => {
        const state = get();
        set({
          pois: [...state.pois, poi],
          optimisticOperations: new Map(state.optimisticOperations).set(poi.id, 'create'),
          error: null,
        });
      },

      rollbackPOICreation: (poiId: string) => {
        const state = get();
        const newOperations = new Map(state.optimisticOperations);
        newOperations.delete(poiId);

        set({
          pois: state.pois.filter(poi => poi.id !== poiId),
          optimisticOperations: newOperations,
        });
      },

      joinPOIOptimistic: (poiId: string, userId: string) => {
        const state = get();
        const poi = state.pois.find(p => p.id === poiId);

        if (!poi || poi.participantCount >= poi.maxParticipants) {
          return false;
        }

        const newOperations = new Map(state.optimisticOperations);
        newOperations.set(`${poiId}-${userId}`, 'join');

        set({
          pois: state.pois.map(p =>
            p.id === poiId
              ? { ...p, participantCount: p.participantCount + 1 }
              : p
          ),
          optimisticOperations: newOperations,
        });

        return true;
      },

      joinPOIOptimisticWithAutoLeave: (poiId: string, userId: string) => {
        const state = get();

        // First, handle auto-leave from current POI
        if (state.currentUserPOI && state.currentUserPOI !== poiId) {
          // Remove from current POI optimistically
          const currentPOI = state.pois.find(p => p.id === state.currentUserPOI);
          if (currentPOI && currentPOI.participantCount > 0) {
            const newOperations = new Map(state.optimisticOperations);
            newOperations.set(`${state.currentUserPOI}-${userId}`, 'leave');

            set({
              pois: state.pois.map(p =>
                p.id === state.currentUserPOI
                  ? { ...p, participantCount: Math.max(0, p.participantCount - 1) }
                  : p
              ),
              optimisticOperations: newOperations,
            });
          }
        }

        // Then join the new POI
        const poi = get().pois.find(p => p.id === poiId);
        if (!poi || poi.participantCount >= poi.maxParticipants) {
          return false;
        }

        const newState = get();
        const newOperations = new Map(newState.optimisticOperations);
        newOperations.set(`${poiId}-${userId}`, 'join');

        set({
          pois: newState.pois.map(p =>
            p.id === poiId
              ? { ...p, participantCount: p.participantCount + 1 }
              : p
          ),
          currentUserPOI: poiId,
          optimisticOperations: newOperations,
        });

        return true;
      },

      rollbackJoinPOI: (poiId: string, userId: string) => {
        const state = get();
        const newOperations = new Map(state.optimisticOperations);
        newOperations.delete(`${poiId}-${userId}`);

        set({
          pois: state.pois.map(p =>
            p.id === poiId
              ? { ...p, participantCount: Math.max(0, p.participantCount - 1) }
              : p
          ),
          optimisticOperations: newOperations,
        });
      },

      confirmJoinPOI: (poiId: string, userId: string) => {
        const state = get();
        const newOperations = new Map(state.optimisticOperations);
        newOperations.delete(`${poiId}-${userId}`);

        set({
          optimisticOperations: newOperations,
        });
      },

      handleRealtimeUpdate: (poi: POIData) => {
        const state = get();
        const existingPOI = state.pois.find(p => p.id === poi.id);

        if (existingPOI) {
          // Update existing POI
          set({
            pois: state.pois.map(p => p.id === poi.id ? poi : p),
          });
        } else {
          // Add new POI
          set({
            pois: [...state.pois, poi],
          });
        }
      },

      handleRealtimeDelete: (poiId: string) => {
        set((state) => ({
          pois: state.pois.filter(poi => poi.id !== poiId),
        }));
      },

      updatePOIParticipantCount: (poiId: string, count: number) => {
        set((state) => ({
          pois: state.pois.map(poi =>
            poi.id === poiId
              ? { ...poi, participantCount: count }
              : poi
          ),
        }));
      },

      updatePOIParticipants: (poiId: string, count: number, participants: any[]) => {
        set((state) => ({
          pois: state.pois.map(poi =>
            poi.id === poiId
              ? {
                ...poi,
                participantCount: count,
                participants: participants.map(p => ({
                  id: p.id,
                  name: p.name,
                  avatarUrl: p.avatarUrl
                }))
              }
              : poi
          ),
        }));
      },

      // Discussion timer methods
      updateDiscussionTimer: (poiId: string, duration: number) => {
        set((state) => ({
          pois: state.pois.map(poi =>
            poi.id === poiId
              ? { ...poi, discussionDuration: duration }
              : poi
          ),
        }));
      },

      getDiscussionTimerState: (poiId: string) => {
        const poi = get().pois.find(p => p.id === poiId);
        if (!poi) return null;

        return {
          isActive: poi.isDiscussionActive || false,
          duration: poi.discussionDuration || 0,
          startTime: poi.discussionStartTime || null
        };
      },

      reset: () => {
        set({
          pois: [],
          isLoading: false,
          error: null,
          currentUserPOI: null,
          optimisticOperations: new Map(),
        });
      },
    }),
    {
      name: 'breakout-globe-pois',
      partialize: (state) => ({
        pois: state.pois,
        currentUserPOI: state.currentUserPOI, // Persist user's current POI membership
      }),
    }
  )
);