import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ProfileCreationModal from '../ProfileCreationModal';

describe('ProfileCreationModal - Fixed Version', () => {
  const mockOnProfileCreated = vi.fn();
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should prevent closing when clicking outside the modal', () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Find the backdrop element
    const backdrop = screen.getByTestId('modal-backdrop');
    
    // Click on the backdrop (outside the modal content)
    fireEvent.click(backdrop);
    
    // Modal should NOT close - onClose should not be called
    expect(mockOnClose).not.toHaveBeenCalled();
    
    // Modal should still be visible
    expect(screen.getByText('Create Your Profile')).toBeInTheDocument();
  });

  it('should not have a Cancel button', () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Should not have a Cancel button
    expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
    
    // Should only have Create Profile button
    expect(screen.getByRole('button', { name: /create profile/i })).toBeInTheDocument();
  });

  it('should show required message', () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Should show that profile creation is required
    expect(screen.getByText(/profile is required to use the app/i)).toBeInTheDocument();
  });

  it('should still allow form input and validation', () => {
    render(
      <ProfileCreationModal
        isOpen={true}
        onProfileCreated={mockOnProfileCreated}
        onClose={mockOnClose}
      />
    );

    // Should be able to type in the display name field
    const displayNameInput = screen.getByLabelText(/display name/i);
    fireEvent.change(displayNameInput, { target: { value: 'Test User' } });
    expect(displayNameInput).toHaveValue('Test User');

    // Should be able to type in the about me field
    const aboutMeInput = screen.getByLabelText(/about me/i);
    fireEvent.change(aboutMeInput, { target: { value: 'Test about me' } });
    expect(aboutMeInput).toHaveValue('Test about me');

    // Form should still be functional
    expect(screen.getByRole('button', { name: /create profile/i })).not.toBeDisabled();
  });
});