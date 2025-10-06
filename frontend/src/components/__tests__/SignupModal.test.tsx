import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import SignupModal from '../SignupModal';

describe('SignupModal', () => {
  const mockProps = {
    isOpen: true,
    onClose: vi.fn(),
    onSignup: vi.fn(),
    onSwitchToLogin: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render when isOpen is true', () => {
      render(<SignupModal {...mockProps} />);
      expect(screen.getByText('Create Full Account')).toBeInTheDocument();
    });

    it('should not render when isOpen is false', () => {
      render(<SignupModal {...mockProps} isOpen={false} />);
      expect(screen.queryByText('Create Full Account')).not.toBeInTheDocument();
    });

    it('should render all form fields', () => {
      render(<SignupModal {...mockProps} />);
      
      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/display name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/about me/i)).toBeInTheDocument();
    });

    it('should render submit button', () => {
      render(<SignupModal {...mockProps} />);
      expect(screen.getByRole('button', { name: /sign up/i })).toBeInTheDocument();
    });

    it('should render link to switch to login', () => {
      render(<SignupModal {...mockProps} />);
      expect(screen.getByText(/already have an account/i)).toBeInTheDocument();
    });
  });

  describe('Form Validation', () => {
    it('should show error for invalid email format', async () => {
      render(<SignupModal {...mockProps} />);
      
      const emailInput = screen.getByLabelText(/email/i);
      fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
      fireEvent.blur(emailInput);
      
      await waitFor(() => {
        expect(screen.getByText(/valid email/i)).toBeInTheDocument();
      });
    });

    it('should show error for weak password', async () => {
      render(<SignupModal {...mockProps} />);
      
      const passwordInput = screen.getByLabelText(/^password$/i);
      fireEvent.change(passwordInput, { target: { value: 'weak' } });
      fireEvent.blur(passwordInput);
      
      await waitFor(() => {
        expect(screen.getByText(/at least 8 characters/i)).toBeInTheDocument();
      });
    });

    it('should show error when passwords do not match', async () => {
      render(<SignupModal {...mockProps} />);
      
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmInput = screen.getByLabelText(/confirm password/i);
      
      fireEvent.change(passwordInput, { target: { value: 'Password123!' } });
      fireEvent.change(confirmInput, { target: { value: 'Different123!' } });
      fireEvent.blur(confirmInput);
      
      await waitFor(() => {
        expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument();
      });
    });

    it('should show error for display name too short', async () => {
      render(<SignupModal {...mockProps} />);
      
      const nameInput = screen.getByLabelText(/display name/i);
      fireEvent.change(nameInput, { target: { value: 'ab' } });
      fireEvent.blur(nameInput);
      
      await waitFor(() => {
        expect(screen.getByText(/at least 3 characters/i)).toBeInTheDocument();
      });
    });
  });

  describe('Form Submission', () => {
    it('should call onSignup with form data when valid', async () => {
      render(<SignupModal {...mockProps} />);
      
      fireEvent.change(screen.getByLabelText(/email/i), { 
        target: { value: 'test@example.com' } 
      });
      fireEvent.change(screen.getByLabelText(/^password$/i), { 
        target: { value: 'Password123!' } 
      });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { 
        target: { value: 'Password123!' } 
      });
      fireEvent.change(screen.getByLabelText(/display name/i), { 
        target: { value: 'Test User' } 
      });
      fireEvent.change(screen.getByLabelText(/about me/i), { 
        target: { value: 'Test bio' } 
      });
      
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      await waitFor(() => {
        expect(mockProps.onSignup).toHaveBeenCalledWith({
          email: 'test@example.com',
          password: 'Password123!',
          displayName: 'Test User',
          aboutMe: 'Test bio',
        });
      });
    });

    it('should not submit when form is invalid', async () => {
      render(<SignupModal {...mockProps} />);
      
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      await waitFor(() => {
        expect(mockProps.onSignup).not.toHaveBeenCalled();
      });
    });
  });

  describe('User Interactions', () => {
    it('should call onSwitchToLogin when login link is clicked', () => {
      render(<SignupModal {...mockProps} />);
      
      const loginLink = screen.getByText(/login/i);
      fireEvent.click(loginLink);
      
      expect(mockProps.onSwitchToLogin).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when close button is clicked', () => {
      render(<SignupModal {...mockProps} />);
      
      const closeButton = screen.getByRole('button', { name: /close/i });
      fireEvent.click(closeButton);
      
      expect(mockProps.onClose).toHaveBeenCalledTimes(1);
    });
  });
});
