import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import { POICreationModal } from '../POICreationModal';

describe('POICreationModal', () => {
  const defaultProps = {
    isOpen: true,
    position: { lat: 40.7128, lng: -74.0060 },
    onCreate: vi.fn(),
    onCancel: vi.fn(),
    isLoading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render image upload field', () => {
    render(<POICreationModal {...defaultProps} />);
    
    expect(screen.getByLabelText(/image \(optional\)/i)).toBeInTheDocument();
    expect(screen.getByText(/supported formats: jpeg, png, webp/i)).toBeInTheDocument();
  });

  it('should validate image file size', async () => {
    render(<POICreationModal {...defaultProps} />);
    
    // Create a mock file that's too large (6MB)
    const largeFile = new File(['x'.repeat(6 * 1024 * 1024)], 'large-image.jpg', {
      type: 'image/jpeg',
    });
    
    const imageInput = screen.getByLabelText(/image \(optional\)/i);
    
    // Mock the files property
    Object.defineProperty(imageInput, 'files', {
      value: [largeFile],
      writable: false,
    });
    
    fireEvent.change(imageInput);
    fireEvent.blur(imageInput);
    
    await waitFor(() => {
      expect(screen.getByText(/image must be smaller than 5mb/i)).toBeInTheDocument();
    });
  });

  it('should validate image file type', async () => {
    render(<POICreationModal {...defaultProps} />);
    
    // Create a mock file with invalid type
    const invalidFile = new File(['content'], 'document.pdf', {
      type: 'application/pdf',
    });
    
    const imageInput = screen.getByLabelText(/image \(optional\)/i);
    
    // Mock the files property
    Object.defineProperty(imageInput, 'files', {
      value: [invalidFile],
      writable: false,
    });
    
    fireEvent.change(imageInput);
    fireEvent.blur(imageInput);
    
    await waitFor(() => {
      expect(screen.getByText(/image must be jpeg, png, or webp format/i)).toBeInTheDocument();
    });
  });

  it('should show selected file information', async () => {
    render(<POICreationModal {...defaultProps} />);
    
    // Create a valid mock file
    const validFile = new File(['content'], 'test-image.jpg', {
      type: 'image/jpeg',
    });
    
    const imageInput = screen.getByLabelText(/image \(optional\)/i);
    
    // Mock the files property
    Object.defineProperty(imageInput, 'files', {
      value: [validFile],
      writable: false,
    });
    
    fireEvent.change(imageInput);
    
    await waitFor(() => {
      expect(screen.getByText(/selected: test-image\.jpg/i)).toBeInTheDocument();
    });
  });

  it('should include image in onCreate callback', async () => {
    render(<POICreationModal {...defaultProps} />);
    
    // Fill in required fields
    fireEvent.change(screen.getByLabelText(/name/i), {
      target: { value: 'Test POI' },
    });
    fireEvent.change(screen.getByLabelText(/description/i), {
      target: { value: 'Test description' },
    });
    
    // Add image
    const validFile = new File(['content'], 'test-image.jpg', {
      type: 'image/jpeg',
    });
    
    const imageInput = screen.getByLabelText(/image \(optional\)/i);
    Object.defineProperty(imageInput, 'files', {
      value: [validFile],
      writable: false,
    });
    fireEvent.change(imageInput);
    
    // Submit form
    fireEvent.click(screen.getByText(/create poi/i));
    
    await waitFor(() => {
      expect(defaultProps.onCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Test POI',
          description: 'Test description',
          image: validFile,
        })
      );
    });
  });

  it('should work without image (optional field)', async () => {
    render(<POICreationModal {...defaultProps} />);
    
    // Fill in required fields only
    fireEvent.change(screen.getByLabelText(/name/i), {
      target: { value: 'Test POI' },
    });
    fireEvent.change(screen.getByLabelText(/description/i), {
      target: { value: 'Test description' },
    });
    
    // Submit form without image
    fireEvent.click(screen.getByText(/create poi/i));
    
    await waitFor(() => {
      expect(defaultProps.onCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Test POI',
          description: 'Test description',
          image: undefined,
        })
      );
    });
  });
});