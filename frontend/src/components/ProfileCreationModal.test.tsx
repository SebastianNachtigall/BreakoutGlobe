import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import ProfileCreationModal from './ProfileCreationModal';

// Mock the API service
vi.mock('../services/api', () => ({
  createGuestProfile: vi.fn(),
  APIError: class APIError extends Error {
    constructor(message: string, public status: number, public code?: string) {
      super(message);
      this.name = 'APIError';
    }
  },
}));

import { createGuestProfile } from '../services/api';
const mockCreateGuestProfile = vi.mocked(createGuestProfile);

describe('ProfileCreationModal', () => {
  const mockOnProfileCreated = vi.fn();
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Form Validation', () => {
    it('should require display name', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/display name is required/i)).toBeInTheDocument();
      });
    });

    it('should validate display name length (minimum 3 characters)', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const displayNameInput = screen.getByLabelText(/display name/i);
      fireEvent.change(displayNameInput, { target: { value: 'ab' } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/display name must be at least 3 characters/i)).toBeInTheDocument();
      });
    });

    it('should validate display name length (maximum 50 characters)', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const displayNameInput = screen.getByLabelText(/display name/i);
      const longName = 'a'.repeat(51);
      fireEvent.change(displayNameInput, { target: { value: longName } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/display name must be 50 characters or less/i)).toBeInTheDocument();
      });
    });

    it('should validate about me text length (maximum 500 characters)', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const aboutMeInput = screen.getByLabelText(/about me/i);
      const longText = 'a'.repeat(501);
      fireEvent.change(aboutMeInput, { target: { value: longText } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/about me must be 500 characters or less/i)).toBeInTheDocument();
      });
    });
  });

  describe('Avatar Upload', () => {
    it('should accept valid image files (JPG/PNG)', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const fileInput = screen.getByLabelText(/avatar image/i);
      const file = new File(['test'], 'avatar.jpg', { type: 'image/jpeg' });
      
      fireEvent.change(fileInput, { target: { files: [file] } });

      // Should not show any error
      expect(screen.queryByText(/invalid file type/i)).not.toBeInTheDocument();
    });

    it('should reject invalid file types', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const fileInput = screen.getByLabelText(/avatar image/i);
      const file = new File(['test'], 'avatar.gif', { type: 'image/gif' });
      
      fireEvent.change(fileInput, { target: { files: [file] } });

      await waitFor(() => {
        expect(screen.getByText(/only jpg and png files are allowed/i)).toBeInTheDocument();
      });
    });

    it('should reject files larger than 2MB', async () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const fileInput = screen.getByLabelText(/avatar image/i);
      // Create a mock file larger than 2MB
      const largeFile = new File(['x'.repeat(2 * 1024 * 1024 + 1)], 'large.jpg', { type: 'image/jpeg' });
      
      fireEvent.change(fileInput, { target: { files: [largeFile] } });

      await waitFor(() => {
        expect(screen.getByText(/file size must be less than 2mb/i)).toBeInTheDocument();
      });
    });
  });

  describe('Guest Profile Creation', () => {
    it('should create guest profile with valid data', async () => {
      const mockProfile = {
        id: 'user-123',
        displayName: 'Test User',
        accountType: 'guest' as const,
        role: 'user' as const,
        isActive: true,
        emailVerified: false,
        createdAt: new Date(),
      };

      mockCreateGuestProfile.mockResolvedValue(mockProfile);

      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const displayNameInput = screen.getByLabelText(/display name/i);
      fireEvent.change(displayNameInput, { target: { value: 'Test User' } });

      const aboutMeInput = screen.getByLabelText(/about me/i);
      fireEvent.change(aboutMeInput, { target: { value: 'Hello, I am a test user!' } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(mockCreateGuestProfile).toHaveBeenCalledWith({
          displayName: 'Test User',
          aboutMe: 'Hello, I am a test user!',
          avatarFile: undefined,
        });
        expect(mockOnProfileCreated).toHaveBeenCalledWith(mockProfile);
      });
    });

    it('should create guest profile with avatar file', async () => {
      const mockProfile = {
        id: 'user-123',
        displayName: 'Test User',
        accountType: 'guest' as const,
        role: 'user' as const,
        isActive: true,
        emailVerified: false,
        createdAt: new Date(),
        avatarURL: 'https://example.com/avatar.jpg',
      };

      mockCreateGuestProfile.mockResolvedValue(mockProfile);

      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const displayNameInput = screen.getByLabelText(/display name/i);
      fireEvent.change(displayNameInput, { target: { value: 'Test User' } });

      const fileInput = screen.getByLabelText(/avatar image/i);
      const file = new File(['test'], 'avatar.jpg', { type: 'image/jpeg' });
      fireEvent.change(fileInput, { target: { files: [file] } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(mockCreateGuestProfile).toHaveBeenCalledWith({
          displayName: 'Test User',
          aboutMe: undefined,
          avatarFile: file,
        });
        expect(mockOnProfileCreated).toHaveBeenCalledWith(mockProfile);
      });
    });

    it('should handle API errors gracefully', async () => {
      mockCreateGuestProfile.mockRejectedValue(new Error('Network error'));

      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const displayNameInput = screen.getByLabelText(/display name/i);
      fireEvent.change(displayNameInput, { target: { value: 'Test User' } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/failed to create profile/i)).toBeInTheDocument();
      });
    });

    it('should show loading state during profile creation', async () => {
      // Make the API call hang to test loading state
      mockCreateGuestProfile.mockImplementation(() => new Promise(() => {}));

      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const displayNameInput = screen.getByLabelText(/display name/i);
      fireEvent.change(displayNameInput, { target: { value: 'Test User' } });

      const createButton = screen.getByRole('button', { name: /create profile/i });
      fireEvent.click(createButton);

      await waitFor(() => {
        expect(screen.getByText(/creating profile/i)).toBeInTheDocument();
        expect(createButton).toBeDisabled();
      });
    });
  });

  describe('Modal Behavior', () => {
    it('should not render when isOpen is false', () => {
      render(
        <ProfileCreationModal
          isOpen={false}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      expect(screen.queryByText(/create your profile/i)).not.toBeInTheDocument();
    });

    it('should call onClose when cancel button is clicked', () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const cancelButton = screen.getByRole('button', { name: /cancel/i });
      fireEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should call onClose when clicking outside modal', () => {
      render(
        <ProfileCreationModal
          isOpen={true}
          onProfileCreated={mockOnProfileCreated}
          onClose={mockOnClose}
        />
      );

      const backdrop = screen.getByTestId('modal-backdrop');
      fireEvent.click(backdrop);

      expect(mockOnClose).toHaveBeenCalled();
    });
  });
});