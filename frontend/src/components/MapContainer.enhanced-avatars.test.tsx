import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MapContainer } from './MapContainer';
import type { UserProfile } from '../types/models';

// Mock MapLibre GL JS
const mockMarker = {
  setLngLat: vi.fn().mockReturnThis(),
  addTo: vi.fn().mockReturnThis(),
  remove: vi.fn().mockReturnThis(),
  getElement: vi.fn(() => document.createElement('div')),
  getLngLat: vi.fn(() => ({ lng: 0, lat: 0 })),
  setPopup: vi.fn().mockReturnThis(),
  togglePopup: vi.fn().mockReturnThis()
};

const mockMap = {
  on: vi.fn(),
  off: vi.fn(),
  remove: vi.fn(),
  addControl: vi.fn(),
  getCanvas: vi.fn(() => ({ style: { cursor: 'default' } })),
  loaded: vi.fn(() => true)
};

vi.mock('maplibre-gl', () => ({
  Map: vi.fn(() => mockMap),
  NavigationControl: vi.fn(() => ({})),
  ScaleControl: vi.fn(() => ({})),
  Marker: vi.fn(() => mockMarker)
}));

// Enhanced AvatarData interface for testing
interface EnhancedAvatarData {
  sessionId: string;
  userId: string;
  displayName: string;
  avatarURL?: string;
  position: {
    lat: number;
    lng: number;
  };
  isCurrentUser: boolean;
  isMoving?: boolean;
  role: 'user' | 'admin' | 'superadmin';
}

