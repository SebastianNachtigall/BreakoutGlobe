# Group Call Participant Limit Analysis

## Current Situation

The group video calls are currently **limited to 2 participants**, but the system was designed to support **6-8 participants** before the refactor.

## Root Cause Analysis

### 1. **POI MaxParticipants Setting** ✅ NOT THE ISSUE
- **Backend Default**: `MaxParticipants: 10` (in `backend/internal/models/poi.go`)
- **Handler Default**: Falls back to `10` if not provided (in `backend/internal/handlers/poi_handler.go`)
- **Frontend Validation**: Allows 1-50 participants (in `frontend/src/types/models.ts`)
- **Conclusion**: The POI itself can support up to 10 participants by default

### 2. **Frontend POI Store Enforcement** ✅ NOT THE ISSUE
The frontend checks if a POI is full before allowing joins:
```typescript
// frontend/src/stores/poiStore.ts
if (!poi || poi.participantCount >= poi.maxParticipants) {
  return false;
}
```
But this only prevents joining when the POI reaches its `maxParticipants` limit (10), not limiting group calls to 2.

### 3. **Group Call Initialization Logic** ⚠️ POTENTIAL ISSUE
In `frontend/src/stores/videoCallStore.ts`, the `checkAndStartGroupCall` function:
```typescript
// 4. Check if multiple participants (need at least 2 for group call)
if (participantCount <= 1) {
  console.log('👤 Only one participant, no group call needed');
  return;
}
```
This correctly requires at least 2 participants to start, but **doesn't limit the maximum**.

### 4. **WebRTC Peer Connection Logic** ⚠️ POTENTIAL ISSUE
The current implementation uses **1-to-1 peer connections** for each participant:
- Each user creates a separate peer connection to every other user
- This is a **mesh topology** (full mesh WebRTC)
- For N participants, each user needs N-1 peer connections

**Mesh Topology Limitations**:
- 2 participants: 1 connection per user ✅
- 3 participants: 2 connections per user ✅
- 4 participants: 3 connections per user ✅
- 6 participants: 5 connections per user ⚠️
- 8 participants: 7 connections per user ⚠️

### 5. **UI Grid Layout** ✅ SUPPORTS UP TO 6+
In `frontend/src/components/GroupCallModal.tsx`:
```typescript
const getGridLayout = (participantCount: number) => {
  if (participantCount <= 1) return { cols: 1, rows: 1, gridClass: 'grid-cols-1', height: 'h-80' };
  if (participantCount <= 2) return { cols: 2, rows: 1, gridClass: 'grid-cols-2', height: 'h-80' };
  if (participantCount <= 4) return { cols: 2, rows: 2, gridClass: 'grid-cols-2', height: 'h-96' };
  if (participantCount <= 6) return { cols: 3, rows: 2, gridClass: 'grid-cols-3', height: 'h-96' };
  // For more than 6, still use 3x2 but some will be hidden/scrollable
  return { cols: 3, rows: 2, gridClass: 'grid-cols-3', height: 'h-96' };
}
```
The UI supports up to 6 participants in a 3x2 grid, with overflow handling for more.

### 6. **Actual Limitation Discovery** 🔍

After reviewing the code, **there is NO explicit 2-participant limit** in the current implementation. The system should theoretically support multiple participants.

## Possible Reasons for Observed 2-Participant Limit

### Theory 1: **Testing Environment**
- You may have only tested with 2 browser windows
- The system might actually support more, but hasn't been tested with 3+ participants

### Theory 2: **POI Configuration**
- The specific POI being tested might have `maxParticipants` set to 2
- Check the POI creation form - it might default to 2 in the UI

### Theory 3: **WebRTC Signaling Issues**
- The lexicographic comparison for offer initiation works for 2 participants
- With 3+ participants, there might be signaling coordination issues
- Each pair needs to establish a connection, which requires proper signaling

### Theory 4: **Race Conditions**
- The `_initializingGroupCall` lock might prevent additional participants from joining
- The 10-second timeout might not be sufficient for complex scenarios

## What Needs to Be Checked

1. **POI Creation Default**:
   - Check `frontend/src/components/POICreationModal.tsx` for default `maxParticipants` value
   - Verify what value is being sent when creating a POI

2. **Database POI Records**:
   - Check the actual `max_participants` value in the database for existing POIs
   - Query: `SELECT id, name, max_participants FROM pois;`

3. **WebRTC Mesh Topology**:
   - Verify that peer connections are being created for ALL participants
   - Check if the signaling works correctly for 3+ participants
   - Monitor console logs when 3rd participant joins

4. **Participant Addition Logic**:
   - Verify that `addGroupCallParticipant` is called for each participant
   - Check if `addPeerToGroupCall` successfully creates peer connections for all

## Recommendations

### Immediate Actions (No Coding)
1. **Test with 3 participants** to see if the issue is real or just untested
2. **Check database** for actual `max_participants` values
3. **Review console logs** when 3rd participant joins to see where it fails
4. **Check POI creation form** for default values

