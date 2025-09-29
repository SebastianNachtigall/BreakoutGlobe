import React from 'react';
import { render, screen, act } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { CircularAvatarPreview } from '../CircularAvatarPreview';

describe('CircularAvatarPreview', () => {
  const mockImageUrl = 'blob:mock-image-url';

  it('should render with default props', () => {
    render(<CircularAvatarPreview imageUrl={mockImageUrl} />);

    const container = screen.getByTestId('circular-avatar-preview');
    const image = screen.getByRole('img');

    expect(container).toBeInTheDocument();
    expect(image).toHaveAttribute('src', mockImageUrl);
    expect(container).toHaveClass('rounded-full');
  });

  it('should show placeholder when no image provided', () => {
    render(<CircularAvatarPreview imageUrl={null} />);

    const placeholder = screen.getByTestId('avatar-placeholder');
    expect(placeholder).toBeInTheDocument();
    expect(placeholder).toHaveTextContent('?');
  });

  it('should show custom placeholder text', () => {
    render(<CircularAvatarPreview imageUrl={null} placeholderText="AB" />);

    const placeholder = screen.getByTestId('avatar-placeholder');
    expect(placeholder).toHaveTextContent('AB');
  });

  it('should apply correct size classes', () => {
    const { rerender } = render(<CircularAvatarPreview imageUrl={mockImageUrl} size="small" />);
    expect(screen.getByTestId('circular-avatar-preview')).toHaveClass('w-8', 'h-8');

    rerender(<CircularAvatarPreview imageUrl={mockImageUrl} size="medium" />);
    expect(screen.getByTestId('circular-avatar-preview')).toHaveClass('w-16', 'h-16');

    rerender(<CircularAvatarPreview imageUrl={mockImageUrl} size="large" />);
    expect(screen.getByTestId('circular-avatar-preview')).toHaveClass('w-32', 'h-32');
  });

  it('should show loading state', () => {
    render(<CircularAvatarPreview imageUrl={mockImageUrl} isLoading />);

    const loadingIndicator = screen.getByTestId('avatar-loading');
    expect(loadingIndicator).toBeInTheDocument();
  });

  it('should apply crop data correctly', () => {
    const cropData = {
      x: 50,
      y: 25,
      width: 200,
      height: 200,
      scale: 1.5,
    };

    render(<CircularAvatarPreview imageUrl={mockImageUrl} cropData={cropData} />);

    const image = screen.getByRole('img');
    expect(image).toHaveStyle({
      transform: 'translate(-50px, -25px) scale(1.5)',
    });
  });

  it('should handle image load errors', async () => {
    render(<CircularAvatarPreview imageUrl="invalid-url" />);

    const image = screen.getByRole('img');
    
    // Simulate image error
    await act(async () => {
      image.dispatchEvent(new Event('error'));
    });

    // Should show placeholder after error
    expect(screen.getByTestId('avatar-placeholder')).toBeInTheDocument();
  });

  it('should apply custom className', () => {
    render(<CircularAvatarPreview imageUrl={mockImageUrl} className="custom-avatar" />);

    const container = screen.getByTestId('circular-avatar-preview');
    expect(container).toHaveClass('custom-avatar');
  });

  it('should show border when specified', () => {
    render(<CircularAvatarPreview imageUrl={mockImageUrl} showBorder />);

    const container = screen.getByTestId('circular-avatar-preview');
    expect(container).toHaveClass('border-2');
  });

  it('should apply hover effects when interactive', () => {
    render(<CircularAvatarPreview imageUrl={mockImageUrl} interactive />);

    const container = screen.getByTestId('circular-avatar-preview');
    expect(container).toHaveClass('cursor-pointer', 'hover:shadow-lg');
  });

  it('should generate initials from name when no image', () => {
    render(<CircularAvatarPreview imageUrl={null} name="John Doe" />);

    const placeholder = screen.getByTestId('avatar-placeholder');
    expect(placeholder).toHaveTextContent('JD');
  });

  it('should handle single name for initials', () => {
    render(<CircularAvatarPreview imageUrl={null} name="John" />);

    const placeholder = screen.getByTestId('avatar-placeholder');
    expect(placeholder).toHaveTextContent('J');
  });

  it('should prioritize placeholderText over name initials', () => {
    render(<CircularAvatarPreview imageUrl={null} name="John Doe" placeholderText="AB" />);

    const placeholder = screen.getByTestId('avatar-placeholder');
    expect(placeholder).toHaveTextContent('AB');
  });
});