import { sessionStore } from '../stores/sessionStore';
import { poiStore } from '../stores/poiStore';
import type { POIData } from '../components/MapContainer';

export enum ConnectionStatus {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  RECONNECTING = 'reconnecting'
}

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: Date;
}

export interface WebSocketError {
  message: string;
  code?: number;
  timestamp: Date;
}

export interface StateSync {
  type: 'avatar' | 'poi' | 'session';
  data: any;
  timestamp: Date;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private sessionId: string;
  private connectionStatus: ConnectionStatus = ConnectionStatus.DISCONNECTED;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private reconnectTimer: number | null = null;
  private messageQueue: WebSocketMessage[] = [];
  private statusChangeCallbacks: ((status: ConnectionStatus) => void)[] = [];
  private messageCallbacks: ((message: WebSocketMessage) => void)[] = [];
  private errorCallbacks: ((error: WebSocketError) => void)[] = [];
  private stateSyncCallbacks: ((data: StateSync) => void)[] = [];

  constructor(url: string, sessionId: string) {
    this.url = url;
    this.sessionId = sessionId;
  }

  // Connection Management
  async connect(): Promise<void> {
    if (this.connectionStatus === ConnectionStatus.CONNECTED || this.connectionStatus === ConnectionStatus.CONNECTING) {
      return;
    }

    this.connectionStatus = ConnectionStatus.CONNECTING;
    this.notifyStatusChange();

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.connectionStatus = ConnectionStatus.CONNECTED;
          this.reconnectAttempts = 0;
          this.reconnectDelay = 1000; // Reset delay
          this.notifyStatusChange();
          this.sendQueuedMessages();
          resolve();
        };

