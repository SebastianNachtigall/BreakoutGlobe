# POI Membership Persistence Fix

## Problem
When a user joined a POI and then refreshed the browser, the POI detail modal would disappear but the POI would still show the user as joined. When the user clicked on the map to move their avatar, they would remain invisible to other users because the system couldn't properly leave the POI.

## Root Cause
The issue was in the POI store persistence configuration. The `currentUserPOI` field was not being persisted across browser refreshes, while the POI data (including participant counts) was being persisted. This created an inconsistent state where:

1. The POI still showed the user as a participant (participant count > 0)
2. But `currentUserPOI` was null after refresh
3. When the user clicked the map, `wsClient.leaveCurrentPOI()` would check `getCurrentUserPOI()` and find null
4. No leave message would be sent to the server
5. The user remained "stuck" in the POI, invisible to others

## Solution
### 1. Fixed POI Store Persistence
Updated the `partialize` function in `poiStore.ts` to include `currentUserPOI`:

```typescript
partialize: (state) => ({
  pois: state.pois,
  currentUserPOI: state.currentUserPOI, // Now persisted!
}),
```

### 2. Improved Edge Case Handling
Enhanced the `leavePOI` method to handle cases where the POI doesn't exist but the user still thinks they're in it:

```typescript
leavePOI: (poiId: string, userId: string) => {
  const state = get();
  const poi = state.pois.find(p => p.id === poiId);
  
  // Always clear currentUserPOI if user is trying to leave this POI
  const shouldClearCurrentPOI = state.currentUserPOI === poiId;
  
  if (!poi || poi.participantCount <= 0) {
    // POI doesn't exist or has no participants, but still clear currentUserPOI if needed
    if (shouldClearCurrentPOI) {
      set((state) => ({
        ...state,
        currentUserPOI: null,
      }));
    }
    return false;
  }
  
  // ... rest of the method
}
```

## Testing
Created comprehensive tests to verify the fix:

1. **poi-membership-persistence.test.tsx** - Reproduces the original issue
2. **poi-membership-persistence-fix.test.tsx** - Verifies the fix works
3. **poi-membership-integration.test.tsx** - Tests the complete flow

All tests pass and the fix has been verified in the browser.

## Files Changed
- `frontend/src/stores/poiStore.ts` - Fixed persistence and edge case handling
- `frontend/src/__tests__/poi-membership-persistence.test.tsx` - Test reproducing the issue
- `frontend/src/__tests__/poi-membership-persistence-fix.test.tsx` - Test verifying the fix
- `frontend/src/__tests__/poi-membership-integration.test.tsx` - Integration tests

## Impact
- Users can now properly leave POIs after browser refresh
- No more invisible users stuck in POIs
- Consistent state management across browser sessions
- Better error handling for edge cases