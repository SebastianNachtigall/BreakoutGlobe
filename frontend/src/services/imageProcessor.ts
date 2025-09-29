export interface ValidationResult {
  isValid: boolean;
  error?: string;
  fileSize: number;
  dimensions?: { width: number; height: number };
}

export interface CropData {
  x: number;
  y: number;
  width: number;
  height: number;
  scale: number;
}

export interface ImageValidationConfig {
  maxFileSize: number; // 10MB = 10 * 1024 * 1024
  allowedTypes: string[]; // ['image/jpeg', 'image/png', 'image/webp']
  minDimensions: { width: number; height: number };
  maxDimensions: { width: number; height: number };
}

export interface ProcessingConfig {
  targetSize: number; // 512px for avatar display
  quality: number; // 0.8 for good quality/size balance
  outputFormat: 'jpeg' | 'png' | 'webp';
}

export class ImageProcessor {
  private readonly config: ImageValidationConfig = {
    maxFileSize: 10 * 1024 * 1024, // 10MB
    allowedTypes: ['image/jpeg', 'image/png', 'image/webp'],
    minDimensions: { width: 50, height: 50 },
    maxDimensions: { width: 4096, height: 4096 },
  };

  private readonly processingConfig: ProcessingConfig = {
    targetSize: 512,
    quality: 0.8,
    outputFormat: 'jpeg',
  };

  /**
   * Validates an image file against size, type, and dimension requirements
   */
  async validateImage(file: File): Promise<ValidationResult> {
    const result: ValidationResult = {
      isValid: false,
      fileSize: file.size,
    };

    // Check file size
    if (file.size > this.config.maxFileSize) {
      result.error = 'File size must be less than 10MB';
      return result;
    }

    // Check file type
    if (!this.config.allowedTypes.includes(file.type)) {
      result.error = 'Only JPEG, PNG, and WebP files are allowed';
      return result;
    }

    // Check image dimensions (if needed)
    try {
      const dimensions = await this.getImageDimensions(file);
      result.dimensions = dimensions;

      if (dimensions.width < this.config.minDimensions.width || 
          dimensions.height < this.config.minDimensions.height) {
        result.error = `Image must be at least ${this.config.minDimensions.width}x${this.config.minDimensions.height} pixels`;
        return result;
      }

      if (dimensions.width > this.config.maxDimensions.width || 
          dimensions.height > this.config.maxDimensions.height) {
        result.error = `Image must be no larger than ${this.config.maxDimensions.width}x${this.config.maxDimensions.height} pixels`;
        return result;
      }
    } catch (error) {
      result.error = 'Unable to read image file';
      return result;
    }

    result.isValid = true;
    return result;
  }

  /**
   * Generates a preview URL for displaying the image
   */
  generatePreviewUrl(file: File): string {
    return URL.createObjectURL(file);
  }

  /**
   * Revokes a previously created preview URL to free memory
   */
  revokePreviewUrl(url: string): void {
    URL.revokeObjectURL(url);
  }

  /**
   * Resizes an image to the specified maximum size while maintaining aspect ratio
   */
  async resizeImage(file: File, maxSize: number = this.processingConfig.targetSize): Promise<File> {
    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      if (!ctx) {
        reject(new Error('Canvas context not available'));
        return;
      }

      const img = new Image();
      
      img.onload = () => {
        try {
          // Calculate new dimensions maintaining aspect ratio
          const { width, height } = this.calculateResizeDimensions(
            img.width, 
            img.height, 
            maxSize
          );

          canvas.width = width;
          canvas.height = height;

          // Configure high-quality rendering
          ctx.imageSmoothingEnabled = true;
          ctx.imageSmoothingQuality = 'high';

          // Draw the resized image
          ctx.drawImage(img, 0, 0, width, height);

          // Convert to blob and create file
          canvas.toBlob(
            (blob) => {
              if (blob) {
                const resizedFile = new File([blob], file.name, {
                  type: `image/${this.processingConfig.outputFormat}`,
                  lastModified: Date.now(),
                });
                resolve(resizedFile);
              } else {
                reject(new Error('Failed to create resized image'));
              }
            },
            `image/${this.processingConfig.outputFormat}`,
            this.processingConfig.quality
          );
        } catch (error) {
          reject(error);
        }
      };

      img.onerror = () => {
        reject(new Error('Failed to load image for resizing'));
      };

      img.src = URL.createObjectURL(file);
    });
  }

  /**
   * Crops an image based on the provided crop data
   */
  async cropImage(file: File, cropData: CropData): Promise<File> {
    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      if (!ctx) {
        reject(new Error('Canvas context not available'));
        return;
      }

      const img = new Image();
      
      img.onload = () => {
        try {
          // Set canvas size to crop dimensions
          canvas.width = cropData.width;
          canvas.height = cropData.height;

          // Configure high-quality rendering
          ctx.imageSmoothingEnabled = true;
          ctx.imageSmoothingQuality = 'high';

          // Draw the cropped portion of the image
          ctx.drawImage(
            img,
            cropData.x, cropData.y, cropData.width, cropData.height, // Source crop area
            0, 0, cropData.width, cropData.height                     // Destination area
          );

          // Convert to blob and create file
          canvas.toBlob(
            (blob) => {
              if (blob) {
                const croppedFile = new File([blob], file.name, {
                  type: `image/${this.processingConfig.outputFormat}`,
                  lastModified: Date.now(),
                });
                resolve(croppedFile);
              } else {
                reject(new Error('Failed to create cropped image'));
              }
            },
            `image/${this.processingConfig.outputFormat}`,
            this.processingConfig.quality
          );
        } catch (error) {
          reject(error);
        }
      };

      img.onerror = () => {
        reject(new Error('Failed to load image for cropping'));
      };

      img.src = URL.createObjectURL(file);
    });
  }

  /**
   * Gets the dimensions of an image file
   */
  private async getImageDimensions(file: File): Promise<{ width: number; height: number }> {
    return new Promise((resolve, reject) => {
      const img = new Image();
      
      img.onload = () => {
        resolve({ width: img.width, height: img.height });
      };

      img.onerror = () => {
        reject(new Error('Failed to load image'));
      };

      img.src = URL.createObjectURL(file);
    });
  }

  /**
   * Calculates new dimensions for resizing while maintaining aspect ratio
   */
  private calculateResizeDimensions(
    originalWidth: number, 
    originalHeight: number, 
    maxSize: number
  ): { width: number; height: number } {
    const aspectRatio = originalWidth / originalHeight;

    if (originalWidth <= maxSize && originalHeight <= maxSize) {
      // Image is already smaller than max size
      return { width: originalWidth, height: originalHeight };
    }

    if (originalWidth > originalHeight) {
      // Landscape orientation
      return {
        width: maxSize,
        height: Math.round(maxSize / aspectRatio),
      };
    } else {
      // Portrait or square orientation
      return {
        width: Math.round(maxSize * aspectRatio),
        height: maxSize,
      };
    }
  }
}

// Export a singleton instance for use throughout the application
export const imageProcessor = new ImageProcessor();