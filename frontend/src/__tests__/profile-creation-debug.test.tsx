import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import userEvent from '@testing-library/user-event';
import ProfileCreationModal from '../components/ProfileCreationModal';

// Mock the API
vi.mock('../services/api', () => ({
  createGuestProfile: vi.fn(),
  APIError: class APIError extends Error {
    constructor(message: string) {
      super(message);
      this.name = 'APIError';
    }
  },
}));

describe('Profile Creation Debug Test', () => {
  const mockOnProfileCreated = vi.fn();
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should handle form submission correctly', async () => {
    const { createGuestProfile } = await import('../services/api');
    
    // Mock successful profile creation
    (createGuestProfile as any).mockResolvedValue({
      id: 'user-123',
      displayName: 'Test User',
      aboutMe: 'Test about me',
      avatarUrl: null,
      createdAt: new Date().toISOString(),
    });

    const user = userEvent.setup();

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Fill in the form
    const displayNameInput = screen.getByLabelText(/display name/i);
    await user.type(displayNameInput, 'Test User');

    const aboutMeInput = screen.getByLabelText(/about me/i);
    await user.type(aboutMeInput, 'Test about me');

    // Submit the form
    const submitButton = screen.getByRole('button', { name: /create profile/i });
    await user.click(submitButton);

    // Wait for the API call
    await waitFor(() => {
      expect(createGuestProfile).toHaveBeenCalledWith({
        displayName: 'Test User',
        aboutMe: 'Test about me',
        avatarFile: undefined,
      });
    });

    // Should call onProfileCreated
    await waitFor(() => {
      expect(mockOnProfileCreated).toHaveBeenCalledWith({
        id: 'user-123',
        displayName: 'Test User',
        aboutMe: 'Test about me',
        avatarUrl: null,
        createdAt: expect.any(String),
      });
    });
  });

  it('should show error when API fails', async () => {
    const { createGuestProfile, APIError } = await import('../services/api');
    
    // Mock API failure
    (createGuestProfile as any).mockRejectedValue(new APIError('Server error'));

    const user = userEvent.setup();

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Fill in the form
    const displayNameInput = screen.getByLabelText(/display name/i);
    await user.type(displayNameInput, 'Test User');

    // Submit the form
    const submitButton = screen.getByRole('button', { name: /create profile/i });
    await user.click(submitButton);

    // Should show error message
    await waitFor(() => {
      expect(screen.getByText('Server error')).toBeInTheDocument();
    });

    // Should not call onProfileCreated
    expect(mockOnProfileCreated).not.toHaveBeenCalled();
  });

  it('should show validation error for empty display name', async () => {
    const user = userEvent.setup();

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Try to submit without filling display name
    const submitButton = screen.getByRole('button', { name: /create profile/i });
    await user.click(submitButton);

    // Should show validation error
    await waitFor(() => {
      expect(screen.getByText('Display name is required')).toBeInTheDocument();
    });

    // Should not call API or onProfileCreated
    const { createGuestProfile } = await import('../services/api');
    expect(createGuestProfile).not.toHaveBeenCalled();
    expect(mockOnProfileCreated).not.toHaveBeenCalled();
  });
});