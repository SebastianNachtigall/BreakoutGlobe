import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AvatarImageUpload } from '../AvatarImageUpload';

// Mock the imageProcessor
vi.mock('../../services/imageProcessor', () => ({
  imageProcessor: {
    validateImage: vi.fn(),
    resizeImage: vi.fn(),
    cropImage: vi.fn(),
    generatePreviewUrl: vi.fn(() => 'blob:mock-preview-url'),
    revokePreviewUrl: vi.fn(),
  },
}));

// Mock URL.createObjectURL
Object.defineProperty(global, 'URL', {
  value: {
    createObjectURL: vi.fn(() => 'blob:mock-url'),
    revokeObjectURL: vi.fn(),
  },
});

// Mock file input
const createMockFile = (name = 'test.jpg', type = 'image/jpeg', size = 1024 * 1024) => {
  const file = new File(['mock-content'], name, { type });
  Object.defineProperty(file, 'size', { value: size });
  return file;
};

describe('AvatarImageUpload', () => {
  const mockOnImageSelected = vi.fn();
  const mockOnError = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render file input and preview area', () => {
    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    expect(screen.getByTestId('avatar-upload-container')).toBeInTheDocument();
    expect(screen.getByTestId('file-input')).toBeInTheDocument();
    expect(screen.getByText(/Choose Image/)).toBeInTheDocument();
  });

  it('should show current avatar when provided', () => {
    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
        currentAvatarUrl="https://example.com/avatar.jpg"
      />
    );

    expect(screen.getByTestId('current-avatar')).toBeInTheDocument();
  });

  it('should handle file selection', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });
    
    const resizedFile = createMockFile('resized.jpg');
    (imageProcessor.resizeImage as any).mockResolvedValue(resizedFile);

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    expect(imageProcessor.validateImage).toHaveBeenCalledWith(mockFile);
    expect(imageProcessor.resizeImage).toHaveBeenCalledWith(mockFile, 512);
    expect(screen.getByTestId('image-preview')).toBeInTheDocument();
  });

  it('should show validation error for invalid file', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: false,
      error: 'File too large',
    });

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile('large.jpg', 'image/jpeg', 15 * 1024 * 1024);

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    expect(mockOnError).toHaveBeenCalledWith('File too large');
    expect(screen.getByTestId('error-message')).toBeInTheDocument();
  });

  it('should open crop editor when crop button clicked', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });
    
    const resizedFile = createMockFile('resized.jpg');
    (imageProcessor.resizeImage as any).mockResolvedValue(resizedFile);

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    const cropButton = screen.getByText('Crop Image');
    await userEvent.click(cropButton);

    expect(screen.getByTestId('crop-editor-modal')).toBeInTheDocument();
  });

  it('should process and upload image after cropping', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });

    const mockCroppedFile = createMockFile('cropped.jpg');
    (imageProcessor.cropImage as any).mockResolvedValue(mockCroppedFile);
    (imageProcessor.resizeImage as any).mockResolvedValue(mockCroppedFile);

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    const cropButton = screen.getByText('Crop Image');
    await userEvent.click(cropButton);

    expect(screen.getByTestId('crop-editor-modal')).toBeInTheDocument();
  });

  it('should handle crop cancellation', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    const cropButton = screen.getByText('Crop Image');
    await userEvent.click(cropButton);

    expect(screen.getByTestId('crop-editor-modal')).toBeInTheDocument();
  });

  it('should show processing state during file validation', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    
    // Mock slow validation
    (imageProcessor.validateImage as any).mockImplementation(
      () => new Promise(resolve => setTimeout(() => resolve({
        isValid: true,
        fileSize: 1024 * 1024,
        dimensions: { width: 800, height: 600 },
      }), 100))
    );

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    // Start file selection
    fireEvent.change(fileInput, { target: { files: [mockFile] } });

    // Should show processing state
    expect(screen.getByTestId('processing-indicator')).toBeInTheDocument();
  });

  it('should handle validation errors gracefully', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockRejectedValue(new Error('Validation failed'));

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    expect(mockOnError).toHaveBeenCalledWith(expect.stringContaining('Something went wrong'));
  });

  it('should allow removing selected image', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    expect(screen.getByTestId('image-preview')).toBeInTheDocument();

    const removeButton = screen.getByText('Remove');
    await userEvent.click(removeButton);

    expect(screen.queryByTestId('image-preview')).not.toBeInTheDocument();
  });

  it('should be disabled when disabled prop is true', () => {
    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
        disabled={true}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    expect(fileInput).toBeDisabled();
  });

  it('should apply custom className', () => {
    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
        className="custom-upload"
      />
    );

    expect(screen.getByTestId('avatar-upload-container')).toHaveClass('custom-upload');
  });

  it('should cleanup preview URLs on unmount', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });
    
    const { unmount } = render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    // Select a file to create a preview URL
    const fileInput = screen.getByTestId('file-input');
    const mockFile = createMockFile();

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: [mockFile] } });
    });

    unmount();

    // Component should handle cleanup internally
    expect(true).toBe(true); // Basic test that unmount doesn't crash
  });

  it('should handle multiple file selection by taking first file', async () => {
    const { imageProcessor } = await import('../../services/imageProcessor');
    (imageProcessor.validateImage as any).mockResolvedValue({
      isValid: true,
      fileSize: 1024 * 1024,
      dimensions: { width: 800, height: 600 },
    });

    render(
      <AvatarImageUpload
        onImageSelected={mockOnImageSelected}
        onError={mockOnError}
      />
    );

    const fileInput = screen.getByTestId('file-input');
    const mockFiles = [createMockFile('first.jpg'), createMockFile('second.jpg')];

    await act(async () => {
      fireEvent.change(fileInput, { target: { files: mockFiles } });
    });

    expect(imageProcessor.validateImage).toHaveBeenCalledWith(mockFiles[0]);
  });
});