import React, { useState, useEffect } from 'react';
import { CropData } from '../types/imageProcessing';

export interface CircularAvatarPreviewProps {
  imageUrl: string | null;
  cropData?: CropData;
  size?: 'small' | 'medium' | 'large';
  name?: string;
  placeholderText?: string;
  isLoading?: boolean;
  showBorder?: boolean;
  interactive?: boolean;
  className?: string;
  onClick?: () => void;
}

export const CircularAvatarPreview: React.FC<CircularAvatarPreviewProps> = ({
  imageUrl,
  cropData,
  size = 'medium',
  name,
  placeholderText,
  isLoading = false,
  showBorder = false,
  interactive = false,
  className = '',
  onClick,
}) => {
  const [imageError, setImageError] = useState(false);
  const [croppedImageUrl, setCroppedImageUrl] = useState<string | null>(null);

  // Generate cropped image when cropData changes
  useEffect(() => {
    if (!imageUrl || !cropData) {
      setCroppedImageUrl(null);
      return;
    }

    const img = new Image();
    img.crossOrigin = 'anonymous';
    img.onload = () => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      if (!ctx) return;

      // Set canvas size to crop dimensions
      canvas.width = cropData.width;
      canvas.height = cropData.height;

      // Draw the cropped portion
      ctx.drawImage(
        img,
        cropData.x, cropData.y, cropData.width, cropData.height, // Source rectangle
        0, 0, canvas.width, canvas.height // Destination rectangle
      );

      // Convert to blob URL
      canvas.toBlob((blob) => {
        if (blob) {
          const url = URL.createObjectURL(blob);
          setCroppedImageUrl(url);
        }
      });
    };

    img.onerror = () => {
      setImageError(true);
    };

    img.src = imageUrl;

    // Cleanup function
    return () => {
      if (croppedImageUrl) {
        URL.revokeObjectURL(croppedImageUrl);
      }
    };
  }, [imageUrl, cropData]);

  // Cleanup cropped image URL when component unmounts
  useEffect(() => {
    return () => {
      if (croppedImageUrl) {
        URL.revokeObjectURL(croppedImageUrl);
      }
    };
  }, [croppedImageUrl]);

  // Size classes mapping
  const sizeClasses = {
    small: 'w-8 h-8 text-xs',
    medium: 'w-16 h-16 text-sm',
    large: 'w-32 h-32 text-xl',
  };

  // Generate initials from name
  const getInitials = (fullName: string): string => {
    const names = fullName.trim().split(' ');
    if (names.length === 1) {
      return names[0].charAt(0).toUpperCase();
    }
    return names[0].charAt(0).toUpperCase() + names[names.length - 1].charAt(0).toUpperCase();
  };

  // Get placeholder content
  const getPlaceholderContent = (): string => {
    if (placeholderText) {
      return placeholderText;
    }
    if (name) {
      return getInitials(name);
    }
    return '?';
  };

  // Container classes
  const containerClasses = [
    sizeClasses[size],
    'rounded-full',
    'overflow-hidden',
    'bg-gray-100',
    'flex items-center justify-center',
    'relative',
    'select-none',
    showBorder ? 'border-2 border-gray-300' : '',
    interactive ? 'cursor-pointer hover:shadow-lg transition-shadow duration-200' : '',
    className,
  ].filter(Boolean).join(' ');

  // Get the appropriate image URL to display
  const getDisplayImageUrl = (): string | null => {
    if (cropData && croppedImageUrl) {
      return croppedImageUrl;
    }
    return imageUrl;
  };

  // Calculate image styles
  const getImageStyles = (): React.CSSProperties => {
    return {
      width: '100%',
      height: '100%',
      objectFit: 'cover' as const,
    };
  };

  const handleImageError = () => {
    setImageError(true);
  };

  const handleClick = () => {
    if (interactive && onClick) {
      onClick();
    }
  };

  return (
    <div 
      className={containerClasses}
      data-testid="circular-avatar-preview"
      onClick={handleClick}
    >
      {isLoading ? (
        <div 
          className="animate-spin rounded-full border-2 border-gray-300 border-t-blue-600 w-1/2 h-1/2"
          data-testid="avatar-loading"
        />
      ) : getDisplayImageUrl() && !imageError ? (
        <img
          src={getDisplayImageUrl()!}
          alt={name ? `${name}'s avatar` : 'Avatar'}
          style={getImageStyles()}
          className="block"
          onError={handleImageError}
        />
      ) : (
        <div 
          className="text-gray-500 font-medium"
          data-testid="avatar-placeholder"
        >
          {getPlaceholderContent()}
        </div>
      )}
    </div>
  );
};