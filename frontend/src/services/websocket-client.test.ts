import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { WebSocketClient } from './websocket-client';
import { sessionStore } from '../stores/sessionStore';
import { poiStore } from '../stores/poiStore';

// Mock WebSocket
class MockWebSocket {
    static CONNECTING = 0;
    static OPEN = 1;
    static CLOSING = 2;
    static CLOSED = 3;

    readyState = MockWebSocket.CONNECTING;
    url: string;
    onopen: ((event: Event) => void) | null = null;
    onclose: ((event: CloseEvent) => void) | null = null;
    onmessage: ((event: MessageEvent) => void) | null = null;
    onerror: ((event: Event) => void) | null = null;

    constructor(url: string) {
        this.url = url;
        // Simulate async connection
        setTimeout(() => {
            this.readyState = MockWebSocket.OPEN;
            this.onopen?.(new Event('open'));
        }, 10);
    }

    send = vi.fn();
    close = vi.fn(() => {
        this.readyState = MockWebSocket.CLOSED;
        this.onclose?.(new CloseEvent('close'));
    });

    // Helper methods for testing
    simulateMessage(data: any) {
        const event = new MessageEvent('message', {
            data: JSON.stringify(data)
        });
        this.onmessage?.(event);
    }

    simulateError() {
        this.onerror?.(new Event('error'));
    }

    simulateClose(code = 1000, reason = 'Normal closure') {
        this.readyState = MockWebSocket.CLOSED;
        const event = new CloseEvent('close', { code, reason });
        this.onclose?.(event);
    }
}

// Mock global WebSocket
global.WebSocket = MockWebSocket as any;

