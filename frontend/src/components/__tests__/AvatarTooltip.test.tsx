import React from 'react';
import { render, screen } from '@testing-library/react';
import { AvatarTooltip } from '../AvatarTooltip';

describe('AvatarTooltip', () => {
  const defaultProps = {
    isOpen: true,
    position: { x: 100, y: 100 },
    avatar: {
      sessionId: 'session-123',
      userId: 'user-123',
      displayName: 'Test User',
      avatarURL: undefined,
      aboutMe: 'Test bio',
      isInCall: false
    },
    onClose: vi.fn(),
    onStartCall: vi.fn()
  };

  it('renders avatar tooltip with call button enabled when user is not in call', () => {
    render(<AvatarTooltip {...defaultProps} />);
    
    expect(screen.getByText('Test User')).toBeInTheDocument();
    expect(screen.getByText('Start Video Call')).toBeInTheDocument();
    
    const callButton = screen.getByRole('button', { name: /start video call/i });
    expect(callButton).not.toBeDisabled();
    expect(callButton).toHaveClass('bg-blue-600');
  });

  it('renders avatar tooltip with call button disabled when user is in call', () => {
    const propsWithInCall = {
      ...defaultProps,
      avatar: {
        ...defaultProps.avatar,
        isInCall: true
      }
    };

    render(<AvatarTooltip {...propsWithInCall} />);
    
    expect(screen.getByText('Test User')).toBeInTheDocument();
    expect(screen.getByText('In Call')).toBeInTheDocument();
    
    const callButton = screen.getByRole('button', { name: /in call/i });
    expect(callButton).toBeDisabled();
    expect(callButton).toHaveClass('bg-gray-300', 'cursor-not-allowed');
  });

  it('handles missing displayName gracefully', () => {
    const propsWithoutName = {
      ...defaultProps,
      avatar: {
        ...defaultProps.avatar,
        displayName: undefined
      }
    };

    render(<AvatarTooltip {...propsWithoutName} />);
    
    expect(screen.getByText('Unknown User')).toBeInTheDocument();
  });

  it('does not render when isOpen is false', () => {
    const closedProps = {
      ...defaultProps,
      isOpen: false
    };

    const { container } = render(<AvatarTooltip {...closedProps} />);
    expect(container.firstChild).toBeNull();
  });
});