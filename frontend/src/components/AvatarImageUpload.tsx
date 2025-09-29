import React, { useState, useRef, useCallback, useEffect } from 'react';
import { CropData, AvatarImageUploadProps } from '../types/imageProcessing';
import { imageProcessor } from '../services/imageProcessor';
import { createImagePreviewUrl, cleanupImagePreviewUrl } from '../utils/imageValidation';
import { getUserFriendlyMessage } from '../utils/imageErrors';
import { CircularAvatarPreview } from './CircularAvatarPreview';
import { ImageCropEditor } from './ImageCropEditor';

interface UploadState {
  selectedFile: File | null;
  originalFile: File | null; // Keep original file for cropping
  previewUrl: string | null;
  isProcessing: boolean;
  showCropEditor: boolean;
  cropData: CropData | null;
  error: string | null;
}

export const AvatarImageUpload: React.FC<AvatarImageUploadProps> = ({
  onImageSelected,
  onError,
  currentAvatarUrl,
  disabled = false,
  className = '',
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [state, setState] = useState<UploadState>({
    selectedFile: null,
    originalFile: null,
    previewUrl: null,
    isProcessing: false,
    showCropEditor: false,
    cropData: null,
    error: null,
  });

  // Cleanup preview URLs on unmount
  useEffect(() => {
    return () => {
      if (state.previewUrl) {
        cleanupImagePreviewUrl(state.previewUrl);
      }
    };
  }, [state.previewUrl]);

  // Handle file selection
  const handleFileSelect = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    const file = files[0]; // Take first file if multiple selected
    
    try {
      // Clear previous state
      if (state.previewUrl) {
        cleanupImagePreviewUrl(state.previewUrl);
      }
      
      setState(prev => ({
        ...prev,
        selectedFile: null,
        originalFile: null,
        previewUrl: null,
        error: null,
        isProcessing: true,
      }));

      // Validate file
      const validationResult = await imageProcessor.validateImage(file);
      
      if (!validationResult.isValid) {
        setState(prev => ({
          ...prev,
          isProcessing: false,
          error: validationResult.error || 'Invalid file',
        }));
        onError(validationResult.error || 'Invalid file');
        return;
      }

      // Create preview URL
      const previewUrl = createImagePreviewUrl(file);
      
      // Automatically resize and process the image
      try {
        console.log('🔄 AvatarImageUpload: Starting resize process for file:', {
          name: file.name,
          size: file.size,
          type: file.type,
          constructor: file.constructor.name,
        });

        const resizedFile = await imageProcessor.resizeImage(file, 512);
        console.log('🖼️ AvatarImageUpload: Auto-resized image:', {
          originalSize: file.size,
          resizedSize: resizedFile?.size || 'unknown',
          name: resizedFile?.name || 'unknown',
        });

        // Create new preview URL for the resized file
        const resizedPreviewUrl = createImagePreviewUrl(resizedFile);
        console.log('🔗 AvatarImageUpload: Created resized preview URL:', resizedPreviewUrl);
        
        // Clean up original preview URL
        console.log('🧹 AvatarImageUpload: Cleaning up original preview URL:', previewUrl);
        cleanupImagePreviewUrl(previewUrl);
        
        // Create a copy of the original file to ensure it's not modified
        console.log('📋 AvatarImageUpload: Creating original file copy...');
        const originalFileCopy = new File([file], file.name, {
          type: file.type,
          lastModified: file.lastModified,
        });
        console.log('📋 AvatarImageUpload: Original file copy created:', {
          name: originalFileCopy.name,
          size: originalFileCopy.size,
          type: originalFileCopy.type,
          constructor: originalFileCopy.constructor.name,
        });

        setState(prev => ({
          ...prev,
          selectedFile: resizedFile,
          originalFile: originalFileCopy, // Keep original file copy for cropping
          previewUrl: resizedPreviewUrl,
          isProcessing: false,
          error: null,
        }));

        // Notify parent component with processed file
        onImageSelected(resizedFile);
        
      } catch (resizeError) {
        console.warn('Auto-resize failed, using original file:', resizeError);
        
        // Create a copy of the original file to ensure it's not modified
        const originalFileCopy = new File([file], file.name, {
          type: file.type,
          lastModified: file.lastModified,
        });

        setState(prev => ({
          ...prev,
          selectedFile: file,
          originalFile: originalFileCopy, // Original file copy is the same when resize fails
          previewUrl,
          isProcessing: false,
          error: null,
        }));

        // Notify parent component with original file
        onImageSelected(file);
      }

    } catch (error) {
      const errorMessage = getUserFriendlyMessage(error);
      setState(prev => ({
        ...prev,
        isProcessing: false,
        error: errorMessage,
      }));
      onError(errorMessage);
    }
  }, [state.previewUrl, onImageSelected, onError]);

  // Handle crop editor open
  const handleOpenCropEditor = useCallback(() => {
    console.log('🎯 AvatarImageUpload: Opening crop editor');
    console.log('📁 AvatarImageUpload: Current state.originalFile:', {
      exists: !!state.originalFile,
      name: state.originalFile?.name,
      size: state.originalFile?.size,
      type: state.originalFile?.type,
      constructor: state.originalFile?.constructor.name,
    });
    console.log('📁 AvatarImageUpload: Current state.selectedFile:', {
      exists: !!state.selectedFile,
      name: state.selectedFile?.name,
      size: state.selectedFile?.size,
      type: state.selectedFile?.type,
    });
    setState(prev => ({ ...prev, showCropEditor: true }));
  }, [state.originalFile, state.selectedFile]);

  // Handle crop confirmation
  const handleCropConfirm = useCallback(async (cropData: CropData) => {
    if (!state.originalFile) return;

    try {
      setState(prev => ({
        ...prev,
        showCropEditor: false,
        isProcessing: true,
        error: null,
      }));

      // Crop the original image (not the resized one)
      const croppedFile = await imageProcessor.cropImage(state.originalFile, cropData);
      
      // Resize the cropped image
      const resizedFile = await imageProcessor.resizeImage(croppedFile, 512);

      // Create new preview URL for the processed file
      const newPreviewUrl = createImagePreviewUrl(resizedFile);
      
      // Clean up old preview URL
      if (state.previewUrl) {
        cleanupImagePreviewUrl(state.previewUrl);
      }

      // Update state with processed file and new preview
      // Clear cropData since the image has already been cropped
      setState(prev => ({
        ...prev,
        selectedFile: resizedFile,
        previewUrl: newPreviewUrl,
        isProcessing: false,
        cropData: null, // Clear crop data since image is already cropped
      }));

      // Notify parent component
      console.log('🖼️ AvatarImageUpload: Sending processed file to parent:', {
        name: resizedFile.name,
        size: resizedFile.size,
        type: resizedFile.type,
      });
      onImageSelected(resizedFile);

    } catch (error) {
      console.error('Crop processing error:', error);
      const errorMessage = 'Failed to process image. Please try again.';
      setState(prev => ({
        ...prev,
        isProcessing: false,
        showCropEditor: false,
        error: errorMessage,
      }));
      onError(errorMessage);
    }
  }, [state.originalFile, state.previewUrl, onImageSelected, onError]);

  // Handle crop cancellation
  const handleCropCancel = useCallback(() => {
    setState(prev => ({ ...prev, showCropEditor: false }));
  }, []);

  // Handle remove image
  const handleRemoveImage = useCallback(() => {
    if (state.previewUrl) {
      cleanupImagePreviewUrl(state.previewUrl);
    }
    
    setState({
      selectedFile: null,
      originalFile: null,
      previewUrl: null,
      isProcessing: false,
      showCropEditor: false,
      cropData: null,
      error: null,
    });

    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }, [state.previewUrl]);

  // Handle click on upload area
  const handleUploadAreaClick = useCallback(() => {
    if (!disabled && fileInputRef.current) {
      fileInputRef.current.click();
    }
  }, [disabled]);

  const containerClasses = [
    'relative',
    'border-2 border-dashed border-gray-300',
    'rounded-lg',
    'p-6',
    'text-center',
    'transition-colors duration-200',
    disabled ? 'opacity-50 cursor-not-allowed' : 'hover:border-gray-400 cursor-pointer',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div className="space-y-4">
      {/* Main upload area */}
      <div 
        className={containerClasses}
        onClick={handleUploadAreaClick}
        data-testid="avatar-upload-container"
      >
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          onChange={handleFileSelect}
          disabled={disabled}
          className="hidden"
          data-testid="file-input"
        />

        {state.isProcessing ? (
          <div className="py-8" data-testid="processing-indicator">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Processing image...</p>
          </div>
        ) : state.selectedFile && state.previewUrl ? (
          <div className="space-y-4" data-testid="image-preview">
            <div className="flex justify-center">
              <CircularAvatarPreview
                imageUrl={state.previewUrl}
                cropData={state.cropData || undefined}
                size="large"
                showBorder
              />
            </div>
            
            <div className="space-y-2">
              <p className="text-sm text-gray-600">
                {state.selectedFile.name} ({Math.round(state.selectedFile.size / 1024)} KB)
              </p>
              
              <div className="flex justify-center space-x-3">
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleOpenCropEditor();
                  }}
                  className="px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 border border-blue-200 rounded-md hover:bg-blue-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  Crop Image
                </button>
                

                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRemoveImage();
                  }}
                  className="px-4 py-2 text-sm font-medium text-gray-600 bg-gray-50 border border-gray-200 rounded-md hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Remove
                </button>
              </div>
            </div>
          </div>
        ) : currentAvatarUrl ? (
          <div className="space-y-4" data-testid="current-avatar">
            <div className="flex justify-center">
              <CircularAvatarPreview
                imageUrl={currentAvatarUrl}
                size="large"
                showBorder
              />
            </div>
            
            <div className="space-y-2">
              <p className="text-sm text-gray-600">Current avatar</p>
              <p className="text-blue-600 font-medium">Choose Image to Change</p>
            </div>
          </div>
        ) : (
          <div className="py-8">
            <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
              <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
            </div>
            
            <h3 className="text-lg font-medium text-gray-900 mb-2">Choose Image</h3>
            <p className="text-sm text-gray-600 mb-4">
              Upload a photo to use as your avatar
            </p>
            <p className="text-xs text-gray-500">
              JPEG, PNG, or WebP • Max 10MB
            </p>
          </div>
        )}
      </div>

      {/* Error message */}
      {state.error && (
        <div 
          className="p-3 bg-red-50 border border-red-200 rounded-md"
          data-testid="error-message"
        >
          <p className="text-sm text-red-600">{state.error}</p>
        </div>
      )}

      {/* Crop editor modal */}
      {state.showCropEditor && state.originalFile && (
        <ImageCropEditor
          imageFile={state.originalFile}
          onCropConfirm={handleCropConfirm}
          onCancel={handleCropCancel}
          isOpen={state.showCropEditor}
        />
      )}
    </div>
  );
};