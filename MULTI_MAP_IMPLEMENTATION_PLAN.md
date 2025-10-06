# Multi-Map Implementation Plan

## Executive Summary

This document outlines the requirements and implementation strategy for adding multi-map support to BreakoutGlobe. The system currently uses a single "default-map" for all users. The goal is to allow users with full accounts to create their own maps with unique URLs, where POIs are isolated per map, and later support map copying/blueprints.

## Current State Analysis

### Database Schema (Already Prepared!)
✅ **Good news**: The database schema is already designed for multi-map support!

**Existing Models:**
- `maps` table with `id`, `name`, `description`, `created_by`, `is_active`
- `pois` table with `map_id` foreign key
- `sessions` table with `map_id` foreign key
- Default map (`default-map`) created automatically on startup

**Key Observations:**
- All POIs are already scoped to a `map_id`
- All sessions are already scoped to a `map_id`
- Map ownership is tracked via `created_by` field
- Permission system exists: `CanBeModifiedBy()`, `CanBeAccessedBy()`

### Current Hardcoded Assumptions

**Backend:**
1. `backend/internal/database/migrations.go` - Creates "default-map" on startup
2. API endpoints assume `default-map` in some places
3. No map creation/listing endpoints exist yet

**Frontend:**
1. `frontend/src/App.tsx` - Hardcoded `mapId: 'default-map'` in multiple places:
   - Session creation
   - POI loading: `getPOIs('default-map')`
   - Map sessions fetch: `/api/maps/default-map/sessions`
2. `frontend/src/services/api.ts` - All POI operations use hardcoded map ID
3. No URL routing for different maps
4. No UI for map creation/selection

## Implementation Requirements

### Phase 1: Core Multi-Map Infrastructure

#### 1.1 Backend API Endpoints

**Map Management Endpoints:**
```
POST   /api/maps                    - Create new map (full accounts only)
GET    /api/maps                    - List all active maps
GET    /api/maps/:mapId             - Get map details
PUT    /api/maps/:mapId             - Update map (owner/admin only)
DELETE /api/maps/:mapId             - Delete map (owner/admin only)
GET    /api/maps/:mapId/sessions    - Get active sessions (already exists)
```

**Required Components:**
- `backend/internal/handlers/map_handler.go` - New handler for map operations
- `backend/internal/services/map_service.go` - Business logic for maps
- `backend/internal/repository/map_repository.go` - Database operations
- Update `backend/internal/server/server.go` to register map routes

**Business Rules:**
- Only full accounts (`AccountTypeFull`) can create maps
- Guest accounts can only join existing maps
- Map creators can modify/delete their maps
- Admins can modify/delete any map
- Default map cannot be deleted

#### 1.2 Frontend URL Routing

**URL Structure:**
```
/                           - Landing page / map selection
/map/:mapId                 - Specific map view
/map/:mapId/create          - Create POI on this map (optional)
```

**Required Changes:**
- Add React Router (if not already present)
- Update `frontend/src/App.tsx` to use URL-based map ID
- Extract map ID from URL params using `useParams()`
- Pass map ID down to all components that need it

**Components to Update:**
- `App.tsx` - Add routing, extract mapId from URL
- `MapContainer.tsx` - Accept mapId as prop
- `POICreationModal.tsx` - Use current mapId
- `POISidebar.tsx` - Filter POIs by current mapId
- All API calls - Replace hardcoded 'default-map' with dynamic mapId

#### 1.3 Map Selection UI

**New Components:**
```
frontend/src/components/MapSelector.tsx       - List/select maps
frontend/src/components/MapCreationModal.tsx  - Create new map
frontend/src/components/MapSettingsModal.tsx  - Edit map settings
```

**Features:**
- Display list of available maps
- Show map name, description, creator, active users
- "Create New Map" button (full accounts only)
- Map search/filter functionality
- Recent maps / favorites (future enhancement)

