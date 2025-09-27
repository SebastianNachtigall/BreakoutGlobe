import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { vi } from 'vitest';
import { POIDetailsPanel } from '../POIDetailsPanel';
import type { POIData } from '../MapContainer';

const mockPOI: POIData = {
  id: 'poi-123',
  name: 'Coffee Shop',
  description: 'Great coffee and wifi',
  position: { lat: 40.7128, lng: -74.0060 },
  participantCount: 2,
  maxParticipants: 10,
  participants: [
    { id: 'user-1', name: 'John Doe' },
    { id: 'user-2', name: 'Jane Smith' }
  ],
  imageUrl: 'http://localhost:8080/uploads/poi-image.jpg',
  createdBy: 'user-1',
  createdAt: new Date('2023-01-01'),
  discussionStartTime: new Date(Date.now() - 120000), // 2 minutes ago
  isDiscussionActive: true,
  discussionDuration: 120 // 2 minutes in seconds
};

const defaultProps = {
  poi: mockPOI,
  currentUserId: 'user-3',
  isUserParticipant: false,
  onJoin: vi.fn(),
  onLeave: vi.fn(),
  onClose: vi.fn(),
  isLoading: false
};

describe('POIDetailsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('POI Image Display', () => {
    it('should display POI image when imageUrl is provided', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      const image = screen.getByAltText('Coffee Shop');
      expect(image).toBeInTheDocument();
      expect(image).toHaveAttribute('src', 'http://localhost:8080/uploads/poi-image.jpg');
    });

    it('should not display image when imageUrl is not provided', () => {
      const poiWithoutImage = { ...mockPOI, imageUrl: undefined };
      render(<POIDetailsPanel {...defaultProps} poi={poiWithoutImage} />);
      
      expect(screen.queryByAltText('Coffee Shop')).not.toBeInTheDocument();
    });

    it('should handle image load error gracefully', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      const image = screen.getByAltText('Coffee Shop');
      fireEvent.error(image);
      
      // Image should still be in DOM but with error handling
      expect(image).toBeInTheDocument();
    });
  });

  describe('User Screen Names Display', () => {
    it('should display user screen names instead of internal usernames', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      // Should show display names, not user IDs
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('Jane Smith')).toBeInTheDocument();
      
      // Should not show internal user IDs
      expect(screen.queryByText('user-1')).not.toBeInTheDocument();
      expect(screen.queryByText('user-2')).not.toBeInTheDocument();
    });

    it('should mark current user with "(You)" indicator', () => {
      const currentUserPOI = {
        ...mockPOI,
        participants: [
          { id: 'user-1', name: 'John Doe' },
          { id: 'user-3', name: 'Current User' }
        ]
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={currentUserPOI} currentUserId="user-3" isUserParticipant={true} />);
      
      expect(screen.getByText('Current User')).toBeInTheDocument();
      expect(screen.getByText('(You)')).toBeInTheDocument();
    });

    it('should handle participants without display names gracefully', () => {
      const poiWithMissingNames = {
        ...mockPOI,
        participants: [
          { id: 'user-1', name: 'John Doe' },
          { id: 'user-2', name: '' } // Empty name
        ]
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={poiWithMissingNames} />);
      
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      // Should show fallback for empty name (first 8 chars of user ID)
      expect(screen.getByText('user-2')).toBeInTheDocument();
    });
  });

  describe('Discussion Timer Logic', () => {
    it('should show active discussion timer when 2 or more participants are present', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      expect(screen.getByText(/Discussion active for:/)).toBeInTheDocument();
      expect(screen.getByText(/2 minutes/)).toBeInTheDocument();
    });

    it('should show "No active discussion" when less than 2 participants', () => {
      const poiWithOneParticipant = {
        ...mockPOI,
        participantCount: 1,
        participants: [{ id: 'user-1', name: 'John Doe' }],
        isDiscussionActive: false,
        discussionDuration: 0
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={poiWithOneParticipant} />);
      
      expect(screen.getByText('No active discussion')).toBeInTheDocument();
      expect(screen.queryByText(/Discussion active for:/)).not.toBeInTheDocument();
    });

    it('should format discussion time correctly for seconds only', () => {
      const poiWithShortDiscussion = {
        ...mockPOI,
        discussionDuration: 45 // 45 seconds
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={poiWithShortDiscussion} />);
      
      expect(screen.getByText(/45 seconds/)).toBeInTheDocument();
    });

    it('should format discussion time correctly for minutes and seconds', () => {
      const poiWithLongDiscussion = {
        ...mockPOI,
        discussionDuration: 125 // 2 minutes 5 seconds
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={poiWithLongDiscussion} />);
      
      expect(screen.getByText(/2 minutes 5 seconds/)).toBeInTheDocument();
    });

    it('should format discussion time correctly for exact minutes', () => {
      const poiWithExactMinutes = {
        ...mockPOI,
        discussionDuration: 180 // exactly 3 minutes
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={poiWithExactMinutes} />);
      
      expect(screen.getByText(/3 minutes/)).toBeInTheDocument();
      expect(screen.queryByText(/seconds/)).not.toBeInTheDocument();
    });

    it('should handle singular vs plural time units correctly', () => {
      const poiWithSingularTime = {
        ...mockPOI,
        discussionDuration: 61 // 1 minute 1 second
      };
      
      render(<POIDetailsPanel {...defaultProps} poi={poiWithSingularTime} />);
      
      expect(screen.getByText(/1 minute 1 second/)).toBeInTheDocument();
    });
  });

  describe('Panel Behavior', () => {
    it('should call onClose when close button is clicked', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      const closeButton = screen.getByLabelText('Close panel');
      fireEvent.click(closeButton);
      
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when escape key is pressed', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      fireEvent.keyDown(document, { key: 'Escape' });
      
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1);
    });

    it('should call onJoin when join button is clicked', () => {
      render(<POIDetailsPanel {...defaultProps} />);
      
      const joinButton = screen.getByText('Join');
      fireEvent.click(joinButton);
      
      expect(defaultProps.onJoin).toHaveBeenCalledWith('poi-123');
    });

    it('should call onLeave when leave button is clicked for participant', () => {
      render(<POIDetailsPanel {...defaultProps} isUserParticipant={true} />);
      
      const leaveButton = screen.getByText('Leave');
      fireEvent.click(leaveButton);
      
      expect(defaultProps.onLeave).toHaveBeenCalledWith('poi-123');
    });
  });
});