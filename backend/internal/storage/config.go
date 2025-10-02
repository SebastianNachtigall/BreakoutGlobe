package storage

import (
	"os"
	"path/filepath"
)

// StorageConfig holds configuration for file storage
type StorageConfig struct {
	UploadPath string
	BaseURL    string
	MaxFileSize int64 // in bytes
}

// GetStorageConfig returns storage configuration based on environment
func GetStorageConfig() StorageConfig {
	var uploadPath string
	var baseURL string

	// Detect environment and set appropriate paths
	if isRailwayEnvironment() {
		// Railway production environment
		uploadPath = "/app/uploads"
		baseURL = getRailwayBaseURL()
	} else {
		// Local development environment
		uploadPath = "./uploads"
		baseURL = getLocalBaseURL()
	}

	return StorageConfig{
		UploadPath:  uploadPath,
		BaseURL:     baseURL,
		MaxFileSize: 5 * 1024 * 1024, // 5MB default
	}
}

// isRailwayEnvironment checks if we're running on Railway
func isRailwayEnvironment() bool {
	return os.Getenv("RAILWAY_ENVIRONMENT") != ""
}

// getRailwayBaseURL constructs the base URL for Railway deployment
func getRailwayBaseURL() string {
	// Railway sets RAILWAY_PUBLIC_DOMAIN automatically
	if domain := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); domain != "" {
		return "https://" + domain
	}
	
	// Fallback to custom domain if set
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}
	
	// Default Railway domain pattern (fallback)
	return "https://breakoutglobe-production.up.railway.app"
}

// getLocalBaseURL returns the base URL for local development
func getLocalBaseURL() string {
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}
	return "http://localhost:8080"
}

// EnsureUploadDirectories creates necessary upload directories
func EnsureUploadDirectories(config StorageConfig) error {
	directories := []string{
		filepath.Join(config.UploadPath, "avatars"),
		filepath.Join(config.UploadPath, "poi-images"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}