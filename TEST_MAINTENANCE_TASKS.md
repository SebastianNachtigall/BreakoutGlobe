# Test Maintenance Task List

## Overview
After removing mock implementations from server.go, several tests are failing due to interface mismatches and missing methods. This document outlines the systematic approach to fix all failing tests.

## Root Cause Analysis

### 1. Interface Evolution Issues
- **POI Service Interface**: Added `GetPOIParticipantsWithInfo` method but mock implementations don't have it
- **Handler Constructors**: POI and WebSocket handlers now require additional parameters
- **Service Constructors**: POI service now requires `UserServiceInterface` parameter

### 2. Mock Service Outdated
- Multiple `MockPOIService` implementations across different packages are missing methods
- Mock constructors have wrong parameter counts
- Test scenarios using outdated constructor signatures

## Task List (Priority Order)

### Phase 1: Fix Interface Mismatches (Critical)

#### Task 1.1: Update MockPOIService Implementations
**Files to Fix:**
- `backend/internal/handlers/poi_handler_test.go` - MockPOIService
- `backend/internal/testdata/mocks.go` - MockPOIService  
- `backend/internal/websocket/handler_test.go` - MockPOIService

**Required Changes:**
- Add missing `GetPOIParticipantsWithInfo(ctx context.Context, poiID string) ([]services.POIParticipantInfo, error)` method
- Ensure all mock methods match the current `POIServiceInterface`

#### Task 1.2: Fix POI Handler Constructor Calls
**Files to Fix:**
- `backend/internal/handlers/poi_handler_migrated_test.go`
- `backend/internal/handlers/poi_handler_test.go`
- `backend/internal/handlers/poi_image_test.go`
- `backend/internal/testdata/scenarios.go`

**Required Changes:**
- Update `NewPOIHandler` calls to include `POIUserServiceInterface` parameter
- Add mock user service where needed

#### Task 1.3: Fix WebSocket Handler Constructor Calls
**Files to Fix:**
- `backend/internal/testdata/scenarios.go`
- `backend/internal/testdata/testws.go`

**Required Changes:**
- Update `websocket.NewHandler` calls to include `POIServiceInterface` parameter
- Add mock POI service where needed

#### Task 1.4: Fix POI Service Constructor Calls
**Files to Fix:**
- `backend/internal/services/poi_discussion_timer_test.go`
- `backend/internal/services/poi_image_service_test.go`

**Required Changes:**
- Update `NewPOIService` and `NewPOIServiceWithImageUploader` calls to include `UserServiceInterface` parameter
- Add mock user service where needed

### Phase 2: Fix Mock Expectations (Medium Priority)

#### Task 2.1: Fix WebSocket Handler Test Mock Expectations
**Files to Fix:**
- `backend/internal/websocket/handler_test.go`

**Issue:**
```
mock: Unexpected Method Call
CheckRateLimit(context.backgroundCtx,string,services.ActionType)
Expected: CheckRateLimit(string,string,services.ActionType)
Action mismatch: join_poi != update_avatar
```

**Required Changes:**
- Update mock expectations to match actual method calls
- Fix context parameter expectations
- Fix action type expectations

### Phase 3: Remove Obsolete Tests (Low Priority)

#### Task 3.1: Evaluate Test Relevance
**Candidates for Removal:**
- Tests that were specifically testing mock implementations that no longer exist
- Duplicate test coverage that's now handled by proper integration tests
- Tests that are no longer relevant after architectural changes

#### Task 3.2: Consolidate Test Coverage
- Ensure proper test coverage exists for all functionality
- Remove redundant tests
- Update test documentation

## Implementation Strategy

### Step-by-Step Approach
1. **Fix compilation errors first** (Phase 1) - Get all tests compiling
2. **Fix runtime errors** (Phase 2) - Get all tests passing
3. **Clean up obsolete tests** (Phase 3) - Remove unnecessary tests

### Testing Strategy
- Fix one package at a time
- Run tests after each fix to ensure no regressions
- Use TDD approach: Red -> Green -> Refactor

### Validation Criteria
- All tests compile successfully
- All tests pass
- No mock-related compilation errors
- Proper test coverage maintained

## Estimated Effort

### Phase 1: 2-3 hours
- Interface fixes are straightforward but require careful attention to method signatures
- Multiple files need updates but changes are similar

### Phase 2: 1-2 hours  
- Mock expectation fixes require understanding test scenarios
- May need to debug specific test failures

### Phase 3: 1 hour
- Test cleanup and consolidation
- Documentation updates

**Total Estimated Time: 4-6 hours**

## Success Metrics
- ✅ All backend tests compile without errors
- ✅ All backend tests pass
- ✅ No interface mismatch errors
- ✅ No missing method errors
- ✅ No constructor parameter errors
- ✅ Proper test coverage maintained
- ✅ Clean, maintainable test code

## Next Steps
1. Start with Phase 1, Task 1.1 (Update MockPOIService implementations)
2. Work through tasks systematically
3. Test after each major change
4. Document any architectural decisions made during fixes