import { useEffect, useRef, useState, useCallback } from 'react';
import { WebSocketClient, ConnectionStatus, WebSocketMessage, WebSocketError, StateSync } from '../services/websocket-client';

export interface UseWebSocketOptions {
  url: string;
  sessionId: string;
  autoConnect?: boolean;
}

export interface UseWebSocketReturn {
  connectionStatus: ConnectionStatus;
  isConnected: boolean;
  isConnecting: boolean;
  isReconnecting: boolean;
  queuedMessageCount: number;
  lastError: WebSocketError | null;
  connect: () => void;
  disconnect: () => void;
  sendMessage: (message: Omit<WebSocketMessage, 'timestamp'>) => void;
  onMessage: (callback: (message: WebSocketMessage) => void) => void;
  onStateSync: (callback: (data: StateSync) => void) => void;
}

export const useWebSocket = (options: UseWebSocketOptions): UseWebSocketReturn => {
  const { url, sessionId, autoConnect = true } = options;
  
  const clientRef = useRef<WebSocketClient | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>(ConnectionStatus.DISCONNECTED);
  const [queuedMessageCount, setQueuedMessageCount] = useState(0);
  const [lastError, setLastError] = useState<WebSocketError | null>(null);

  // Initialize WebSocket client
  useEffect(() => {
    if (!clientRef.current) {
      clientRef.current = new WebSocketClient(url, sessionId);
      
      // Set up status change listener
      clientRef.current.onStatusChange((status) => {
        setConnectionStatus(status);
      });
      
      // Set up error listener
      clientRef.current.onError((error) => {
        setLastError(error);
      });
    }
    
    return () => {
      if (clientRef.current) {
        clientRef.current.disconnect();
        clientRef.current = null;
      }
    };
  }, [url, sessionId]);

  // Auto-connect if enabled
  useEffect(() => {
    if (autoConnect && clientRef.current && connectionStatus === ConnectionStatus.DISCONNECTED) {
      clientRef.current.connect();
    }
  }, [autoConnect, connectionStatus]);

  // Update queued message count periodically
  useEffect(() => {
    const interval = setInterval(() => {
      if (clientRef.current) {
        setQueuedMessageCount(clientRef.current.getQueuedMessageCount());
      }
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const connect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.connect();
    }
  }, []);

  const disconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.disconnect();
    }
  }, []);

  const sendMessage = useCallback((message: Omit<WebSocketMessage, 'timestamp'>) => {
    if (clientRef.current) {
      clientRef.current.send({
        ...message,
        timestamp: new Date()
      });
    }
  }, []);

  const onMessage = useCallback((callback: (message: WebSocketMessage) => void) => {
    if (clientRef.current) {
      clientRef.current.onMessage(callback);
    }
  }, []);

  const onStateSync = useCallback((callback: (data: StateSync) => void) => {
    if (clientRef.current) {
      clientRef.current.onStateSync(callback);
    }
  }, []);

  return {
    connectionStatus,
    isConnected: connectionStatus === ConnectionStatus.CONNECTED,
    isConnecting: connectionStatus === ConnectionStatus.CONNECTING,
    isReconnecting: connectionStatus === ConnectionStatus.RECONNECTING,
    queuedMessageCount,
    lastError,
    connect,
    disconnect,
    sendMessage,
    onMessage,
    onStateSync
  };
};