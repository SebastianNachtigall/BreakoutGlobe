import { render, screen } from '@testing-library/react';
import { POIMarker } from '../POIMarker';
import type { POIData } from '../MapContainer';

describe('POIMarker Avatar Badges', () => {
  const mockPOI: POIData = {
    id: 'poi-123',
    name: 'Coffee Shop',
    description: 'A great place for coffee',
    position: { lat: 40.7128, lng: -74.0060 },
    participantCount: 3,
    maxParticipants: 10,
    imageUrl: 'https://example.com/coffee.jpg',
    createdBy: 'user-456',
    createdAt: new Date('2023-01-01'),
    participants: [
      { id: 'user-1', name: 'Alice', avatarUrl: 'https://example.com/alice.jpg' },
      { id: 'user-2', name: 'Bob', avatarUrl: 'https://example.com/bob.jpg' },
      { id: 'user-3', name: 'Charlie', avatarUrl: null }
    ]
  };

  it('should render avatar badges around POI when users are joined', () => {
    render(<POIMarker poi={mockPOI} />);
    
    // Should show the POI marker
    expect(screen.getByTestId('poi-marker')).toBeInTheDocument();
    
    // Should show avatar badges for each participant
    expect(screen.getByTestId('avatar-badges-container')).toBeInTheDocument();
    
    // Should show 3 avatar badges
    const avatarBadges = screen.getAllByTestId('avatar-badge');
    expect(avatarBadges).toHaveLength(3);
    
    // Check that badges have the correct positioning transforms
    const firstTransform = avatarBadges[0].style.transform;
    const secondTransform = avatarBadges[1].style.transform;
    const thirdTransform = avatarBadges[2].style.transform;
    
    expect(firstTransform).toContain('rotate(-60deg)');
    expect(secondTransform).toContain('rotate(-105deg)');
    expect(thirdTransform).toContain('rotate(-150deg)');
  });

  it('should show avatar images in badges when available', () => {
    render(<POIMarker poi={mockPOI} />);
    
    const avatarBadges = screen.getAllByTestId('avatar-badge');
    
    // First badge should have Alice's avatar image
    const aliceImg = avatarBadges[0].querySelector('img');
    expect(aliceImg).toHaveAttribute('src', 'https://example.com/alice.jpg');
    expect(aliceImg).toHaveAttribute('alt', 'Alice');
    
    // Second badge should have Bob's avatar image
    const bobImg = avatarBadges[1].querySelector('img');
    expect(bobImg).toHaveAttribute('src', 'https://example.com/bob.jpg');
    expect(bobImg).toHaveAttribute('alt', 'Bob');
    
    // Third badge should show initials for Charlie (no avatar)
    expect(avatarBadges[2]).toHaveTextContent('C');
  });

  it('should not render avatar badges when no participants', () => {
    const emptyPOI = { ...mockPOI, participantCount: 0, participants: [] };
    render(<POIMarker poi={emptyPOI} />);
    
    expect(screen.getByTestId('poi-marker')).toBeInTheDocument();
    expect(screen.queryByTestId('avatar-badges-container')).not.toBeInTheDocument();
  });

  it('should keep participant count badge visible alongside avatar badges', () => {
    render(<POIMarker poi={mockPOI} />);
    
    // Should still show the participant count badge
    expect(screen.getByTestId('participant-badge')).toBeInTheDocument();
    expect(screen.getByTestId('participant-badge')).toHaveTextContent('3');
    
    // Should also show avatar badges
    expect(screen.getByTestId('avatar-badges-container')).toBeInTheDocument();
  });

  it('should handle maximum of 8 avatar badges with proper positioning', () => {
    const manyParticipantsPOI = {
      ...mockPOI,
      participantCount: 8,
      participants: Array.from({ length: 8 }, (_, i) => ({
        id: `user-${i + 1}`,
        name: `User ${i + 1}`,
        avatarUrl: `https://example.com/user${i + 1}.jpg`
      }))
    };
    
    render(<POIMarker poi={manyParticipantsPOI} />);
    
    const avatarBadges = screen.getAllByTestId('avatar-badge');
    expect(avatarBadges).toHaveLength(8);
    
    // Check that all 8 positions are used (45-degree increments starting at 11 o'clock)
    const expectedAngles = [-60, -105, -150, -195, -240, -285, -330, -15];
    avatarBadges.forEach((badge, index) => {
      const transform = badge.style.transform;
      expect(transform).toContain(`rotate(${expectedAngles[index]}deg)`);
    });
  });
});