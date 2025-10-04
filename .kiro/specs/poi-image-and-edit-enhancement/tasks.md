# Implementation Plan

- [x] 1. Backend: Add database migration for thumbnail URL
  - Add `thumbnail_url` column to `pois` table in migrations
  - Update POI model struct to include `ThumbnailURL` field
  - _Requirements: 2.2, 2.5_

- [x] 2. Backend: Implement image processing service
  - [x] 2.1 Create `image_processor.go` with thumbnail generation
    - Implement `GenerateThumbnail()` method to create 200x200px thumbnails
    - Implement `ProcessPOIImage()` to handle original + thumbnail storage
    - Use `github.com/disintegration/imaging` library for image processing
    - Support JPEG, PNG, WebP input formats
    - Output thumbnails as JPEG with 85% quality
    - _Requirements: 2.1, 2.2_
  
  - [ ]* 2.2 Write unit tests for image processor
    - Test thumbnail generation with various image formats
    - Test aspect ratio handling and center cropping
    - Test error cases (corrupt images, unsupported formats)
    - _Requirements: 2.1_

- [x] 3. Backend: Enhance POI service for thumbnail support
  - [x] 3.1 Add ImageProcessor to POIService
    - Update `NewPOIServiceWithImageUploader` to include image processor
    - Modify `CreatePOIWithImage` to call `ProcessPOIImage`
    - Store both `ImageURL` and `ThumbnailURL` in POI model
    - _Requirements: 2.1, 2.2, 2.5_
  
  - [ ]* 3.2 Write service tests for thumbnail creation
    - Test POI creation with image generates both URLs
    - Test POI creation without image works as before
    - Test error handling when thumbnail generation fails
    - _Requirements: 2.1, 2.2_

- [x] 4. Backend: Update POI handler responses
  - Update `CreatePOIResponse` to include `ThumbnailURL` field
  - Update `GetPOIResponse` to include `ThumbnailURL` field
  - Update `POIInfo` struct to include `ThumbnailURL` field
  - Ensure all POI endpoints return thumbnail URL when available
  - _Requirements: 2.5_

- [x] 5. Frontend: Integrate AvatarImageUpload in POI creation
  - [x] 5.1 Update POISidebar component
    - Import and integrate `AvatarImageUpload` component
    - Add state management for selected image file
    - Add error handling for image validation
    - Remove old basic file input
    - _Requirements: 1.1, 1.3, 1.4_
  
  - [x] 5.2 Update POI creation form submission
    - Change from JSON to multipart/form-data
    - Include image file in form data when selected
    - Handle submission with and without image
    - _Requirements: 1.1, 1.4_
  
  - [ ]* 5.3 Write component tests for POI creation with image
    - Test image selection shows preview
    - Test form submission includes image file
    - Test validation error display
    - Test creation without image still works
    - _Requirements: 1.1, 1.3_

- [x] 6. Frontend: Update POI store for image upload
  - Modify `createPOI` method to accept optional image file parameter
  - Build FormData with POI fields and image
  - Send multipart/form-data request instead of JSON
  - Handle response with both `imageUrl` and `thumbnailUrl`
  - _Requirements: 1.1, 1.4_

- [x] 7. Frontend: Update POI display components for thumbnails
  - [x] 7.1 Update POISidebar list view
    - Use `thumbnailUrl` for POI list items
    - Fall back to `imageUrl` if thumbnail missing
    - Maintain circular display with CSS
    - _Requirements: 2.3, 3.3_
  
  - [x] 7.2 Update AvatarMarker for map display
    - Use `thumbnailUrl` for map markers
    - Fall back to `imageUrl` if thumbnail missing
    - Apply circular clipping with CSS
    - _Requirements: 2.3, 3.3_
  
  - [x] 7.3 Update POIDetailsPanel for full image
    - Use `imageUrl` (original) for detail view
    - Display as rectangular image
    - Handle missing image gracefully
    - _Requirements: 2.4, 3.3_
  
  - [ ]* 7.4 Write display component tests
    - Test thumbnail display in list and map
    - Test full image display in details
    - Test fallback behavior when thumbnail missing
    - _Requirements: 2.3, 2.4, 3.3_

- [x] 8. Frontend: Update TypeScript types
  - Add `thumbnailUrl?: string` to POI interface
  - Update all POI-related type definitions
  - Ensure type safety across components
  - _Requirements: 2.5_

- [ ]* 9. Integration testing
  - [ ]* 9.1 Test complete POI creation flow with image
    - Test end-to-end: select image → preview → submit → display
    - Verify both original and thumbnail are stored
    - Verify correct URLs returned in response
    - _Requirements: 1.1, 2.1, 2.2_
  
  - [ ]* 9.2 Test backward compatibility
    - Test existing POIs without thumbnails display correctly
    - Test fallback to original image works
    - Test new POIs with thumbnails display correctly
    - _Requirements: 3.1, 3.2, 3.3_
