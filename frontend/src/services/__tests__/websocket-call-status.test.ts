import { WebSocketClient } from '../websocket-client';
import { avatarStore } from '../../stores/avatarStore';

// Mock WebSocket
class MockWebSocket {
  onopen: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  
  send = vi.fn();
  close = vi.fn();
  
  // Helper to simulate receiving messages
  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) } as MessageEvent);
    }
  }
  
  // Helper to simulate connection
  simulateOpen() {
    if (this.onopen) {
      this.onopen({} as Event);
    }
  }
}

// Mock global WebSocket
(global as any).WebSocket = MockWebSocket;

describe.skip('WebSocket Call Status Handling', () => {
  let client: WebSocketClient;
  let mockWS: MockWebSocket;

  beforeEach(() => {
    // Reset avatar store
    avatarStore.getState().clearAllAvatars();
    
    // Create client
    client = new WebSocketClient('ws://test', 'test-session');
    
    // Get the mock WebSocket instance
    mockWS = (client as any).ws as MockWebSocket;
  });

  afterEach(() => {
    client.disconnect();
  });

  it('should handle user_call_status messages and update avatar store', async () => {
    // Add a test avatar to the store
    avatarStore.getState().addOrUpdateAvatar({
      sessionId: 'session-123',
      userId: 'user-123',
      displayName: 'Test User',
      position: { lat: 0, lng: 0 },
      isCurrentUser: false,
      isInCall: false
    });

    // Connect the client
    await client.connect();
    mockWS.simulateOpen();

    // Simulate receiving a call status message
    mockWS.simulateMessage({
      type: 'user_call_status',
      data: {
        userId: 'user-123',
        isInCall: true
      },
      timestamp: new Date()
    });

    // Wait for the message to be processed
    await new Promise(resolve => setTimeout(resolve, 100));

    // Check that the avatar store was updated
    const avatar = avatarStore.getState().getAvatarBySessionId('session-123');
    expect(avatar?.isInCall).toBe(true);
  });

  it('should handle call status false and update avatar store', async () => {
    // Add a test avatar to the store (initially in call)
    avatarStore.getState().addOrUpdateAvatar({
      sessionId: 'session-123',
      userId: 'user-123',
      displayName: 'Test User',
      position: { lat: 0, lng: 0 },
      isCurrentUser: false,
      isInCall: true
    });

    // Connect the client
    await client.connect();
    mockWS.simulateOpen();

    // Simulate receiving a call status message (call ended)
    mockWS.simulateMessage({
      type: 'user_call_status',
      data: {
        userId: 'user-123',
        isInCall: false
      },
      timestamp: new Date()
    });

    // Wait for the message to be processed
    await new Promise(resolve => setTimeout(resolve, 100));

    // Check that the avatar store was updated
    const avatar = avatarStore.getState().getAvatarBySessionId('session-123');
    expect(avatar?.isInCall).toBe(false);
  });

  it('should handle unknown user gracefully', async () => {
    // Connect the client
    await client.connect();
    mockWS.simulateOpen();

    // Simulate receiving a call status message for unknown user
    mockWS.simulateMessage({
      type: 'user_call_status',
      data: {
        userId: 'unknown-user',
        isInCall: true
      },
      timestamp: new Date()
    });

    // Wait for the message to be processed
    await new Promise(resolve => setTimeout(resolve, 100));

    // Should not crash and store should remain empty
    expect(avatarStore.getState().getOtherUsersAvatars()).toHaveLength(0);
  });
});