        this.ws.onclose = (event) => {
          this.connectionStatus = ConnectionStatus.DISCONNECTED;
          this.ws = null;
          this.notifyStatusChange();

          // Auto-reconnect on unexpected closure
          if (event.code !== 1000 && event.code !== 1001) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          this.connectionStatus = ConnectionStatus.DISCONNECTED;
          const wsError: WebSocketError = {
            message: 'WebSocket connection error',
            timestamp: new Date()
          };
          this.notifyError(wsError);
          reject(error);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };

      } catch (error) {
        this.connectionStatus = ConnectionStatus.DISCONNECTED;
        const wsError: WebSocketError = {
          message: error instanceof Error ? error.message : 'Unknown connection error',
          timestamp: new Date()
        };
        this.notifyError(wsError);
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Normal closure');
      this.ws = null;
    }

    this.connectionStatus = ConnectionStatus.DISCONNECTED;
    this.notifyStatusChange();
  }

  isConnected(): boolean {
    return this.connectionStatus === ConnectionStatus.CONNECTED;
  }

  isConnecting(): boolean {
    return this.connectionStatus === ConnectionStatus.CONNECTING;
  }

  isReconnecting(): boolean {
    return this.connectionStatus === ConnectionStatus.RECONNECTING;
  }

  getConnectionStatus(): ConnectionStatus {
    return this.connectionStatus;
  }

  getQueuedMessageCount(): number {
    return this.messageQueue.length;
  }

  // Callback registration methods
  onStatusChange(callback: (status: ConnectionStatus) => void): void {
    this.statusChangeCallbacks.push(callback);
  }

  onMessage(callback: (message: WebSocketMessage) => void): void {
    this.messageCallbacks.push(callback);
  }

  onError(callback: (error: WebSocketError) => void): void {
    this.errorCallbacks.push(callback);
  }

  onStateSync(callback: (data: StateSync) => void): void {
    this.stateSyncCallbacks.push(callback);
  }

  // Helper methods for notifications
  private notifyStatusChange(): void {
    this.statusChangeCallbacks.forEach(callback => {
      try {
        callback(this.connectionStatus);
      } catch (error) {
        console.error('Error in status change callback:', error);
      }
    });
  }

  private notifyMessage(message: WebSocketMessage): void {
    this.messageCallbacks.forEach(callback => {
      try {
        callback(message);
      } catch (error) {
        console.error('Error in message callback:', error);
      }
    });
  }

  private notifyError(error: WebSocketError): void {
    this.errorCallbacks.forEach(callback => {
      try {
        callback(error);
      } catch (error) {
        console.error('Error in error callback:', error);
      }
    });
  }

  private notifyStateSync(data: StateSync): void {
    this.stateSyncCallbacks.forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error('Error in state sync callback:', error);
      }
    });
  }

  // Reconnection Logic
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return;
    }

    this.connectionStatus = ConnectionStatus.RECONNECTING;
    this.reconnectAttempts++;
    this.notifyStatusChange();

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1); // Exponential backoff

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch(() => {
        // Connection failed, will trigger another reconnect attempt
      });
    }, delay);
  }

  setMaxReconnectAttempts(max: number): void {
    this.maxReconnectAttempts = max;
  }

  getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }

  // Message Handling
  send(message: WebSocketMessage): void {
    if (this.isConnected() && this.ws) {
      this.ws.send(JSON.stringify(message));
    } else {
      // Queue message for later
      this.messageQueue.push(message);
    }
  }

  private sendQueuedMessages(): void {
    while (this.messageQueue.length > 0 && this.isConnected() && this.ws) {
      const message = this.messageQueue.shift()!;
      this.ws.send(JSON.stringify(message));
    }
  }

  getQueuedMessages(): WebSocketMessage[] {
    return [...this.messageQueue];
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      if (!message.timestamp) {
        message.timestamp = new Date();
      }
      this.notifyMessage(message);
      this.processMessage(message);
    } catch (error) {
      this.notifyError({
        message: 'Failed to parse WebSocket message',
        timestamp: new Date()
      });
    }
  }

  private processMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'avatar_update':
        this.handleAvatarUpdate(message.data);
        break;
      case 'avatar_move_confirmed':
        this.handleAvatarMoveConfirmed(message.data);
        break;
      case 'avatar_move_rejected':
        this.handleAvatarMoveRejected(message.data);
        break;
      case 'poi_update':
        this.handlePOIUpdate(message.data);
        break;
      case 'poi_delete':
        this.handlePOIDelete(message.data);
        break;
      case 'poi_create_confirmed':
        this.handlePOICreateConfirmed(message.data);
        break;
      case 'poi_create_rejected':
        this.handlePOICreateRejected(message.data);
        break;
      default:
        // Unknown message type, ignore or log
        break;
    }
  }

  // Store Integration Handlers
  private handleAvatarUpdate(data: any): void {
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().confirmAvatarPosition(data.position);
      this.notifyStateSync({
        type: 'avatar',
        data: data.position,
        timestamp: new Date()
      });
    }
  }

  private handleAvatarMoveConfirmed(data: any): void {
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().confirmAvatarPosition(data.position);
      this.notifyStateSync({
        type: 'avatar',
        data: data.position,
        timestamp: new Date()
      });
    }
  }

  private handleAvatarMoveRejected(data: any): void {
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().rollbackAvatarPosition();
      this.notifyStateSync({
        type: 'avatar',
        data: { rejected: true, reason: data.reason },
        timestamp: new Date()
      });
    }
  }

  private handlePOIUpdate(data: POIData): void {
    poiStore.getState().handleRealtimeUpdate(data);
    this.notifyStateSync({
      type: 'poi',
      data,
      timestamp: new Date()
    });
  }

  private handlePOIDelete(data: { id: string }): void {
    poiStore.getState().handleRealtimeDelete(data.id);
    this.notifyStateSync({
      type: 'poi',
      data: { action: 'delete', id: data.id },
      timestamp: new Date()
    });
  }

  private handlePOICreateConfirmed(data: { tempId: string; poi: POIData }): void {
    // Remove the temporary POI and add the real one
    poiStore.getState().rollbackPOICreation(data.tempId);
    poiStore.getState().addPOI(data.poi);
    this.notifyStateSync({
      type: 'poi',
      data: { action: 'create_confirmed', poi: data.poi },
      timestamp: new Date()
    });
  }

  private handlePOICreateRejected(data: { tempId: string; reason: string }): void {
    poiStore.getState().rollbackPOICreation(data.tempId);
    this.notifyStateSync({
      type: 'poi',
      data: { action: 'create_rejected', tempId: data.tempId, reason: data.reason },
      timestamp: new Date()
    });
  }

  // Optimistic Update Methods
  moveAvatar(position: { lat: number; lng: number }): void {
    // Perform optimistic update
    sessionStore.getState().updateAvatarPosition(position, true);

    // Send to server
    this.send({
      type: 'avatar_move',
      data: {
        sessionId: this.sessionId,
        position
      },
      timestamp: new Date()
    });

    // Emit state sync
    this.notifyStateSync({
      type: 'avatar',
      data: { position },
      timestamp: new Date()
    });
  }

  createPOI(poi: POIData): void {
    // Perform optimistic update
    poiStore.getState().createPOIOptimistic(poi);

    // Send to server
    this.send({
      type: 'poi_create',
      data: {
        tempId: poi.id,
        poi: {
          ...poi,
          id: undefined // Server will assign real ID
        }
      },
      timestamp: new Date()
    });

    // Emit state sync
    this.notifyStateSync({
      type: 'poi',
      data: poi,
      timestamp: new Date()
    });
  }

  joinPOI(poiId: string): void {
    const userId = this.sessionId;

    // Perform optimistic update
    const success = poiStore.getState().joinPOIOptimistic(poiId, userId);

    if (success) {
      // Send to server
      this.send({
        type: 'poi_join',
        data: { poiId, userId },
        timestamp: new Date()
      });

      // Emit state sync
      this.notifyStateSync({
        type: 'poi',
        data: { action: 'join', poiId, userId },
        timestamp: new Date()
      });
    }
  }

  leavePOI(poiId: string): void {
    const userId = this.sessionId;

    // Perform optimistic update
    const success = poiStore.getState().leavePOI(poiId, userId);

    if (success) {
      // Send to server
      this.send({
        type: 'poi_leave',
        data: { poiId, userId },
        timestamp: new Date()
      });

      // Emit state sync
      this.notifyStateSync({
        type: 'poi',
        data: { action: 'leave', poiId, userId },
        timestamp: new Date()
      });
    }
  }

  // Auto-leave functionality for POI switching
  joinPOIWithAutoLeave(poiId: string): void {
    const userId = this.sessionId;

    // Perform optimistic update with auto-leave
    const success = poiStore.getState().joinPOIOptimisticWithAutoLeave(poiId, userId);

    if (success) {
      // Send to server
      this.send({
        type: 'poi_join',
        data: { poiId, userId },
        timestamp: new Date()
      });

      // Emit state sync
      this.notifyStateSync({
        type: 'poi',
        data: { action: 'join', poiId, userId },
        timestamp: new Date()
      });
    }
  }

  // Leave current POI (for map clicks)
  leaveCurrentPOI(): void {
    const userId = this.sessionId;
    const currentPOI = poiStore.getState().getCurrentUserPOI();

    if (currentPOI) {
      this.leavePOI(currentPOI);
    }
  }


}