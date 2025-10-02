package storage

import "context"

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	GetFileURL(key string) string
	FileExists(key string) bool
	GenerateUniqueKey(prefix, userID, originalFilename string) string
}

// NewFileStorage creates a new file storage instance based on configuration
func NewFileStorage(config StorageConfig) FileStorage {
	// For now, we only support local storage
	// In the future, we can add S3 or other storage backends here
	return NewLocalFileStorage(config)
}