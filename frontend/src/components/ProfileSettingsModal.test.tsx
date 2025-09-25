import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import ProfileSettingsModal from './ProfileSettingsModal';
import { useUserProfileStore } from '../stores/userProfileStore';
import * as api from '../services/api';

// Mock the user profile store
vi.mock('../stores/userProfileStore', () => ({
  useUserProfileStore: vi.fn()
}));
const mockUseUserProfileStore = vi.mocked(useUserProfileStore);

// Mock the API
vi.mock('../services/api');
const mockApi = vi.mocked(api);

describe('ProfileSettingsModal', () => {
  const mockOnClose = vi.fn();
  const mockUpdateProfile = vi.fn();
  
  const mockGuestProfile = {
    id: 'user-123',
    displayName: 'Test User',
    accountType: 'guest' as const,
    role: 'user' as const,
    isActive: true,
    createdAt: '2023-01-01T00:00:00Z',
    avatarUrl: '',
    aboutMe: 'Hello, I am a test user'
  };

  const mockFullProfile = {
    ...mockGuestProfile,
    accountType: 'full' as const,
    email: 'test@example.com'
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseUserProfileStore.mockReturnValue({
      profile: mockGuestProfile,
      updateProfile: mockUpdateProfile,
      isLoading: false,
      error: null,
      createProfile: vi.fn(),
      uploadAvatar: vi.fn(),
      clearError: vi.fn(),
      syncWithBackend: vi.fn()
    });
  });

  describe('Profile Settings Access', () => {
    it('should render profile settings modal for guest accounts', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      expect(screen.getByText('Profile Settings')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Test User')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Hello, I am a test user')).toBeInTheDocument();
    });

    it('should show display name as read-only for guest accounts', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const displayNameInput = screen.getByDisplayValue('Test User');
      expect(displayNameInput).toBeDisabled();
      expect(screen.getByText('Display name cannot be changed for guest accounts')).toBeInTheDocument();
    });

    it('should allow editing about me for guest accounts', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      expect(aboutMeTextarea).not.toBeDisabled();
      expect(aboutMeTextarea).toHaveAttribute('placeholder', 'Tell others about yourself...');
    });

    it('should allow editing display name for full accounts', () => {
      mockUseUserProfileStore.mockReturnValue({
        profile: mockFullProfile,
        updateProfile: mockUpdateProfile,
        isLoading: false,
        error: null,
        createProfile: vi.fn(),
        uploadAvatar: vi.fn(),
        clearError: vi.fn(),
        syncWithBackend: vi.fn()
      });

      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const displayNameInput = screen.getByDisplayValue('Test User');
      expect(displayNameInput).not.toBeDisabled();
      expect(screen.queryByText('Display name cannot be changed for guest accounts')).not.toBeInTheDocument();
    });
  });

  describe('Profile Update Success', () => {
    it('should update about me for guest profile', async () => {
      mockApi.updateUserProfile.mockResolvedValue({
        ...mockGuestProfile,
        aboutMe: 'Updated about me text'
      });

      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      fireEvent.change(aboutMeTextarea, { target: { value: 'Updated about me text' } });
      
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
      
      await waitFor(() => {
        expect(mockApi.updateUserProfile).toHaveBeenCalledWith({
          aboutMe: 'Updated about me text'
        });
      });
      
      expect(mockUpdateProfile).toHaveBeenCalledWith({
        ...mockGuestProfile,
        aboutMe: 'Updated about me text'
      });
      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should update both display name and about me for full profile', async () => {
      mockUseUserProfileStore.mockReturnValue({
        profile: mockFullProfile,
        updateProfile: mockUpdateProfile,
        isLoading: false,
        error: null,
        createProfile: vi.fn(),
        uploadAvatar: vi.fn(),
        clearError: vi.fn(),
        syncWithBackend: vi.fn()
      });

      mockApi.updateUserProfile.mockResolvedValue({
        ...mockFullProfile,
        displayName: 'Updated Name',
        aboutMe: 'Updated about me'
      });

      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const displayNameInput = screen.getByDisplayValue('Test User');
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      
      fireEvent.change(displayNameInput, { target: { value: 'Updated Name' } });
      fireEvent.change(aboutMeTextarea, { target: { value: 'Updated about me' } });
      
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
      
      await waitFor(() => {
        expect(mockApi.updateUserProfile).toHaveBeenCalledWith({
          displayName: 'Updated Name',
          aboutMe: 'Updated about me'
        });
      });
    });

    it('should show loading state during update', async () => {
      mockApi.updateUserProfile.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      fireEvent.change(aboutMeTextarea, { target: { value: 'Updated text' } });
      
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
      
      expect(screen.getByText('Saving...')).toBeInTheDocument();
      expect(saveButton).toBeDisabled();
    });

    it('should handle update errors gracefully', async () => {
      mockApi.updateUserProfile.mockRejectedValue(new Error('Update failed'));

      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      fireEvent.change(aboutMeTextarea, { target: { value: 'Updated text' } });
      
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
      
      await waitFor(() => {
        expect(screen.getByText('Failed to update profile. Please try again.')).toBeInTheDocument();
      });
      
      expect(mockOnClose).not.toHaveBeenCalled();
    });
  });

  describe('Form Validation', () => {
    it('should validate display name length for full accounts', async () => {
      mockUseUserProfileStore.mockReturnValue({
        profile: mockFullProfile,
        updateProfile: mockUpdateProfile,
        isLoading: false,
        error: null,
        createProfile: vi.fn(),
        uploadAvatar: vi.fn(),
        clearError: vi.fn(),
        syncWithBackend: vi.fn()
      });

      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const displayNameInput = screen.getByDisplayValue('Test User');
      fireEvent.change(displayNameInput, { target: { value: 'AB' } }); // Too short
      
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
      
      expect(screen.getByText('Display name must be at least 3 characters')).toBeInTheDocument();
      expect(mockApi.updateUserProfile).not.toHaveBeenCalled();
    });

    it('should validate about me length', async () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      const longText = 'A'.repeat(1001); // Too long
      fireEvent.change(aboutMeTextarea, { target: { value: longText } });
      
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
      
      expect(screen.getByText('About me must be less than 1000 characters')).toBeInTheDocument();
      expect(mockApi.updateUserProfile).not.toHaveBeenCalled();
    });

    it('should disable save button when no changes are made', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const saveButton = screen.getByText('Save Changes');
      expect(saveButton).toBeDisabled();
    });

    it('should enable save button when changes are made', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const aboutMeTextarea = screen.getByDisplayValue('Hello, I am a test user');
      fireEvent.change(aboutMeTextarea, { target: { value: 'Updated text' } });
      
      const saveButton = screen.getByText('Save Changes');
      expect(saveButton).not.toBeDisabled();
    });
  });

  describe('Modal Behavior', () => {
    it('should not render when isOpen is false', () => {
      render(<ProfileSettingsModal isOpen={false} onClose={mockOnClose} />);
      
      expect(screen.queryByText('Profile Settings')).not.toBeInTheDocument();
    });

    it('should close modal when cancel button is clicked', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const cancelButton = screen.getByText('Cancel');
      fireEvent.click(cancelButton);
      
      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should close modal when clicking outside', () => {
      render(<ProfileSettingsModal isOpen={true} onClose={mockOnClose} />);
      
      const backdrop = screen.getByTestId('modal-backdrop');
      fireEvent.click(backdrop);
      
      expect(mockOnClose).toHaveBeenCalled();
    });
  });
});