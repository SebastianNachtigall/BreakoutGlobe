import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { MapContainer } from './MapContainer';

// Mock MapLibre GL JS - Enhanced to actually add elements to DOM
let createdMarkerElements: HTMLElement[] = [];

const mockMarker = {
  setLngLat: vi.fn().mockReturnThis(),
  addTo: vi.fn((map) => {
    // Actually add the marker element to the DOM to simulate real behavior
    const element = mockMarker.getElement();
    element.classList.add('maplibregl-marker');
    document.body.appendChild(element);
    return mockMarker;
  }),
  remove: vi.fn(() => {
    const element = mockMarker.getElement();
    if (element.parentNode) {
      element.parentNode.removeChild(element);
    }
    return mockMarker;
  }),
  getElement: vi.fn(() => {
    // This should return the actual element created by MapContainer
    // But since we're mocking, we need to track the real elements
    const existingElement = createdMarkerElements[createdMarkerElements.length - 1];
    return existingElement || document.createElement('div');
  }),
  getLngLat: vi.fn(() => ({ lng: -74.0060, lat: 40.7128 })),
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
  Marker: vi.fn((options) => {
    // Capture the element that MapContainer creates
    if (options && options.element) {
      createdMarkerElements.push(options.element);
    }
    return mockMarker;
  })
}));

describe('Avatar Visibility Issue', () => {
  const mockAvatarData = {
    sessionId: 'session-123',
    userId: 'user-456',
    displayName: 'AnotherProfile',
    avatarURL: undefined,
    position: { lat: 40.7128, lng: -74.0060 },
    isCurrentUser: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Clear any existing marker elements from previous tests
    createdMarkerElements.forEach(element => {
      if (element.parentNode) {
        element.parentNode.removeChild(element);
      }
    });
    createdMarkerElements = [];
  });

  it('should render visible initials for guest users without avatar images', async () => {
    // This test should FAIL initially, reproducing the bug
    render(
      <MapContainer
        avatars={[mockAvatarData]}
        pois={[]}
        onPOICreate={() => {}}
      />
    );

    await waitFor(() => {
      // Check the elements that MapContainer actually created
      console.log('🔍 Created marker elements:', createdMarkerElements.length);
      
      let foundInitials = false;
      let elementDetails: string[] = [];
      
      createdMarkerElements.forEach((element, index) => {
        const textContent = element.textContent || '';
        const className = element.className || '';
        const computedStyle = window.getComputedStyle(element);
        
        elementDetails.push(`Marker ${index}: text="${textContent}", class="${className}", display="${computedStyle.display}", visibility="${computedStyle.visibility}"`);
        
        if (textContent.includes('AN')) {
          foundInitials = true;
        }
      });
      
      console.log('🎯 Element details:', elementDetails);
      
      // Also check DOM for maplibre markers
      const domMarkers = document.querySelectorAll('.maplibregl-marker');
      console.log('🗺️ DOM markers found:', domMarkers.length);
      
      // This should reveal the issue - elements exist but may not be visible
      expect(foundInitials).toBe(true);
      
      // Also check if React Testing Library can find the initials
      const initialsElement = screen.queryByText('AN');
      console.log('🔤 RTL found initials element:', !!initialsElement);
      expect(initialsElement).toBeInTheDocument();
    });
  });

  it('should position avatar initials correctly on the map', async () => {
    render(
      <MapContainer
        avatars={[mockAvatarData]}
        pois={[]}
        onPOICreate={() => {}}
      />
    );

    await waitFor(() => {
      const initialsElement = screen.queryByText('AN');
      
      if (initialsElement) {
        const computedStyle = window.getComputedStyle(initialsElement);
        
        // These should FAIL - positioning likely incorrect
        expect(computedStyle.position).toBe('absolute');
        expect(computedStyle.zIndex).not.toBe('');
        expect(parseInt(computedStyle.zIndex)).toBeGreaterThan(0);
      }
    });
  });

  it('should apply consistent styling between initials and image avatars', async () => {
    // Test with image avatar first
    const imageAvatarData = {
      ...mockAvatarData,
      avatarURL: 'https://example.com/avatar.jpg',
    };

    const { rerender } = render(
      <MapContainer
        avatars={[imageAvatarData]}
        pois={[]}
        onPOICreate={() => {}}
      />
    );

    // Then test with initials avatar
    rerender(
      <MapContainer
        avatars={[mockAvatarData]}
        pois={[]}
        onPOICreate={() => {}}
      />
    );

    await waitFor(() => {
      const initialsElement = screen.queryByText('AN');
      
      if (initialsElement) {
        const computedStyle = window.getComputedStyle(initialsElement);
        
        // These should FAIL - styling likely inconsistent
        expect(computedStyle.width).not.toBe('');
        expect(computedStyle.height).not.toBe('');
        expect(computedStyle.borderRadius).not.toBe('');
      }
    });
  });
});