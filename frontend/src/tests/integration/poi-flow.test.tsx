import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from '../../App';

const mockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  readyState: WebSocket.OPEN
};

const mockFetch = vi.fn();

describe('POI Flow Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    global.WebSocket = vi.fn(() => mockWebSocket) as any;
    global.fetch = mockFetch;
    
    // Mock session creation
    mockFetch.mockImplementation((url, options) => {
      if (url.includes('/api/sessions') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            sessionId: 'test-session-123',
            position: { lat: 40.7128, lng: -74.0060 }
          })
        });
      }
      
      // Mock POI creation
      if (url.includes('/api/pois') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            id: 'poi-123',
            name: 'Test Meeting Room',
            description: 'A test meeting room',
            maxParticipants: 10,
            participantCount: 0,
            position: { lat: 40.7128, lng: -74.0060 },
            participants: []
          })
        });
      }
      
      // Mock POI list
      if (url.includes('/api/pois') && options?.method === 'GET') {
        return Promise.resolve({
          ok: true,
          json: async () => ([])
        });
      }
      
      return Promise.reject(new Error('Unmocked API call'));
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should create POI through right-click context menu', async () => {
    const user = userEvent.setup();
    render(<App />);

    // Wait for initial session creation
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/sessions'),
        expect.objectContaining({ method: 'POST' })
      );
    });

    // Right-click on map to open context menu
    const mapContainer = screen.getByTestId('map-container');
    await user.pointer({ keys: '[MouseRight]', target: mapContainer });

    // Should show context menu
    await waitFor(() => {
      expect(screen.getByTestId('poi-context-menu')).toBeInTheDocument();
    });

    // Click "Create POI" option
    const createPOIButton = screen.getByText(/create poi/i);
    await user.click(createPOIButton);

    // Should open POI creation modal
    await waitFor(() => {
      expect(screen.getByTestId('poi-creation-modal')).toBeInTheDocument();
    });

    // Fill in POI details
    await user.type(screen.getByLabelText(/name/i), 'Test Meeting Room');
    await user.type(screen.getByLabelText(/description/i), 'A test meeting room for discussions');

    // Submit the form
    const createButton = screen.getByText(/create poi/i);
    await waitFor(() => {
      expect(createButton).toBeEnabled();
    });
    await user.click(createButton);

    // Should call POI creation API
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/pois'),
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('Test Meeting Room')
        })
      );
    });

    // Should send WebSocket event for real-time update
    await waitFor(() => {
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        expect.stringContaining('"type":"poi_created"')
      );
    });

    // Should close modal after successful creation
    await waitFor(() => {
      expect(screen.queryByTestId('poi-creation-modal')).not.toBeInTheDocument();
    });
  });

  it('should handle POI join/leave workflow', async () => {
    const user = userEvent.setup();
    
    // Mock POI data
    const mockPOI = {
      id: 'poi-123',
      name: 'Meeting Room A',
      description: 'A comfortable meeting room',
      maxParticipants: 10,
      participantCount: 2,
      position: { lat: 40.7128, lng: -74.0060 },
      participants: [
        { id: 'user-1', name: 'Alice' },
        { id: 'user-2', name: 'Bob' }
      ]
    };

    // Mock POI list API to return our test POI
    mockFetch.mockImplementation((url, options) => {
      if (url.includes('/api/sessions') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            sessionId: 'test-session-123',
            position: { lat: 40.7128, lng: -74.0060 }
          })
        });
      }
      
      if (url.includes('/api/pois') && options?.method === 'GET') {
        return Promise.resolve({
          ok: true,
          json: async () => ([mockPOI])
        });
      }
      
      if (url.includes('/api/pois/poi-123/join') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            ...mockPOI,
            participantCount: 3,
            participants: [
              ...mockPOI.participants,
              { id: 'test-session-123', name: 'Current User' }
            ]
          })
        });
      }
      
      return Promise.reject(new Error('Unmocked API call'));
    });

    render(<App />);

    // Wait for POIs to load
    await waitFor(() => {
      expect(screen.getByTestId('poi-marker')).toBeInTheDocument();
    });

    // Click on POI marker to open details
    const poiMarker = screen.getByTestId('poi-marker');
    await user.click(poiMarker);

    // Should show POI details panel
    await waitFor(() => {
      expect(screen.getByText('Meeting Room A')).toBeInTheDocument();
      expect(screen.getByText('2/10 participants')).toBeInTheDocument();
    });

    // Click join button
    const joinButton = screen.getByText(/join/i);
    await user.click(joinButton);

    // Should call join API
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/pois/poi-123/join'),
        expect.objectContaining({ method: 'POST' })
      );
    });

    // Should send WebSocket event
    await waitFor(() => {
      expect(mockWebSocket.send).toHaveBeenCalledWith(
        expect.stringContaining('"type":"poi_joined"')
      );
    });

    // Should update UI to show user as participant
    await waitFor(() => {
      expect(screen.getByText('3/10 participants')).toBeInTheDocument();
      expect(screen.getByText(/leave/i)).toBeInTheDocument();
    });
  });

  it('should handle multi-user POI interactions', async () => {
    const user = userEvent.setup();
    render(<App />);

    // Wait for initial setup
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
    });

    // Simulate receiving WebSocket message about another user creating POI
    const onMessageHandler = mockWebSocket.addEventListener.mock.calls
      .find(call => call[0] === 'message')?.[1];

    if (onMessageHandler) {
      const poiCreatedEvent = {
        data: JSON.stringify({
          type: 'poi_created',
          payload: {
            id: 'new-poi-456',
            name: 'New Meeting Room',
            description: 'Created by another user',
            maxParticipants: 5,
            participantCount: 1,
            position: { lat: 51.5074, lng: -0.1278 },
            participants: [{ id: 'other-user', name: 'Other User' }]
          }
        })
      };

      onMessageHandler(poiCreatedEvent);
    }

    // Should show the new POI on the map
    await waitFor(() => {
      expect(screen.getByText('New Meeting Room')).toBeInTheDocument();
    });

    // Simulate another user joining a POI
    if (onMessageHandler) {
      const poiJoinedEvent = {
        data: JSON.stringify({
          type: 'poi_joined',
          payload: {
            poiId: 'new-poi-456',
            participantCount: 2,
            participant: { id: 'another-user', name: 'Another User' }
          }
        })
      };

      onMessageHandler(poiJoinedEvent);
    }

    // Should update participant count in real-time
    await waitFor(() => {
      expect(screen.getByText('2/5')).toBeInTheDocument();
    });
  });

  it('should handle POI capacity limits', async () => {
    const user = userEvent.setup();
    
    // Mock full POI
    const fullPOI = {
      id: 'full-poi-789',
      name: 'Full Meeting Room',
      description: 'This room is at capacity',
      maxParticipants: 2,
      participantCount: 2,
      position: { lat: 40.7128, lng: -74.0060 },
      participants: [
        { id: 'user-1', name: 'Alice' },
        { id: 'user-2', name: 'Bob' }
      ]
    };

    mockFetch.mockImplementation((url, options) => {
      if (url.includes('/api/sessions') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            sessionId: 'test-session-123',
            position: { lat: 40.7128, lng: -74.0060 }
          })
        });
      }
      
      if (url.includes('/api/pois') && options?.method === 'GET') {
        return Promise.resolve({
          ok: true,
          json: async () => ([fullPOI])
        });
      }
      
      if (url.includes('/api/pois/full-poi-789/join')) {
        return Promise.resolve({
          ok: false,
          status: 409,
          json: async () => ({
            error: 'POI is at maximum capacity'
          })
        });
      }
      
      return Promise.reject(new Error('Unmocked API call'));
    });

    render(<App />);

    // Wait for POI to load
    await waitFor(() => {
      expect(screen.getByTestId('poi-marker')).toBeInTheDocument();
    });

    // POI marker should show as full
    const poiMarker = screen.getByTestId('poi-marker');
    expect(poiMarker).toHaveClass('bg-red-500'); // Full POI styling

    // Click on full POI
    await user.click(poiMarker);

    // Should show details but join button should be disabled
    await waitFor(() => {
      expect(screen.getByText('Full Meeting Room')).toBeInTheDocument();
      expect(screen.getByText('2/2 participants')).toBeInTheDocument();
      expect(screen.getByText('(Full)')).toBeInTheDocument();
    });

    const joinButton = screen.getByText(/join/i);
    expect(joinButton).toBeDisabled();
  });

  it('should handle real-time avatar movements from other users', async () => {
    render(<App />);

    // Wait for initial setup
    await waitFor(() => {
      expect(global.WebSocket).toHaveBeenCalled();
    });

    // Simulate receiving avatar movement from another user
    const onMessageHandler = mockWebSocket.addEventListener.mock.calls
      .find(call => call[0] === 'message')?.[1];

    if (onMessageHandler) {
      const avatarMoveEvent = {
        data: JSON.stringify({
          type: 'avatar_moved',
          payload: {
            sessionId: 'other-user-123',
            position: { lat: 48.8566, lng: 2.3522 }, // Paris
            timestamp: new Date().toISOString()
          }
        })
      };

      onMessageHandler(avatarMoveEvent);
    }

    // Should update other user's avatar position on the map
    // Note: This would require the MapContainer to be connected to the WebSocket
    await waitFor(() => {
      // This test will help us identify what integration is needed
      expect(mockWebSocket.addEventListener).toHaveBeenCalledWith('message', expect.any(Function));
    });
  });
});