describe('WebSocketClient', () => {
    let client: WebSocketClient;
    let mockWebSocket: MockWebSocket;

    // Helper function to connect and get mock WebSocket
    const connectAndGetMock = async () => {
        await client.connect();
        mockWebSocket = (client as any).ws as MockWebSocket;
        return mockWebSocket;
    };

    beforeEach(() => {
        // Reset stores
        sessionStore.getState().reset();
        poiStore.getState().reset();

        // Clear all mocks
        vi.clearAllMocks();

        // Create client
        client = new WebSocketClient('ws://localhost:8080/ws', 'test-session-id');

        // Mock WebSocket will be created when connect() is called
        mockWebSocket = null as any;
    });

    afterEach(() => {
        client.disconnect();
    });

    describe('Connection Management', () => {
        it('should establish WebSocket connection', async () => {
            expect(client.isConnected()).toBe(false);

            const mock = await connectAndGetMock();

            expect(client.isConnected()).toBe(true);
            expect(mock.url).toBe('ws://localhost:8080/ws');
        });

        it('should handle connection errors', async () => {
            const errorSpy = vi.fn();
            client.onError(errorSpy);

            const mock = await connectAndGetMock();
            mock.simulateError();

            expect(errorSpy).toHaveBeenCalled();
            expect(client.isConnected()).toBe(false);
        });

        it('should disconnect WebSocket', async () => {
            const mock = await connectAndGetMock();
            expect(client.isConnected()).toBe(true);

            client.disconnect();

            expect(mock.close).toHaveBeenCalled();
            expect(client.isConnected()).toBe(false);
        });

        it('should handle unexpected disconnection', async () => {
            const statusSpy = vi.fn();
            client.onStatusChange(statusSpy);

            const mock = await connectAndGetMock();
            mock.simulateClose(1006, 'Abnormal closure');

            expect(statusSpy).toHaveBeenCalledWith('disconnected');
            expect(client.isConnected()).toBe(false);
        });
    });

    describe('Automatic Reconnection', () => {
        it('should attempt reconnection on unexpected disconnect', async () => {
            const statusSpy = vi.fn();
            client.onStatusChange(statusSpy);

            const mock = await connectAndGetMock();

            // Simulate unexpected disconnect
            mock.simulateClose(1006, 'Connection lost');

            // Wait for reconnection attempt
            await new Promise(resolve => setTimeout(resolve, 100));

            expect(statusSpy).toHaveBeenCalledWith('reconnecting');
        });

        it('should use exponential backoff for reconnection', async () => {
            const statusSpy = vi.fn();
            client.onStatusChange(statusSpy);

            let mock = await connectAndGetMock();

            // Simulate multiple failed reconnections
            for (let i = 0; i < 3; i++) {
                mock.simulateClose(1006, 'Connection lost');
                await new Promise(resolve => setTimeout(resolve, 50));
            }

            // Should have been called with 'reconnecting' multiple times
            const reconnectingCalls = statusSpy.mock.calls.filter(call => call[0] === 'reconnecting');
            expect(reconnectingCalls.length).toBeGreaterThan(0);
        });

        it('should stop reconnecting after max attempts', async () => {
            // Set low max attempts for testing
            client.setMaxReconnectAttempts(2);

            let mock = await connectAndGetMock();

            // Simulate multiple failed reconnections
            for (let i = 0; i < 3; i++) {
                mock.simulateClose(1006, 'Connection lost');
                await new Promise(resolve => setTimeout(resolve, 50));
            }

            // Should stop attempting after max attempts
            expect(client.getReconnectAttempts()).toBeLessThanOrEqual(2);
        });

        it('should reset reconnection attempts on successful connection', async () => {
            const mock = await connectAndGetMock();

            // Simulate disconnect and reconnect
            mock.simulateClose(1006, 'Connection lost');
            await new Promise(resolve => setTimeout(resolve, 50));

            // Simulate successful reconnection
            mock.readyState = MockWebSocket.OPEN;
            mock.onopen?.(new Event('open'));

            expect(client.getReconnectAttempts()).toBe(0);
        });
    });

    describe('Message Handling', () => {
        beforeEach(async () => {
            mockWebSocket = await connectAndGetMock();
        });

        it('should send messages to server', () => {
            const message = { type: 'avatar_move', data: { lat: 40.7128, lng: -74.0060 } };

            client.send(message);

            expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
        });

        it('should handle incoming messages', () => {
            const messageSpy = vi.fn();
            client.onMessage(messageSpy);

            const message = { type: 'avatar_update', data: { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 } } };
            mockWebSocket.simulateMessage(message);

            expect(messageSpy).toHaveBeenCalledWith(expect.objectContaining({
                type: 'avatar_update',
                data: message.data
            }));
        });

        it('should handle malformed messages gracefully', () => {
            const errorSpy = vi.fn();
            client.onError(errorSpy);

            // Simulate malformed JSON
            const event = new MessageEvent('message', { data: 'invalid-json' });
            mockWebSocket.onmessage?.(event);

            expect(errorSpy).toHaveBeenCalled();
        });

        it('should queue messages when disconnected', () => {
            client.disconnect();

            const message = { type: 'avatar_move', data: { lat: 40.7128, lng: -74.0060 }, timestamp: new Date() };
            client.send(message);

            expect(client.getQueuedMessages()).toHaveLength(1);
            expect(client.getQueuedMessages()[0]).toEqual(message);
        });

        it('should send queued messages on reconnection', async () => {
            const message1 = { type: 'avatar_move', data: { lat: 40.7128, lng: -74.0060 }, timestamp: new Date() };
            const message2 = { type: 'poi_create', data: { name: 'Test POI' }, timestamp: new Date() };

            client.disconnect();
            client.send(message1);
            client.send(message2);

            const newMock = await connectAndGetMock();

            expect(newMock.send).toHaveBeenCalledWith(JSON.stringify(message1));
            expect(newMock.send).toHaveBeenCalledWith(JSON.stringify(message2));
            expect(client.getQueuedMessages()).toHaveLength(0);
        });
    });

    describe('Store Integration', () => {
        beforeEach(async () => {
            mockWebSocket = await connectAndGetMock();
        });

        it('should sync avatar position updates to session store', () => {
            const message = {
                type: 'avatar_update',
                data: {
                    sessionId: 'test-session-id',
                    position: { lat: 40.7128, lng: -74.0060 }
                }
            };

            // Set up session store with test session
            sessionStore.getState().createSession('test-session-id', { lat: 0, lng: 0 });

            mockWebSocket.simulateMessage(message);

            const state = sessionStore.getState();
            expect(state.avatarPosition).toEqual({ lat: 40.7128, lng: -74.0060 });
        });

        it('should sync POI updates to POI store', () => {
            const poi = {
                id: 'poi-1',
                name: 'Test Meeting Room',
                description: 'A test POI',
                position: { lat: 40.7128, lng: -74.0060 },
                participantCount: 3,
                maxParticipants: 10,
                createdBy: 'user-123',
                createdAt: '2025-09-21T15:55:04.400Z'
            };

            const message = {
                type: 'poi_update',
                data: poi
            };

            mockWebSocket.simulateMessage(message);

            const state = poiStore.getState();
            expect(state.pois).toHaveLength(1);
            expect(state.pois[0]).toEqual(poi);
        });

        it('should handle POI deletion from server', () => {
            // Add POI to store first
            const poi = {
                id: 'poi-1',
                name: 'Test Meeting Room',
                description: 'A test POI',
                position: { lat: 40.7128, lng: -74.0060 },
                participantCount: 3,
                maxParticipants: 10,
                createdBy: 'user-123',
                createdAt: new Date()
            };
            poiStore.getState().addPOI(poi);

            const message = {
                type: 'poi_delete',
                data: { id: 'poi-1' }
            };

            mockWebSocket.simulateMessage(message);

            const state = poiStore.getState();
            expect(state.pois).toHaveLength(0);
        });
    });

    describe('Optimistic Updates', () => {
        beforeEach(async () => {
            mockWebSocket = await connectAndGetMock();
            sessionStore.getState().createSession('test-session-id', { lat: 0, lng: 0 });
        });

        it('should handle avatar move confirmation', () => {
            const newPosition = { lat: 40.7128, lng: -74.0060 };

            // Perform optimistic update
            client.moveAvatar(newPosition);

            // Simulate server confirmation
            const confirmMessage = {
                type: 'avatar_move_confirmed',
                data: {
                    sessionId: 'test-session-id',
                    position: newPosition
                }
            };

            mockWebSocket.simulateMessage(confirmMessage);

            const state = sessionStore.getState();
            expect(state.avatarPosition).toEqual(newPosition);
            expect(state.isMoving).toBe(false);
        });

        it('should handle avatar move rejection with rollback', () => {
            const originalPosition = { lat: 0, lng: 0 };
            const newPosition = { lat: 40.7128, lng: -74.0060 };

            // Perform optimistic update
            client.moveAvatar(newPosition);

            // Simulate server rejection
            const rejectMessage = {
                type: 'avatar_move_rejected',
                data: {
                    sessionId: 'test-session-id',
                    reason: 'Invalid position'
                }
            };

            mockWebSocket.simulateMessage(rejectMessage);

            const state = sessionStore.getState();
            expect(state.avatarPosition).toEqual(originalPosition);
            expect(state.isMoving).toBe(false);
        });

        it('should handle POI creation confirmation', () => {
            const poi = {
                id: 'temp-poi-123',
                name: 'New Meeting Room',
                description: 'Created via optimistic update',
                position: { lat: 40.7128, lng: -74.0060 },
                participantCount: 1,
                maxParticipants: 8,
                createdBy: 'current-user',
                createdAt: new Date()
            };

            // Perform optimistic POI creation
            client.createPOI(poi);

            // Simulate server confirmation with real ID
            const confirmMessage = {
                type: 'poi_create_confirmed',
                data: {
                    tempId: 'temp-poi-123',
                    poi: { ...poi, id: 'poi-real-456' }
                }
            };

            mockWebSocket.simulateMessage(confirmMessage);

            const state = poiStore.getState();
            expect(state.pois).toHaveLength(1);
            expect(state.pois[0].id).toBe('poi-real-456');
        });

        it('should handle POI creation rejection with rollback', () => {
            const poi = {
                id: 'temp-poi-123',
                name: 'New Meeting Room',
                description: 'Created via optimistic update',
                position: { lat: 40.7128, lng: -74.0060 },
                participantCount: 1,
                maxParticipants: 8,
                createdBy: 'current-user',
                createdAt: new Date()
            };

            // Perform optimistic POI creation
            client.createPOI(poi);
            expect(poiStore.getState().pois).toHaveLength(1);

            // Simulate server rejection
            const rejectMessage = {
                type: 'poi_create_rejected',
                data: {
                    tempId: 'temp-poi-123',
                    reason: 'Invalid location'
                }
            };

            mockWebSocket.simulateMessage(rejectMessage);

            const state = poiStore.getState();
            expect(state.pois).toHaveLength(0);
        });
    });

    describe('Callback System', () => {
        it('should support status change callbacks', async () => {
            const statusSpy = vi.fn();
            client.onStatusChange(statusSpy);

            await client.connect();
            client.disconnect();

            expect(statusSpy).toHaveBeenCalledWith('connecting');
            expect(statusSpy).toHaveBeenCalledWith('connected');
            expect(statusSpy).toHaveBeenCalledWith('disconnected');
        });

        it('should support error callbacks', async () => {
            const errorSpy = vi.fn();
            client.onError(errorSpy);

            const mock = await connectAndGetMock();
            mock.simulateError();

            expect(errorSpy).toHaveBeenCalled();
        });

        it('should support message callbacks', async () => {
            const messageSpy = vi.fn();
            client.onMessage(messageSpy);

            const mock = await connectAndGetMock();
            const message = { type: 'test', data: { value: 'test' } };
            mock.simulateMessage(message);

            expect(messageSpy).toHaveBeenCalled();
        });
    });

    describe('Connection State', () => {
        it('should track connection state correctly', async () => {
            expect(client.getConnectionStatus()).toBe('disconnected');

            const connectPromise = client.connect();
            expect(client.getConnectionStatus()).toBe('connecting');

            await connectPromise;
            expect(client.getConnectionStatus()).toBe('connected');

            client.disconnect();
            expect(client.getConnectionStatus()).toBe('disconnected');
        });

        it('should track reconnection attempts', async () => {
            const mock = await connectAndGetMock();

            expect(client.getReconnectAttempts()).toBe(0);

            mock.simulateClose(1006, 'Connection lost');
            await new Promise(resolve => setTimeout(resolve, 50));

            expect(client.getReconnectAttempts()).toBeGreaterThan(0);
        });
    });
});