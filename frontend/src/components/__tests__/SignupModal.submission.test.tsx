import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import SignupModal from '../SignupModal';

describe('SignupModal - Submission Integration', () => {
  const mockProps = {
    isOpen: true,
    onClose: vi.fn(),
    onSignup: vi.fn(),
    onSwitchToLogin: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading state during submission', async () => {
      const slowSignup = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)));
      render(<SignupModal {...mockProps} onSignup={slowSignup} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/display name/i), { target: { value: 'Test User' } });
      
      // Submit
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      // Should show loading state
      await waitFor(() => {
        expect(screen.getByText(/signing up/i)).toBeInTheDocument();
      });
    });

    it('should disable submit button during loading', async () => {
      const slowSignup = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)));
      render(<SignupModal {...mockProps} onSignup={slowSignup} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/display name/i), { target: { value: 'Test User' } });
      
      const submitButton = screen.getByRole('button', { name: /sign up/i });
      fireEvent.click(submitButton);
      
      await waitFor(() => {
        expect(submitButton).toBeDisabled();
      });
    });
  });

  describe('Error Handling', () => {
    it('should display error message when signup fails', async () => {
      const failingSignup = vi.fn().mockRejectedValue(new Error('Email already exists'));
      render(<SignupModal {...mockProps} onSignup={failingSignup} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/display name/i), { target: { value: 'Test User' } });
      
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/email already exists/i)).toBeInTheDocument();
      });
    });

    it('should display generic error for network failures', async () => {
      const failingSignup = vi.fn().mockRejectedValue(new Error('Network error'));
      render(<SignupModal {...mockProps} onSignup={failingSignup} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/display name/i), { target: { value: 'Test User' } });
      
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });

    it('should clear error when user starts typing', async () => {
      const failingSignup = vi.fn().mockRejectedValue(new Error('Email already exists'));
      render(<SignupModal {...mockProps} onSignup={failingSignup} />);
      
      // Fill and submit
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/display name/i), { target: { value: 'Test User' } });
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/email already exists/i)).toBeInTheDocument();
      });
      
      // Start typing
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test2@example.com' } });
      
      await waitFor(() => {
        expect(screen.queryByText(/email already exists/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Success Handling', () => {
    it('should call onClose after successful signup', async () => {
      const successfulSignup = vi.fn().mockResolvedValue({ success: true });
      render(<SignupModal {...mockProps} onSignup={successfulSignup} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/confirm password/i), { target: { value: 'Password123!' } });
      fireEvent.change(screen.getByLabelText(/display name/i), { target: { value: 'Test User' } });
      
      fireEvent.click(screen.getByRole('button', { name: /sign up/i }));
      
      await waitFor(() => {
        expect(mockProps.onClose).toHaveBeenCalled();
      });
    });
  });
});
