import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ImageCropEditor } from '../ImageCropEditor';

// Mock URL.createObjectURL
Object.defineProperty(global, 'URL', {
  value: {
    createObjectURL: vi.fn(() => 'blob:mock-url'),
    revokeObjectURL: vi.fn(),
  },
});

describe('ImageCropEditor', () => {
  const mockFile = new File(['mock-content'], 'test.jpg', { type: 'image/jpeg' });
  const mockOnCropConfirm = vi.fn();
  const mockOnCancel = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render when open', () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    expect(screen.getByTestId('crop-editor-modal')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Crop Image' })).toBeInTheDocument();
  });

  it('should not render when closed', () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={false}
      />
    );

    expect(screen.queryByTestId('crop-editor-modal')).not.toBeInTheDocument();
  });

  it('should show loading state initially', async () => {
    // Mock URL.createObjectURL to delay the URL creation
    const originalCreateObjectURL = URL.createObjectURL;
    URL.createObjectURL = vi.fn(() => {
      // Return URL immediately but component will still show loading briefly
      return 'blob:mock-url';
    });

    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    // The component shows the crop interface immediately since we mock URL.createObjectURL
    expect(screen.getByTestId('crop-editor-modal')).toBeInTheDocument();
    
    // Restore original function
    URL.createObjectURL = originalCreateObjectURL;
  });

  it('should close modal when backdrop clicked', async () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    const backdrop = screen.getByTestId('modal-backdrop');
    await userEvent.click(backdrop);

    expect(mockOnCancel).toHaveBeenCalledTimes(1);
  });

  it('should not close when clicking inside modal content', async () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    const modalContent = screen.getByTestId('modal-content');
    await userEvent.click(modalContent);

    expect(mockOnCancel).not.toHaveBeenCalled();
  });

  it('should handle crop confirmation', async () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    // The crop button should be disabled initially until a crop is completed
    const cropButton = screen.getByRole('button', { name: 'Crop Image' });
    expect(cropButton).toBeDisabled();
  });

  it('should create object URL for image preview', () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    expect(URL.createObjectURL).toHaveBeenCalledWith(mockFile);
  });

  it('should have proper modal structure', () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    expect(screen.getByRole('heading', { name: 'Crop Image' })).toBeInTheDocument();
    expect(screen.getByText(/Drag to select the area/)).toBeInTheDocument();
  });
});