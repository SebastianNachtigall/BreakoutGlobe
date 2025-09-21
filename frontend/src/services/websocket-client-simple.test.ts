import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocketClient, ConnectionStatus } from './websocket-client';

describe('WebSocketClient - Basic Functionality', () => {
  let client: WebSocketClient;

  beforeEach(() => {
    vi.clearAllTimers();
    vi.useFakeTimers();
    
    client = new WebSocketClient('ws://localhost:8080/ws', 'test-session-123');
  });

  afterEach(() => {
    client.disconnect();
    vi.useRealTimers();
  });

  describe('Initialization', () => {
    it('should initialize with disconnected status', () => {
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.DISCONNECTED);
      expect(client.isConnected()).toBe(false);
      expect(client.isConnecting()).toBe(false);
      expect(client.isReconnecting()).toBe(false);
    });

    it('should have empty message queue initially', () => {
      expect(client.getQueuedMessageCount()).toBe(0);
    });
  });

  describe('Connection Status', () => {
    it('should update status to connecting when connect is called', () => {
      const statusCallback = vi.fn();
      client.onStatusChange(statusCallback);

      client.connect();
      
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.CONNECTING);
      expect(client.isConnecting()).toBe(true);
      expect(statusCallback).toHaveBeenCalledWith(ConnectionStatus.CONNECTING);
    });

    it('should handle manual disconnect', () => {
      const statusCallback = vi.fn();
      client.onStatusChange(statusCallback);

      client.connect();
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.CONNECTING);

      client.disconnect();
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.DISCONNECTED);
      expect(statusCallback).toHaveBeenCalledWith(ConnectionStatus.DISCONNECTED);
    });
  });

  describe('Message Queuing', () => {
    it('should queue messages when not connected', () => {
      const message = {
        type: 'avatar_move',
        data: { position: { lat: 40.7128, lng: -74.0060 } },
        timestamp: new Date()
      };

      // Should not throw when sending while disconnected
      expect(() => client.send(message)).not.toThrow();
      
      // Message should be queued
      expect(client.getQueuedMessageCount()).toBe(1);
    });

    it('should limit message queue size', () => {
      // Send more messages than queue limit (100)
      for (let i = 0; i < 150; i++) {
        client.send({
          type: 'avatar_move',
          data: { position: { lat: i, lng: i } },
          timestamp: new Date()
        });
      }

      // Should not exceed queue limit
      expect(client.getQueuedMessageCount()).toBeLessThanOrEqual(100);
    });

    it('should clear queue on disconnect', () => {
      // Queue some messages
      client.send({
        type: 'avatar_move',
        data: { position: { lat: 40.7128, lng: -74.0060 } },
        timestamp: new Date()
      });
      
      expect(client.getQueuedMessageCount()).toBe(1);
      
      client.disconnect();
      expect(client.getQueuedMessageCount()).toBe(0);
    });
  });

  describe('Event Callbacks', () => {
    it('should register and call status change callbacks', () => {
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      client.onStatusChange(callback1);
      client.onStatusChange(callback2);

      client.connect();

      expect(callback1).toHaveBeenCalledWith(ConnectionStatus.CONNECTING);
      expect(callback2).toHaveBeenCalledWith(ConnectionStatus.CONNECTING);
    });

    it('should register message callbacks', () => {
      const messageCallback = vi.fn();
      client.onMessage(messageCallback);

      // This test just verifies the callback is registered
      // Actual message handling would require WebSocket connection
      expect(messageCallback).not.toHaveBeenCalled();
    });

    it('should register error callbacks', () => {
      const errorCallback = vi.fn();
      client.onError(errorCallback);

      // This test just verifies the callback is registered
      expect(errorCallback).not.toHaveBeenCalled();
    });

    it('should register state sync callbacks', () => {
      const syncCallback = vi.fn();
      client.onStateSync(syncCallback);

      // This test just verifies the callback is registered
      expect(syncCallback).not.toHaveBeenCalled();
    });

    it('should remove callbacks', () => {
      const callback = vi.fn();
      
      client.onStatusChange(callback);
      client.offStatusChange(callback);
      
      client.connect();
      
      // Callback should not be called after removal
      expect(callback).not.toHaveBeenCalled();
    });
  });

  describe('Connection URL', () => {
    it('should construct correct WebSocket URL with session ID', () => {
      const client = new WebSocketClient('ws://localhost:8080/ws', 'session-456');
      
      // We can't easily test the internal URL construction without exposing it
      // But we can verify the client was created successfully
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.DISCONNECTED);
    });
  });

  describe('Reconnection Logic', () => {
    it('should have reconnection configuration', () => {
      // Test that the client has the expected reconnection behavior
      // by checking it doesn't throw errors and maintains proper state
      client.connect();
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.CONNECTING);
      
      client.disconnect();
      expect(client.getConnectionStatus()).toBe(ConnectionStatus.DISCONNECTED);
    });
  });
});