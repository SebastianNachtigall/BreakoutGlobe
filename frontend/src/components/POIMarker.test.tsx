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

    it('should render POI marker with circular image and name underneath', () => {
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
      expect(image).toHaveClass('rounded-full'); // Circular image
      
      const nameElement = screen.getByText('Test Meeting Room');
      expect(nameElement).toBeInTheDocument();
      
      // Check that the structure has image above and name below
      const marker = screen.getByTestId('poi-marker');
      expect(marker).toHaveClass('flex-col'); // Vertical layout
    });

    it('should display red badge with participant count', () => {
      const poiWithImage = { ...mockPOI, imageUrl: 'https://example.com/test-image.jpg' };
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={poiWithImage} 
          onPOIClick={onPOIClick}
        />
      );
      
      const badge = screen.getByTestId('participant-badge');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveClass('bg-red-500'); // Red badge
      expect(badge).toHaveTextContent('3'); // Participant count
    });

    it('should position badge in top-right corner of circular image', () => {
      const poiWithImage = { ...mockPOI, imageUrl: 'https://example.com/test-image.jpg' };
      const onPOIClick = vi.fn();
      
      render(
        <POIMarker 
          poi={poiWithImage} 
          onPOIClick={onPOIClick}
        />
      );
      
      const badge = screen.getByTestId('participant-badge');
      expect(badge).toHaveClass('absolute', 'top-0', 'right-0'); // Positioned in top-right corner
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