#### 1.4 WebSocket Updates

**Current State:**
- WebSocket connection uses `sessionId` in URL: `/ws?sessionId=xxx`
- Session already has `map_id` field
- PubSub events are not scoped to maps

**Required Changes:**
- Scope WebSocket broadcasts to map participants only
- Update `backend/internal/websocket/handler.go`:
  - Filter avatar updates by map
  - Filter POI events by map
  - Add map-scoped rooms/channels
- Update Redis PubSub channels to include map ID:
  - `poi:created:mapId`
  - `poi:updated:mapId`
  - `poi:deleted:mapId`
  - `avatar:moved:mapId`

### Phase 2: Map Sharing & Access Control

#### 2.1 Shareable URLs

**Implementation:**
- Maps are accessible via `/map/:mapId`
- Anyone with the URL can view/join (if map is active)
- Map ID is a UUID (already implemented)
- No additional authentication needed for public maps

**Optional Enhancements:**
- Private maps (require invitation/password)
- Map visibility settings (public/unlisted/private)
- Invitation links with tokens

#### 2.2 Map Discovery

**Features:**
- Public map directory (optional)
- Search maps by name/description
- Featured/popular maps
- User's created maps list
- Recently visited maps

### Phase 3: Map Copying/Blueprints

#### 3.1 Copy Map Functionality

**API Endpoint:**
```
POST /api/maps/:mapId/copy
Body: { name: string, description: string }
Response: { newMapId: string, ... }
```

**Implementation:**
1. Verify user has permission to copy (full account)
2. Create new map with new ID
3. Copy all POIs from source map to new map
4. Update POI `created_by` to copying user (or keep original?)
5. Do NOT copy sessions (maps start empty)
6. Do NOT copy POI images (or copy them?)

**Business Rules:**
- Only full accounts can copy maps
- Source map must be active
- POI images: decide whether to copy or reference
- Participant counts reset to 0
- Discussion timers reset

#### 3.2 Map Templates

**Future Enhancement:**
- Mark maps as "templates"
- Template gallery
- Pre-built map templates (city tours, conference venues, etc.)

## Data Migration Strategy

### No Migration Needed!
The database schema already supports multiple maps. All existing data will continue to work:
- Existing POIs are already associated with "default-map"
- Existing sessions are already associated with "default-map"
- No schema changes required

### Backward Compatibility
- Keep "default-map" as the default landing page
- Redirect `/` to `/map/default-map` initially
- Existing bookmarks/links continue to work

## Implementation Checklist

### Backend Tasks

**Map Service & Repository:**
- [ ] Create `backend/internal/repository/map_repository.go`
  - [ ] `Create(ctx, map) (*Map, error)`
  - [ ] `GetByID(ctx, id) (*Map, error)`
  - [ ] `GetAll(ctx) ([]*Map, error)`
  - [ ] `GetByCreator(ctx, userId) ([]*Map, error)`
  - [ ] `Update(ctx, map) error`
  - [ ] `Delete(ctx, id) error`
  - [ ] `CopyMap(ctx, sourceId, newMap) (*Map, error)`

- [ ] Create `backend/internal/services/map_service.go`
  - [ ] `CreateMap(ctx, userId, name, desc) (*Map, error)`
  - [ ] `GetMap(ctx, mapId) (*Map, error)`
  - [ ] `ListMaps(ctx) ([]*Map, error)`
  - [ ] `UpdateMap(ctx, mapId, userId, updates) (*Map, error)`
  - [ ] `DeleteMap(ctx, mapId, userId) error`
  - [ ] `CopyMap(ctx, sourceMapId, userId, newName) (*Map, error)`
  - [ ] Validate permissions (full account check)
  - [ ] Prevent deletion of default map

