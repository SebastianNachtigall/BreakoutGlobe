import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import WelcomeScreen from '../WelcomeScreen';

describe('WelcomeScreen Authentication', () => {
  const mockProps = {
    isOpen: true,
    onCreateProfile: vi.fn(),
    onSignup: vi.fn(),
    onLogin: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render all three authentication options', () => {
    render(<WelcomeScreen {...mockProps} />);
    
    expect(screen.getByText('Create Full Account')).toBeInTheDocument();
    expect(screen.getByText('Login')).toBeInTheDocument();
    expect(screen.getByText('Continue as Guest')).toBeInTheDocument();
  });

  it('should call onSignup when Create Full Account is clicked', () => {
    render(<WelcomeScreen {...mockProps} />);
    
    fireEvent.click(screen.getByText('Create Full Account'));
    expect(mockProps.onSignup).toHaveBeenCalledTimes(1);
  });

  it('should call onLogin when Login is clicked', () => {
    render(<WelcomeScreen {...mockProps} />);
    
    fireEvent.click(screen.getByText('Login'));
    expect(mockProps.onLogin).toHaveBeenCalledTimes(1);
  });

  it('should call onCreateProfile when Continue as Guest is clicked', () => {
    render(<WelcomeScreen {...mockProps} />);
    
    fireEvent.click(screen.getByText('Continue as Guest'));
    expect(mockProps.onCreateProfile).toHaveBeenCalledTimes(1);
  });
});
