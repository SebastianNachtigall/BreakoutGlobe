import { formatFileSize } from './imageValidation';

export type ErrorCode = 
  | 'FILE_TOO_LARGE'
  | 'INVALID_FILE_TYPE'
  | 'INVALID_DIMENSIONS'
  | 'PROCESSING_FAILED'
  | 'BROWSER_NOT_SUPPORTED'
  | 'NETWORK_ERROR'
  | 'UNKNOWN_ERROR';

export class ImageProcessingError extends Error {
  public readonly code: ErrorCode;
  public readonly context?: Record<string, any>;

  constructor(code: ErrorCode, message: string, context?: Record<string, any>) {
    super(message);
    this.name = 'ImageProcessingError';
    this.code = code;
    this.context = context;
  }
}

/**
 * Creates a standardized image processing error with appropriate message
 */
export function createImageError(
  code: ErrorCode, 
  context?: Record<string, any>
): ImageProcessingError {
  const message = generateErrorMessage(code, context);
  return new ImageProcessingError(code, message, context);
}

/**
 * Generates user-friendly error messages based on error code and context
 */
function generateErrorMessage(code: ErrorCode, context?: Record<string, any>): string {
  switch (code) {
    case 'FILE_TOO_LARGE':
      if (context?.fileSize && context?.maxSize) {
        const currentSize = formatFileSize(context.fileSize);
        const maxSize = formatFileSize(context.maxSize);
        return `File size (${currentSize}) exceeds the maximum allowed size of ${maxSize}`;
      }
      return 'File size is too large';

    case 'INVALID_FILE_TYPE':
      if (context?.fileType && context?.allowedTypes) {
        const allowedTypesText = context.allowedTypes
          .map((type: string) => type.replace('image/', '').toUpperCase())
          .join(', ');
        return `File type ${context.fileType} is not supported. Please use ${allowedTypesText} files`;
      }
      return 'File type is not supported';

    case 'INVALID_DIMENSIONS':
      if (context?.width && context?.height && context?.minWidth && context?.minHeight) {
        return `Image dimensions (${context.width}x${context.height}) are too small. Minimum required: ${context.minWidth}x${context.minHeight}`;
      }
      if (context?.width && context?.height && context?.maxWidth && context?.maxHeight) {
        return `Image dimensions (${context.width}x${context.height}) are too large. Maximum allowed: ${context.maxWidth}x${context.maxHeight}`;
      }
      return 'Image dimensions are invalid';

    case 'PROCESSING_FAILED':
      if (context?.operation) {
        return `Failed to ${context.operation} image. Please try again or use a different image`;
      }
      return 'Image processing failed. Please try again';

    case 'BROWSER_NOT_SUPPORTED':
      if (context?.missingFeature) {
        return `Your browser doesn't support ${context.missingFeature}, which is required for image processing`;
      }
      return 'Your browser doesn\'t support the required features for image processing';

    case 'NETWORK_ERROR':
      if (context?.operation) {
        return `Network error occurred during image ${context.operation}. Please check your connection and try again`;
      }
      return 'Network error occurred. Please check your connection and try again';

    case 'UNKNOWN_ERROR':
    default:
      return 'An unexpected error occurred while processing the image';
  }
}

/**
 * Extracts error message from various error types
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  
  if (typeof error === 'string') {
    return error;
  }
  
  return 'An unknown error occurred';
}

/**
 * Type guard to check if error is an ImageProcessingError
 */
export function isImageProcessingError(error: unknown): error is ImageProcessingError {
  return error instanceof ImageProcessingError;
}

/**
 * Wraps a function to catch and convert errors to ImageProcessingError
 */
export function withImageErrorHandling<T extends any[], R>(
  fn: (...args: T) => Promise<R>,
  operation: string
) {
  return async (...args: T): Promise<R> => {
    try {
      return await fn(...args);
    } catch (error) {
      if (isImageProcessingError(error)) {
        throw error;
      }
      
      throw createImageError('PROCESSING_FAILED', {
        operation,
        originalError: error,
      });
    }
  };
}

/**
 * User-friendly error messages for common scenarios
 */
export const USER_FRIENDLY_MESSAGES = {
  FILE_TOO_LARGE: 'Please choose a smaller image (max 10MB)',
  INVALID_FILE_TYPE: 'Please choose a JPEG, PNG, or WebP image',
  INVALID_DIMENSIONS: 'Image is too small or too large',
  PROCESSING_FAILED: 'Something went wrong. Please try a different image',
  BROWSER_NOT_SUPPORTED: 'Please use a modern browser for image editing',
  NETWORK_ERROR: 'Connection problem. Please try again',
  UNKNOWN_ERROR: 'Something went wrong. Please try again',
} as const;

/**
 * Gets a simplified, user-friendly error message
 */
export function getUserFriendlyMessage(error: unknown): string {
  if (isImageProcessingError(error)) {
    return USER_FRIENDLY_MESSAGES[error.code] || USER_FRIENDLY_MESSAGES.UNKNOWN_ERROR;
  }
  
  return USER_FRIENDLY_MESSAGES.UNKNOWN_ERROR;
}