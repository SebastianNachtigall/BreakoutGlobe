package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

// ImageProcessor handles image processing operations like thumbnail generation
type ImageProcessor struct {
	storage FileStorage
}

// NewImageProcessor creates a new ImageProcessor instance
func NewImageProcessor(storage FileStorage) *ImageProcessor {
	return &ImageProcessor{
		storage: storage,
	}
}

// ProcessPOIImage processes a POI image by saving the original and generating a thumbnail
// Returns the original URL and thumbnail URL
func (ip *ImageProcessor) ProcessPOIImage(
	ctx context.Context,
	poiID string,
	imageFile *multipart.FileHeader,
) (originalURL, thumbnailURL string, err error) {
	// Open the uploaded file
	file, err := imageFile.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image file: %w", err)
	}

	// Detect content type
	contentType := imageFile.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // Default fallback
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(imageFile.Filename))
	if ext == "" {
		ext = ".jpg" // Default fallback
	}

	// Generate keys for original and thumbnail
	originalKey := fmt.Sprintf("pois/%s-original%s", poiID, ext)
	thumbnailKey := fmt.Sprintf("pois/%s-thumb.jpg", poiID)

	// Save original image
	originalURL, err = ip.storage.UploadFile(ctx, originalKey, data, contentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload original image: %w", err)
	}

	// Generate thumbnail
	thumbnailData, err := ip.GenerateThumbnail(data, contentType)
	if err != nil {
		// Clean up original if thumbnail generation fails
		_ = ip.storage.DeleteFile(ctx, originalKey)
		return "", "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Save thumbnail
	thumbnailURL, err = ip.storage.UploadFile(ctx, thumbnailKey, thumbnailData, "image/jpeg")
	if err != nil {
		// Clean up original if thumbnail upload fails
		_ = ip.storage.DeleteFile(ctx, originalKey)
		return "", "", fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	return originalURL, thumbnailURL, nil
}

// GenerateThumbnail creates a 200x200px thumbnail from image data
// The thumbnail is center-cropped to maintain aspect ratio and always outputs JPEG
func (ip *ImageProcessor) GenerateThumbnail(data []byte, contentType string) ([]byte, error) {
	// Decode image based on content type
	img, err := ip.decodeImage(bytes.NewReader(data), contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize and crop to 200x200
	thumbnail := imaging.Fill(img, 200, 200, imaging.Center, imaging.Lanczos)

	// Encode as JPEG with 85% quality
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// decodeImage decodes an image from a reader based on content type
func (ip *ImageProcessor) decodeImage(r io.Reader, contentType string) (image.Image, error) {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return jpeg.Decode(r)
	case "image/png":
		return png.Decode(r)
	case "image/webp":
		return webp.Decode(r)
	default:
		// Try generic decode as fallback
		img, _, err := image.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("unsupported image format: %s", contentType)
		}
		return img, nil
	}
}

// DeletePOIImages deletes both original and thumbnail images for a POI
func (ip *ImageProcessor) DeletePOIImages(ctx context.Context, poiID string) error {
	// Try to delete both original and thumbnail
	// We don't know the original extension, so we try common ones
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	
	var lastErr error
	for _, ext := range extensions {
		originalKey := fmt.Sprintf("pois/%s-original%s", poiID, ext)
		if err := ip.storage.DeleteFile(ctx, originalKey); err != nil {
			lastErr = err
		}
	}

	// Delete thumbnail (always .jpg)
	thumbnailKey := fmt.Sprintf("pois/%s-thumb.jpg", poiID)
	if err := ip.storage.DeleteFile(ctx, thumbnailKey); err != nil {
		lastErr = err
	}

	// Return last error if any (non-critical since files might not exist)
	return lastErr
}
