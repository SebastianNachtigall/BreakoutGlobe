import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import LoginModal from '../LoginModal';

describe('LoginModal', () => {
  const mockProps = {
    isOpen: true,
    onClose: vi.fn(),
    onLogin: vi.fn(),
    onSwitchToSignup: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render when isOpen is true', () => {
      render(<LoginModal {...mockProps} />);
      expect(screen.getByRole('heading', { name: /login/i })).toBeInTheDocument();
    });

    it('should not render when isOpen is false', () => {
      render(<LoginModal {...mockProps} isOpen={false} />);
      expect(screen.queryByText('Login')).not.toBeInTheDocument();
    });

    it('should render email and password fields', () => {
      render(<LoginModal {...mockProps} />);
      
      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    });

    it('should render login button', () => {
      render(<LoginModal {...mockProps} />);
      expect(screen.getByRole('button', { name: /^login$/i })).toBeInTheDocument();
    });

    it('should render link to switch to signup', () => {
      render(<LoginModal {...mockProps} />);
      expect(screen.getByText(/don't have an account/i)).toBeInTheDocument();
    });

    it('should render show/hide password toggle', () => {
      render(<LoginModal {...mockProps} />);
      expect(screen.getByText(/show/i)).toBeInTheDocument();
    });
  });

  describe('Form Validation', () => {
    it('should show error for invalid email format', async () => {
      render(<LoginModal {...mockProps} />);
      
      const emailInput = screen.getByLabelText(/email/i);
      fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
      fireEvent.blur(emailInput);
      
      await waitFor(() => {
        expect(screen.getByText(/valid email/i)).toBeInTheDocument();
      });
    });

    it('should show error for empty password', async () => {
      render(<LoginModal {...mockProps} />);
      
      const passwordInput = screen.getByLabelText(/password/i);
      fireEvent.change(passwordInput, { target: { value: '' } });
      fireEvent.blur(passwordInput);
      
      await waitFor(() => {
        expect(screen.getByText(/password is required/i)).toBeInTheDocument();
      });
    });
  });

  describe('Form Submission', () => {
    it('should call onLogin with credentials when valid', async () => {
      render(<LoginModal {...mockProps} />);
      
      fireEvent.change(screen.getByLabelText(/email/i), { 
        target: { value: 'test@example.com' } 
      });
      fireEvent.change(screen.getByLabelText(/password/i), { 
        target: { value: 'password123' } 
      });
      
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      await waitFor(() => {
        expect(mockProps.onLogin).toHaveBeenCalledWith('test@example.com', 'password123');
      });
    });

    it('should not submit when form is invalid', async () => {
      render(<LoginModal {...mockProps} />);
      
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      await waitFor(() => {
        expect(mockProps.onLogin).not.toHaveBeenCalled();
      });
    });
  });

  describe('User Interactions', () => {
    it('should toggle password visibility', () => {
      render(<LoginModal {...mockProps} />);
      
      const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement;
      const toggleButton = screen.getByText(/show/i);
      
      expect(passwordInput.type).toBe('password');
      
      fireEvent.click(toggleButton);
      expect(passwordInput.type).toBe('text');
      expect(screen.getByText(/hide/i)).toBeInTheDocument();
      
      fireEvent.click(screen.getByText(/hide/i));
      expect(passwordInput.type).toBe('password');
    });

    it('should call onSwitchToSignup when signup link is clicked', () => {
      render(<LoginModal {...mockProps} />);
      
      const signupLink = screen.getByText(/sign up/i);
      fireEvent.click(signupLink);
      
      expect(mockProps.onSwitchToSignup).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when close button is clicked', () => {
      render(<LoginModal {...mockProps} />);
      
      const closeButton = screen.getByRole('button', { name: /close/i });
      fireEvent.click(closeButton);
      
      expect(mockProps.onClose).toHaveBeenCalledTimes(1);
    });
  });
});