### If Issue Confirmed - Potential Fixes

#### Option 1: Increase POI Default (Simple)
- Change default `maxParticipants` in POI creation form to 8
- Update database records for existing POIs

#### Option 2: Fix WebRTC Mesh Topology (Complex)
- Ensure all peer connections are established correctly
- Fix any signaling coordination issues for 3+ participants
- Add better error handling and retry logic

#### Option 3: Implement SFU Architecture (Advanced)
- For 6-8 participants, mesh topology becomes inefficient
- Consider using a Selective Forwarding Unit (SFU) server
- This would require significant backend changes

## Test Plan

### Test Case 1: Verify Current Limit
1. Open 3 browser windows
2. Create profiles in each
3. Join the same POI
4. Observe if group call starts for all 3
5. Check console logs for errors

### Test Case 2: Check POI Configuration
1. Create a new POI
2. Check what `maxParticipants` value is set
3. Verify in database
4. Try joining with that many participants

### Test Case 3: WebRTC Peer Connections
1. Join with 3 participants
2. Check browser console for peer connection logs
3. Verify that each user has 2 peer connections (to the other 2 users)
4. Check for ICE candidate exchange logs

## Conclusion

**The system is NOT explicitly limited to 2 participants**. The limitation is likely due to:
1. Untested scenarios (only tested with 2 participants)
2. POI configuration issues (maxParticipants set too low)
3. WebRTC signaling coordination issues for 3+ participants

**Next Steps**: Test with 3 participants and review console logs to identify the actual bottleneck.


## 🎯 ROOT CAUSE CONFIRMED

### The Bug

In `frontend/src/stores/videoCallStore.ts`, the `checkAndStartGroupCall` function has faulty logic:

```typescript
// 2. Check if group call already active for this POI
const state = get();
if (state.isGroupCallActive && state.currentPOI === poiId) {
  console.log('✅ Group call already active for this POI, skipping');
  return;  // ❌ BUG: This prevents new users from joining!
}
```

### What Happens:

1. **User 1 & 2** join POI → Group call initializes successfully
2. **User 3** joins POI → Detects call is active → **SKIPS initialization**
3. **User 3** never joins the WebRTC mesh
4. **User 3's call immediately ends** (cleanup triggered)

### Console Evidence from User 3:

```
✅ Group call already active for this POI, skipping
🚪 Leaving POI group call
🧹 WebRTC: Cleaning up group call resources...
```

### The Logic Flaw:

The check is meant to prevent **duplicate initialization by the same user**, but it incorrectly prevents **new users from joining an existing call**.

### Why It Appears Limited to 2 Participants:

- Only the **first 2 users** who trigger group call initialization can join
- **Any subsequent users** are blocked by this check
- They skip initialization → Never connect → Call ends immediately

### The Fix Required:

The logic needs to differentiate between:
1. **Same user re-joining** → Should skip ✅
2. **New user joining existing call** → Should initialize and connect ❌ Currently broken

### Correct Behavior for New User Joining:

1. Initialize their own WebRTC service
2. Get list of existing participants  
3. Create peer connections to ALL existing participants
4. Notify existing participants to create peer connections back

### Impact:

This bug **completely prevents more than 2 participants** from joining group calls, regardless of the POI's `maxParticipants` setting or UI grid layout capabilities.

## Solution Approach

### Option 1: Remove the Check (Simple but Risky)
Remove the `isGroupCallActive` check entirely and rely on other guards.

**Pros**: Quick fix
**Cons**: Might allow duplicate initializations

### Option 2: Check User's Own State (Recommended)
Instead of checking if "a call is active for this POI", check if "THIS USER is already in this call":

```typescript
// Check if THIS USER is already in this specific call
if (state.isGroupCallActive && state.currentPOI === poiId && state.groupWebRTCService) {
  // Only skip if we're already initialized and connected
  console.log('✅ Already in this group call, skipping');
  return;
}
```

### Option 3: Add Participant to Existing Call (Best)
When a new user joins an existing call, add them to the mesh:

```typescript
if (state.isGroupCallActive && state.currentPOI === poiId) {
  // We're already in this call, but a new participant might have joined
  // Add the new participant to our existing call
  console.log('📞 Adding new participant to existing call');
  // Logic to add new peer connection
  return;
}
```

### Option 4: Event-Driven Architecture (Most Robust)
Listen for `poi_joined` WebSocket events and dynamically add participants:
- When ANY user joins the POI, ALL existing participants receive an event
- Each participant creates a peer connection to the new user
- This works regardless of who joined first

## Recommendation

**Implement Option 4** (Event-Driven) as it's the most scalable and robust solution for true multi-participant support (3-8 users).

This requires:
1. WebSocket listener for `poi_joined` events
2. Dynamic peer connection creation when new participants join
3. Proper cleanup when participants leave
4. Signaling coordination for the new peer connections
