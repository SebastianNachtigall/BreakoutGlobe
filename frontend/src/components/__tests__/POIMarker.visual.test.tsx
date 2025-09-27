import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { POIMarker } from '../POIMarker';

describe('POIMarker Visual Design', () => {
  const mockPOI = {
    id: 'poi-1',
    name: 'Coffee Shop',
    description: 'A cozy coffee shop',
    position: { lat: 40.7128, lng: -74.0060 },
    participantCount: 3,
    maxParticipants: 10,
    createdBy: 'user-123',
    createdAt: new Date()
  };

  it('should render circular image design with name underneath and red badge', () => {
    const poiWithImage = { 
      ...mockPOI, 
      imageUrl: 'https://example.com/coffee-shop.jpg' 
    };
    
    render(<POIMarker poi={poiWithImage} />);
    
    // Check circular image
    const image = screen.getByAltText('Coffee Shop');
    expect(image).toHaveClass('rounded-full', 'w-16', 'h-16');
    
    // Check name is displayed underneath
    const nameElement = screen.getByText('Coffee Shop');
    expect(nameElement).toBeInTheDocument();
    expect(nameElement).toHaveClass('text-xs', 'font-semibold', 'text-gray-800');
    
    // Check red badge with participant count
    const badge = screen.getByTestId('participant-badge');
    expect(badge).toHaveClass('bg-red-500', 'rounded-full', 'w-6', 'h-6');
    expect(badge).toHaveTextContent('3');
    
    // Check overall layout is vertical
    const marker = screen.getByTestId('poi-marker');
    expect(marker).toHaveClass('flex', 'flex-col', 'items-center');
  });

  it('should maintain fallback design for POIs without images', () => {
    render(<POIMarker poi={mockPOI} />);
    
    // Should still show name and participant count
    expect(screen.getByText('Coffee Shop')).toBeInTheDocument();
    expect(screen.getByText('3/10')).toBeInTheDocument();
    
    // Should use default colored background
    const marker = screen.getByTestId('poi-marker');
    expect(marker).toHaveClass('bg-green-500');
  });

  it('should show different badge counts correctly', () => {
    const scenarios = [
      { count: 0, expected: '0' },
      { count: 5, expected: '5' },
      { count: 12, expected: '12' },
      { count: 99, expected: '99' }
    ];

    scenarios.forEach(({ count, expected }) => {
      const poi = { 
        ...mockPOI, 
        participantCount: count,
        imageUrl: 'https://example.com/test.jpg'
      };
      
      const { unmount } = render(<POIMarker poi={poi} />);
      
      const badge = screen.getByTestId('participant-badge');
      expect(badge).toHaveTextContent(expected);
      
      unmount();
    });
  });

  it('should maintain circular design when participant count changes', () => {
    const poiWithImage = { 
      ...mockPOI, 
      imageUrl: 'https://example.com/coffee-shop.jpg',
      participantCount: 1
    };
    
    const { rerender } = render(<POIMarker poi={poiWithImage} />);
    
    // Initial state - should be circular
    expect(screen.getByAltText('Coffee Shop')).toHaveClass('rounded-full');
    expect(screen.getByTestId('participant-badge')).toHaveTextContent('1');
    
    // Update participant count
    const updatedPOI = { ...poiWithImage, participantCount: 5 };
    rerender(<POIMarker poi={updatedPOI} />);
    
    // Should still be circular after update
    expect(screen.getByAltText('Coffee Shop')).toHaveClass('rounded-full');
    expect(screen.getByTestId('participant-badge')).toHaveTextContent('5');
    
    // Layout should still be vertical
    const marker = screen.getByTestId('poi-marker');
    expect(marker).toHaveClass('flex', 'flex-col', 'items-center');
  });
});