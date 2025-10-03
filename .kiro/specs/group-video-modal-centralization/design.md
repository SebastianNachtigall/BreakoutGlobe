# Design Document

## Overview

This design eliminates the race condition in POI group video modal triggering by centralizing all group call decision logic into a single method in the video call store. The approach is surgical and minimal - we only change what's necessary to fix the core issue.

## Architecture

### Current Problem Architecture
```
App.tsx handleJoinPOI() ──┐
                          ├──> Race Condition ──> Inconsistent Modal Display
WebSocket handlePOIJoined() ──┘
```

### Fixed Architecture
```
App.tsx handleJoinPOI() ──┐
                          ├──> videoCallStore.checkAndStartGroupCall() ──> Reliable Modal Display
WebSocket handlePOIJoined() ──┘
```

## Components and Interfaces

### Core Method: checkAndStartGroupCall()

**Location**: `frontend/src/stores/videoCallStore.ts`

**Signature**:
```typescript
checkAndStartGroupCall: (poiId: string, participantCount: number, triggerUserId: string) => void
```

**Responsibilities**:
1. Check if group call should be started
2. Prevent duplicate initializations
3. Set modal display state atomically
4. Initialize WebRTC if needed

**Logic Flow**:
```typescript
checkAndStartGroupCall(poiId, participantCount, triggerUserId) {
  // 1. Check if already initializing (race condition prevention)
  if (this._initializingGroupCall) return;
  
  // 2. Check if group call already active for this POI
  if (this.isGroupCallActive && this.currentPOI === poiId) return;
  
  // 3. Check if current user is in this POI
  const currentUserPOI = poiStore.getState().getCurrentUserPOI();
  if (currentUserPOI !== poiId) return;
  
  // 4. Check if multiple participants
  if (participantCount <= 1) return;
  
  // 5. Start group call
  this._initializingGroupCall = true;
  this.joinPOICall(poiId);
  // ... WebRTC initialization
  this._initializingGroupCall = false;
}
```

### State Changes

**New State Variables**:
```typescript
interface VideoCallState {
  // ... existing state
  _initializingGroupCall: boolean; // Private lock flag
}
```

**State Transitions**:
- `_initializingGroupCall`: `false` → `true` → `false`
- `isGroupCallActive`: `false` → `true` (atomic with currentPOI)
- `currentPOI`: `null` → `poiId` (atomic with isGroupCallActive)

## Data Models

No changes to data models. We only modify the control flow logic.

## Error Handling

### Initialization Lock Timeout
If initialization takes too long, implement a timeout:
```typescript
setTimeout(() => {
  if (this._initializingGroupCall) {
    console.warn('Group call initialization timeout, releasing lock');
    this._initializingGroupCall = false;
  }
}, 10000); // 10 second timeout
```

### WebRTC Initialization Failure
```typescript
try {
  await this.initializeGroupWebRTC();
} catch (error) {
  console.error('Group WebRTC initialization failed:', error);
  this._initializingGroupCall = false;
  this.leavePOICall(); // Clean up state
}
```

## Testing Strategy

### Unit Tests
1. Test `checkAndStartGroupCall()` with various parameter combinations
2. Test race condition prevention with concurrent calls
3. Test state cleanup on initialization failure

### Integration Tests
1. Test complete POI join → group call flow
2. Test WebSocket event → group call flow
3. Test simultaneous triggers from both paths

### Manual Testing Scenarios
1. Two users join POI within 100ms
2. User joins POI while API is slow (5+ seconds)
3. WebSocket events arrive before API responses
4. Network failures during group call initialization

## Implementation Plan

### Phase 1: Add Centralized Method
1. Add `checkAndStartGroupCall()` to videoCallStore
2. Add `_initializingGroupCall` state variable
3. Implement core logic and locking

### Phase 2: Update Call Sites
1. Replace App.tsx group call logic with centralized call
2. Replace WebSocket handler group call logic with centralized call
3. Remove dynamic imports from WebSocket handlers

### Phase 3: Clean Up Old Code
1. Remove all old group call triggering code
2. Remove unused imports and variables
3. Update any remaining references

### Phase 4: State Synchronization Fix
1. Ensure `currentUserPOI` is set immediately in optimistic updates
2. Test that centralized method uses correct state

## Rollback Plan

If issues arise, the changes can be rolled back by:
1. Reverting the centralized method
2. Restoring the original dual-path logic
3. Re-adding dynamic imports to WebSocket handlers

The changes are isolated to specific methods and don't affect the overall architecture, making rollback straightforward.