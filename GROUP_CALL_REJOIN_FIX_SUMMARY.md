# Group Call Rejoin Fix Summary

## Problem

When a user rejoined a POI with an active group call, they could not receive video streams from existing participants.

### Root Cause

The WebRTC offer coordination uses lexicographic ID comparison to decide who initiates peer connections:

```typescript
if (currentUserId && currentUserId < userId) {
  // Create offer and send to peer
} else {
  // Wait for offer from peer
}
```

**The Issue:**
- When User 3 rejoins, they create peer connections and wait for offers from users with "smaller" IDs
- But Users 1 and 2 don't know User 3 rejoined, so they never create offers
- Result: User 3 sees no video streams

### Example Scenario

1. **Users 1, 2, 3 in group call** - All working ✅
2. **User 3 leaves** - Users 1 and 2 continue call ✅
3. **User 3 rejoins** - User 3 creates peer connections but waits for offers ❌
4. **Users 1 and 2** - Never notified, never send offers ❌
5. **Result** - User 3 has no video streams ❌

## Solution

Added logic to notify existing group call participants when a new user joins the POI.

### Implementation

**File:** `frontend/src/services/websocket-client.ts`

**Change:** In `handlePOIJoined()` method, added:

```typescript
// If this user is already in an active group call for this POI, add the new participant
const videoCallState = videoCallStore.getState();
if (videoCallState.isGroupCallActive && videoCallState.currentPOI === poiId && videoCallState.groupWebRTCService) {
  console.log('🔗 Adding new participant to existing group call:', userId);
  
  // Find participant info from the participants list
  const participantInfo = participants.find((p: any) => p.id === userId);
  if (participantInfo) {
    // Add participant to the group call
    videoCallStore.getState().addGroupCallParticipant(userId, {
      userId: userId,
      displayName: participantInfo.name || 'Unknown User',
      avatarURL: participantInfo.avatarUrl || undefined
    });
    
    // Add peer connection for the new participant
    videoCallStore.getState().addPeerToGroupCall(userId).catch((error) => {
      console.error('❌ Failed to add peer for new participant:', userId, error);
    });
  }
}
```

### How It Works

1. **Backend broadcasts** `poi_joined` event when any user joins a POI
2. **All clients receive** the event via WebSocket
3. **Existing participants check** if they're in an active group call for that POI
4. **If yes**, they:
   - Add the new user to their participants list
   - Create a peer connection for the new user
   - Since their ID is "smaller", they initiate the offer
5. **New joiner receives** the offers and establishes connections
6. **Result**: Video streams flow! ✅

### Flow Diagram

```
User 3 Rejoins POI
       ↓
Backend broadcasts "poi_joined" event
       ↓
   ┌───┴───┐
   ↓       ↓
User 1   User 2
   ↓       ↓
Check: In group call for this POI?
   ↓       ↓
  YES     YES
   ↓       ↓
Add User 3 as participant
   ↓       ↓
Create peer connection
   ↓       ↓
ID comparison: user1 < user3? YES
               user2 < user3? YES
   ↓       ↓
Create & send offers to User 3
       ↓
User 3 receives offers
       ↓
User 3 creates answers
       ↓
✅ Video streams established!
```

## Testing

### Expected Behavior

**Scenario: 3-User Rejoin**
1. Users 1, 2, 3 join POI → All see each other's video ✅
2. User 3 leaves → Users 1, 2 continue call ✅
3. User 3 rejoins → All three see each other's video ✅

### Logs to Verify

**User 3 (Rejoining):**
```
🏢 Joining POI group call: [poi-id]
🔗 Initializing group WebRTC service
👥 Adding existing participants to group call: 2
📞 WebRTC: Waiting for offer from user (user-1)
📞 WebRTC: Waiting for offer from user (user-2)
📥 WebRTC: Received offer from user-1  // ✅ NEW!
📥 WebRTC: Received offer from user-2  // ✅ NEW!
📺 Remote stream received from user: user-1
📺 Remote stream received from user: user-2
```

**Users 1 & 2 (Existing):**
```
👥 WebSocket: User joined POI {userId: 'user-3', ...}
🔗 Adding new participant to existing group call: user-3  // ✅ NEW!
👥 Adding group call participant: {userId: 'user-3', ...}
👥 Adding peer to group call: user-3
📝 WebRTC: Current user should initiate call to user (user-3)  // ✅ NEW!
📤 WebRTC: Offer created for user: user-3, sending via WebSocket
```

## Files Changed

- `frontend/src/services/websocket-client.ts` - Added peer connection logic in `handlePOIJoined()`

## Related Issues

- Group call initialization race condition (previously fixed)
- Participant limit bug (previously fixed)
- This completes the group call rejoin functionality

## Status

✅ **FIXED** - Existing participants now automatically add peer connections for rejoining users
