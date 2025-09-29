// Re-export types from imageProcessor for easier imports
export type {
  ValidationResult as ProcessorValidationResult,
  CropData,
  ImageValidationConfig,
  ProcessingConfig,
} from '../services/imageProcessor';

// Re-export types from validation utilities
export type {
  ValidationResult,
  ValidationOptions,
  ImageFileInfo,
} from '../utils/imageValidation';

// Re-export error types
export type {
  ErrorCode,
  ImageProcessingError,
} from '../utils/imageErrors';

// Additional UI-specific types for components
export interface ImageUploadState {
  selectedFile: File | null;
  previewUrl: string | null;
  isProcessing: boolean;
  showCropEditor: boolean;
  cropData: CropData | null;
  error: string | null;
}

export interface ImagePreviewProps {
  imageUrl: string;
  cropData?: CropData;
  size?: 'small' | 'medium' | 'large';
  showCircularPreview?: boolean;
  className?: string;
}

export interface ImageCropEditorProps {
  imageFile: File;
  onCropConfirm: (cropData: CropData) => void;
  onCancel: () => void;
  isOpen: boolean;
}

export interface AvatarImageUploadProps {
  onImageSelected: (processedFile: File) => void;
  onError: (error: string) => void;
  currentAvatarUrl?: string;
  disabled?: boolean;
  className?: string;
}

// Constants for consistent sizing across components
export const AVATAR_SIZES = {
  small: 32,
  medium: 64,
  large: 128,
} as const;

export type AvatarSize = keyof typeof AVATAR_SIZES;