describe.skip('MapContainer - Enhanced Avatar Display', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Avatar Display with User Profiles', () => {
    it('should display user avatar image when avatarURL is provided', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          avatarURL: 'https://example.com/avatar.jpg',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        // Should create marker with avatar image
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that marker element contains avatar image
      const markerElement = mockMarker.getElement();
      expect(markerElement.querySelector('img')).toBeTruthy();
      expect(markerElement.querySelector('img')?.src).toBe('https://example.com/avatar.jpg');
    });

    it('should display initials when no avatar image is provided', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that marker element contains initials
      const markerElement = mockMarker.getElement();
      expect(markerElement.textContent).toBe('JD'); // John Doe initials
    });

    it('should display visually styled initials for guest profiles without avatars', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'ANewProfile1',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that marker element contains correct initials
      const markerElement = mockMarker.getElement();
      expect(markerElement.textContent).toBe('AN'); // ANewProfile1 initials
      
      // CRITICAL: Check that initials are visually styled and visible
      expect(markerElement.className).toContain('text-white'); // Text should be white
      expect(markerElement.className).toContain('text-xs'); // Text should have size
      expect(markerElement.className).toContain('font-bold'); // Text should be bold
      expect(markerElement.className).toContain('flex'); // Should use flexbox for centering
      expect(markerElement.className).toContain('items-center'); // Should center vertically
      expect(markerElement.className).toContain('justify-center'); // Should center horizontally
      expect(markerElement.className).toContain('bg-gray-500'); // Should have background color
      expect(markerElement.className).toContain('w-8'); // Should have width
      expect(markerElement.className).toContain('h-8'); // Should have height
      
      // Verify no image element exists (should be text only)
      expect(markerElement.querySelector('img')).toBeNull();
      
      // Verify the element is not empty or hidden
      expect(markerElement.textContent).not.toBe('');
      expect(markerElement.style.display).not.toBe('none');
      expect(markerElement.style.visibility).not.toBe('hidden');
    });

    it('should show display name on hover', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that marker element has correct title attribute
      const markerElement = mockMarker.getElement();
      expect(markerElement.title).toBe('John Doe');
    });

    it('should handle fallback for single name initials', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'Madonna',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that marker element contains first two characters for single name
      const markerElement = mockMarker.getElement();
      expect(markerElement.textContent).toBe('MA'); // First two characters
    });

    it('should display different styling for current user', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'Current User',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: true,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that current user marker has different styling
      const markerElement = mockMarker.getElement();
      expect(markerElement.className).toContain('bg-blue-500');
      expect(markerElement.className).toContain('border-blue-600');
    });

    it('should display role-based styling for admin users', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'Admin User',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'admin'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Check that admin user marker has special styling
      const markerElement = mockMarker.getElement();
      expect(markerElement.className).toContain('ring-yellow-400'); // Admin indicator
    });
  });

  describe('Profile Card Display', () => {
    it('should show profile card when avatar is clicked', async () => {
      const user = userEvent.setup();
      const onAvatarClick = vi.fn();
      
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          avatarURL: 'https://example.com/avatar.jpg',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} onAvatarClick={onAvatarClick} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Simulate clicking on avatar marker
      const markerElement = mockMarker.getElement();
      await user.click(markerElement);

      expect(onAvatarClick).toHaveBeenCalledWith('user-1');
    });

    it('should display profile card with user information', async () => {
      const userProfile: UserProfile = {
        id: 'user-1',
        displayName: 'John Doe',
        avatarURL: 'https://example.com/avatar.jpg',
        aboutMe: 'Software developer from NYC',
        accountType: 'full',
        role: 'user',
        isActive: true,
        emailVerified: true,
        createdAt: new Date('2024-01-01')
      };

      render(
        <MapContainer 
          showProfileCard={true}
          selectedUserProfile={userProfile}
          onProfileCardClose={vi.fn()}
        />
      );

      // Should display profile card
      expect(screen.getByTestId('profile-card')).toBeInTheDocument();
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('Software developer from NYC')).toBeInTheDocument();
      expect(screen.getByRole('img')).toHaveAttribute('src', 'https://example.com/avatar.jpg');
    });

    it('should close profile card when close button is clicked', async () => {
      const user = userEvent.setup();
      const onProfileCardClose = vi.fn();
      
      const userProfile: UserProfile = {
        id: 'user-1',
        displayName: 'John Doe',
        accountType: 'guest',
        role: 'user',
        isActive: true,
        emailVerified: false,
        createdAt: new Date('2024-01-01')
      };

      render(
        <MapContainer 
          showProfileCard={true}
          selectedUserProfile={userProfile}
          onProfileCardClose={onProfileCardClose}
        />
      );

      const closeButton = screen.getByRole('button', { name: /close/i });
      await user.click(closeButton);

      expect(onProfileCardClose).toHaveBeenCalled();
    });
  });

  describe('Avatar Image Loading', () => {
    it('should handle avatar image loading errors gracefully', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          avatarURL: 'https://invalid-url.com/avatar.jpg',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Simulate image load error
      const markerElement = mockMarker.getElement();
      const avatarImage = markerElement.querySelector('img');
      
      if (avatarImage) {
        fireEvent.error(avatarImage);
        
        // Should fallback to initials
        await waitFor(() => {
          expect(markerElement.textContent).toBe('JD');
        });
      }
    });

    it('should show loading state while avatar image loads', async () => {
      const avatarsWithProfiles: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          avatarURL: 'https://example.com/avatar.jpg',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      render(<MapContainer avatars={avatarsWithProfiles} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Should show loading indicator initially
      const markerElement = mockMarker.getElement();
      expect(markerElement.className).toContain('animate-pulse');
    });
  });

  describe('Real-time Avatar Updates', () => {
    it('should update avatar display when user profile changes', async () => {
      const initialAvatars: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      const { rerender } = render(<MapContainer avatars={initialAvatars} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Update avatar with new image
      const updatedAvatars: EnhancedAvatarData[] = [
        {
          ...initialAvatars[0],
          avatarURL: 'https://example.com/new-avatar.jpg'
        }
      ];

      rerender(<MapContainer avatars={updatedAvatars} />);

      await waitFor(() => {
        const markerElement = mockMarker.getElement();
        const avatarImage = markerElement.querySelector('img');
        expect(avatarImage?.src).toBe('https://example.com/new-avatar.jpg');
      });
    });

    it('should update display name when user profile changes', async () => {
      const initialAvatars: EnhancedAvatarData[] = [
        {
          sessionId: 'session-1',
          userId: 'user-1',
          displayName: 'John Doe',
          position: { lat: 40.7128, lng: -74.0060 },
          isCurrentUser: false,
          role: 'user'
        }
      ];

      const { rerender } = render(<MapContainer avatars={initialAvatars} />);

      await waitFor(() => {
        expect(mockMarker.getElement).toHaveBeenCalled();
      });

      // Update display name
      const updatedAvatars: EnhancedAvatarData[] = [
        {
          ...initialAvatars[0],
          displayName: 'Jane Smith'
        }
      ];

      rerender(<MapContainer avatars={updatedAvatars} />);

      await waitFor(() => {
        const markerElement = mockMarker.getElement();
        expect(markerElement.title).toBe('Jane Smith');
        expect(markerElement.textContent).toBe('JS'); // New initials
      });
    });
  });
});