package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalFileStorage implements file storage for local filesystem
type LocalFileStorage struct {
	config StorageConfig
}

// NewLocalFileStorage creates a new local file storage instance
func NewLocalFileStorage(config StorageConfig) *LocalFileStorage {
	return &LocalFileStorage{
		config: config,
	}
}

// UploadFile saves a file to the local filesystem
func (l *LocalFileStorage) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	// Validate file size
	if int64(len(data)) > l.config.MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", l.config.MaxFileSize)
	}

	// Sanitize the key to prevent path traversal
	key = sanitizeFilePath(key)
	
	// Full file path
	filePath := filepath.Join(l.config.UploadPath, key)
	
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	
	// Return public URL
	return l.GetFileURL(key), nil
}

// DeleteFile removes a file from the local filesystem
func (l *LocalFileStorage) DeleteFile(ctx context.Context, key string) error {
	key = sanitizeFilePath(key)
	filePath := filepath.Join(l.config.UploadPath, key)
	
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// GetFileURL returns the public URL for a file
func (l *LocalFileStorage) GetFileURL(key string) string {
	return fmt.Sprintf("%s/uploads/%s", l.config.BaseURL, key)
}

// FileExists checks if a file exists
func (l *LocalFileStorage) FileExists(key string) bool {
	key = sanitizeFilePath(key)
	filePath := filepath.Join(l.config.UploadPath, key)
	_, err := os.Stat(filePath)
	return err == nil
}

// GenerateUniqueKey generates a unique file key with timestamp
func (l *LocalFileStorage) GenerateUniqueKey(prefix, userID, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/%s_%d%s", prefix, userID, timestamp, ext)
}

// sanitizeFilePath prevents path traversal attacks
func sanitizeFilePath(path string) string {
	// Remove any path traversal attempts
	path = strings.ReplaceAll(path, "..", "")
	path = strings.ReplaceAll(path, `\`, "/")
	
	// Remove leading slashes
	path = strings.TrimPrefix(path, "/")
	
	return path
}