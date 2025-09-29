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
    expect(screen.getByText('Crop Image')).toBeInTheDocument();
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

  it('should show loading state initially', () => {
    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    expect(screen.getByTestId('crop-editor-loading')).toBeInTheDocument();
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

  it('should handle image load error', async () => {
    // Mock Image constructor to simulate error
    const mockImage = {
      width: 0,
      height: 0,
      onload: null as (() => void) | null,
      onerror: null as (() => void) | null,
      set src(value: string) {
        setTimeout(() => {
          if (this.onerror) {
            this.onerror();
          }
        }, 0);
      },
    };

    vi.spyOn(global, 'Image').mockImplementation(() => mockImage as any);

    render(
      <ImageCropEditor
        imageFile={mockFile}
        onCropConfirm={mockOnCropConfirm}
        onCancel={mockOnCancel}
        isOpen={true}
      />
    );

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 10));
    });

    expect(screen.getByTestId('crop-editor-error')).toBeInTheDocument();
    expect(screen.getByText(/Failed to load image/)).toBeInTheDocument();
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

    expect(screen.getByText('Crop Image')).toBeInTheDocument();
    expect(screen.getByText(/Drag to select the area/)).toBeInTheDocument();
  });
});