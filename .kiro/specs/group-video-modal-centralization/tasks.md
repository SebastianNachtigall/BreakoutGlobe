# Implementation Plan

## Overview
This plan implements a surgical fix for the POI group video modal race condition by centralizing all group call decision logic. The approach is minimal and focused - we only change what's necessary to eliminate the race condition.

- [x] 1. Add centralized group call decision method
  - Add `checkAndStartGroupCall()` method to videoCallStore
  - Add `_initializingGroupCall` private state variable for race condition prevention
  - Implement core decision logic: check POI membership, participant count, and existing call state
  - Add initialization locking to prevent duplicate WebRTC service creation
  - _Requirements: 1.1, 1.2, 4.1, 4.2_

- [x] 2. Fix immediate state synchronization in POI store
  - Modify `joinPOIOptimisticWithAutoLeave()` to set `currentUserPOI` immediately
  - Ensure centralized method can rely on synchronous state checks
  - Remove dependency on API response timing for group call decisions
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 3. Replace App.tsx group call logic with centralized call
  - Remove group call triggering logic from `handleJoinPOI()` method
  - Replace with single call to `videoCallStore.checkAndStartGroupCall()`
  - Pass POI ID, participant count, and current user ID to centralized method
  - Remove all WebRTC initialization code from App.tsx
  - _Requirements: 2.1, 2.2, 6.1_

- [x] 4. Replace WebSocket handler group call logic with centralized call
  - Remove dynamic imports from `websocket-client.ts`
  - Add direct import of videoCallStore at top of file
  - Remove group call triggering logic from `handlePOIJoined()` method
  - Replace with single call to `videoCallStore.checkAndStartGroupCall()`
  - _Requirements: 2.3, 2.4, 5.1, 5.2, 6.2_

- [x] 5. Clean up all artifacts of old dual-path system
  - Remove unused imports related to old group call logic
  - Remove commented-out code from previous implementations
  - Remove any remaining references to old group call triggering methods
  - Verify no group call logic remains outside the centralized method
  - _Requirements: 6.3, 6.4_

- [x] 6. Add error handling and timeout protection
  - Add timeout mechanism for initialization lock (10 second timeout)
  - Add proper error handling for WebRTC initialization failures
  - Ensure state is cleaned up properly on initialization errors
  - Add logging for debugging race condition issues
  - _Requirements: 4.3, 4.4_

## Success Criteria

After implementation:
1. Only one method (`checkAndStartGroupCall`) can trigger group calls
2. No race conditions occur when multiple events trigger simultaneously
3. Group video modal shows reliably for both users joining a POI
4. No duplicate WebRTC services are created
5. All old group call triggering code paths are removed
6. WebSocket handlers use direct imports, no dynamic imports for core functionality

## Testing Approach

**Note**: Existing tests will likely fail during refactoring. We will fix tests after completing the implementation.

**Manual Testing Priority**:
1. Two users join same POI within 100ms - both should see modal
2. User joins POI while API is slow - modal should show immediately
3. WebSocket events arrive before API responses - modal should still show
4. Page refresh during group call - state should recover correctly

**Post-Implementation Test Fixes**:
- Update unit tests for videoCallStore to test centralized method
- Update integration tests to expect single code path
- Fix any tests that relied on old dual-path behavior
## Post-Im
plementation Issues Found

- [x] 7. Fix WebRTC signaling race conditions and duplicate participants
  - Fix WebRTC state machine violation: "Called in wrong state: stable"
  - Prevent duplicate participant creation in UI (showing 2 large + 1 small video areas)
  - Implement proper offer/answer coordination to prevent both peers creating offers
  - Add participant deduplication logic in addGroupCallParticipant
  - Fix peer connection state management to prevent signaling conflicts
  - _Requirements: 1.1, 1.2, 4.1_