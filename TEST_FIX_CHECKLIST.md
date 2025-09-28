# Test Fix Checklist - Actionable Tasks

## Phase 1: Critical Interface Fixes

### ✅ Task 1.1: Add Missing Method to MockPOIService (handlers)
**File**: `backend/internal/handlers/poi_handler_test.go`
**Issue**: Missing `GetPOIParticipantsWithInfo` method
**Action**: Add method to MockPOIService struct

### ✅ Task 1.2: Add Missing Method to MockPOIService (testdata)  
**File**: `backend/internal/testdata/mocks.go`
**Issue**: Missing `GetPOIParticipantsWithInfo` method
**Action**: Add method to MockPOIService struct

### ✅ Task 1.3: Add Missing Method to MockPOIService (websocket)
**File**: `backend/internal/websocket/handler_test.go` 
**Issue**: Missing `GetPOIParticipantsWithInfo` method
**Action**: Add method to MockPOIService struct

### ✅ Task 1.4: Fix POI Handler Constructor (handlers - migrated test)
**File**: `backend/internal/handlers/poi_handler_migrated_test.go`
**Issue**: `NewPOIHandler` needs POIUserServiceInterface parameter
**Action**: Add mock user service parameter

### ✅ Task 1.5: Fix POI Handler Constructor (handlers - main test)
**File**: `backend/internal/handlers/poi_handler_test.go`
**Issue**: `NewPOIHandler` needs POIUserServiceInterface parameter  
**Action**: Add mock user service parameter

### ✅ Task 1.6: Fix POI Handler Constructor (handlers - image test)
**File**: `backend/internal/handlers/poi_image_test.go`
**Issue**: `NewPOIHandler` needs POIUserServiceInterface parameter
**Action**: Add mock user service parameter

### ✅ Task 1.7: Fix POI Handler Constructor (testdata scenarios)
**File**: `backend/internal/testdata/scenarios.go`
**Issue**: `NewPOIHandler` needs POIUserServiceInterface parameter
**Action**: Add mock user service parameter

### ✅ Task 1.8: Fix WebSocket Handler Constructor (testdata scenarios)
**File**: `backend/internal/testdata/scenarios.go`
**Issue**: `websocket.NewHandler` needs POIServiceInterface parameter
**Action**: Add mock POI service parameter

### ✅ Task 1.9: Fix WebSocket Handler Constructor (testdata testws)
**File**: `backend/internal/testdata/testws.go`
**Issue**: `websocket.NewHandler` needs POIServiceInterface parameter
**Action**: Add mock POI service parameter

### ✅ Task 1.10: Fix POI Service Constructor (services - timer test)
**File**: `backend/internal/services/poi_discussion_timer_test.go`
**Issue**: `NewPOIService` needs UserServiceInterface parameter
**Action**: Add mock user service parameter

### ✅ Task 1.11: Fix POI Service Constructor (services - image test)
**File**: `backend/internal/services/poi_image_service_test.go`
**Issue**: `NewPOIServiceWithImageUploader` needs UserServiceInterface parameter
**Action**: Add mock user service parameter

## Phase 2: Runtime Error Fixes

### ✅ Task 2.1: Fix WebSocket Mock Expectations
**File**: `backend/internal/websocket/handler_test.go`
**Issue**: Mock expectations don't match actual calls
**Actions**:
- Fix context parameter expectations (context.Context vs string)
- Fix action type expectations (join_poi vs update_avatar)
- Update mock setup to match actual usage

## Phase 3: Test Cleanup (Optional)

### ✅ Task 3.1: Review Test Relevance
**Action**: Identify tests that may no longer be needed after mock removal

### ✅ Task 3.2: Consolidate Duplicate Coverage
**Action**: Remove redundant tests if any exist

## Quick Commands to Check Progress

```bash
# Check compilation status
go test ./internal/handlers -v --dry-run
go test ./internal/services -v --dry-run  
go test ./internal/websocket -v --dry-run
go test ./internal/testdata -v --dry-run

# Run specific failing packages
go test ./internal/handlers -v
go test ./internal/services -v
go test ./internal/websocket -v

# Run all tests
go test ./... -v
```

## Implementation Order

1. **Start with MockPOIService fixes** (Tasks 1.1-1.3) - These are blocking multiple packages
2. **Fix constructor calls** (Tasks 1.4-1.11) - These depend on mock fixes
3. **Fix runtime expectations** (Task 2.1) - These need working mocks
4. **Clean up** (Tasks 3.1-3.2) - Optional improvements

## Success Criteria for Each Task

- ✅ **Task Complete**: No compilation errors for that specific file/package
- ✅ **Phase Complete**: All tests in phase compile and run
- ✅ **All Complete**: `go test ./...` passes completely

## Ready to Start?

The first task should be **Task 1.1**: Add the missing `GetPOIParticipantsWithInfo` method to the MockPOIService in `backend/internal/handlers/poi_handler_test.go`.