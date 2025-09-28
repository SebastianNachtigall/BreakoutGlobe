import { render, screen, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect } from 'vitest';
import { GroupCallModal } from '../GroupCallModal';

describe('GroupCallModal', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    callState: 'connecting' as const,
    poiId: 'poi-123',
    poiName: 'Test POI',
    onEndCall: vi.fn()
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('should render when open', () => {
      render(<GroupCallModal {...defaultProps} />);
      
      expect(screen.getByText('POI Group Call')).toBeInTheDocument();
      expect(screen.getByText('Test POI')).toBeInTheDocument();
      expect(screen.getAllByText('Connecting to group call...')).toHaveLength(2); // Header and content
    });

    it('should not render when closed', () => {
      render(<GroupCallModal {...defaultProps} isOpen={false} />);
      
      expect(screen.queryByText('POI Group Call')).not.toBeInTheDocument();
    });

    it('should show POI ID when no name provided', () => {
      render(<GroupCallModal {...defaultProps} poiName={undefined} />);
      
      expect(screen.getByText('POI poi-123')).toBeInTheDocument();
    });

    it('should show basic group call message', () => {
      render(<GroupCallModal {...defaultProps} />);
      
      expect(screen.getByText('Group call active in this POI! Video functionality coming soon.')).toBeInTheDocument();
    });
  });

  describe('call states', () => {
    it('should show connecting state', () => {
      render(<GroupCallModal {...defaultProps} callState="connecting" />);
      
      expect(screen.getAllByText('Connecting to group call...')).toHaveLength(2);
      expect(screen.getByTitle('Leave call')).toBeInTheDocument();
    });

    it('should show connected state', () => {
      render(<GroupCallModal {...defaultProps} callState="connected" />);
      
      expect(screen.getAllByText('Group call active')).toHaveLength(2);
      expect(screen.getByTitle('Leave call')).toBeInTheDocument();
    });

    it('should show ended state', () => {
      render(<GroupCallModal {...defaultProps} callState="ended" />);
      
      expect(screen.getAllByText('Group call ended')).toHaveLength(2);
      expect(screen.getByRole('button', { name: /close/i })).toBeInTheDocument();
    });

    it('should show loading animation for connecting state', () => {
      render(<GroupCallModal {...defaultProps} callState="connecting" />);
      
      // Check for loading spinner (has animate-spin class)
      const spinner = document.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('interactions', () => {
    it('should call onClose when close button clicked', () => {
      const onClose = vi.fn();
      render(<GroupCallModal {...defaultProps} onClose={onClose} />);
      
      const closeButton = screen.getByTitle('Close');
      fireEvent.click(closeButton);
      
      expect(onClose).toHaveBeenCalledOnce();
    });

    it('should call onEndCall when leave call button clicked', () => {
      const onEndCall = vi.fn();
      render(<GroupCallModal {...defaultProps} onEndCall={onEndCall} callState="connecting" />);
      
      const leaveButton = screen.getByTitle('Leave call');
      fireEvent.click(leaveButton);
      
      expect(onEndCall).toHaveBeenCalledOnce();
    });

    it('should call onClose when close button clicked in ended state', () => {
      const onClose = vi.fn();
      render(<GroupCallModal {...defaultProps} onClose={onClose} callState="ended" />);
      
      const closeButton = screen.getByRole('button', { name: /close/i });
      fireEvent.click(closeButton);
      
      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA labels', () => {
      render(<GroupCallModal {...defaultProps} />);
      
      expect(screen.getByTitle('Close')).toBeInTheDocument();
      expect(screen.getByTitle('Leave call')).toBeInTheDocument();
    });

    it('should have proper modal structure', () => {
      render(<GroupCallModal {...defaultProps} />);
      
      // Modal should have high z-index for overlay
      const modal = document.querySelector('.fixed.inset-0');
      expect(modal).toBeInTheDocument();
      expect(modal).toHaveClass('z-[9999]');
    });
  });

  describe('dual peer video support', () => {
    const mockParticipants = new Map([
      ['user-456', { userId: 'user-456', displayName: 'Alice', avatarURL: 'https://example.com/alice.jpg' }],
      ['user-789', { userId: 'user-789', displayName: 'Bob' }]
    ]);

    const mockStreams = new Map([
      ['user-456', {} as MediaStream],
      ['user-789', {} as MediaStream]
    ]);

    const propsWithVideo = {
      ...defaultProps,
      callState: 'connected' as const,
      participants: mockParticipants,
      remoteStreams: mockStreams,
      localStream: {} as MediaStream,
      isAudioEnabled: true,
      isVideoEnabled: true,
      onToggleAudio: vi.fn(),
      onToggleVideo: vi.fn()
    };

    it('should show two video streams side by side when connected', () => {
      render(<GroupCallModal {...propsWithVideo} />);
      
      // Should show video grid container
      const videoGrid = document.querySelector('.grid');
      expect(videoGrid).toBeInTheDocument();
      
      // Should show participant names
      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.getByText('Bob')).toBeInTheDocument();
    });

    it('should show local video in picture-in-picture style', () => {
      render(<GroupCallModal {...propsWithVideo} />);
      
      // Should have local video container
      const localVideo = document.querySelector('[data-testid="local-video"]');
      expect(localVideo).toBeInTheDocument();
    });

    it('should show audio/video controls when connected', () => {
      render(<GroupCallModal {...propsWithVideo} />);
      
      expect(screen.getByTitle('Mute')).toBeInTheDocument();
      expect(screen.getByTitle('Turn off camera')).toBeInTheDocument();
      expect(screen.getByTitle('Leave call')).toBeInTheDocument();
    });

    it('should call toggle functions when controls clicked', () => {
      const onToggleAudio = vi.fn();
      const onToggleVideo = vi.fn();
      
      render(<GroupCallModal {...propsWithVideo} onToggleAudio={onToggleAudio} onToggleVideo={onToggleVideo} />);
      
      fireEvent.click(screen.getByTitle('Mute'));
      expect(onToggleAudio).toHaveBeenCalledOnce();
      
      fireEvent.click(screen.getByTitle('Turn off camera'));
      expect(onToggleVideo).toHaveBeenCalledOnce();
    });

    it('should show placeholder when no video stream available', () => {
      const propsWithoutStreams = {
        ...propsWithVideo,
        remoteStreams: new Map()
      };
      
      render(<GroupCallModal {...propsWithoutStreams} />);
      
      const waitingTexts = screen.getAllByText('Waiting for video...');
      expect(waitingTexts).toHaveLength(2); // One for each participant
      expect(waitingTexts[0]).toBeInTheDocument();
    });
  });
});