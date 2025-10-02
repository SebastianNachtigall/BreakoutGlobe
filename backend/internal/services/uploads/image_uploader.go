package uploads

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"breakoutglobe/internal/storage"
	"github.com/google/uuid"
)

// ImageUploader handles POI image uploads
type ImageUploader struct {
	storage storage.FileStorage
}

// NewImageUploader creates a new ImageUploader instance
func NewImageUploader(fileStorage storage.FileStorage) *ImageUploader {
	return &ImageUploader{
		storage: fileStorage,
	}
}

// UploadPOIImage uploads a POI image and returns the URL
func (u *ImageUploader) UploadPOIImage(ctx context.Context, imageFile *multipart.FileHeader) (string, error) {
	if imageFile == nil {
		return "", fmt.Errorf("no image file provided")
	}

	// Validate file type
	contentType := imageFile.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return "", fmt.Errorf("invalid image type: %s", contentType)
	}

	// Validate file size (max 5MB)
	if imageFile.Size > 5*1024*1024 {
		return "", fmt.Errorf("image file too large: %d bytes (max 5MB)", imageFile.Size)
	}

	// Generate unique filename
	ext := getFileExtension(imageFile.Filename)
	filename := fmt.Sprintf("poi-%s-%d%s", uuid.New().String(), time.Now().Unix(), ext)
	
	// Open uploaded file
	src, err := imageFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file data
	fileData, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file data: %w", err)
	}

	// Upload using storage system
	imageURL, err := u.storage.UploadFile(ctx, filename, fileData, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload POI image: %w", err)
	}

	return imageURL, nil
}

// isValidImageType checks if the content type is a valid image type
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/webp",
	}
	
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// getFileExtension returns the file extension from filename
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg" // Default extension
	}
	return strings.ToLower(ext)
}