- [ ] Create `backend/internal/handlers/map_handler.go`
  - [ ] `CreateMap` handler
  - [ ] `GetMap` handler
  - [ ] `ListMaps` handler
  - [ ] `UpdateMap` handler
  - [ ] `DeleteMap` handler
  - [ ] `CopyMap` handler
  - [ ] Rate limiting integration
  - [ ] Error handling

- [ ] Update `backend/internal/server/server.go`
  - [ ] Add `setupMapRoutes()` method
  - [ ] Register map routes in `setupRoutes()`

**WebSocket Scoping:**
- [ ] Update `backend/internal/websocket/handler.go`
  - [ ] Add map-scoped broadcast methods
  - [ ] Filter avatar updates by map
  - [ ] Filter POI events by map
  - [ ] Add `GetMapParticipants()` method

- [ ] Update `backend/internal/redis/pubsub.go`
  - [ ] Add map ID to PubSub channel names
  - [ ] Update publish methods to include map scope
  - [ ] Update subscribe methods to filter by map

**Testing:**
- [ ] Unit tests for map repository
- [ ] Unit tests for map service
- [ ] Integration tests for map API endpoints
- [ ] Test map permissions (guest vs full account)
- [ ] Test map copying functionality
- [ ] Test WebSocket map scoping

### Frontend Tasks

**Routing Infrastructure:**
- [ ] Install/configure React Router (if not present)
- [ ] Create route structure in `App.tsx`
- [ ] Add `useParams()` to extract mapId from URL
- [ ] Update all hardcoded 'default-map' references

**Map Management UI:**
- [ ] Create `frontend/src/components/MapSelector.tsx`
  - [ ] List available maps
  - [ ] Show map details (name, description, active users)
  - [ ] "Create Map" button (conditional on account type)
  - [ ] "Join Map" action (navigate to map URL)

- [ ] Create `frontend/src/components/MapCreationModal.tsx`
  - [ ] Form: name, description
  - [ ] Validation
  - [ ] API integration
  - [ ] Redirect to new map on success

- [ ] Create `frontend/src/components/MapSettingsModal.tsx`
  - [ ] Edit map name/description
  - [ ] Delete map (with confirmation)
  - [ ] Copy map functionality
  - [ ] Permission checks

**Component Updates:**
- [ ] Update `App.tsx`
  - [ ] Remove hardcoded 'default-map'
  - [ ] Get mapId from URL params
  - [ ] Pass mapId to child components
  - [ ] Update session creation to use URL mapId
  - [ ] Update POI loading to use URL mapId

- [ ] Update `MapContainer.tsx`
  - [ ] Accept mapId as prop
  - [ ] Pass to child components

- [ ] Update `POICreationModal.tsx`
  - [ ] Accept mapId as prop
  - [ ] Use in API calls

- [ ] Update `POISidebar.tsx`
  - [ ] Filter POIs by current mapId
  - [ ] Show map name in header

- [ ] Update `frontend/src/services/api.ts`
  - [ ] Remove hardcoded mapId from all functions
  - [ ] Add mapId parameter to all POI functions
  - [ ] Add new map management functions

**State Management:**
- [ ] Create `frontend/src/stores/mapStore.ts`
  - [ ] Current map state
  - [ ] Available maps list
  - [ ] Map loading/error states
  - [ ] Map CRUD operations

**Testing:**
- [ ] Test map selection flow
- [ ] Test map creation (full account)
- [ ] Test map creation blocked (guest account)
- [ ] Test URL navigation between maps
- [ ] Test POI isolation per map
- [ ] Test session isolation per map
- [ ] Test map copying

### Documentation Tasks

- [ ] Update README.md with multi-map features
- [ ] API documentation for map endpoints
- [ ] User guide for creating/sharing maps
- [ ] Developer guide for map-scoped features

## Security Considerations

### Access Control
- ✅ Map creation restricted to full accounts
- ✅ Map modification restricted to owner/admin
- ✅ Map deletion restricted to owner/admin
- ✅ Default map cannot be deleted
- ⚠️ Consider: Should maps be public by default?

