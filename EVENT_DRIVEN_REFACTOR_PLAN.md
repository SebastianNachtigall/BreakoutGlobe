# Event-Driven Architecture Refactor Plan

## Goal
Replace imperative/polling approach with event-driven architecture to eliminate race conditions in group video calls.

## Implementation Steps

### Step 1: Create Event Bus Infrastructure
**File:** `frontend/src/utils/eventBus.ts`

- Create simple EventBus class with `on()`, `off()`, `emit()`
- Define event constants for type safety
- Add error handling for listeners
- Add logging for debugging

**Events to Define:**
- `POI_USER_JOINED` - When user joins a POI
- `POI_USER_LEFT` - When user leaves a POI
- `WEBRTC_INITIALIZED` - When WebRTC service is ready
- `GROUP_CALL_READY` - When group call is fully initialized
- `PEER_ADDED` - When peer connection is added

### Step 2: Update Video Call Store
**File:** `frontend/src/stores/videoCallStore.ts`

**Changes:**
1. Add event listeners in store initialization
2. Add `pendingPeers` queue for users who join during initialization
3. Emit `WEBRTC_INITIALIZED` event when WebRTC is ready
4. Listen for `POI_USER_JOINED` and handle based on state:
   - If WebRTC ready → add peer immediately
   - If WebRTC initializing → queue in `pendingPeers`
5. When `WEBRTC_INITIALIZED` fires → process all `pendingPeers`
6. Remove polling/interval logic

**Key Methods:**
- `setupEventListeners()` - Subscribe to events
- `handlePOIUserJoined(data)` - React to user joins
- `handleWebRTCInitialized()` - Process pending peers
- `cleanupEventListeners()` - Unsubscribe on cleanup

### Step 3: Update WebSocket Client
**File:** `frontend/src/services/websocket-client.ts`

**Changes:**
1. Import eventBus
2. In `handlePOIJoined()`:
   - Remove direct videoCallStore calls
   - Just emit `POI_USER_JOINED` event
3. In `handlePOILeft()`:
   - Emit `POI_USER_LEFT` event
4. Remove all polling/interval logic
5. Keep POI store updates (still needed)

**Simplified Flow:**
```typescript
handlePOIJoined(data) {
  // Update POI participants (still needed)
  poiStore.updateParticipants(...)
  
  // Emit event (let listeners handle it)
  eventBus.emit(Events.POI_USER_JOINED, {
    poiId: data.poiId,
    userId: data.userId,
    participants: data.participants
  })
}
```

### Step 4: Update WebRTC Initialization
**File:** `frontend/src/stores/videoCallStore.ts`

**Changes:**
1. In `initializeGroupWebRTC()`:
   - After successful initialization
   - Emit `WEBRTC_INITIALIZED` event
2. In `joinPOICall()`:
   - After call setup complete
   - Emit `GROUP_CALL_READY` event

### Step 5: Testing & Verification

**Test Scenarios:**
1. **3 users join sequentially** - All should see each other
2. **User joins during initialization** - Should be queued and added after ready
3. **User leaves and rejoins** - Should work without issues
4. **Multiple rapid joins** - Should handle gracefully

**Verification Points:**
- No more polling/intervals in code
- No race conditions
- Clean event logs showing flow
- All peers connected properly

## Event Flow Diagram

```
User 3 Joins POI
       ↓
Backend: poi_joined event
       ↓
WebSocket Client receives
       ↓
eventBus.emit(POI_USER_JOINED)
       ↓
   ┌───┴───────────┐
   ↓               ↓
User 1 Store    User 2 Store
   ↓               ↓
Check: WebRTC ready?
   ↓               ↓
YES: Add peer   NO: Queue in pendingPeers
   ↓               ↓
Done            Wait for WEBRTC_INITIALIZED
                    ↓
                Process pendingPeers
                    ↓
                Add peer
                    ↓
                Done
```

## Files to Modify

1. ✅ **NEW:** `frontend/src/utils/eventBus.ts` - Event bus implementation
2. 🔧 **MODIFY:** `frontend/src/stores/videoCallStore.ts` - Add event listeners & pending queue
3. 🔧 **MODIFY:** `frontend/src/services/websocket-client.ts` - Emit events instead of direct calls
4. 🧪 **TEST:** Manual testing with 3+ users

## Rollback Plan

If issues arise:
- Event bus is additive (doesn't break existing code)
- Can keep both approaches temporarily
- Easy to disable by not emitting events
- Git revert if needed

## Success Criteria

✅ No polling/intervals in codebase
✅ User 3 can see User 1's video stream
✅ All users can see each other regardless of join timing
✅ Clean, debuggable event logs
✅ No race conditions

## Estimated Time

- Step 1 (Event Bus): 10 minutes
- Step 2 (Video Store): 20 minutes
- Step 3 (WebSocket): 10 minutes
- Step 4 (Emit Events): 5 minutes
- Step 5 (Testing): 15 minutes

**Total: ~60 minutes**

## Next Steps

1. Review this plan
2. Implement Step 1 (Event Bus)
3. Implement Step 2 (Video Store)
4. Implement Step 3 (WebSocket)
5. Test thoroughly
6. Commit and push

Ready to proceed? 🚀
