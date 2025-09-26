export enum ConnectionStatus {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  RECONNECTING = 'reconnecting'
}

export class WebSocketClient {
  private url: string;
  private sessionId: string;

  constructor(url: string, sessionId: string) {
    this.url = url;
    this.sessionId = sessionId;
    console.log('✅ WebSocketClient created successfully');
  }

  async connect(): Promise<void> {
    console.log('🔌 WebSocket: Connecting...');
  }

  isConnected(): boolean {
    return false;
  }

  onStatusChange(callback: (status: ConnectionStatus) => void): void {
    // Mock implementation
  }

  onError(callback: (error: any) => void): void {
    // Mock implementation
  }

  onStateSync(callback: (data: any) => void): void {
    // Mock implementation
  }

  requestInitialUsers(): void {
    // Mock implementation
  }
}