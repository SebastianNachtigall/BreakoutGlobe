export interface ValidationOptions {
  maxFileSize?: number;
  allowedTypes?: string[];
  allowedExtensions?: string[];
}

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

export interface ImageFileInfo {
  name: string;
  size: number;
  type: string;
  formattedSize: string;
  extension: string;
  lastModified: Date;
}

// Default validation configuration
const DEFAULT_OPTIONS: Required<ValidationOptions> = {
  maxFileSize: 10 * 1024 * 1024, // 10MB
  allowedTypes: ['image/jpeg', 'image/png', 'image/webp'],
  allowedExtensions: ['jpg', 'jpeg', 'png', 'webp'],
};

/**
 * Validates an image file against size, type, and extension requirements
 */
export function validateImageFile(
  file: File, 
  options: ValidationOptions = {}
): ValidationResult {
  const config = { ...DEFAULT_OPTIONS, ...options };
  const errors: string[] = [];

  // Check file size
  if (file.size > config.maxFileSize) {
    const maxSizeMB = Math.round(config.maxFileSize / (1024 * 1024));
    errors.push(`File size must be less than ${maxSizeMB}MB`);
  }

  // Check MIME type
  if (!config.allowedTypes.includes(file.type)) {
    errors.push('Only JPEG, PNG, and WebP files are allowed');
  }

  // Check file extension
  const extension = getFileExtension(file.name);
  if (!extension) {
    errors.push('File must have a valid image extension');
  } else if (!config.allowedExtensions.includes(extension.toLowerCase())) {
    errors.push(`File extension .${extension} is not supported`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

/**
 * Creates a preview URL for an image file
 */
export function createImagePreviewUrl(file: File | null): string | null {
  if (!file) {
    return null;
  }
  
  return URL.createObjectURL(file);
}

/**
 * Cleans up a preview URL to free memory
 */
export function cleanupImagePreviewUrl(url: string | null): void {
  if (url && url.startsWith('blob:')) {
    URL.revokeObjectURL(url);
  }
}

/**
 * Formats file size in human-readable format
 */
export function formatFileSize(bytes: number): string {
  if (bytes < 0) {
    return '0 B';
  }

  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  // Format with appropriate decimal places
  const formatted = unitIndex === 0 
    ? size.toString() 
    : size.toFixed(1);

  return `${formatted} ${units[unitIndex]}`;
}

/**
 * Extracts comprehensive information about an image file
 */
export function getImageFileInfo(file: File): ImageFileInfo {
  return {
    name: file.name,
    size: file.size,
    type: file.type,
    formattedSize: formatFileSize(file.size),
    extension: getFileExtension(file.name),
    lastModified: new Date(file.lastModified),
  };
}

/**
 * Checks if a file is an image based on MIME type
 */
export function isImageFile(file: File): boolean {
  return file.type.startsWith('image/') && 
         DEFAULT_OPTIONS.allowedTypes.includes(file.type);
}

/**
 * Extracts file extension from filename
 */
function getFileExtension(filename: string): string {
  const lastDotIndex = filename.lastIndexOf('.');
  if (lastDotIndex === -1 || lastDotIndex === filename.length - 1) {
    return '';
  }
  return filename.substring(lastDotIndex + 1);
}

/**
 * Validates multiple files at once
 */
export function validateImageFiles(
  files: File[], 
  options: ValidationOptions = {}
): { valid: File[]; invalid: Array<{ file: File; errors: string[] }> } {
  const valid: File[] = [];
  const invalid: Array<{ file: File; errors: string[] }> = [];

  files.forEach(file => {
    const result = validateImageFile(file, options);
    if (result.isValid) {
      valid.push(file);
    } else {
      invalid.push({ file, errors: result.errors });
    }
  });

  return { valid, invalid };
}

/**
 * Creates a safe filename by removing or replacing invalid characters
 */
export function sanitizeFilename(filename: string): string {
  // Remove or replace invalid characters
  return filename
    .replace(/[<>:"/\\|?*]/g, '_') // Replace invalid chars with underscore
    .replace(/\s+/g, '_') // Replace spaces with underscore
    .replace(/_+/g, '_') // Replace multiple underscores with single
    .replace(/^_|_$/g, '') // Remove leading/trailing underscores
    .toLowerCase();
}

/**
 * Generates a unique filename by appending timestamp if needed
 */
export function generateUniqueFilename(originalName: string): string {
  const extension = getFileExtension(originalName);
  const nameWithoutExt = originalName.substring(0, originalName.lastIndexOf('.'));
  const sanitizedName = sanitizeFilename(nameWithoutExt);
  const timestamp = Date.now();
  
  return extension 
    ? `${sanitizedName}_${timestamp}.${extension}`
    : `${sanitizedName}_${timestamp}`;
}