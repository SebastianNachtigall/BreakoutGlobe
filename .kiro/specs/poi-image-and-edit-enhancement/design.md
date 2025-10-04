# Design Document

## Overview

This design enhances the POI creation experience by adding image preview functionality and optimizing image storage. The solution reuses existing avatar upload components (`AvatarImageUpload`, `ImagePreview`) and adds backend thumbnail generation to improve performance when displaying many POIs on the map.

## Architecture

### Component Layers

```
Frontend (React/TypeScript)
├── POISidebar.tsx (Enhanced)
│   └── Integrates AvatarImageUpload for preview
├── AvatarImageUpload.tsx (Reused)
│   └── Handles file selection, validation, preview
└── POIDetailsPanel.tsx (Enhanced)
    └── Displays full image vs thumbnail

Backend (Go)
├── handlers/poi_handler.go (Enhanced)
│   └── Handles multipart form with image
├── services/poi_service.go (Enhanced)
│   └── Orchestrates thumbnail generation
└── storage/image_processor.go (New)
    └── Generates thumbnails
```

### Data Flow

**POI Creation with Image:**
1. User selects image → `AvatarImageUpload` validates and shows preview
2. User submits form → Frontend sends multipart/form-data
3. Backend receives image → Validates file
4. Backend generates thumbnail (200x200px)
5. Backend saves both original and thumbnail
6. Backend returns POI with `imageUrl` and `thumbnailUrl`
7. Frontend displays POI on map using thumbnail

**Image Display:**
- Map markers: Use `thumbnailUrl` with CSS `border-radius: 50%`
- Details panel: Use `imageUrl` (full rectangular image)
- Fallback: If `thumbnailUrl` missing, use `imageUrl`

## Components and Interfaces

### Frontend Changes

#### 1. POISidebar.tsx Enhancement

**Current State:**
- Basic file input without preview
- No validation feedback
- No reuse of existing components

**Changes:**
```typescript
// Add state for image handling
const [selectedImage, setSelectedImage] = useState<File | null>(null);
const [imageError, setImageError] = useState<string | null>(null);

// Replace basic file input with AvatarImageUpload
<AvatarImageUpload
  onImageSelected={(file) => setSelectedImage(file)}
  onError={(error) => setImageError(error)}
  disabled={isSubmitting}
/>
```

**Integration Points:**
- Import `AvatarImageUpload` component
- Handle image file in form submission
- Send as multipart/form-data instead of JSON

#### 2. POI Store Enhancement

**Changes to `poiStore.ts`:**
```typescript
// Update createPOI to handle image files
async createPOI(data: CreatePOIData, imageFile?: File) {
  const formData = new FormData();
  formData.append('mapId', data.mapId);
  formData.append('name', data.name);
  // ... other fields
  if (imageFile) {
    formData.append('image', imageFile);
  }
  
  const response = await fetch('/api/pois', {
    method: 'POST',
    body: formData, // No Content-Type header - browser sets it
  });
  // ...
}
```

#### 3. POI Display Components

**POISidebar.tsx (List View):**
```typescript
// Use thumbnail for list items
<img
  src={poi.thumbnailUrl || poi.imageUrl}
  className="w-10 h-10 rounded object-cover"
/>
```

**AvatarMarker.tsx (Map Markers):**
```typescript
// Use thumbnail with circular clipping
<img
  src={poi.thumbnailUrl || poi.imageUrl}
  className="w-full h-full rounded-full object-cover"
/>
```

**POIDetailsPanel.tsx (Detail View):**
```typescript
// Use full original image
<img
  src={poi.imageUrl}
  className="w-full h-48 object-cover rounded-lg"
/>
```

### Backend Changes

#### 1. Image Processor Service (New)

**File:** `backend/internal/storage/image_processor.go`

```go
type ImageProcessor struct {
    storage FileStorage
}

// GenerateThumbnail creates a 200x200px thumbnail
func (ip *ImageProcessor) GenerateThumbnail(
    ctx context.Context,
    originalData []byte,
    contentType string,
) ([]byte, error) {
    // Decode image
    // Resize to 200x200 (maintaining aspect ratio, center crop)
    // Encode as JPEG with quality 85
    // Return thumbnail data
}

// ProcessPOIImage handles both original and thumbnail
func (ip *ImageProcessor) ProcessPOIImage(
    ctx context.Context,
    poiID string,
    imageFile *multipart.FileHeader,
) (originalURL, thumbnailURL string, err error) {
    // Read file data
    // Save original: {poiID}-original.{ext}
    // Generate thumbnail
    // Save thumbnail: {poiID}-thumb.jpg
    // Return both URLs
}
```

**Dependencies:**
- Use `github.com/disintegration/imaging` for image processing
- Supports JPEG, PNG, WebP input
- Always output thumbnails as JPEG for consistency

#### 2. POI Service Enhancement

**File:** `backend/internal/services/poi_service.go`

**Changes:**
```go
type POIService struct {
    // ... existing fields
    imageProcessor *ImageProcessor // Add image processor
}

// Enhance CreatePOIWithImage
func (s *POIService) CreatePOIWithImage(...) (*models.POI, error) {
    // Existing validation...
    
    if imageFile != nil && s.imageProcessor != nil {
        // Process image (original + thumbnail)
        originalURL, thumbnailURL, err := s.imageProcessor.ProcessPOIImage(
            ctx, poi.ID, imageFile,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to process image: %w", err)
        }
        
        poi.ImageURL = originalURL
        poi.ThumbnailURL = thumbnailURL
    }
    
    // Save to database...
}
```

#### 3. POI Model Enhancement

**File:** `backend/internal/models/poi.go`

