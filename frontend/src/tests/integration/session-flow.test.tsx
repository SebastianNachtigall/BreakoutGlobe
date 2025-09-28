import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from '../../App';

// Mock the WebSocket and API calls for now
const mockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  readyState: WebSocket.OPEN
};

const mockFetch = vi.fn();

describe('Session Flow Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock WebSocket constructor
    global.WebSocket = vi.fn(() => mockWebSocket) as any;
    
    // Mock fetch for API calls
    global.fetch = mockFetch;
    
    // Mock localStorage to provide a user profile (bypass profile creation)
    const mockProfile = {
      id: 'test-user-123',
      displayName: 'Test User',
      aboutMe: 'Test user for integration tests',
      avatarUrl: null,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    
    global.localStorage = {
      getItem: vi.fn((key) => {
        if (key === 'breakoutglobe_user_profile') {
          return JSON.stringify({
            profile: mockProfile,
            timestamp: Date.now()
          });
        }
        if (key === 'breakoutglobe-session') {
          return JSON.stringify({
            sessionId: 'test-session-123',
            position: { lat: 40.7128, lng: -74.0060 }
          });
        }
        return null;
      }),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn(),
      length: 0,
      key: vi.fn()
    } as any;
    
    // Mock API responses
    mockFetch.mockImplementation((url, options) => {
      // Mock user profile API
      if (url.includes('/api/users/profile')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockProfile
        });
      }
      
      // Mock session creation
      if (url.includes('/api/sessions')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            sessionId: 'test-session-123',
            position: { lat: 40.7128, lng: -74.0060 }
          })
        });
      }
      
      // Mock POI listing
      if (url.includes('/api/pois')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            pois: [],
            count: 0
          })
        });
      }
      
      // Default response
      return Promise.resolve({
        ok: true,
        json: async () => ({})
      });
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should create user session and place avatar on map load', async () => {
    render(<App />);

    // Should show the map container
    expect(screen.getByTestId('map-container')).toBeInTheDocument();

    // Should attempt to create a session
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/sessions'),
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );
    });

    // Should establish WebSocket connection
    await waitFor(() => {
      expect(global.WebSocket).toHaveBeenCalledWith(
        expect.stringContaining('ws://localhost:8080/ws')
      );
    });

    // Should show connection status as connected
    await waitFor(() => {
      const connectionStatus = screen.getByTestId('connection-indicator');
      expect(connectionStatus).toHaveClass('connected');
    });
  });

  it('should handle avatar movement with real-time synchronization', async () => {
    const user = userEvent.setup();
    render(<App />);

    // Wait for initial session creation
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
    });

    // Simulate clicking on the map to move avatar
    const mapContainer = screen.getByTestId('map-container');
    await user.click(mapContainer);

    // Should send avatar position update via WebSocket
    await waitFor(() => {
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        expect.stringContaining('"type":"avatar_move"')
      );
    });

    // Should also send HTTP request to update position
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringMatching(/\/api\/sessions\/.*\/avatar/),
        expect.objectContaining({
          method: 'PUT',
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          }),
          body: expect.stringContaining('"lat":')
        })
      );
    });
  });

  it('should handle session reconnection after connection loss', async () => {
    render(<App />);

    // Wait for initial connection
    await waitFor(() => {
      expect(global.WebSocket).toHaveBeenCalled();
    });

    // Simulate connection loss
    const onCloseHandler = mockWebSocket.addEventListener.mock.calls
      .find(call => call[0] === 'close')?.[1];
    
    if (onCloseHandler) {
      onCloseHandler({ code: 1006, reason: 'Connection lost' });
    }

    // Should show reconnecting status
    await waitFor(() => {
      const connectionStatus = screen.getByTestId('connection-indicator');
      expect(connectionStatus).toHaveClass('reconnecting');
    });

    // Should attempt to reconnect
    await waitFor(() => {
      expect(global.WebSocket).toHaveBeenCalledTimes(2);
    });
  });

  it('should handle API errors gracefully', async () => {
    // Mock API error response
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    render(<App />);

    // Should show error notification
    await waitFor(() => {
      expect(screen.getByTestId('notification-center')).toBeInTheDocument();
      expect(screen.getByText(/network error/i)).toBeInTheDocument();
    });

    // Should show retry option
    expect(screen.getByText(/retry/i)).toBeInTheDocument();
  });

  it('should maintain session state across page refreshes', async () => {
    // Mock localStorage
    const mockLocalStorage = {
      getItem: vi.fn(),
      setItem: vi.fn(),
      removeItem: vi.fn()
    };
    Object.defineProperty(window, 'localStorage', { value: mockLocalStorage });

    // Mock existing session in localStorage
    mockLocalStorage.getItem.mockReturnValue(JSON.stringify({
      sessionId: 'existing-session-456',
      position: { lat: 51.5074, lng: -0.1278 }
    }));

    render(<App />);

    // Should restore session from localStorage
    expect(mockLocalStorage.getItem).toHaveBeenCalledWith('breakoutglobe-session');

    // Should not create new session if one exists
    await waitFor(() => {
      const sessionCreationCalls = mockFetch.mock.calls.filter(call => 
        call[0].includes('/api/sessions') && call[1]?.method === 'POST'
      );
      expect(sessionCreationCalls).toHaveLength(0);
    });

    // Should connect WebSocket with existing session
    await waitFor(() => {
      expect(global.WebSocket).toHaveBeenCalledWith(
        expect.stringContaining('sessionId=existing-session-456')
      );
    });
  });
});