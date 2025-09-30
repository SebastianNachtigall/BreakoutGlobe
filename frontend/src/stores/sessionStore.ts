import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface Position {
  lat: number;
  lng: number;
}

export interface SessionState {
  // Session data
  sessionId: string | null;
  isConnected: boolean;
  avatarPosition: Position;
  isMoving: boolean;
  lastHeartbeat: Date | null;
  
  // Previous position for rollback
  previousPosition: Position | null;
  
  // Actions
  createSession: (sessionId: string, initialPosition: Position) => void;
  updateAvatarPosition: (position: Position, optimistic?: boolean) => void;
  setMoving: (moving: boolean) => void;
  updateHeartbeat: () => void;
  disconnect: () => void;
  reset: () => void;
  
  // Optimistic update actions
  rollbackAvatarPosition: () => void;
  confirmAvatarPosition: (position: Position) => void;
}

const initialState = {
  sessionId: null,
  isConnected: false,
  avatarPosition: { lat: 52.5200, lng: 13.4050 }, // Berlin, Germany
  isMoving: false,
  lastHeartbeat: null,
  previousPosition: null,
};

export const sessionStore = create<SessionState>()(
  persist(
    (set, get) => ({
      ...initialState,
      
      createSession: (sessionId: string, initialPosition: Position) => {
        set({
          sessionId,
          avatarPosition: initialPosition,
          isConnected: true,
          lastHeartbeat: new Date(),
          previousPosition: null,
        });
      },
      
      updateAvatarPosition: (position: Position, optimistic = false) => {
        const currentState = get();
        
        if (optimistic) {
          // Store current position for potential rollback
          set({
            previousPosition: currentState.avatarPosition,
            avatarPosition: position,
            isMoving: true,
          });
        } else {
          // Regular update
          set({
            avatarPosition: position,
            isMoving: true,
            previousPosition: null,
          });
        }
      },
      
      setMoving: (moving: boolean) => {
        set({ isMoving: moving });
      },
      
      updateHeartbeat: () => {
        set({ lastHeartbeat: new Date() });
      },
      
      disconnect: () => {
        set({
          isConnected: false,
          sessionId: null,
        });
      },
      
      reset: () => {
        set(initialState);
      },
      
      rollbackAvatarPosition: () => {
        const { previousPosition } = get();
        if (previousPosition) {
          set({
            avatarPosition: previousPosition,
            previousPosition: null,
            isMoving: false,
          });
        }
      },
      
      confirmAvatarPosition: (position: Position) => {
        set({
          avatarPosition: position,
          previousPosition: null,
          isMoving: false,
        });
      },
    }),
    {
      name: 'breakout-globe-session',
      partialize: (state) => ({
        sessionId: state.sessionId,
        avatarPosition: state.avatarPosition,
        isConnected: state.isConnected,
        isMoving: state.isMoving,
        lastHeartbeat: state.lastHeartbeat,
      }),
    }
  )
);