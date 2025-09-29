import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { ImageProcessor } from '../imageProcessor';

// Mock canvas and image APIs
const mockCanvas = {
  width: 0,
  height: 0,
  getContext: vi.fn(() => ({
    imageSmoothingEnabled: true,
    imageSmoothingQuality: 'high',
    drawImage: vi.fn(),
  })),
  toBlob: vi.fn(),
};

const mockImage = {
  width: 0,
  height: 0,
  onload: null as (() => void) | null,
  onerror: null as (() => void) | null,
  set src(value: string) {
    // Simulate immediate load success
    setTimeout(() => {
      if (this.onload) {
        this.onload();
      }
    }, 0);
  },
};

// Mock DOM APIs
Object.defineProperty(global, 'HTMLCanvasElement', {
  value: vi.fn(() => mockCanvas),
});

Object.defineProperty(global, 'Image', {
  value: vi.fn(() => mockImage),
});

Object.defineProperty(global, 'URL', {
  value: {
    createObjectURL: vi.fn(() => 'blob:mock-url'),
    revokeObjectURL: vi.fn(),
  },
});

describe('ImageProcessor', () => {
  let imageProcessor: ImageProcessor;
  let mockFile: File;

  beforeEach(() => {
    imageProcessor = new ImageProcessor();
    mockFile = new File(['mock-content'], 'test.jpg', { type: 'image/jpeg' });
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('validateImage', () => {
    it('should validate a valid JPEG image', async () => {
      const validFile = new File(['content'], 'test.jpg', { 
        type: 'image/jpeg' 
      });
      Object.defineProperty(validFile, 'size', { value: 5 * 1024 * 1024 }); // 5MB

      // Set up mock image dimensions
      mockImage.width = 800;
      mockImage.height = 600;

      const result = await imageProcessor.validateImage(validFile);

      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
      expect(result.fileSize).toBe(5 * 1024 * 1024);
    });

    it('should reject files larger than 10MB', async () => {
      const largeFile = new File(['content'], 'large.jpg', { 
        type: 'image/jpeg' 
      });
      Object.defineProperty(largeFile, 'size', { value: 15 * 1024 * 1024 }); // 15MB

      const result = await imageProcessor.validateImage(largeFile);

      expect(result.isValid).toBe(false);
      expect(result.error).toBe('File size must be less than 10MB');
    });

    it('should reject unsupported file types', async () => {
      const invalidFile = new File(['content'], 'test.gif', { 
        type: 'image/gif' 
      });

      const result = await imageProcessor.validateImage(invalidFile);

      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Only JPEG, PNG, and WebP files are allowed');
    });

    it('should accept PNG files', async () => {
      const pngFile = new File(['content'], 'test.png', { 
        type: 'image/png' 
      });
      Object.defineProperty(pngFile, 'size', { value: 2 * 1024 * 1024 }); // 2MB

      // Set up mock image dimensions
      mockImage.width = 800;
      mockImage.height = 600;

      const result = await imageProcessor.validateImage(pngFile);

      expect(result.isValid).toBe(true);
    });

    it('should accept WebP files', async () => {
      const webpFile = new File(['content'], 'test.webp', { 
        type: 'image/webp' 
      });
      Object.defineProperty(webpFile, 'size', { value: 3 * 1024 * 1024 }); // 3MB

      // Set up mock image dimensions
      mockImage.width = 800;
      mockImage.height = 600;

      const result = await imageProcessor.validateImage(webpFile);

      expect(result.isValid).toBe(true);
    });
  });

  describe('generatePreviewUrl', () => {
    it('should generate a preview URL for a file', () => {
      const url = imageProcessor.generatePreviewUrl(mockFile);

      expect(URL.createObjectURL).toHaveBeenCalledWith(mockFile);
      expect(url).toBe('blob:mock-url');
    });
  });

  describe('revokePreviewUrl', () => {
    it('should revoke a preview URL', () => {
      const url = 'blob:mock-url';
      
      imageProcessor.revokePreviewUrl(url);

      expect(URL.revokeObjectURL).toHaveBeenCalledWith(url);
    });
  });

  describe('resizeImage', () => {
    it('should resize an image to target dimensions', async () => {
      // Setup mock image with dimensions
      mockImage.width = 1024;
      mockImage.height = 768;
      
      // Setup mock canvas context
      const mockContext = {
        imageSmoothingEnabled: true,
        imageSmoothingQuality: 'high',
        drawImage: vi.fn(),
      };
      mockCanvas.getContext = vi.fn(() => mockContext);
      
      // Setup mock blob result
      const mockBlob = new Blob(['resized'], { type: 'image/jpeg' });
      mockCanvas.toBlob = vi.fn((callback) => callback(mockBlob));

      // Mock document.createElement
      vi.spyOn(document, 'createElement').mockReturnValue(mockCanvas as any);

      const resizedFile = await imageProcessor.resizeImage(mockFile, 512);

      expect(document.createElement).toHaveBeenCalledWith('canvas');
      expect(mockCanvas.width).toBe(512);
      expect(mockCanvas.height).toBe(384); // Maintains aspect ratio
      expect(mockContext.drawImage).toHaveBeenCalled();
      expect(resizedFile).toBeInstanceOf(File);
    });

    it('should handle square images correctly', async () => {
      // Setup mock square image
      mockImage.width = 800;
      mockImage.height = 800;
      
      const mockContext = {
        imageSmoothingEnabled: true,
        imageSmoothingQuality: 'high',
        drawImage: vi.fn(),
      };
      mockCanvas.getContext = vi.fn(() => mockContext);
      
      const mockBlob = new Blob(['resized'], { type: 'image/jpeg' });
      mockCanvas.toBlob = vi.fn((callback) => callback(mockBlob));

      vi.spyOn(document, 'createElement').mockReturnValue(mockCanvas as any);

      await imageProcessor.resizeImage(mockFile, 512);

      expect(mockCanvas.width).toBe(512);
      expect(mockCanvas.height).toBe(512);
    });

    it('should handle resize errors gracefully', async () => {
      // Override the src setter to trigger error instead of load
      const errorImage = {
        ...mockImage,
        set src(value: string) {
          setTimeout(() => {
            if (this.onerror) {
              this.onerror();
            }
          }, 0);
        },
      };
      
      vi.spyOn(global, 'Image').mockImplementation(() => errorImage as any);
      vi.spyOn(document, 'createElement').mockReturnValue(mockCanvas as any);

      await expect(imageProcessor.resizeImage(mockFile, 512)).rejects.toThrow();
    });
  });

  describe('cropImage', () => {
    it('should crop an image with specified crop data', async () => {
      const cropData = {
        x: 100,
        y: 100,
        width: 300,
        height: 300,
        scale: 1,
      };

      mockImage.width = 800;
      mockImage.height = 600;
      
      const mockContext = {
        imageSmoothingEnabled: true,
        imageSmoothingQuality: 'high',
        drawImage: vi.fn(),
      };
      mockCanvas.getContext = vi.fn(() => mockContext);
      
      const mockBlob = new Blob(['cropped'], { type: 'image/jpeg' });
      mockCanvas.toBlob = vi.fn((callback) => callback(mockBlob));

      vi.spyOn(document, 'createElement').mockReturnValue(mockCanvas as any);

      const croppedFile = await imageProcessor.cropImage(mockFile, cropData);

      expect(mockCanvas.width).toBe(300);
      expect(mockCanvas.height).toBe(300);
      expect(mockContext.drawImage).toHaveBeenCalledWith(
        mockImage,
        100, 100, 300, 300, // Source crop area
        0, 0, 300, 300       // Destination area
      );
      expect(croppedFile).toBeInstanceOf(File);
    });

    it('should handle crop errors gracefully', async () => {
      const cropData = {
        x: 0,
        y: 0,
        width: 100,
        height: 100,
        scale: 1,
      };

      // Override the src setter to trigger error instead of load
      const errorImage = {
        ...mockImage,
        set src(value: string) {
          setTimeout(() => {
            if (this.onerror) {
              this.onerror();
            }
          }, 0);
        },
      };
      
      vi.spyOn(global, 'Image').mockImplementation(() => errorImage as any);
      vi.spyOn(document, 'createElement').mockReturnValue(mockCanvas as any);

      await expect(imageProcessor.cropImage(mockFile, cropData)).rejects.toThrow();
    });
  });
});