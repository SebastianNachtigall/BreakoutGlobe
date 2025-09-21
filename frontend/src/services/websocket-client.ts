export enum ConnectionStatus {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  RECONNECTING = 'reconnecting'
}

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: Date | string;
}

export interface WebSocketError {
  type: 'connection_error' | 'parse_error' | 'send_error';
  message: string;
  originalError?: Error;
}

export interface StateSync {
  avatars?: Array<{ sessionId: string; position: { lat: number; lng: number } }>;
  pois?: Array<{ id: string; name: string; position: { lat: number; lng: number } }>;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private status: ConnectionStatus = ConnectionStatus.DISCONNECTED;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private baseReconnectDelay = 1000; // 1 second
  private maxReconnectDelay = 30000; // 30 seconds
  private reconnectTimer: number | null = null;
  private messageQueue: WebSocketMessage[] = [];
  private maxQueueSize = 100;
  private shouldReconnect = false;

  // Event callbacks
  private statusCallbacks: ((status: ConnectionStatus) => void)[] = [];
  private messageCallbacks: ((message: WebSocketMessage) => void)[] = [];
  private errorCallbacks: ((error: WebSocketError) => void)[] = [];
  private stateSyncCallbacks: ((data: StateSync) => void)[] = [];

  constructor(
    private url: string,
    private sessionId: string
  ) {}

  // Connection management
  connect(): void {
    if (this.status === ConnectionStatus.CONNECTED || this.status === ConnectionStatus.CONNECTING) {
      return;
    }

    this.shouldReconnect = true;
    this.setStatus(ConnectionStatus.CONNECTING);
    this.createWebSocket();
  }

  disconnect(): void {
    this.shouldReconnect = false;
    this.clearReconnectTimer();
    
    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect');
      this.ws = null;
    }
    
    this.setStatus(ConnectionStatus.DISCONNECTED);
    this.messageQueue = [];
  }

  private createWebSocket(): void {
    try {
      const wsUrl = `${this.url}?sessionId=${this.sessionId}`;
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = this.handleOpen.bind(this);
      this.ws.onclose = this.handleClose.bind(this);
      this.ws.onerror = this.handleError.bind(this);
      this.ws.onmessage = this.handleMessage.bind(this);
    } catch (error) {
      this.handleConnectionError(error as Error);
    }
  }

  private handleOpen(): void {
    const wasReconnecting = this.status === ConnectionStatus.RECONNECTING;
    this.setStatus(ConnectionStatus.CONNECTED);
    this.reconnectAttempts = 0;
    this.sendQueuedMessages();
    
    // Request state synchronization after reconnection
    if (wasReconnecting) {
      this.requestStateSync();
    }
  }

  private handleClose(event: CloseEvent): void {
    this.ws = null;
    
    if (!this.shouldReconnect) {
      this.setStatus(ConnectionStatus.DISCONNECTED);
      return;
    }

    // Don't reconnect for normal closure or manual disconnect
    if (event.code === 1000) {
      this.setStatus(ConnectionStatus.DISCONNECTED);
      return;
    }

    this.setStatus(ConnectionStatus.RECONNECTING);
    this.scheduleReconnect();
  }

  private handleError(event: Event): void {
    const error: WebSocketError = {
      type: 'connection_error',
      message: 'WebSocket error occurred'
    };
    
    this.notifyError(error);
    
    if (this.status === ConnectionStatus.CONNECTING) {
      this.setStatus(ConnectionStatus.DISCONNECTED);
      if (this.shouldReconnect) {
        this.setStatus(ConnectionStatus.RECONNECTING);
        this.scheduleReconnect();
      }
    }
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      
      // Handle special message types
      if (message.type === 'sync_response') {
        this.notifyStateSync(message.data);
      } else {
        this.notifyMessage(message);
      }
    } catch (error) {
      const parseError: WebSocketError = {
        type: 'parse_error',
        message: 'Failed to parse message: ' + (error as Error).message,
        originalError: error as Error
      };
      this.notifyError(parseError);
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.setStatus(ConnectionStatus.DISCONNECTED);
      this.shouldReconnect = false;
      return;
    }

    const delay = Math.min(
      this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts),
      this.maxReconnectDelay
    );

    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectAttempts++;
      this.setStatus(ConnectionStatus.CONNECTING);
      this.createWebSocket();
    }, delay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // Message handling
  send(message: WebSocketMessage): void {
    if (this.status === ConnectionStatus.CONNECTED && this.ws) {
      try {
        this.ws.send(JSON.stringify(message));
      } catch (error) {
        const sendError: WebSocketError = {
          type: 'send_error',
          message: 'Failed to send message: ' + (error as Error).message,
          originalError: error as Error
        };
        this.notifyError(sendError);
        this.queueMessage(message);
      }
    } else {
      this.queueMessage(message);
    }
  }

  private queueMessage(message: WebSocketMessage): void {
    if (this.messageQueue.length >= this.maxQueueSize) {
      // Remove oldest message to make room
      this.messageQueue.shift();
    }
    this.messageQueue.push(message);
  }

  private sendQueuedMessages(): void {
    while (this.messageQueue.length > 0 && this.status === ConnectionStatus.CONNECTED) {
      const message = this.messageQueue.shift();
      if (message) {
        this.send(message);
      }
    }
  }

  private requestStateSync(): void {
    const syncRequest: WebSocketMessage = {
      type: 'sync_request',
      data: {},
      timestamp: new Date().toISOString()
    };
    this.send(syncRequest);
  }

  // Status management
  private setStatus(status: ConnectionStatus): void {
    if (this.status !== status) {
      this.status = status;
      this.notifyStatusChange(status);
    }
  }

  getConnectionStatus(): ConnectionStatus {
    return this.status;
  }

  isConnected(): boolean {
    return this.status === ConnectionStatus.CONNECTED;
  }

  isConnecting(): boolean {
    return this.status === ConnectionStatus.CONNECTING;
  }

  isReconnecting(): boolean {
    return this.status === ConnectionStatus.RECONNECTING;
  }

  getQueuedMessageCount(): number {
    return this.messageQueue.length;
  }

  // Event handling
  onStatusChange(callback: (status: ConnectionStatus) => void): void {
    this.statusCallbacks.push(callback);
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

  // Remove event listeners
  offStatusChange(callback: (status: ConnectionStatus) => void): void {
    const index = this.statusCallbacks.indexOf(callback);
    if (index > -1) {
      this.statusCallbacks.splice(index, 1);
    }
  }

  offMessage(callback: (message: WebSocketMessage) => void): void {
    const index = this.messageCallbacks.indexOf(callback);
    if (index > -1) {
      this.messageCallbacks.splice(index, 1);
    }
  }

  offError(callback: (error: WebSocketError) => void): void {
    const index = this.errorCallbacks.indexOf(callback);
    if (index > -1) {
      this.errorCallbacks.splice(index, 1);
    }
  }

  offStateSync(callback: (data: StateSync) => void): void {
    const index = this.stateSyncCallbacks.indexOf(callback);
    if (index > -1) {
      this.stateSyncCallbacks.splice(index, 1);
    }
  }

  // Notification methods
  private notifyStatusChange(status: ConnectionStatus): void {
    this.statusCallbacks.forEach(callback => {
      try {
        callback(status);
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

  private handleConnectionError(error: Error): void {
    const connectionError: WebSocketError = {
      type: 'connection_error',
      message: 'Failed to create WebSocket connection: ' + error.message,
      originalError: error
    };
    this.notifyError(connectionError);
    this.setStatus(ConnectionStatus.DISCONNECTED);
  }
}