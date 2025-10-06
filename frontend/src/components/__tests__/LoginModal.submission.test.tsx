import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import LoginModal from '../LoginModal';

describe('LoginModal - Submission Integration', () => {
  const mockProps = {
    isOpen: true,
    onClose: vi.fn(),
    onLogin: vi.fn(),
    onSwitchToSignup: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading state during submission', async () => {
      const slowLogin = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)));
      render(<LoginModal {...mockProps} onLogin={slowLogin} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'password123' } });
      
      // Submit
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      // Should show loading state
      await waitFor(() => {
        expect(screen.getByText(/logging in/i)).toBeInTheDocument();
      });
    });

    it('should disable submit button during loading', async () => {
      const slowLogin = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)));
      render(<LoginModal {...mockProps} onLogin={slowLogin} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'password123' } });
      
      const submitButton = screen.getByRole('button', { name: /^login$/i });
      fireEvent.click(submitButton);
      
      await waitFor(() => {
        expect(submitButton).toBeDisabled();
      });
    });
  });

  describe('Error Handling', () => {
    it('should display error message when login fails', async () => {
      const failingLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'));
      render(<LoginModal {...mockProps} onLogin={failingLogin} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'wrongpassword' } });
      
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
      });
    });

    it('should display generic error for network failures', async () => {
      const failingLogin = vi.fn().mockRejectedValue(new Error('Network error'));
      render(<LoginModal {...mockProps} onLogin={failingLogin} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'password123' } });
      
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });

    it('should clear error when user starts typing', async () => {
      const failingLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'));
      render(<LoginModal {...mockProps} onLogin={failingLogin} />);
      
      // Fill and submit
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'wrongpassword' } });
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
      });
      
      // Start typing
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'newpassword' } });
      
      await waitFor(() => {
        expect(screen.queryByText(/invalid credentials/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Success Handling', () => {
    it('should call onClose after successful login', async () => {
      const successfulLogin = vi.fn().mockResolvedValue({ success: true });
      render(<LoginModal {...mockProps} onLogin={successfulLogin} />);
      
      // Fill form
      fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
      fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'password123' } });
      
      fireEvent.click(screen.getByRole('button', { name: /^login$/i }));
      
      await waitFor(() => {
        expect(mockProps.onClose).toHaveBeenCalled();
      });
    });
  });
});
