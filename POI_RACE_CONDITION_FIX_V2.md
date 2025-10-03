# POI Group Call Race Condition Fix V2

## Problem Description

User 3 (Emma) cannot see User 1 (Basti) video and vice versa in a 3-user group call scenario.

### Root Cause

The issue occurs because WebRTC offers arrive **before** the `currentPOI` state is set in the videoCallStore. The sequence is:

1. User 3 joins POI via API call
2. User 1 receives notification and immediately sends WebRTC offer
3. **User 3 receives offer BEFORE `currentPOI` is set** ❌
4. Offer is rejected with "not in matching POI"
5. User 3 completes POI join and WebRTC initialization
6. User 3 waits for offer that will never come again

### Evidence from Logs

**User 3 (Emma) logs:**
```
23:03:08.679 - poi_joined message received
23:03:08.772 - ⚠️ Received POI call offer but not in matching POI (REJECTED)
23:03:08.841+ - Multiple ICE candidates rejected
Later - WebRTC initialized, but offer was already lost
```

**User 1 (Basti) logs:**
```
23:03:08.679 - User joined POI event
23:03:08.772 - Sending POI call offer to User 3
(No answer received from User 3)
```

## Solution

### 1. Queue Offers When POI Not Set

Modified `handlePOICallOffer` to queue offers when **either**:
- WebRTC service doesn't exist yet, OR
- Current POI doesn't match (user still joining)

**Before:**
```typescript
// Check if we're in the right POI
if (state.currentPOI !== poiId) {
  console.warn('⚠️ Received POI call offer but not in matching POI');
  return; // ❌ Offer lost forever
}

// If WebRTC service doesn't exist yet, queue the offer
if (!state.groupWebRTCService) {
  this.pendingOffers.set(fromUserId, data);
  return;
}
```

**After:**
```typescript
// If WebRTC service doesn't exist yet OR we're not in the POI yet, queue the offer
// This handles the race condition where offers arrive before POI join completes
if (!state.groupWebRTCService || state.currentPOI !== poiId) {
  console.log('⏳ WebRTC not ready or POI not set yet, queueing offer from:', fromUserId, {
    hasWebRTC: !!state.groupWebRTCService,
    currentPOI: state.currentPOI,
    offerPOI: poiId
  });
  this.pendingOffers.set(fromUserId, data);
  return;
}
```

### 2. Added processPendingOffers Method

Added public method to process queued offers:

```typescript
public processPendingOffers(): void {
  if (this.pendingOffers.size === 0) {
    return;
  }

  console.log(`🔄 Processing ${this.pendingOffers.size} pending offers`);
  const offers = Array.from(this.pendingOffers.values());
  this.pendingOffers.clear();

  offers.forEach(offerData => {
    console.log('📝 Processing queued offer from:', offerData.fromUserId);
    this.handlePOICallOffer(offerData);
  });
}
```

### 3. Process Pending Offers After WebRTC Init

Modified `initializeGroupWebRTC` in videoCallStore to process pending offers:

```typescript
// Process any pending offers that arrived before WebRTC was ready
const wsClient = (window as any).wsClient;
if (wsClient && wsClient.processPendingOffers) {
  console.log('🔄 Processing pending offers from WebSocket');
  wsClient.processPendingOffers();
}
```

## Fixed Flow

```
User 3 joins POI
    ↓
User 1 sends offer immediately
    ↓
User 3 receives offer → POI not set yet → QUEUED ⏳
    ↓
User 3 completes POI join
    ↓
User 3 initializes WebRTC
    ↓
Process pending offers → Offer delivered ✅
    ↓
User 3 sends answer → Connection established 🎉
```

## Files Modified

1. **frontend/src/services/websocket-client.ts**
   - Modified `handlePOICallOffer` to queue offers when POI not set
   - Added `processPendingOffers()` public method

2. **frontend/src/stores/videoCallStore.ts**
   - Added call to `processPendingOffers()` after WebRTC initialization

## Testing

To test:
1. Have User 1 and User 2 join a POI (group call starts)
2. Have User 3 join the same POI
3. Verify all 3 users can see each other's video streams
4. Check console logs for "Processing pending offers" message

## Expected Logs

**User 3 should now see:**
```
⏳ WebRTC not ready or POI not set yet, queueing offer from: <User1-ID>
✅ Group WebRTC service initialized
🔄 Processing pending offers from WebSocket
🔄 Processing 1 pending offers
📝 Processing queued offer from: <User1-ID>
✅ Remote description (answer) set for peer: <User1-ID>
🎉 WebRTC: Peer connection established with user: <User1-ID>
```

## Related Issues

This is an extension of the previous race condition fix that only handled the case where WebRTC service didn't exist. This fix also handles the case where the POI state isn't set yet.

---

**Status:** ✅ Fixed  
**Date:** October 4, 2025  
**Impact:** Critical - Enables reliable 3+ user group video calls