### Data Isolation
- ✅ POIs are already scoped to maps (database level)
- ✅ Sessions are already scoped to maps (database level)
- ⚠️ Ensure WebSocket events are properly scoped
- ⚠️ Ensure Redis cache keys include map ID

### URL Security
- ✅ Map IDs are UUIDs (not guessable)
- ⚠️ Consider: Rate limiting on map access
- ⚠️ Consider: Abuse prevention (spam maps)

## Performance Considerations

### Database Queries
- ✅ Indexes already exist on `map_id` columns
- ⚠️ Add index on `maps.created_by` for user's maps list
- ⚠️ Consider pagination for map listings

### WebSocket Scaling
- ⚠️ Current implementation broadcasts to all connections
- ⚠️ Need map-scoped rooms to reduce unnecessary traffic
- ⚠️ Consider Redis pub/sub channels per map

### Caching
- ⚠️ Cache map metadata (name, description, creator)
- ⚠️ Cache active user counts per map
- ⚠️ Invalidate cache on map updates

## Future Enhancements

### Phase 4: Advanced Features
- [ ] Map analytics (views, active users, POI count)
- [ ] Map ratings/reviews
- [ ] Map categories/tags
- [ ] Map search with filters
- [ ] Favorite maps
- [ ] Map activity feed
- [ ] Map collaboration (multiple owners)
- [ ] Map permissions (view/edit/admin roles)

### Phase 5: Premium Features
- [ ] Private maps (password protected)
- [ ] Map size limits (POI count, user count)
- [ ] Custom map styling
- [ ] Map export/import
- [ ] Map versioning
- [ ] Map backups

## Estimated Effort

### Phase 1: Core Multi-Map (MVP)
- Backend: 2-3 days
  - Map service/repository: 1 day
  - API endpoints: 0.5 day
  - WebSocket scoping: 1 day
  - Testing: 0.5 day

- Frontend: 2-3 days
  - Routing setup: 0.5 day
  - Map selector UI: 1 day
  - Component updates: 1 day
  - Testing: 0.5 day

**Total: 4-6 days**

### Phase 2: Sharing & Discovery
- Backend: 1 day
- Frontend: 1-2 days
**Total: 2-3 days**

### Phase 3: Map Copying
- Backend: 1 day
- Frontend: 0.5 day
**Total: 1.5 days**

**Grand Total: 7.5-10.5 days**

## Success Criteria

### MVP (Phase 1)
- [ ] Users can create new maps (full accounts only)
- [ ] Each map has a unique URL
- [ ] POIs are isolated per map
- [ ] Sessions are isolated per map
- [ ] WebSocket events are scoped to maps
- [ ] Users can switch between maps
- [ ] Default map continues to work

### Phase 2
- [ ] Maps can be shared via URL
- [ ] Users can discover public maps
- [ ] Map metadata is displayed correctly

### Phase 3
- [ ] Maps can be copied to create new maps
- [ ] Copied maps are independent
- [ ] POIs are duplicated correctly

## Risk Assessment

### Low Risk
- ✅ Database schema already supports multi-map
- ✅ No data migration required
- ✅ Backward compatible with existing data

### Medium Risk
- ⚠️ WebSocket scoping complexity
- ⚠️ Frontend routing changes (potential bugs)
- ⚠️ Testing coverage for all map scenarios

### High Risk
- ❌ None identified

## Conclusion

The BreakoutGlobe codebase is **already well-prepared** for multi-map support! The database schema, models, and core architecture are designed with multi-map in mind. The main work involves:

1. **Removing hardcoded assumptions** about "default-map"
2. **Adding map management APIs** (CRUD operations)
3. **Implementing URL-based routing** on the frontend
4. **Scoping WebSocket events** to specific maps
5. **Building map selection/creation UI**

The implementation is straightforward and low-risk, with an estimated effort of 7.5-10.5 days for all three phases.
