# Group Call UX Improvements

## Changes Made

### 1. Improved Hang-Up Button Design

**Before:** Red button with 📵 emoji (unclear/weird appearance)

**After:** Red button with proper phone hang-up icon (SVG)

**Implementation:**
- Replaced emoji with Heroicons phone hang-up SVG icon
- Consistent sizing: 56px × 56px for connecting state, 48px × 48px for connected state
- Clean, professional appearance with proper stroke width

**Files Changed:**
- `frontend/src/components/GroupCallModal.tsx`

**Visual:**
```
Before: [📵]  (emoji, inconsistent rendering)
After:  [📞✕] (SVG icon, consistent across browsers)
```

### 2. Auto-Leave POI on Hang-Up

**Rationale:** There's nothing to do in a POI when not in a video call, so hanging up should also leave the POI.

**Behavior:**
- When user clicks hang-up button → Ends video call AND leaves POI
- User returns to normal map view
- Avatar becomes visible again on the map
- Triggers full POI leave flow (API call, state updates, data refresh)

**Implementation:**
- Added `handleEndGroupCall()` handler in `App.tsx`
- Calls `leavePOICall()` to clean up WebRTC
- Calls `handleLeavePOI()` to properly leave the POI (same as manual leave button)
- Ensures all visual updates happen (participant count, avatar visibility, etc.)

**Files Changed:**
- `frontend/src/App.tsx`

**Flow:**
```
User clicks hang-up
       ↓
Clean up WebRTC (streams, peer connections)
       ↓
Clear video call state
       ↓
Leave POI (if in one)
       ↓
User back on map, avatar visible
```

## Code Changes

### GroupCallModal.tsx

**Hang-Up Button (Connecting State):**
```tsx
<button
  onClick={onEndCall}
  className="bg-red-500 hover:bg-red-600 text-white p-4 rounded-full transition-colors flex items-center justify-center"
  title="Leave call"
  style={{ width: '56px', height: '56px' }}
>
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-7 h-7">
    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 3.75L18 6m0 0l2.25 2.25M18 6l2.25-2.25M18 6l-2.25 2.25m1.5 13.5c-8.284 0-15-6.716-15-15V4.5A2.25 2.25 0 014.5 2.25h1.372c.516 0 .966.351 1.091.852l1.106 4.423c.11.44-.054.902-.417 1.173l-1.293.97a1.062 1.062 0 00-.38 1.21 12.035 12.035 0 007.143 7.143c.441.162.928-.004 1.21-.38l.97-1.293a1.125 1.125 0 011.173-.417l4.423 1.106c.5.125.852.575.852 1.091V19.5a2.25 2.25 0 01-2.25 2.25h-2.25z" />
  </svg>
</button>
```

**Hang-Up Button (Connected State):**
```tsx
<button
  onClick={onEndCall}
  className="bg-red-500 hover:bg-red-600 text-white p-3 rounded-full transition-colors flex items-center justify-center"
  title="Leave call"
  style={{ width: '48px', height: '48px' }}
>
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-6 h-6">
    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 3.75L18 6m0 0l2.25 2.25M18 6l2.25-2.25M18 6l-2.25 2.25m1.5 13.5c-8.284 0-15-6.716-15-15V4.5A2.25 2.25 0 014.5 2.25h1.372c.516 0 .966.351 1.091.852l1.106 4.423c.11.44-.054.902-.417 1.173l-1.293.97a1.062 1.062 0 00-.38 1.21 12.035 12.035 0 007.143 7.143c.441.162.928-.004 1.21-.38l.97-1.293a1.125 1.125 0 011.173-.417l4.423 1.106c.5.125.852.575.852 1.091V19.5a2.25 2.25 0 01-2.25 2.25h-2.25z" />
  </svg>
</button>
```

### App.tsx

**handleEndGroupCall Function:**
```typescript
const handleEndGroupCall = useCallback(async () => {
  const videoState = videoCallStore.getState()
  const poiId = videoState.currentPOI
  
  // End the group call
  videoState.leavePOICall()
  
  // Also leave the POI since there's nothing to do without a call
  if (poiId && userProfile) {
    await handleLeavePOI(poiId)
  }
}, [userProfile, handleLeavePOI])
```

**GroupCallModal Usage:**
```typescript
<GroupCallModal
  isOpen={true}
  onClose={handleEndGroupCall}  // ✅ Now uses proper handler
  callState={videoCallState.callState}
  poiId={videoCallState.currentPOI}
  poiName={poiState.pois.find(p => p.id === videoCallState.currentPOI)?.name}
  participants={videoCallState.groupCallParticipants}
  remoteStreams={videoCallState.remoteStreams}
  localStream={videoCallState.localStream}
  isAudioEnabled={videoCallState.isAudioEnabled}
  isVideoEnabled={videoCallState.isVideoEnabled}
  onEndCall={handleEndGroupCall}  // ✅ Now uses proper handler
  onToggleAudio={() => videoCallStore.getState().toggleAudio()}
  onToggleVideo={() => videoCallStore.getState().toggleVideo()}
/>
```

## User Experience

### Before:
1. User in group call
2. Clicks weird red emoji button 📵
3. Call ends but user still in POI
4. User has to manually leave POI (nothing to do there)

### After:
1. User in group call
2. Clicks clear red hang-up button with phone icon
3. Call ends AND user automatically leaves POI
4. User back on map, ready to continue

## Benefits

1. **Clearer UI:** Professional hang-up icon instead of emoji
2. **Better UX:** One action (hang up) does everything needed
3. **Logical Flow:** No reason to stay in POI without a call
4. **Consistent:** Matches user expectations from other video call apps

## Testing

**Manual Test:**
1. Join a POI with other users
2. Group call starts automatically
3. Click the red hang-up button
4. ✅ Call should end
5. ✅ User should leave POI
6. ✅ Avatar should reappear on map
7. ✅ Button should show proper phone hang-up icon

## Files Modified

- `frontend/src/components/GroupCallModal.tsx` - Updated hang-up button UI with SVG icon
- `frontend/src/App.tsx` - Added `handleEndGroupCall` handler that properly leaves POI

## Status

✅ **IMPLEMENTED** - Both improvements are live
