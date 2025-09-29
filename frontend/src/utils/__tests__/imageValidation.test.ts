import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  validateImageFile,
  createImagePreviewUrl,
  cleanupImagePreviewUrl,
  formatFileSize,
  getImageFileInfo,
  isImageFile,
  ImageFileInfo,
} from '../imageValidation';

// Mock URL API
Object.defineProperty(global, 'URL', {
  value: {
    createObjectURL: vi.fn(() => 'blob:mock-url-123'),
    revokeObjectURL: vi.fn(),
  },
});

describe('imageValidation utilities', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('validateImageFile', () => {
    it('should validate a valid image file', () => {
      const validFile = new File(['content'], 'test.jpg', { 
        type: 'image/jpeg' 
      });
      Object.defineProperty(validFile, 'size', { value: 5 * 1024 * 1024 }); // 5MB

      const result = validateImageFile(validFile);

      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject files that are too large', () => {
      const largeFile = new File(['content'], 'large.jpg', { 
        type: 'image/jpeg' 
      });
      Object.defineProperty(largeFile, 'size', { value: 15 * 1024 * 1024 }); // 15MB

      const result = validateImageFile(largeFile);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('File size must be less than 10MB');
    });

    it('should reject unsupported file types', () => {
      const invalidFile = new File(['content'], 'test.gif', { 
        type: 'image/gif' 
      });

      const result = validateImageFile(invalidFile);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Only JPEG, PNG, and WebP files are allowed');
    });

    it('should reject files with no extension', () => {
      const noExtFile = new File(['content'], 'image', { 
        type: 'image/jpeg' 
      });

      const result = validateImageFile(noExtFile);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('File must have a valid image extension');
    });

    it('should accept all supported formats', () => {
      const jpegFile = new File(['content'], 'test.jpg', { type: 'image/jpeg' });
      const pngFile = new File(['content'], 'test.png', { type: 'image/png' });
      const webpFile = new File(['content'], 'test.webp', { type: 'image/webp' });

      [jpegFile, pngFile, webpFile].forEach(file => {
        Object.defineProperty(file, 'size', { value: 1024 * 1024 }); // 1MB
        const result = validateImageFile(file);
        expect(result.isValid).toBe(true);
      });
    });

    it('should handle custom validation options', () => {
      const file = new File(['content'], 'test.jpg', { type: 'image/jpeg' });
      Object.defineProperty(file, 'size', { value: 3 * 1024 * 1024 }); // 3MB

      const result = validateImageFile(file, {
        maxFileSize: 2 * 1024 * 1024, // 2MB limit
        allowedTypes: ['image/jpeg'],
      });

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('File size must be less than 2MB');
    });
  });

  describe('createImagePreviewUrl', () => {
    it('should create a preview URL for an image file', () => {
      const file = new File(['content'], 'test.jpg', { type: 'image/jpeg' });

      const url = createImagePreviewUrl(file);

      expect(URL.createObjectURL).toHaveBeenCalledWith(file);
      expect(url).toBe('blob:mock-url-123');
    });

    it('should handle null file gracefully', () => {
      const url = createImagePreviewUrl(null);

      expect(url).toBeNull();
      expect(URL.createObjectURL).not.toHaveBeenCalled();
    });
  });

  describe('cleanupImagePreviewUrl', () => {
    it('should revoke a preview URL', () => {
      const url = 'blob:mock-url-123';

      cleanupImagePreviewUrl(url);

      expect(URL.revokeObjectURL).toHaveBeenCalledWith(url);
    });

    it('should handle null URL gracefully', () => {
      cleanupImagePreviewUrl(null);

      expect(URL.revokeObjectURL).not.toHaveBeenCalled();
    });

    it('should handle empty string gracefully', () => {
      cleanupImagePreviewUrl('');

      expect(URL.revokeObjectURL).not.toHaveBeenCalled();
    });
  });

  describe('formatFileSize', () => {
    it('should format bytes correctly', () => {
      expect(formatFileSize(0)).toBe('0 B');
      expect(formatFileSize(512)).toBe('512 B');
      expect(formatFileSize(1023)).toBe('1023 B');
    });

    it('should format kilobytes correctly', () => {
      expect(formatFileSize(1024)).toBe('1.0 KB');
      expect(formatFileSize(1536)).toBe('1.5 KB');
      expect(formatFileSize(2048)).toBe('2.0 KB');
    });

    it('should format megabytes correctly', () => {
      expect(formatFileSize(1024 * 1024)).toBe('1.0 MB');
      expect(formatFileSize(1.5 * 1024 * 1024)).toBe('1.5 MB');
      expect(formatFileSize(10 * 1024 * 1024)).toBe('10.0 MB');
    });

    it('should format gigabytes correctly', () => {
      expect(formatFileSize(1024 * 1024 * 1024)).toBe('1.0 GB');
      expect(formatFileSize(2.5 * 1024 * 1024 * 1024)).toBe('2.5 GB');
    });

    it('should handle negative numbers', () => {
      expect(formatFileSize(-1024)).toBe('0 B');
    });
  });

  describe('getImageFileInfo', () => {
    it('should extract file information correctly', () => {
      const file = new File(['content'], 'test-image.jpg', { 
        type: 'image/jpeg',
        lastModified: 1640995200000, // 2022-01-01
      });
      Object.defineProperty(file, 'size', { value: 2 * 1024 * 1024 }); // 2MB

      const info: ImageFileInfo = getImageFileInfo(file);

      expect(info.name).toBe('test-image.jpg');
      expect(info.size).toBe(2 * 1024 * 1024);
      expect(info.type).toBe('image/jpeg');
      expect(info.formattedSize).toBe('2.0 MB');
      expect(info.extension).toBe('jpg');
      expect(info.lastModified).toBeInstanceOf(Date);
    });

    it('should handle files without extensions', () => {
      const file = new File(['content'], 'image', { type: 'image/png' });

      const info = getImageFileInfo(file);

      expect(info.extension).toBe('');
    });

    it('should handle files with multiple dots in name', () => {
      const file = new File(['content'], 'my.test.image.png', { type: 'image/png' });

      const info = getImageFileInfo(file);

      expect(info.extension).toBe('png');
      expect(info.name).toBe('my.test.image.png');
    });
  });

  describe('isImageFile', () => {
    it('should identify image files correctly', () => {
      const jpegFile = new File(['content'], 'test.jpg', { type: 'image/jpeg' });
      const pngFile = new File(['content'], 'test.png', { type: 'image/png' });
      const webpFile = new File(['content'], 'test.webp', { type: 'image/webp' });

      expect(isImageFile(jpegFile)).toBe(true);
      expect(isImageFile(pngFile)).toBe(true);
      expect(isImageFile(webpFile)).toBe(true);
    });

    it('should reject non-image files', () => {
      const textFile = new File(['content'], 'test.txt', { type: 'text/plain' });
      const pdfFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });

      expect(isImageFile(textFile)).toBe(false);
      expect(isImageFile(pdfFile)).toBe(false);
    });

    it('should handle files with no MIME type', () => {
      const unknownFile = new File(['content'], 'test.jpg', { type: '' });

      expect(isImageFile(unknownFile)).toBe(false);
    });
  });
});