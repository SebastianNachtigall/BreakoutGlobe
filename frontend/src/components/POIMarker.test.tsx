import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { POIMarker } from './POIMarker';

// Mock MapLibre GL JS
vi.mock('maplibre-gl', () => ({
  Marker: vi.fn(() => ({
    setLngLat: vi.fn().mockReturnThis(),
    addTo: vi.fn().mockReturnThis(),
    remove: vi.fn().mockReturnThis(),
    getElement: vi.fn(() => document.createElement('div'))
  }))
}));

describe('POIMarker', () => {
  const mockPOI = {
    id: 'poi-1',
    name: 'Test Meeting Room',
    description: 'A test POI for meetings',
    position: { lat: 40.7128, lng: -74.0060 },
    participantCount: 3,
    maxParticipants: 10,
    createdBy: 'user-123',
    createdAt: new Date()
  };

  describe('Rendering', () => {
    it('should render POI marker with name', () => {
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={mockPOI} 
          onPOIClick={onPOIClick}
        />
      );
      
      expect(screen.getByText('Test Meeting Room')).toBeInTheDocument();
    });

    it('should display participant count', () => {
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={mockPOI} 
          onPOIClick={onPOIClick}
        />
      );
      
      expect(screen.getByText('3/10')).toBeInTheDocument();
    });

    it('should show full indicator when at capacity', () => {
      const fullPOI = { ...mockPOI, participantCount: 10 };
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={fullPOI} 
          onPOIClick={onPOIClick}
        />
      );
      
      const marker = screen.getByTestId('poi-marker');
      expect(marker).toHaveClass('bg-red-500'); // Full indicator
    });

    it('should render POI marker with image when imageUrl is provided', () => {
      const poiWithImage = { ...mockPOI, imageUrl: 'https://example.com/test-image.jpg' };
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={poiWithImage} 
          onPOIClick={onPOIClick}
        />
      );
      
      const image = screen.getByAltText('Test Meeting Room');
      expect(image).toBeInTheDocument();
      expect(image).toHaveAttribute('src', 'https://example.com/test-image.jpg');
      expect(screen.getByText('Test Meeting Room')).toBeInTheDocument();
      expect(screen.getByText('3/10')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('should call onPOIClick when marker is clicked', () => {
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={mockPOI} 
          onPOIClick={onPOIClick}
        />
      );
      
      fireEvent.click(screen.getByTestId('poi-marker'));
      expect(onPOIClick).toHaveBeenCalledWith(mockPOI.id);
    });

    it('should prevent click when POI is at capacity', () => {
      const fullPOI = { ...mockPOI, participantCount: 10 };
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={fullPOI} 
          onPOIClick={onPOIClick}
        />
      );
      
      fireEvent.click(screen.getByTestId('poi-marker'));
      expect(onPOIClick).not.toHaveBeenCalled();
    });
  });
});