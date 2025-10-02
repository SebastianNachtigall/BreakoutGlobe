import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from '../App';

// Mock all the services and stores
vi.mock('../services/api', () => ({
  getCurrentUserProfile: vi.fn(),
  createGuestProfile: vi.fn(),
  getPOIs: vi.fn().mockResolvedValue([]),
  createPOI: vi.fn(),
  joinPOI: vi.fn(),
  leavePOI: vi.fn(),
  deletePOI: vi.fn(),
  clearAllPOIs: vi.fn(),
  clearAllUsers: vi.fn(),
  transformToCreatePOIRequest: vi.fn(),
  transformFromPOIResponse: vi.fn(),
}));

vi.mock('../services/websocket-client', () => ({
  WebSocketClient: vi.fn().mockImplementation(() => ({
    connect: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(true),
    onStatusChange: vi.fn(),
    onError: vi.fn(),
    onStateSync: vi.fn(),
    requestInitialUsers: vi.fn(),
    moveAvatar: vi.fn(),
    leaveCurrentPOI: vi.fn(),
  })),
  ConnectionStatus: {
    CONNECTED: 'connected',
    DISCONNECTED: 'disconnected',
    CONNECTING: 'connecting',
  },
}));

vi.mock('../services/session-service', () => ({
  SessionService: vi.fn().mockImplementation(() => ({
    startHeartbeat: vi.fn(),
    stopHeartbeat: vi.fn(),
  })),
}));

// Mock stores
vi.mock('../stores/userProfileStore', () => ({
  userProfileStore: {
    getState: vi.fn(() => ({
      getProfileOffline: vi.fn().mockReturnValue(null),
      setProfile: vi.fn(),
      clearProfile: vi.fn(),
    })),
  },
}));

vi.mock('../stores/sessionStore', () => ({
  sessionStore: vi.fn(() => ({
    sessionId: null,
  })),
}));

vi.mock('../stores/poiStore', () => ({
  poiStore: vi.fn(() => ({
    pois: [],
    isLoading: false,
    error: null,
    currentUserPOI: null,
  })),
}));

vi.mock('../stores/errorStore', () => ({
  errorStore: {
    getState: vi.fn(() => ({
      addError: vi.fn(),
    })),
  },
}));

vi.mock('../stores/avatarStore', () => ({
  avatarStore: {
    subscribe: vi.fn(() => vi.fn()),
    getState: vi.fn(() => ({
      getAvatarBySessionId: vi.fn(),
    })),
  },
}));

vi.mock('../stores/videoCallStore', () => ({
  videoCallStore: vi.fn(() => ({
    callState: 'idle',
    currentCall: null,
    isGroupCallActive: false,
    currentPOI: null,
    groupCallParticipants: [],
    remoteStreams: {},
    localStream: null,
    isAudioEnabled: true,
    isVideoEnabled: true,
  })),
  setWebSocketClient: vi.fn(),
}));

// Mock environment variables
Object.defineProperty(import.meta, 'env', {
  value: {
    VITE_API_BASE_URL: 'http://localhost:8080',
    VITE_WS_URL: 'ws://localhost:8080',
  },
});

// Mock fetch for session creation
global.fetch = vi.fn();

describe('Welcome Screen Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock localStorage
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
    });

    // Mock fetch responses
    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        sessionId: 'test-session-123',
        position: { lat: 52.5200, lng: 13.4050 },
      }),
    });
  });

  it('shows welcome screen for new users before profile creation', async () => {
    const { getCurrentUserProfile } = await import('../services/api');
    
    // Mock no existing profile (404 error)
    (getCurrentUserProfile as any).mockRejectedValue(new Error('404'));

    render(<App />);

    // Wait for profile check to complete
    await waitFor(() => {
      expect(screen.getByText('Welcome')).toBeInTheDocument();
    });

    // Verify welcome screen content
    expect(screen.getByText('Join POIs on the map to initiate video calls.')).toBeInTheDocument();
    expect(screen.getByText('Useful for user-driven breakout sessions in a workshop scenario.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Get Started' })).toBeInTheDocument();

    // Verify profile creation modal is NOT shown yet
    expect(screen.queryByText('Create Your Profile')).not.toBeInTheDocument();
  });

  it('transitions from welcome screen to profile creation modal', async () => {
    const { getCurrentUserProfile } = await import('../services/api');
    
    // Mock no existing profile
    (getCurrentUserProfile as any).mockRejectedValue(new Error('404'));

    render(<App />);

    // Wait for welcome screen
    await waitFor(() => {
      expect(screen.getByText('Welcome')).toBeInTheDocument();
    });

    // Click "Get Started" button
    const getStartedButton = screen.getByRole('button', { name: 'Get Started' });
    fireEvent.click(getStartedButton);

    // Wait for profile creation modal to appear
    await waitFor(() => {
      expect(screen.getByText('Create Your Profile')).toBeInTheDocument();
    });

    // Verify welcome screen is no longer visible
    expect(screen.queryByText('Welcome')).not.toBeInTheDocument();
    
    // Verify profile creation modal content
    expect(screen.getByLabelText('Display Name *')).toBeInTheDocument();
    expect(screen.getByLabelText('About Me')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Create Profile' })).toBeInTheDocument();
  });

  it('skips welcome screen for existing users', async () => {
    const { getCurrentUserProfile } = await import('../services/api');
    
    // Mock existing profile
    const mockProfile = {
      id: 'user-123',
      displayName: 'Test User',
      aboutMe: 'Test about me',
      avatarUrl: null,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    (getCurrentUserProfile as any).mockResolvedValue(mockProfile);

    render(<App />);

    // Wait for initialization
    await waitFor(() => {
      expect(screen.getByText('Initializing BreakoutGlobe...')).toBeInTheDocument();
    });

    // Verify welcome screen is NOT shown
    expect(screen.queryByText('Welcome')).not.toBeInTheDocument();
    expect(screen.queryByText('Create Your Profile')).not.toBeInTheDocument();
  });
});