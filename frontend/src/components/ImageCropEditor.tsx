import React, { useState, useRef, useCallback } from 'react';
import ReactCrop, { Crop, PixelCrop } from 'react-image-crop';
import 'react-image-crop/dist/ReactCrop.css';
import { CropData } from '../types/imageProcessing';
import { CircularAvatarPreview } from './CircularAvatarPreview';

export interface ImageCropEditorProps {
  imageFile: File;
  onCropConfirm: (cropData: CropData) => void;
  onCancel: () => void;
  isOpen: boolean;
}

export const ImageCropEditor: React.FC<ImageCropEditorProps> = ({
  imageFile,
  onCropConfirm,
  onCancel,
  isOpen,
}) => {
  const imgRef = useRef<HTMLImageElement>(null);
  const [crop, setCrop] = useState<Crop>({
    unit: '%',
    width: 80,
    height: 80,
    x: 10,
    y: 10,
  });
  const [completedCrop, setCompletedCrop] = useState<PixelCrop>();
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);

  // Create preview URL when component mounts
  React.useEffect(() => {
    if (imageFile) {
      const url = URL.createObjectURL(imageFile);
      setPreviewUrl(url);
      setIsLoading(false);
      
      return () => {
        URL.revokeObjectURL(url);
      };
    }
  }, [imageFile]);

  // Handle crop change - ensure square aspect ratio
  const onCropChange = useCallback((crop: Crop, percentCrop: Crop) => {
    // Force square aspect ratio
    const size = Math.min(crop.width || 0, crop.height || 0);
    setCrop({
      ...crop,
      width: size,
      height: size,
    });
  }, []);

  // Handle crop complete
  const onCropComplete = useCallback((crop: PixelCrop) => {
    setCompletedCrop(crop);
  }, []);

  // Convert crop to our CropData format and confirm
  const handleConfirm = useCallback(() => {
    if (!completedCrop || !imgRef.current) return;

    const image = imgRef.current;
    
    // completedCrop is always in display pixels (PixelCrop type)
    const scaleX = image.naturalWidth / image.width;
    const scaleY = image.naturalHeight / image.height;

    const cropData: CropData = {
      x: Math.round(completedCrop.x * scaleX),
      y: Math.round(completedCrop.y * scaleY),
      width: Math.round(completedCrop.width * scaleX),
      height: Math.round(completedCrop.height * scaleY),
      scale: 1,
    };

    onCropConfirm(cropData);
  }, [completedCrop, onCropConfirm]);

  // Reset crop to center
  const handleReset = useCallback(() => {
    setCrop({
      unit: '%',
      width: 80,
      height: 80,
      x: 10,
      y: 10,
    });
  }, []);

  // Handle backdrop click
  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onCancel();
    }
  };

  if (!isOpen) {
    return null;
  }

  return (
    <div 
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[9999]"
      data-testid="modal-backdrop"
      onClick={handleBackdropClick}
    >
      <div 
        className="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto"
        data-testid="crop-editor-modal"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6" data-testid="modal-content">
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              Crop Image
            </h2>
            <p className="text-gray-600">
              Drag to select the area you want to use for your avatar. The selection will be cropped to a square.
            </p>
          </div>

          {isLoading && (
            <div 
              className="flex items-center justify-center py-12"
              data-testid="crop-editor-loading"
            >
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              <span className="ml-3 text-gray-600">Loading image...</span>
            </div>
          )}

          {!isLoading && (
            <div className="space-y-6">
              {/* Main crop area */}
              <div className="flex gap-6">
                <div className="flex-1">
                  <ReactCrop
                    crop={crop}
                    onChange={onCropChange}
                    onComplete={onCropComplete}
                    aspect={1} // Force square aspect ratio
                    minWidth={50}
                    minHeight={50}
                  >
                    <img
                      ref={imgRef}
                      src={previewUrl}
                      alt="Crop preview"
                      className="max-w-full max-h-96 object-contain"
                      onLoad={() => {
                        // Set initial crop when image loads
                        const img = imgRef.current;
                        if (img) {
                          const { width, height } = img;
                          const size = Math.min(width, height) * 0.8;
                          const x = (width - size) / 2;
                          const y = (height - size) / 2;
                          
                          setCrop({
                            unit: 'px',
                            width: size,
                            height: size,
                            x,
                            y,
                          });
                        }
                      }}
                    />
                  </ReactCrop>
                </div>

                {/* Preview panel */}
                <div className="w-48 space-y-4">
                  <div>
                    <h3 className="text-sm font-medium text-gray-700 mb-2">Preview</h3>
                    <div className="flex justify-center">
                      <div data-testid="circular-preview">
                        <CircularAvatarPreview
                          imageUrl={previewUrl}
                          cropData={completedCrop && imgRef.current ? (() => {
                            const img = imgRef.current!;
                            
                            // completedCrop is always in display pixels (PixelCrop type)
                            const scaleX = img.naturalWidth / img.width;
                            const scaleY = img.naturalHeight / img.height;
                            
                            return {
                              x: Math.round(completedCrop.x * scaleX),
                              y: Math.round(completedCrop.y * scaleY),
                              width: Math.round(completedCrop.width * scaleX),
                              height: Math.round(completedCrop.height * scaleY),
                              scale: 1,
                            };
                          })() : undefined}
                          size="large"
                        />
                      </div>
                    </div>
                  </div>

                  {completedCrop && (
                    <div data-testid="crop-info">
                      <h3 className="text-sm font-medium text-gray-700 mb-2">Crop Info</h3>
                      <div className="text-xs text-gray-600 space-y-1">
                        <div>Size: {Math.round(completedCrop.width)} × {Math.round(completedCrop.height)}</div>
                        <div>Position: ({Math.round(completedCrop.x)}, {Math.round(completedCrop.y)})</div>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Controls */}
              <div className="flex justify-between items-center pt-4 border-t border-gray-200">
                <button
                  type="button"
                  onClick={handleReset}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Reset
                </button>

                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={onCancel}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handleConfirm}
                    disabled={!completedCrop}
                    className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Crop Image
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};