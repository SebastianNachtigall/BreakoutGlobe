package uploads

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ImageUploader handles POI image uploads
type ImageUploader struct {
	uploadDir string
	baseURL   string
}

// NewImageUploader creates a new ImageUploader instance
func NewImageUploader(uploadDir, baseURL string) *ImageUploader {
	return &ImageUploader{
		uploadDir: uploadDir,
		baseURL:   baseURL,
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

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(u.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := getFileExtension(imageFile.Filename)
	filename := fmt.Sprintf("poi-%s-%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(u.uploadDir, filename)

	// Open uploaded file
	src, err := imageFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return public URL
	imageURL := fmt.Sprintf("%s/uploads/%s", u.baseURL, filename)
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