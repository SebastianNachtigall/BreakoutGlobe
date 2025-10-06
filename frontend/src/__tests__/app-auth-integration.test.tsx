import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { authStore } from '../stores/authStore';
import SignupModal from '../components/SignupModal';
import LoginModal from '../components/LoginModal';

// Mock components
vi.mock('../components/SignupModal', () => ({
  default: ({ isOpen, onClose, onSignup, onSwitchToLogin }: any) => 
    isOpen ? (
      <div data-testid="signup-modal">
        <button onClick={() => onSignup({ email: 'test@example.com', password: 'Pass123!', displayName: 'Test' })}>
          Submit Signup
        </button>
        <button onClick={onSwitchToLogin}>Switch to Login</button>
        <button onClick={onClose}>Close</button>
      </div>
    ) : null
}));

vi.mock('../components/LoginModal', () => ({
  default: ({ isOpen, onClose, onLogin, onSwitchToSignup }: any) => 
    isOpen ? (
      <div data-testid="login-modal">
        <button onClick={() => onLogin('test@example.com', 'password')}>
          Submit Login
        </button>
        <button onClick={onSwitchToSignup}>Switch to Signup</button>
        <button onClick={onClose}>Close</button>
      </div>
    ) : null
}));

// Simple test component that mimics App's auth modal logic
function AuthModalTestComponent() {
  const [showSignup, setShowSignup] = React.useState(false);
  const [showLogin, setShowLogin] = React.useState(false);
  const { signup, login } = authStore();

  const handleSignup = async (data: any) => {
    await signup(data);
    setShowSignup(false);
  };

  const handleLogin = async (email: string, password: string) => {
    await login(email, password);
    setShowLogin(false);
  };

  return (
    <div>
      <button onClick={() => setShowSignup(true)}>Open Signup</button>
      <button onClick={() => setShowLogin(true)}>Open Login</button>
      
      <SignupModal
        isOpen={showSignup}
        onClose={() => setShowSignup(false)}
        onSignup={handleSignup}
        onSwitchToLogin={() => {
          setShowSignup(false);
          setShowLogin(true);
        }}
      />
      
      <LoginModal
        isOpen={showLogin}
        onClose={() => setShowLogin(false)}
        onLogin={handleLogin}
        onSwitchToSignup={() => {
          setShowLogin(false);
          setShowSignup(true);
        }}
      />
    </div>
  );
}

describe('App Auth Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    authStore.getState().logout();
  });

  describe('Modal State Management', () => {
    it('should open signup modal when button is clicked', () => {
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Signup'));
      
      expect(screen.getByTestId('signup-modal')).toBeInTheDocument();
    });

    it('should open login modal when button is clicked', () => {
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Login'));
      
      expect(screen.getByTestId('login-modal')).toBeInTheDocument();
    });

    it('should switch from signup to login', () => {
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Signup'));
      expect(screen.getByTestId('signup-modal')).toBeInTheDocument();
      
      fireEvent.click(screen.getByText('Switch to Login'));
      
      expect(screen.queryByTestId('signup-modal')).not.toBeInTheDocument();
      expect(screen.getByTestId('login-modal')).toBeInTheDocument();
    });

    it('should switch from login to signup', () => {
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Login'));
      expect(screen.getByTestId('login-modal')).toBeInTheDocument();
      
      fireEvent.click(screen.getByText('Switch to Signup'));
      
      expect(screen.queryByTestId('login-modal')).not.toBeInTheDocument();
      expect(screen.getByTestId('signup-modal')).toBeInTheDocument();
    });

    it('should close signup modal', () => {
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Signup'));
      expect(screen.getByTestId('signup-modal')).toBeInTheDocument();
      
      fireEvent.click(screen.getByText('Close'));
      
      expect(screen.queryByTestId('signup-modal')).not.toBeInTheDocument();
    });

    it('should close login modal', () => {
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Login'));
      expect(screen.getByTestId('login-modal')).toBeInTheDocument();
      
      fireEvent.click(screen.getByText('Close'));
      
      expect(screen.queryByTestId('login-modal')).not.toBeInTheDocument();
    });
  });

  describe('Auth Flow Integration', () => {
    it('should close signup modal after successful signup', async () => {
      // Mock successful signup
      vi.spyOn(authStore.getState(), 'signup').mockResolvedValue();
      
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Signup'));
      fireEvent.click(screen.getByText('Submit Signup'));
      
      await waitFor(() => {
        expect(screen.queryByTestId('signup-modal')).not.toBeInTheDocument();
      });
    });

    it('should close login modal after successful login', async () => {
      // Mock successful login
      vi.spyOn(authStore.getState(), 'login').mockResolvedValue();
      
      render(<AuthModalTestComponent />);
      
      fireEvent.click(screen.getByText('Open Login'));
      fireEvent.click(screen.getByText('Submit Login'));
      
      await waitFor(() => {
        expect(screen.queryByTestId('login-modal')).not.toBeInTheDocument();
      });
    });
  });
});
