# Requirements Document

## Introduction

This feature enhances the POI (Point of Interest) creation experience by adding image preview functionality and optimizing image storage. Currently, POI creation has a basic image upload without preview, and images are stored at full size. This enhancement will provide immediate image previews during creation by reusing the polished avatar upload components, and optimize storage by generating thumbnails for map display while keeping originals for detail views.

## Requirements

### Requirement 1: Image Preview During POI Creation

**User Story:** As a user creating a POI, I want to see a preview of my selected image immediately, so that I can confirm it looks correct before submitting.

#### Acceptance Criteria

1. WHEN a user selects an image file during POI creation THEN the system SHALL display an immediate preview of the selected image
2. WHEN a user uploads a POI image THEN the system SHALL validate file type, size, and dimensions using the same rules as avatar uploads
3. WHEN an invalid image is selected THEN the system SHALL display clear error messages explaining the validation failure
4. WHEN a POI image is uploaded THEN the system SHALL reuse existing image upload components from the avatar system (AvatarImageUpload, ImagePreview)

### Requirement 2: POI Image Storage Optimization

**User Story:** As a system administrator, I want POI images to be optimized for different display contexts, so that the application performs well with many POIs on the map.

#### Acceptance Criteria

1. WHEN a POI image is uploaded THEN the system SHALL generate a thumbnail version (200x200px) for map display
2. WHEN a POI image is uploaded THEN the system SHALL store both the original image and thumbnail with naming pattern `{poi-id}-original.{ext}` and `{poi-id}-thumb.{ext}`
3. WHEN a POI is displayed on the map THEN the system SHALL use the thumbnail image with circular CSS clipping
4. WHEN a POI details panel is opened THEN the system SHALL display the full original rectangular image
5. WHEN the backend returns POI data THEN the system SHALL include both `imageUrl` (original) and `thumbnailUrl` fields

### Requirement 3: Backward Compatibility for Existing POIs

**User Story:** As a system, I need to handle existing POIs that don't have thumbnails, so that the application continues to work with legacy data.

#### Acceptance Criteria

1. WHEN a POI without a thumbnail is displayed THEN the system SHALL fall back to using the original image
2. WHEN the backend detects a POI image without a thumbnail THEN the system SHALL generate the thumbnail on-demand
3. WHEN displaying POIs on the map THEN the system SHALL gracefully handle missing thumbnail URLs
