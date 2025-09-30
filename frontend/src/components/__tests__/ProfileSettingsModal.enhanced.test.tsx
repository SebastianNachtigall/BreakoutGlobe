import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import ProfileSettingsModal from '../ProfileSettingsModal';
import { userProfileStore } from '../../stores/userProfileStore';
import * as api from '../../services/api';

// Mock the stores and services
vi.mock('../../stores/userProfileStore');
vi.mock('../../services/api');
vi.mock('../AvatarImageUpload', () => ({
  AvatarImageUpload: ({ onImageSelected, onError, currentAvatarUrl, disabled }: any) => (
    <div data-testid="avatar-upload">
      <div data-testid="current-avatar-url">{currentAvatarUrl || 'no-avatar'}</div>
      <div data-testid="upload-disabled">{disabled ? 'disabled' : 'enabled'}</div>
      <button
        data-testid="mock-select-image"
        onClick={() => {
          const mockFile = new File(['test'], 'test.jpg', { type: 'image/jpeg' });
          onImageSelected(mockFile);
        }}
      >
        Select Image
      </button>
      <button
        data-testid="mock-upload-error"
        onClick={() => onError('Upload failed')}
      >
        Trigger Error
      </button>
    </div>
  ),
}));

const mockProfile = {
  id: 'user-123',
  displayName: 'Test User',
  aboutMe: 'Test about me',
  avatarURL: 'https://example.com/avatar.jpg',
  accountType: 'full' as const,
  role: 'user' as const,
  isActive: true,
  emailVerified: true,
  createdAt: new Date(),
};

const mockUserProfileStore = {
  profile: mockProfile,
  setProfile: vi.fn(),
};

describe('ProfileSettingsModal - Enhanced with Avatar Upload', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (userProfileStore as any).mockReturnValue(mockUserProfileStore);
    (api.updateUserProfile as any).mockResolvedValue(mockProfile);
    (api.uploadAvatar as any).mockResolvedValue({
      ...mockProfile,
      avatarURL: 'https://example.com/new-avatar.jpg',
    });
  });

  it('should render avatar upload section', () => {
    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    expect(screen.getByText('Avatar Image')).toBeInTheDocument();
    expect(screen.getByTestId('avatar-upload')).toBeInTheDocument();
    expect(screen.getByTestId('current-avatar-url')).toHaveTextContent('https://example.com/avatar.jpg');
  });

  it('should pass current avatar URL to AvatarImageUpload', () => {
    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    expect(screen.getByTestId('current-avatar-url')).toHaveTextContent('https://example.com/avatar.jpg');
  });

  it('should handle avatar selection and enable save button', async () => {
    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    const saveButton = screen.getByRole('button', { name: /save changes/i });
    expect(saveButton).toBeDisabled();

    fireEvent.click(screen.getByTestId('mock-select-image'));

    await waitFor(() => {
      expect(saveButton).toBeEnabled();
    });
  });

  it('should upload avatar when saving with new avatar', async () => {
    const mockOnClose = vi.fn();
    render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);

    // Select a new avatar
    fireEvent.click(screen.getByTestId('mock-select-image'));

    // Save changes
    const saveButton = screen.getByRole('button', { name: /save changes/i });
    fireEvent.click(saveButton);

    // Should show uploading state
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /uploading avatar/i })).toBeInTheDocument();
    });

    // Should call uploadAvatar API
    await waitFor(() => {
      expect(api.uploadAvatar).toHaveBeenCalledWith(expect.any(File));
    });

    // Should update profile store
    await waitFor(() => {
      expect(mockUserProfileStore.setProfile).toHaveBeenCalledWith({
        ...mockProfile,
        avatarURL: 'https://example.com/new-avatar.jpg',
      });
    });

    // Should close modal
    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it('should handle avatar upload error', async () => {
    (api.uploadAvatar as any).mockRejectedValue(new Error('Upload failed'));
    
    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    // Select a new avatar
    fireEvent.click(screen.getByTestId('mock-select-image'));

    // Save changes
    const saveButton = screen.getByRole('button', { name: /save changes/i });
    fireEvent.click(saveButton);

    // Should show error message
    await waitFor(() => {
      expect(screen.getByText('Failed to upload avatar. Please try again.')).toBeInTheDocument();
    });
  });

  it('should handle avatar selection error', async () => {
    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    // Trigger upload error
    fireEvent.click(screen.getByTestId('mock-upload-error'));

    // Should show error message
    await waitFor(() => {
      expect(screen.getByText('Upload failed')).toBeInTheDocument();
    });
  });

  it('should disable avatar upload when loading', () => {
    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    // Select avatar and start saving
    fireEvent.click(screen.getByTestId('mock-select-image'));
    fireEvent.click(screen.getByRole('button', { name: /save changes/i }));

    // Avatar upload should be disabled during upload
    expect(screen.getByTestId('upload-disabled')).toHaveTextContent('disabled');
  });

  it('should save profile fields and avatar together', async () => {
    const mockOnClose = vi.fn();
    render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);

    // Change display name
    const displayNameInput = screen.getByLabelText(/display name/i);
    fireEvent.change(displayNameInput, { target: { value: 'Updated Name' } });

    // Select new avatar
    fireEvent.click(screen.getByTestId('mock-select-image'));

    // Save changes
    fireEvent.click(screen.getByRole('button', { name: /save changes/i }));

    // Should upload avatar first
    await waitFor(() => {
      expect(api.uploadAvatar).toHaveBeenCalled();
    });

    // Should update profile fields
    await waitFor(() => {
      expect(api.updateUserProfile).toHaveBeenCalledWith(
        { displayName: 'Updated Name' },
        'user-123'
      );
    });
  });

  it('should work with guest accounts (no display name changes)', () => {
    const guestProfile = { ...mockProfile, accountType: 'guest' as const };
    (userProfileStore as any).mockReturnValue({
      profile: guestProfile,
      setProfile: vi.fn(),
    });

    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    // Display name should be disabled for guest accounts
    const displayNameInput = screen.getByLabelText(/display name/i);
    expect(displayNameInput).toBeDisabled();

    // But avatar upload should still work
    expect(screen.getByTestId('avatar-upload')).toBeInTheDocument();
  });

  it('should handle no current avatar URL', () => {
    const profileWithoutAvatar = { ...mockProfile, avatarURL: undefined };
    (userProfileStore as any).mockReturnValue({
      profile: profileWithoutAvatar,
      setProfile: vi.fn(),
    });

    render(<ProfileSettingsModal isOpen={true} onClose={vi.fn()} />);

    expect(screen.getByTestId('current-avatar-url')).toHaveTextContent('no-avatar');
  });
});