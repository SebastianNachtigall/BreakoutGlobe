import React from 'react';
import { CropData } from '../types/imageProcessing';

export interface ImagePreviewProps {
  imageUrl: string | null;
  cropData?: CropData;
  size?: 'small' | 'medium' | 'large';
  showCircularPreview?: boolean;
  className?: string;
  alt?: string;
}

export const ImagePreview: React.FC<ImagePreviewProps> = ({
  imageUrl,
  cropData,
  size = 'medium',
  showCircularPreview = true,
  className = '',
  alt = 'Image preview',
}) => {
  // Size classes mapping
  const sizeClasses = {
    small: 'w-8 h-8',
    medium: 'w-16 h-16',
    large: 'w-32 h-32',
  };

  // Shape classes
  const shapeClasses = showCircularPreview ? 'rounded-full' : 'rounded-lg';

  // Container classes
  const containerClasses = [
    sizeClasses[size],
    shapeClasses,
    'overflow-hidden',
    'bg-gray-100',
    'border-2 border-gray-200',
    'flex items-center justify-center',
    'relative',
    className,
  ].filter(Boolean).join(' ');

  // If no image URL, show placeholder
  if (!imageUrl) {
    return (
      <div 
        className={containerClasses}
        data-testid="image-preview-container"
      >
        <div 
          className="text-gray-400 text-xs text-center"
          data-testid="image-preview-placeholder"
        >
          No image
        </div>
      </div>
    );
  }

  // Calculate image styles based on crop data
  const getImageStyles = (): React.CSSProperties => {
    if (!cropData) {
      return {
        width: '100%',
        height: '100%',
        objectFit: 'cover' as const,
      };
    }

    return {
      transform: `translate(-${cropData.x}px, -${cropData.y}px) scale(${cropData.scale})`,
      width: `${cropData.width}px`,
      height: `${cropData.height}px`,
      objectFit: 'cover' as const,
      transformOrigin: 'top left',
    };
  };

  return (
    <div 
      className={`${containerClasses} ${showCircularPreview ? 'aspect-square' : ''}`}
      data-testid="image-preview-container"
    >
      <img
        src={imageUrl}
        alt={alt}
        style={getImageStyles()}
        className="block"
        onError={(e) => {
          // Handle image load errors gracefully
          const target = e.target as HTMLImageElement;
          target.style.display = 'none';
        }}
      />
    </div>
  );
};