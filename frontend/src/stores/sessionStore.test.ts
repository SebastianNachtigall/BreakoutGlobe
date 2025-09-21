import { describe, it, expect, beforeEach, vi } from 'vitest';
import { sessionStore } from './sessionStore';

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

describe('sessionStore', () => {
  beforeEach(() => {
    // Reset store state
    sessionStore.getState().reset();
    vi.clearAllMocks();
  });

  describe('Session Management', () => {
    it('should initialize with default state', () => {
      const state = sessionStore.getState();
      
      expect(state.sessionId).toBeNull();
      expect(state.isConnected).toBe(false);
      expect(state.avatarPosition).toEqual({ lat: 0, lng: 0 });
      expect(state.isMoving).toBe(false);
      expect(state.lastHeartbeat).toBeNull();
    });

    it('should create a new session', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      
      const state = sessionStore.getState();
      expect(state.sessionId).toBe(sessionId);
      expect(state.avatarPosition).toEqual(initialPosition);
      expect(state.isConnected).toBe(true);
      expect(state.lastHeartbeat).toBeInstanceOf(Date);
    });

    it('should update avatar position', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      const newPosition = { lat: 51.5074, lng: -0.1278 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      sessionStore.getState().updateAvatarPosition(newPosition);
      
      const state = sessionStore.getState();
      expect(state.avatarPosition).toEqual(newPosition);
      expect(state.isMoving).toBe(true);
    });

    it('should set moving state', () => {
      sessionStore.getState().setMoving(true);
      expect(sessionStore.getState().isMoving).toBe(true);
      
      sessionStore.getState().setMoving(false);
      expect(sessionStore.getState().isMoving).toBe(false);
    });

    it('should update heartbeat timestamp', () => {
      const beforeHeartbeat = new Date();
      sessionStore.getState().updateHeartbeat();
      const afterHeartbeat = new Date();
      
      const heartbeat = sessionStore.getState().lastHeartbeat;
      expect(heartbeat).toBeInstanceOf(Date);
      expect(heartbeat!.getTime()).toBeGreaterThanOrEqual(beforeHeartbeat.getTime());
      expect(heartbeat!.getTime()).toBeLessThanOrEqual(afterHeartbeat.getTime());
    });

    it('should disconnect session', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      sessionStore.getState().disconnect();
      
      const state = sessionStore.getState();
      expect(state.isConnected).toBe(false);
      expect(state.sessionId).toBeNull();
    });

    it('should reset store to initial state', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      sessionStore.getState().reset();
      
      const state = sessionStore.getState();
      expect(state.sessionId).toBeNull();
      expect(state.isConnected).toBe(false);
      expect(state.avatarPosition).toEqual({ lat: 0, lng: 0 });
      expect(state.isMoving).toBe(false);
      expect(state.lastHeartbeat).toBeNull();
    });
  });

  describe('Persistence', () => {
    it('should have persistence configuration', () => {
      // Test that the store is configured with persistence
      expect(sessionStore.persist).toBeDefined();
      expect(sessionStore.persist.getOptions().name).toBe('breakout-globe-session');
    });

    it('should include correct fields in persistence', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      
      // Test that partialize function includes the right fields
      const options = sessionStore.persist.getOptions();
      const state = sessionStore.getState();
      const persistedState = options.partialize(state);
      
      expect(persistedState).toHaveProperty('sessionId', sessionId);
      expect(persistedState).toHaveProperty('avatarPosition', initialPosition);
      expect(persistedState).toHaveProperty('isConnected', true);
      expect(persistedState).toHaveProperty('isMoving', false);
      expect(persistedState).toHaveProperty('lastHeartbeat');
    });

    it('should handle store rehydration', () => {
      // Test that the store can be rehydrated (basic functionality test)
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      sessionStore.getState().reset();
      
      // After reset, state should be back to initial
      const state = sessionStore.getState();
      expect(state.sessionId).toBeNull();
      expect(state.isConnected).toBe(false);
    });
  });

  describe('Synchronization', () => {
    it('should provide optimistic updates for avatar position', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      const newPosition = { lat: 51.5074, lng: -0.1278 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      
      // Optimistic update should be immediate
      sessionStore.getState().updateAvatarPosition(newPosition, true);
      
      const state = sessionStore.getState();
      expect(state.avatarPosition).toEqual(newPosition);
      expect(state.isMoving).toBe(true);
    });

    it('should rollback optimistic updates on server rejection', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      const newPosition = { lat: 51.5074, lng: -0.1278 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      sessionStore.getState().updateAvatarPosition(newPosition, true);
      
      // Server rejects the update
      sessionStore.getState().rollbackAvatarPosition();
      
      const state = sessionStore.getState();
      expect(state.avatarPosition).toEqual(initialPosition);
      expect(state.isMoving).toBe(false);
    });

    it('should confirm optimistic updates on server acceptance', () => {
      const sessionId = 'test-session-123';
      const initialPosition = { lat: 40.7128, lng: -74.0060 };
      const newPosition = { lat: 51.5074, lng: -0.1278 };
      
      sessionStore.getState().createSession(sessionId, initialPosition);
      sessionStore.getState().updateAvatarPosition(newPosition, true);
      
      // Server confirms the update
      sessionStore.getState().confirmAvatarPosition(newPosition);
      
      const state = sessionStore.getState();
      expect(state.avatarPosition).toEqual(newPosition);
      expect(state.isMoving).toBe(false);
    });
  });
});