**Changes:**
```go
type POI struct {
    // ... existing fields
    ImageURL     string `json:"imageUrl,omitempty" gorm:"type:varchar(500)"`
    ThumbnailURL string `json:"thumbnailUrl,omitempty" gorm:"type:varchar(500)"` // NEW
    // ... rest of fields
}
```

**Migration:**
```go
// Add migration to add thumbnail_url column
ALTER TABLE pois ADD COLUMN thumbnail_url VARCHAR(500);
```

#### 4. Handler Enhancement

**File:** `backend/internal/handlers/poi_handler.go`

**Changes:**
```go
type CreatePOIResponse struct {
    // ... existing fields
    ImageURL     string `json:"imageUrl,omitempty"`
    ThumbnailURL string `json:"thumbnailUrl,omitempty"` // NEW
}

type POIInfo struct {
    // ... existing fields
    ImageURL     string `json:"imageUrl,omitempty"`
    ThumbnailURL string `json:"thumbnailUrl,omitempty"` // NEW
}
```

## Data Models

### POI Model (Enhanced)

```go
type POI struct {
    ID              string
    MapID           string
    Name            string
    Description     string
    Position        LatLng
    CreatedBy       string
    MaxParticipants int
    ImageURL        string  // Original image URL
    ThumbnailURL    string  // Thumbnail URL (NEW)
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### Frontend POI Type (Enhanced)

```typescript
interface POIData {
    id: string;
    mapId: string;
    name: string;
    description: string;
    position: { lat: number; lng: number };
    createdBy: string;
    maxParticipants: number;
    participantCount: number;
    imageUrl?: string;
    thumbnailUrl?: string; // NEW
    createdAt: string;
}
```

## Error Handling

### Frontend Errors

1. **Image Validation Errors**
   - Reuse existing `imageErrors.ts` utilities
   - Display user-friendly messages from `AvatarImageUpload`
   - Handle: file too large, invalid type, invalid dimensions

2. **Upload Errors**
   - Network failures → Retry mechanism
   - Server errors → Display error message
   - Timeout → Cancel and notify user

### Backend Errors

1. **Image Processing Errors**
   - Invalid image format → 400 Bad Request
   - Image too large → 413 Payload Too Large
   - Processing failure → 500 Internal Server Error
   - Cleanup partial uploads on failure

2. **Storage Errors**
   - Disk full → 507 Insufficient Storage
   - Permission denied → 500 Internal Server Error
   - Log errors for monitoring

## Testing Strategy

### Frontend Tests

1. **Component Tests**
   - `POISidebar.test.tsx`: Image upload integration
   - Test image selection triggers preview
   - Test form submission with image
   - Test error handling

2. **Integration Tests**
   - Test complete POI creation flow with image
   - Test image display in different contexts (map, list, details)
   - Test fallback when thumbnail missing

### Backend Tests

1. **Unit Tests**
   - `image_processor_test.go`: Thumbnail generation
   - Test various image formats (JPEG, PNG, WebP)
   - Test aspect ratio handling
   - Test error cases (corrupt images, unsupported formats)

2. **Integration Tests**
   - Test complete POI creation with image upload
   - Test file storage and retrieval
   - Test thumbnail generation in request flow
   - Test backward compatibility (POIs without thumbnails)

3. **Handler Tests**
   - Test multipart form parsing
   - Test response includes both URLs
   - Test error responses

## Performance Considerations

### Image Processing

- **Thumbnail Size:** 200x200px balances quality and file size
- **Format:** JPEG with 85% quality for thumbnails
- **Processing Time:** ~50-200ms per image (acceptable for creation flow)
- **Memory:** Process images in streaming mode to limit memory usage

### Storage

- **File Naming:** `{poi-id}-original.{ext}` and `{poi-id}-thumb.jpg`
- **Directory Structure:** `uploads/pois/`
- **Cleanup:** Delete both files when POI deleted

### Frontend

- **Lazy Loading:** Load images only when visible
- **Caching:** Browser caches images by URL
- **Fallback:** Graceful degradation if thumbnail missing

## Migration Strategy

### Database Migration

```sql
-- Add thumbnail_url column
ALTER TABLE pois ADD COLUMN thumbnail_url VARCHAR(500);

-- No data migration needed - new field is optional
-- Existing POIs will have NULL thumbnail_url
-- Frontend will fall back to imageUrl
```

### Backward Compatibility

1. **Frontend:** Check for `thumbnailUrl` existence, fall back to `imageUrl`
2. **Backend:** Generate thumbnails on-demand for existing POIs (optional future enhancement)
3. **API:** Both fields optional in responses

## Security Considerations

1. **File Validation**
   - Validate file type (JPEG, PNG, WebP only)
   - Validate file size (max 10MB)
   - Validate image dimensions (reasonable limits)

2. **Path Traversal Prevention**
   - Sanitize file names
   - Use UUID-based naming
   - Store in dedicated upload directory

3. **Content Type Verification**
   - Verify actual file content matches extension
   - Use magic number detection

4. **Rate Limiting**
   - Existing rate limiting applies to POI creation
   - No additional limits needed for image upload

## Deployment Considerations

1. **Storage Requirements**
   - Each POI with image: ~100KB (original) + ~20KB (thumbnail)
   - Estimate storage needs based on expected POI count

2. **Dependencies**
   - Add `github.com/disintegration/imaging` to Go dependencies
   - No new frontend dependencies (reusing existing components)

3. **Configuration**
   - Thumbnail size configurable via environment variable
   - JPEG quality configurable via environment variable

4. **Monitoring**
   - Log image processing failures
   - Monitor storage usage
   - Track thumbnail generation performance
