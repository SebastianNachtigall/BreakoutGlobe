import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { POIDetailsPanel } from './POIDetailsPanel';

const mockOnJoin = vi.fn();
const mockOnLeave = vi.fn();
const mockOnClose = vi.fn();

const mockPOI = {
  id: 'poi-1',
  name: 'Meeting Room A',
  description: 'A comfortable meeting room for team discussions',
  maxParticipants: 10,
  participantCount: 3,
  position: { lat: 40.7128, lng: -74.0060 },
  participants: [
    { id: 'user-1', name: 'Alice Johnson' },
    { id: 'user-2', name: 'Bob Smith' },
    { id: 'user-3', name: 'Carol Davis' }
  ]
};

const defaultProps = {
  poi: mockPOI,
  currentUserId: 'user-4',
  isUserParticipant: false,
  onJoin: mockOnJoin,
  onLeave: mockOnLeave,
  onClose: mockOnClose,
  isLoading: false
};

describe('POIDetailsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render POI details', () => {
    render(<POIDetailsPanel {...defaultProps} />);
    
    expect(screen.getByText('Meeting Room A')).toBeInTheDocument();
    expect(screen.getByText('A comfortable meeting room for team discussions')).toBeInTheDocument();
    expect(screen.getByText('3/10 participants')).toBeInTheDocument();
  });

  it('should display participant list', () => {
    render(<POIDetailsPanel {...defaultProps} />);
    
    expect(screen.getByText('Participants')).toBeInTheDocument();
    expect(screen.getByText('Alice Johnson')).toBeInTheDocument();
    expect(screen.getByText('Bob Smith')).toBeInTheDocument();
    expect(screen.getByText('Carol Davis')).toBeInTheDocument();
  });

  it('should show join button when user is not a participant', () => {
    render(<POIDetailsPanel {...defaultProps} />);
    
    const joinButton = screen.getByText(/join/i);
    expect(joinButton).toBeInTheDocument();
    expect(joinButton).toBeEnabled();
  });

  it('should show leave button when user is a participant', () => {
    render(<POIDetailsPanel {...defaultProps} isUserParticipant={true} />);
    
    const leaveButton = screen.getByText(/leave/i);
    expect(leaveButton).toBeInTheDocument();
    expect(leaveButton).toBeEnabled();
  });

  it('should disable join button when POI is at capacity', () => {
    const fullPOI = {
      ...mockPOI,
      participantCount: 10,
      participants: Array.from({ length: 10 }, (_, i) => ({
        id: `user-${i + 1}`,
        name: `User ${i + 1}`
      }))
    };

    render(<POIDetailsPanel {...defaultProps} poi={fullPOI} />);
    
    const joinButton = screen.getByText('Join (Full)');
    expect(joinButton).toBeDisabled();
    expect(screen.getByText('(Full)')).toBeInTheDocument();
  });

  it('should call onJoin when join button is clicked', async () => {
    const user = userEvent.setup();
    render(<POIDetailsPanel {...defaultProps} />);
    
    const joinButton = screen.getByText(/join/i);
    await user.click(joinButton);
    
    expect(mockOnJoin).toHaveBeenCalledWith('poi-1');
  });

  it('should call onLeave when leave button is clicked', async () => {
    const user = userEvent.setup();
    render(<POIDetailsPanel {...defaultProps} isUserParticipant={true} />);
    
    const leaveButton = screen.getByText(/leave/i);
    await user.click(leaveButton);
    
    expect(mockOnLeave).toHaveBeenCalledWith('poi-1');
  });

  it('should call onClose when close button is clicked', async () => {
    const user = userEvent.setup();
    render(<POIDetailsPanel {...defaultProps} />);
    
    const closeButton = screen.getByLabelText(/close/i);
    await user.click(closeButton);
    
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should call onClose when escape key is pressed', async () => {
    const user = userEvent.setup();
    render(<POIDetailsPanel {...defaultProps} />);
    
    await user.keyboard('{Escape}');
    
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should show loading state when joining', () => {
    render(<POIDetailsPanel {...defaultProps} isLoading={true} />);
    
    const joinButton = screen.getByText(/joining.../i);
    expect(joinButton).toBeDisabled();
  });

  it('should show loading state when leaving', () => {
    render(<POIDetailsPanel {...defaultProps} isUserParticipant={true} isLoading={true} />);
    
    const leaveButton = screen.getByText(/leaving.../i);
    expect(leaveButton).toBeDisabled();
  });

  it('should display coordinates', () => {
    render(<POIDetailsPanel {...defaultProps} />);
    
    expect(screen.getByText(/40\.7128/)).toBeInTheDocument();
    expect(screen.getByText(/-74\.0060/)).toBeInTheDocument();
  });

  it('should handle empty participant list', () => {
    const emptyPOI = {
      ...mockPOI,
      participantCount: 0,
      participants: []
    };

    render(<POIDetailsPanel {...defaultProps} poi={emptyPOI} />);
    
    expect(screen.getByText('0/10 participants')).toBeInTheDocument();
    expect(screen.getByText(/no participants yet/i)).toBeInTheDocument();
  });

  it('should highlight current user in participant list', () => {
    const poiWithCurrentUser = {
      ...mockPOI,
      participants: [
        ...mockPOI.participants,
        { id: 'user-4', name: 'Current User' }
      ]
    };

    render(<POIDetailsPanel {...defaultProps} poi={poiWithCurrentUser} isUserParticipant={true} />);
    
    const currentUserElement = screen.getByTestId('current-user');
    expect(currentUserElement).toBeInTheDocument();
    expect(currentUserElement).toHaveTextContent('Current User (You)');
  });

  it('should show capacity warning when near full', () => {
    const nearFullPOI = {
      ...mockPOI,
      participantCount: 9,
      participants: Array.from({ length: 9 }, (_, i) => ({
        id: `user-${i + 1}`,
        name: `User ${i + 1}`
      }))
    };

    render(<POIDetailsPanel {...defaultProps} poi={nearFullPOI} />);
    
    expect(screen.getByText(/almost full/i)).toBeInTheDocument();
  });
});