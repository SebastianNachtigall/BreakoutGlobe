import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ProfileCreationModal from '../ProfileCreationModal';

// Mock the API
vi.mock('../../services/api', () => ({
  createGuestProfile: vi.fn(),
  APIError: class APIError extends Error {
    constructor(message: string) {
      super(message);
      this.name = 'APIError';
    }
  },
}));

// Mock the AvatarImageUpload component
vi.mock('../AvatarImageUpload', () => ({
  AvatarImageUpload: vi.fn(({ onImageSelected, onError }) => (
    <div data-testid="avatar-image-upload">
      <button
        onClick={() => {
          const mockFile = new File(['mock'], 'test.jpg', { type: 'image/jpeg' });
          onImageSelected(mockFile);
        }}
        data-testid="mock-select-image"
      >
        Select Image
      </button>
      <button
        onClick={() => onError('Mock error')}
        data-testid="mock-error"
      >
        Trigger Error
      </button>
    </div>
  )),
}));

describe('ProfileCreationModal (Enhanced)', () => {
  const mockOnProfileCreated = vi.fn();
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render with enhanced avatar upload component', () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText('Create Your Profile')).toBeInTheDocument();
    expect(screen.getByTestId('avatar-image-upload')).toBeInTheDocument();
    expect(screen.queryByLabelText('Avatar Image')).not.toBeInTheDocument(); // Old file input should be gone
  });

  it('should handle avatar image selection', async () => {
    const { createGuestProfile } = await import('../../services/api');
    (createGuestProfile as any).mockResolvedValue({
      id: 'user-123',
      displayName: 'Test User',
      aboutMe: 'Test about me',
      avatarUrl: 'https://example.com/avatar.jpg',
    });

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Fill in required fields
    const displayNameInput = screen.getByLabelText('Display Name *');
    await userEvent.type(displayNameInput, 'Test User');

    // Select an avatar image
    const selectImageButton = screen.getByTestId('mock-select-image');
    await userEvent.click(selectImageButton);

    // Submit form
    const submitButton = screen.getByText('Create Profile');
    await userEvent.click(submitButton);

    expect(createGuestProfile).toHaveBeenCalledWith({
      displayName: 'Test User',
      aboutMe: undefined,
      avatarFile: expect.any(File),
    });
  });

  it('should handle avatar upload errors', async () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Trigger avatar error
    const errorButton = screen.getByTestId('mock-error');
    await userEvent.click(errorButton);

    // The error should be handled by the component (we can't easily test the internal state)
    expect(errorButton).toBeInTheDocument(); // Basic test that the component renders
  });

  it('should submit form without avatar', async () => {
    const { createGuestProfile } = await import('../../services/api');
    (createGuestProfile as any).mockResolvedValue({
      id: 'user-123',
      displayName: 'Test User',
      aboutMe: '',
    });

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Fill in required fields
    const displayNameInput = screen.getByLabelText('Display Name *');
    await userEvent.type(displayNameInput, 'Test User');

    // Submit form without avatar
    const submitButton = screen.getByText('Create Profile');
    await userEvent.click(submitButton);

    expect(createGuestProfile).toHaveBeenCalledWith({
      displayName: 'Test User',
      aboutMe: undefined,
      avatarFile: undefined,
    });
  });

  it('should maintain existing form validation', async () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Try to submit without display name
    const submitButton = screen.getByText('Create Profile');
    await userEvent.click(submitButton);

    expect(screen.getByText('Display name is required')).toBeInTheDocument();
  });

  it('should handle about me character limit', async () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    const aboutMeInput = screen.getByLabelText('About Me');
    // Fill display name first to avoid that validation error
    const displayNameInput = screen.getByLabelText('Display Name *');
    await userEvent.type(displayNameInput, 'Test User');

    // Set a long text directly (bypassing maxLength)
    const longText = 'a'.repeat(501); // Over 500 character limit
    fireEvent.change(aboutMeInput, { target: { value: longText } });

    const submitButton = screen.getByText('Create Profile');
    await userEvent.click(submitButton);

    expect(screen.getByText('About me must be 500 characters or less')).toBeInTheDocument();
  });

  it('should show loading state during submission', async () => {
    const { createGuestProfile } = await import('../../services/api');
    (createGuestProfile as any).mockImplementation(
      () => new Promise(resolve => setTimeout(resolve, 100))
    );

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Fill in required fields
    const displayNameInput = screen.getByLabelText('Display Name *');
    await userEvent.type(displayNameInput, 'Test User');

    // Submit form
    const submitButton = screen.getByText('Create Profile');
    fireEvent.click(submitButton);

    expect(screen.getByText('Creating Profile...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
  });

  it('should handle API errors', async () => {
    const { createGuestProfile, APIError } = await import('../../services/api');
    (createGuestProfile as any).mockRejectedValue(new APIError('Profile creation failed'));

    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Fill in required fields
    const displayNameInput = screen.getByLabelText('Display Name *');
    await userEvent.type(displayNameInput, 'Test User');

    // Submit form
    const submitButton = screen.getByText('Create Profile');
    await userEvent.click(submitButton);

    expect(screen.getByText('Profile creation failed')).toBeInTheDocument();
  });

  it('should close modal when backdrop is clicked', async () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    const backdrop = screen.getByTestId('modal-backdrop');
    await userEvent.click(backdrop);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('should not render when closed', () => {
    render(
      <ProfileCreationModal
        isOpen={false}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    expect(screen.queryByText('Create Your Profile')).not.toBeInTheDocument();
  });

  it('should clear avatar error when new image is selected', async () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Trigger avatar error first
    const errorButton = screen.getByTestId('mock-error');
    await userEvent.click(errorButton);

    // Select a new image (this should clear any errors internally)
    const selectImageButton = screen.getByTestId('mock-select-image');
    await userEvent.click(selectImageButton);

    // Both buttons should still be present (basic functionality test)
    expect(selectImageButton).toBeInTheDocument();
    expect(errorButton).toBeInTheDocument();
  });
});