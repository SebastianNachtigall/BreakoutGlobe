# File Storage System

This package provides a flexible file storage system that works both locally and on Railway with persistent volumes.

## Features

- **Environment Detection**: Automatically detects Railway vs local development
- **Persistent Storage**: Uses Railway volumes in production, local filesystem in development
- **Security**: Path traversal protection and file type validation
- **Flexible**: Easy to extend with cloud storage backends (S3, etc.)

## Configuration

The storage system automatically configures itself based on environment:

### Local Development
- **Path**: `./uploads`
- **URL**: `http://localhost:8080/uploads/...`
- **Volume**: Bind mount in docker-compose

### Railway Production
- **Path**: `/app/uploads` (persistent volume)
- **URL**: `https://yourapp.railway.app/uploads/...`
- **Volume**: Railway persistent volume

## Environment Variables

- `RAILWAY_ENVIRONMENT`: Automatically set by Railway
- `RAILWAY_PUBLIC_DOMAIN`: Automatically set by Railway
- `BASE_URL`: Optional override for base URL

## Usage

```go
// Get storage configuration
config := storage.GetStorageConfig()

// Create storage instance
fileStorage := storage.NewFileStorage(config)

// Upload a file
url, err := fileStorage.UploadFile(ctx, "avatars/user123.jpg", data, "image/jpeg")

// Delete a file
err := fileStorage.DeleteFile(ctx, "avatars/user123.jpg")
```

## Directory Structure

```
uploads/
├── avatars/          # User avatar images
└── poi-images/       # POI images
```

## File Serving

Files are served via Gin static middleware:
- Route: `/uploads/*filepath`
- Security: Path traversal protection
- Caching: Proper cache headers for images

## Testing

The package includes mock implementations for testing:

```go
mockStorage := &MockFileStorage{}
mockStorage.On("UploadFile", ...).Return("url", nil)
```