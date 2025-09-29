import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { ImagePreview } from '../ImagePreview';

describe('ImagePreview', () => {
  const mockImageUrl = 'blob:mock-image-url';

  it('should render image with default props', () => {
    render(<ImagePreview imageUrl={mockImageUrl} />);

    const image = screen.getByRole('img');
    expect(image).toBeInTheDocument();
    expect(image).toHaveAttribute('src', mockImageUrl);
    expect(image).toHaveAttribute('alt', 'Image preview');
  });

  it('should apply circular preview styling when showCircularPreview is true', () => {
    render(<ImagePreview imageUrl={mockImageUrl} showCircularPreview />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('rounded-full');
  });

  it('should apply rectangular styling when showCircularPreview is false', () => {
    render(<ImagePreview imageUrl={mockImageUrl} showCircularPreview={false} />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('rounded-lg');
    expect(container).not.toHaveClass('rounded-full');
  });

  it('should apply small size styling', () => {
    render(<ImagePreview imageUrl={mockImageUrl} size="small" />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('w-8', 'h-8');
  });

  it('should apply medium size styling', () => {
    render(<ImagePreview imageUrl={mockImageUrl} size="medium" />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('w-16', 'h-16');
  });

  it('should apply large size styling', () => {
    render(<ImagePreview imageUrl={mockImageUrl} size="large" />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('w-32', 'h-32');
  });

  it('should apply custom className', () => {
    render(<ImagePreview imageUrl={mockImageUrl} className="custom-class" />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('custom-class');
  });

  it('should handle crop data by applying transform styles', () => {
    const cropData = {
      x: 100,
      y: 50,
      width: 200,
      height: 200,
      scale: 1,
    };

    render(<ImagePreview imageUrl={mockImageUrl} cropData={cropData} />);

    const image = screen.getByRole('img');
    expect(image).toHaveStyle({
      transform: 'translate(-100px, -50px) scale(1)',
      width: '200px',
      height: '200px',
    });
  });

  it('should show loading placeholder when imageUrl is empty', () => {
    render(<ImagePreview imageUrl="" />);

    const placeholder = screen.getByTestId('image-preview-placeholder');
    expect(placeholder).toBeInTheDocument();
    expect(placeholder).toHaveTextContent('No image');
  });

  it('should show loading placeholder when imageUrl is null', () => {
    render(<ImagePreview imageUrl={null as any} />);

    const placeholder = screen.getByTestId('image-preview-placeholder');
    expect(placeholder).toBeInTheDocument();
  });

  it('should apply correct aspect ratio for crop preview', () => {
    const cropData = {
      x: 0,
      y: 0,
      width: 300,
      height: 200, // Different aspect ratio
      scale: 1,
    };

    render(<ImagePreview imageUrl={mockImageUrl} cropData={cropData} showCircularPreview />);

    const container = screen.getByTestId('image-preview-container');
    // Should maintain square aspect ratio for circular preview
    expect(container).toHaveClass('aspect-square');
  });

  it('should handle missing alt text gracefully', () => {
    render(<ImagePreview imageUrl={mockImageUrl} />);

    const image = screen.getByRole('img');
    expect(image).toHaveAttribute('alt', 'Image preview');
  });

  it('should apply overflow hidden for cropped images', () => {
    const cropData = {
      x: 50,
      y: 50,
      width: 100,
      height: 100,
      scale: 1,
    };

    render(<ImagePreview imageUrl={mockImageUrl} cropData={cropData} />);

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('overflow-hidden');
  });

  it('should combine size and circular preview classes correctly', () => {
    render(
      <ImagePreview 
        imageUrl={mockImageUrl} 
        size="large" 
        showCircularPreview 
      />
    );

    const container = screen.getByTestId('image-preview-container');
    expect(container).toHaveClass('w-32', 'h-32', 'rounded-full');
  });
});