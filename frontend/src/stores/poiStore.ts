import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { POIData } from '../components/MapContainer';

export interface POIState {
  // POI data
  pois: POIData[];
  isLoading: boolean;
  error: string | null;
  
  // Optimistic update tracking
  optimisticOperations: Map<string, 'create' | 'join' | 'leave'>;
  
  // Actions
  addPOI: (poi: POIData) => void;
  updatePOI: (id: string, updates: Partial<POIData>) => void;
  removePOI: (id: string) => void;
  getPOIById: (id: string) => POIData | undefined;
  setPOIs: (pois: POIData[]) => void;
  
  // Participant management
  joinPOI: (poiId: string, userId: string) => boolean;
  leavePOI: (poiId: string, userId: string) => boolean;
  
  // Loading and error states
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // Optimistic updates
  createPOIOptimistic: (poi: POIData) => void;
  rollbackPOICreation: (poiId: string) => void;
  joinPOIOptimistic: (poiId: string, userId: string) => boolean;
  rollbackJoinPOI: (poiId: string, userId: string) => void;
  confirmJoinPOI: (poiId: string, userId: string) => void;
  
  // Real-time updates
  handleRealtimeUpdate: (poi: POIData) => void;
  handleRealtimeDelete: (poiId: string) => void;
  
  // Store management
  reset: () => void;
}

const initialState = {
  pois: [],
  isLoading: false,
  error: null,
  optimisticOperations: new Map<string, 'create' | 'join' | 'leave'>(),
};

export const poiStore = create<POIState>()(
  persist(
    (set, get) => ({
      ...initialState,
      
      addPOI: (poi: POIData) => {
        set((state) => ({
          pois: [...state.pois, poi],
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
        
        set((state) => ({
          pois: state.pois.map(p => 
            p.id === poiId 
              ? { ...p, participantCount: p.participantCount + 1 }
              : p
          ),
        }));
        
        return true;
      },
      
      leavePOI: (poiId: string, userId: string) => {
        const state = get();
        const poi = state.pois.find(p => p.id === poiId);
        
        if (!poi || poi.participantCount <= 0) {
          return false;
        }
        
        set((state) => ({
          pois: state.pois.map(p => 
            p.id === poiId 
              ? { ...p, participantCount: Math.max(0, p.participantCount - 1) }
              : p
          ),
        }));
        
        return true;
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
      
      reset: () => {
        set({
          pois: [],
          isLoading: false,
          error: null,
          optimisticOperations: new Map(),
        });
      },
    }),
    {
      name: 'breakout-globe-pois',
      partialize: (state) => ({
        pois: state.pois,
      }),
    }
  )
);