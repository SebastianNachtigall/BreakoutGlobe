# POI Creation Fix - Design Document

## Overview

Fix the broken POI creation workflow by completing API integration, fixing interface mismatches, and ensuring proper data flow from frontend form submission to backend persistence and real-time updates.

## Architecture

```
User Right-Click → Context Menu → Modal Form → API Service → Backend → Database
                                      ↓
                              Optimistic Update → Store → Map Markers
```

## Components and Interfaces

### Frontend API Service (`frontend/src/services/api.ts`)

Add POI-specific API functions:

```typescript
// POI API Types
export interface CreatePOIRequest {
  mapId: string;
  name: string;
  description: string;
  position: { lat: number; lng: number };
  createdBy: string;
  maxParticipants: number;
}

export interface POIResponse {
  id: string;
  mapId: string;
  name: string;
  description: string;
  position: { lat: number; lng: number };
  createdBy: string;
  maxParticipants: number;
  participantCount: number;
  participants: Array<{ id: string; name: string }>;
  createdAt: string;
}

// API Functions
export async function createPOI(request: CreatePOIRequest): Promise<POIResponse>
export async function getPOIs(mapId: string): Promise<POIResponse[]>
export async function updatePOI(poiId: string, updates: Partial<CreatePOIRequest>): Promise<POIResponse>
export async function deletePOI(poiId: string): Promise<void>
```

### App Component Integration Fix

Fix the interface mismatch and complete API integration:

1. **Fix prop name**: Change `onSubmit` to `onCreate` in POICreationModal usage
2. **Add proper API call**: Replace WebSocket call with HTTP API call
3. **Add loading states**: Show loading during API calls
4. **Add error handling**: Display errors to user

### Data Transformation

Map between frontend and backend data formats:

```typescript
// Frontend → Backend
const transformToAPIRequest = (formData: POICreationData, userId: string): CreatePOIRequest => ({
  mapId: 'default-map',
  name: formData.name,
  description: formData.description,
  position: formData.position,
  createdBy: userId,
  maxParticipants: formData.maxParticipants
});

// Backend → Frontend
const transformFromAPIResponse = (apiResponse: POIResponse): POIData => ({
  id: apiResponse.id,
  name: apiResponse.name,
  description: apiResponse.description,
  position: apiResponse.position,
  participantCount: apiResponse.participantCount,
  maxParticipants: apiResponse.maxParticipants,
  participants: apiResponse.participants,
  createdBy: apiResponse.createdBy,
  createdAt: new Date(apiResponse.createdAt)
});
```

## Data Models

### Frontend POI Data Structure
Ensure consistency with existing `POIData` interface in MapContainer.tsx - no changes needed.

### Backend API Contract
Use existing backend POI handler structure - no changes needed to backend.

## Error Handling

### Error Types and Responses
- **Network Errors**: Show retry option
- **Validation Errors**: Highlight specific form fields
- **Server Errors**: Show generic error message
- **Rate Limiting**: Show retry timing

### Error Recovery
- Rollback optimistic updates on failure
- Provide retry mechanisms for transient failures
- Clear error states on successful retry

## Testing Strategy

### Unit Tests
- API service functions with mock responses
- Data transformation functions
- Error handling scenarios

### Integration Tests
- Complete POI creation workflow
- Optimistic update behavior
- Error recovery flows

### Manual Testing
- Right-click → Create POI → Verify on map
- Form validation edge cases
- Network failure scenarios