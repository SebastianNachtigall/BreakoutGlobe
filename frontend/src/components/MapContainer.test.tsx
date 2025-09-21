import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MapContainer } from './MapContainer';

// Mock MapLibre GL JS
Object.defineProperty(window.URL, 'createObjectURL', {
  writable: true,
  value: vi.fn(() => 'mock-url')
});

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
  resize: vi.fn(),
  addControl: vi.fn(),
  getCanvas: vi.fn(() => ({
    style: { cursor: 'default' }
  })),
  getContainer: vi.fn(() => document.createElement('div')),
  loaded: vi.fn(() => true),
  getCenter: vi.fn(() => ({ lng: 0, lat: 0 })),
  getZoom: vi.fn(() => 10),
  getBounds: vi.fn(() => ({
    getNorthEast: () => ({ lng: 1, lat: 1 }),
    getSouthWest: () => ({ lng: -1, lat: -1 })
  }))
};

vi.mock('maplibre-gl', () => ({
  Map: vi.fn(() => mockMap),
  NavigationControl: vi.fn(() => ({})),
  ScaleControl: vi.fn(() => ({})),
  Marker: vi.fn(() => mockMarker)
}));

describe('MapContainer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Map Initialization', () => {
    it('should render map container with correct dimensions', () => {
      render(<MapContainer />);
      
      const mapContainer = screen.getByTestId('map-container');
      expect(mapContainer).toBeInTheDocument();
      expect(mapContainer).toHaveClass('w-full', 'h-full');
    });

    it('should initialize MapLibre GL map with default configuration', async () => {
      const { Map } = vi.mocked(await import('maplibre-gl'));
      
      render(<MapContainer />);
      
      expect(Map).toHaveBeenCalledWith({
        container: expect.any(HTMLElement),
        style: 'https://demotiles.maplibre.org/style.json',
        center: [0, 0],
        zoom: 2,
        attributionControl: false,
        preserveDrawingBuffer: false,
        antialias: false,
        maxZoom: 18,
        minZoom: 1,
        renderWorldCopies: false,
        fadeDuration: 0,
        crossSourceCollisions: false
      });
    });

    it('should initialize map with custom center and zoom', async () => {
      const { Map } = vi.mocked(await import('maplibre-gl'));
      
      render(
        <MapContainer 
          initialCenter={[10, 20]} 
          initialZoom={8} 
        />
      );
      
      expect(Map).toHaveBeenCalledWith(
        expect.objectContaining({
          center: [10, 20],
          zoom: 8
        })
      );
    });

    it('should add navigation and scale controls', async () => {
      const { NavigationControl, ScaleControl } = vi.mocked(await import('maplibre-gl'));
      
      render(<MapContainer />);
      
      expect(NavigationControl).toHaveBeenCalled();
      expect(ScaleControl).toHaveBeenCalled();
    });
  });

  describe('Map Interaction Handling', () => {
    it('should handle map click events', () => {
      const onMapClick = vi.fn();
      
      render(<MapContainer onMapClick={onMapClick} />);
      
      // Verify click event listener was added
      expect(mockMap.on).toHaveBeenCalledWith('click', expect.any(Function));
    });

    it('should call onMapReady when map is initialized', () => {
      const onMapReady = vi.fn();
      
      render(<MapContainer onMapReady={onMapReady} />);
      
      expect(onMapReady).toHaveBeenCalledWith(mockMap);
    });

    it('should handle click-to-move functionality', () => {
      const onAvatarMove = vi.fn();
      const onMapClick = vi.fn();
      
      render(
        <MapContainer 
          onMapClick={onMapClick}
          onAvatarMove={onAvatarMove}
        />
      );
      
      // Simulate map click event
      const clickEvent = {
        lngLat: { lng: -74.0060, lat: 40.7128 }
      };
      
      // Get the click handler that was registered
      const clickHandler = mockMap.on.mock.calls.find(call => call[0] === 'click')[1];
      clickHandler(clickEvent);
      
      expect(onMapClick).toHaveBeenCalledWith(clickEvent);
      expect(onAvatarMove).toHaveBeenCalledWith({ lat: 40.7128, lng: -74.0060 });
    });

    it('should convert coordinates correctly for avatar movement', () => {
      const onAvatarMove = vi.fn();
      
      render(<MapContainer onAvatarMove={onAvatarMove} />);
      
      // Simulate click with different coordinate formats
      const clickEvent = {
        lngLat: { lng: -122.4194, lat: 37.7749 } // San Francisco
      };
      
      const clickHandler = mockMap.on.mock.calls.find(call => call[0] === 'click')[1];
      clickHandler(clickEvent);
      
      expect(onAvatarMove).toHaveBeenCalledWith({ lat: 37.7749, lng: -122.4194 });
    });
  });

  describe('Marker Management', () => {
    it('should create markers for provided avatars', async () => {
      const { Marker } = vi.mocked(await import('maplibre-gl'));
      
      const avatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false },
        { sessionId: 'user-2', position: { lat: 51.5074, lng: -0.1278 }, isCurrentUser: true }
      ];
      
      render(<MapContainer avatars={avatars} />);
      
      expect(Marker).toHaveBeenCalledTimes(2);
      expect(mockMarker.setLngLat).toHaveBeenCalledWith([-74.0060, 40.7128]);
      expect(mockMarker.setLngLat).toHaveBeenCalledWith([-0.1278, 51.5074]);
      expect(mockMarker.addTo).toHaveBeenCalledTimes(2);
    });

    it('should update marker positions when avatars change', () => {
      const initialAvatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false }
      ];
      
      const { rerender } = render(<MapContainer avatars={initialAvatars} />);
      
      const updatedAvatars = [
        { sessionId: 'user-1', position: { lat: 41.8781, lng: -87.6298 }, isCurrentUser: false }
      ];
      
      rerender(<MapContainer avatars={updatedAvatars} />);
      
      expect(mockMarker.setLngLat).toHaveBeenCalledWith([-87.6298, 41.8781]);
    });

    it('should remove markers when avatars are removed', () => {
      const initialAvatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false }
      ];
      
      const { rerender } = render(<MapContainer avatars={initialAvatars} />);
      
      rerender(<MapContainer avatars={[]} />);
      
      expect(mockMarker.remove).toHaveBeenCalled();
    });

    it('should apply smooth movement animations to avatar markers', async () => {
      const { Marker } = vi.mocked(await import('maplibre-gl'));
      
      const movingAvatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false, isMoving: true }
      ];
      
      render(<MapContainer avatars={movingAvatars} />);
      
      // Check that Marker was called with proper options
      expect(Marker).toHaveBeenCalled();
      const markerCall = (Marker as any).mock.calls[0];
      const markerOptions = markerCall[0];
      
      // Verify marker has performance optimizations
      expect(markerOptions.pitchAlignment).toBe('viewport');
      expect(markerOptions.rotationAlignment).toBe('viewport');
      expect(markerOptions.draggable).toBe(false);
      
      // Verify element has proper styling
      const element = markerOptions.element;
      expect(element.style.willChange).toBe('transform');
      expect(element.style.backfaceVisibility).toBe('hidden');
    });

    it('should handle avatar collision detection', () => {
      const overlappingAvatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false },
        { sessionId: 'user-2', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: true } // Same position
      ];
      
      render(<MapContainer avatars={overlappingAvatars} />);
      
      // Verify both markers are created but positioned to avoid overlap
      expect(mockMarker.setLngLat).toHaveBeenCalledTimes(2);
      
      // Second marker should be slightly offset
      const calls = mockMarker.setLngLat.mock.calls;
      const [lng1, lat1] = calls[0][0];
      const [lng2, lat2] = calls[1][0];
      
      // Should have small offset to prevent overlap
      expect(Math.abs(lng1 - lng2) > 0 || Math.abs(lat1 - lat2) > 0).toBe(true);
    });

    it('should optimize marker positioning for better visibility', () => {
      const clusteredAvatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false },
        { sessionId: 'user-2', position: { lat: 40.7129, lng: -74.0061 }, isCurrentUser: false },
        { sessionId: 'user-3', position: { lat: 40.7127, lng: -74.0059 }, isCurrentUser: true }
      ];
      
      render(<MapContainer avatars={clusteredAvatars} />);
      
      // All markers should be created
      expect(mockMarker.setLngLat).toHaveBeenCalledTimes(3);
      
      // Current user marker should be prioritized (rendered last/on top)
      const calls = mockMarker.setLngLat.mock.calls;
      const lastCall = calls[calls.length - 1];
      expect(lastCall[0]).toEqual([-74.0059, 40.7127]); // Current user position
    });
  });

  describe('Cleanup', () => {
    it('should remove map instance on unmount', () => {
      const { unmount } = render(<MapContainer />);
      
      unmount();
      
      expect(mockMap.remove).toHaveBeenCalled();
    });

    it('should remove all markers on unmount', () => {
      const avatars = [
        { sessionId: 'user-1', position: { lat: 40.7128, lng: -74.0060 }, isCurrentUser: false }
      ];
      
      const { unmount } = render(<MapContainer avatars={avatars} />);
      
      unmount();
      
      expect(mockMarker.remove).toHaveBeenCalled();
    });
  });
});