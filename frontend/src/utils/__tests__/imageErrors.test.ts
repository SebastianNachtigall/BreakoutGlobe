import { describe, it, expect } from 'vitest';
import {
  ImageProcessingError,
  createImageError,
  getErrorMessage,
  isImageProcessingError,
  ErrorCode,
} from '../imageErrors';

describe('imageErrors utilities', () => {
  describe('ImageProcessingError', () => {
    it('should create error with code and message', () => {
      const error = new ImageProcessingError('FILE_TOO_LARGE', 'File is too large');

      expect(error.code).toBe('FILE_TOO_LARGE');
      expect(error.message).toBe('File is too large');
      expect(error.name).toBe('ImageProcessingError');
      expect(error).toBeInstanceOf(Error);
    });

    it('should create error with additional context', () => {
      const error = new ImageProcessingError(
        'INVALID_DIMENSIONS', 
        'Invalid dimensions',
        { width: 100, height: 50 }
      );

      expect(error.context).toEqual({ width: 100, height: 50 });
    });
  });

  describe('createImageError', () => {
    it('should create FILE_TOO_LARGE error', () => {
      const error = createImageError('FILE_TOO_LARGE', { 
        fileSize: 15 * 1024 * 1024,
        maxSize: 10 * 1024 * 1024 
      });

      expect(error.code).toBe('FILE_TOO_LARGE');
      expect(error.message).toContain('15.0 MB');
      expect(error.message).toContain('10.0 MB');
    });

    it('should create INVALID_FILE_TYPE error', () => {
      const error = createImageError('INVALID_FILE_TYPE', { 
        fileType: 'image/gif',
        allowedTypes: ['image/jpeg', 'image/png']
      });

      expect(error.code).toBe('INVALID_FILE_TYPE');
      expect(error.message).toContain('image/gif');
      expect(error.message).toContain('JPEG, PNG');
    });

    it('should create PROCESSING_FAILED error', () => {
      const error = createImageError('PROCESSING_FAILED', { 
        operation: 'resize',
        originalError: new Error('Canvas error')
      });

      expect(error.code).toBe('PROCESSING_FAILED');
      expect(error.message).toContain('resize');
    });

    it('should create INVALID_DIMENSIONS error', () => {
      const error = createImageError('INVALID_DIMENSIONS', { 
        width: 10,
        height: 10,
        minWidth: 50,
        minHeight: 50
      });

      expect(error.code).toBe('INVALID_DIMENSIONS');
      expect(error.message).toContain('10x10');
      expect(error.message).toContain('50x50');
    });

    it('should create BROWSER_NOT_SUPPORTED error', () => {
      const error = createImageError('BROWSER_NOT_SUPPORTED', { 
        missingFeature: 'Canvas API'
      });

      expect(error.code).toBe('BROWSER_NOT_SUPPORTED');
      expect(error.message).toContain('Canvas API');
    });

    it('should create NETWORK_ERROR error', () => {
      const error = createImageError('NETWORK_ERROR', { 
        operation: 'upload'
      });

      expect(error.code).toBe('NETWORK_ERROR');
      expect(error.message).toContain('upload');
    });

    it('should create UNKNOWN_ERROR with default message', () => {
      const error = createImageError('UNKNOWN_ERROR');

      expect(error.code).toBe('UNKNOWN_ERROR');
      expect(error.message).toBe('An unexpected error occurred while processing the image');
    });
  });

  describe('getErrorMessage', () => {
    it('should return message from ImageProcessingError', () => {
      const error = new ImageProcessingError('FILE_TOO_LARGE', 'Custom message');

      expect(getErrorMessage(error)).toBe('Custom message');
    });

    it('should return message from regular Error', () => {
      const error = new Error('Regular error message');

      expect(getErrorMessage(error)).toBe('Regular error message');
    });

    it('should return string error as-is', () => {
      expect(getErrorMessage('String error')).toBe('String error');
    });

    it('should handle unknown error types', () => {
      expect(getErrorMessage({ unknown: 'object' })).toBe('An unknown error occurred');
      expect(getErrorMessage(null)).toBe('An unknown error occurred');
      expect(getErrorMessage(undefined)).toBe('An unknown error occurred');
    });
  });

  describe('isImageProcessingError', () => {
    it('should identify ImageProcessingError correctly', () => {
      const imageError = new ImageProcessingError('FILE_TOO_LARGE', 'Message');
      const regularError = new Error('Regular error');

      expect(isImageProcessingError(imageError)).toBe(true);
      expect(isImageProcessingError(regularError)).toBe(false);
      expect(isImageProcessingError('string')).toBe(false);
      expect(isImageProcessingError(null)).toBe(false);
    });
  });
});