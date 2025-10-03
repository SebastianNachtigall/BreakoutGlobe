# "Unknown User" Display Name Investigation

## Problem

Users who join a POI later are sometimes displayed as "Unknown User" in the video stream instead of their actual display name.

## Root Cause Analysis

### Data Flow

1. **Backend (POI Service)** - `JoinPOI` method:
   - Calls `GetPOIParticipantsWithInfo()` to get participant details
   - This method fetches user info from the user service
   - Publishes `POIJoinedEventWithParticipants` with participant array

2. **Frontend (WebSocket Client)** - `handlePOIJoined`:
   ```typescript
   const participantInfo = participants.find((p: any) => p.id === userId);
   
   eventBus.emit(GroupCallEvents.USER_JOINED_POI, {
     poiId,
     userId,
     displayName: participantInfo?.name || 'Unknown User',  // ⚠️ FALLBACK
     avatarURL: participantInfo?.avatarUrl,
     participants
   });
   ```

3. **Frontend (VideoCallStore)** - Event handler:
   ```typescript
   videoCallStore.getState().addGroupCallParticipant(data.userId, {
     userId: data.userId,
     displayName: data.displayName,  // Uses the fallback if not found
     avatarURL: data.avatarURL
   });
   ```

## Why "Unknown User" Appears

### Scenario 1: Empty Participants Array (Race Condition)
**When it happens:** User 3 joins while User 1 and User 2 are already in the POI

**The issue:**
- Backend publishes `poi_joined` event with `participants` array
- Frontend receives the event
- The `participants` array might be **empty** or **not yet populated** due to timing
- `participantInfo` lookup fails: `participants.find((p: any) => p.id === userId)` returns `undefined`
- Falls back to `'Unknown User'`

**Evidence from logs:**
```
eventBus.ts:40 📢 Event: group_call:user_joined_poi {
  poiId: '6d967b63-bc78-480a-b638-05afea95969f',
  userId: '79702bbe-0ca5-4e4b-a2cc-211b3d783118',
  displayName: 'Unknown User',  // ⚠️ Fallback used
  avatarURL: undefined,
  participants: Array(0)  // ⚠️ Empty array!
}
```

### Scenario 2: Participant Not in Array
**When it happens:** The joining user's info hasn't been added to the participants array yet

**The issue:**
- Backend calls `GetPOIParticipantsWithInfo()` **before** the user is fully added
- Or there's a timing issue where the user is added but the info isn't fetched yet
- The `participants` array doesn't include the joining user
- Lookup fails, falls back to `'Unknown User'`

### Scenario 3: User Service Failure
**When it happens:** Backend can't fetch user details from the user service

**The issue in backend:**
```go
// In GetPOIParticipantsWithInfo
participantInfo := POIParticipantInfo{
    ID:        userID,
    Name:      userID, // ⚠️ Fallback to userID
    AvatarURL: "",
}

if s.userService != nil {
    user, err := s.userService.GetUser(ctx, userID)
    if err == nil && user != nil {
        participantInfo.Name = user.DisplayName
        // ...
    }
    // If user service fails, we continue with fallback values
}
```

If `GetUser()` fails, the participant name is set to the `userID` (UUID), not "Unknown User", but this could still look wrong.

## Solutions

### Option 1: Fetch User Info from Avatar Store (Frontend)
When the `participants` array is empty or doesn't contain the user, fall back to the avatar store which already has user information.

**Pros:**
- Quick fix on frontend
- Avatar store is already populated with user data
- No backend changes needed

**Cons:**
- Relies on avatar store being up-to-date
- Doesn't fix the root cause

### Option 2: Include Joining User in Participants Array (Backend)
Modify the backend to ensure the joining user's info is included in the `participants` array of the `poi_joined` event.

**Pros:**
- Fixes root cause
- More reliable data flow
- Consistent with expectations

**Cons:**
- Requires backend changes
- Need to ensure user info is fetched before publishing event

### Option 3: Separate User Info in Event (Backend)
Add a separate `joiningUser` field to the event with the user's display name and avatar.

**Pros:**
- Clear separation of concerns
- Joining user info is always available
- Doesn't rely on participants array

**Cons:**
- Changes event structure
- Requires both backend and frontend changes

### Option 4: Retry with Avatar Store Lookup (Frontend)
When `participantInfo` is not found, look up the user in the avatar store before falling back to "Unknown User".

**Pros:**
- Simple frontend fix
- Uses existing data
- Graceful degradation

**Cons:**
- Still a workaround
- Doesn't fix backend timing issue

## Recommended Solution

**Hybrid Approach:**

1. **Frontend Fix (Immediate):** Add avatar store lookup before "Unknown User" fallback
2. **Backend Fix (Long-term):** Ensure joining user info is included in participants array or add separate `joiningUser` field

### Implementation Plan

#### Phase 1: Frontend Quick Fix
```typescript
// In websocket-client.ts handlePOIJoined
const participantInfo = participants.find((p: any) => p.id === userId);

// If not found in participants, try avatar store
let displayName = participantInfo?.name;
let avatarURL = participantInfo?.avatarUrl;

if (!displayName) {
  const avatar = avatarStore.getState().getAvatarByUserId(userId);
  if (avatar) {
    displayName = avatar.displayName;
    avatarURL = avatar.avatarURL;
  }
}

eventBus.emit(GroupCallEvents.USER_JOINED_POI, {
  poiId,
  userId,
  displayName: displayName || 'Unknown User',
  avatarURL: avatarURL,
  participants
});
```

#### Phase 2: Backend Enhancement
Add `joiningUserInfo` to the event:
```go
type POIJoinedEventWithParticipants struct {
    POIID        string
    MapID        string
    UserID       string
    SessionID    string
    CurrentCount int
    Participants []POIParticipant
    JoiningUser  POIParticipant  // ✅ Add this
    Timestamp    time.Time
}
```

## Testing Checklist

- [ ] Test with 2 users joining simultaneously
- [ ] Test with 3rd user joining after 2 are already in call
- [ ] Test with rapid join/leave scenarios
- [ ] Test with user service temporarily unavailable
- [ ] Verify display names appear correctly in all scenarios
- [ ] Check avatar URLs are also populated correctly

## Related Files

- `frontend/src/services/websocket-client.ts` - Event emission
- `frontend/src/stores/videoCallStore.ts` - Event handling
- `backend/internal/services/poi_service.go` - Participant info fetching
- `backend/internal/redis/pubsub.go` - Event publishing

---

**Status:** Investigation Complete  
**Priority:** Medium (UX issue, not blocking functionality)  
**Estimated Fix Time:** 30 minutes (frontend), 1-2 hours (backend)
