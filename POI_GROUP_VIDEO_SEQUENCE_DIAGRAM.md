# POI Group Video Call - Sequence Diagram

## Scenario: 2 Users, 1 POI - WebRTC Flow Analysis

Based on the codebase analysis, here's the detailed sequence diagram for the POI group video call functionality:

```mermaid
sequenceDiagram
    participant U1 as User 1 (First)
    participant WS1 as WebSocket Client 1
    participant BE as Backend WebSocket Handler
    participant WS2 as WebSocket Client 2
    participant U2 as User 2 (Second)
    participant WebRTC1 as WebRTC Service 1
    participant WebRTC2 as WebRTC Service 2

    Note over U1, U2: Initial Setup - Both users connected to map

    %% User 1 joins POI first
    U1->>WS1: Click POI to join
    WS1->>BE: poi_join message
    BE->>WS1: poi_join_ack
    Note over U1: User 1 in POI, no group call yet (single participant)

    %% User 2 joins the same POI
    U2->>WS2: Click same POI to join
    WS2->>BE: poi_join message
    BE->>WS2: poi_join_ack
    BE->>WS1: poi_joined broadcast (currentCount: 2)
    
    %% Group call initialization triggered for User 1
    Note over WS1: Receives poi_joined, sees currentCount > 1
    WS1->>WS1: videoCallStore.joinPOICall(poiId)
    WS1->>WS1: videoCallStore.initializeGroupWebRTC()
    WS1->>WebRTC1: new GroupWebRTCService()
    WebRTC1->>WebRTC1: getUserMedia() - get local stream
    
    %% User 1 adds User 2 as peer
    WS1->>WebRTC1: addPeer(user2Id)
    WebRTC1->>WebRTC1: new RTCPeerConnection()
    WebRTC1->>WebRTC1: createOffer()
    WebRTC1->>WS1: onIceCandidate callback
    WS1->>BE: poi_call_offer message
    BE->>WS2: poi_call_offer forwarded

    %% User 2 receives offer and initializes WebRTC
    Note over WS2: Receives poi_call_offer from User 1
    WS2->>WS2: videoCallStore.joinPOICall(poiId)
    WS2->>WS2: videoCallStore.initializeGroupWebRTC()
    WS2->>WebRTC2: new GroupWebRTCService()
    WebRTC2->>WebRTC2: getUserMedia() - get local stream
    
    %% User 2 processes offer and creates answer
    WS2->>WebRTC2: addPeer(user1Id)
    WebRTC2->>WebRTC2: new RTCPeerConnection()
    WebRTC2->>WebRTC2: setRemoteDescription(offer)
    WebRTC2->>WebRTC2: createAnswer()
    WebRTC2->>WS2: onIceCandidate callback
    WS2->>BE: poi_call_answer message
    BE->>WS1: poi_call_answer forwarded

    %% User 1 processes answer
    WS1->>WebRTC1: setRemoteDescription(answer)
    
    %% ICE candidate exchange
    WebRTC1->>WS1: ICE candidate generated
    WS1->>BE: poi_call_ice_candidate
    BE->>WS2: poi_call_ice_candidate forwarded
    WS2->>WebRTC2: addIceCandidate()
    
    WebRTC2->>WS2: ICE candidate generated
    WS2->>BE: poi_call_ice_candidate
    BE->>WS1: poi_call_ice_candidate forwarded
    WS1->>WebRTC1: addIceCandidate()

    %% Connection established
    WebRTC1->>WebRTC1: ICE connection established
    WebRTC2->>WebRTC2: ICE connection established
    WebRTC1->>WS1: onRemoteStreamForUser callback
    WebRTC2->>WS2: onRemoteStreamForUser callback
    
    Note over U1, U2: Group video call active - both users can see each other

    %% User leaves POI
    U1->>WS1: Leave POI
    WS1->>BE: poi_leave message
    BE->>WS1: poi_leave_ack
    BE->>WS2: poi_left broadcast
    WS1->>WS1: videoCallStore.leavePOICall()
    WS1->>WebRTC1: cleanup()
    WS2->>WS2: removePeerFromGroupCall(user1Id)
    WS2->>WebRTC2: removePeer(user1Id)
    
    Note over U2: User 2 alone in POI, group call ends
    WS2->>WS2: videoCallStore.leavePOICall()
    WS2->>WebRTC2: cleanup()
```

## Key Issues Identified

### 1. **Race Condition in Group Call Initialization**
- **Problem**: When User 2 joins POI, both users might try to initialize group calls simultaneously
- **Location**: `frontend/src/services/websocket-client.ts:handlePOIJoined()`
- **Impact**: Duplicate WebRTC service creation, connection failures

### 2. **Peer Connection Timing Issues**
- **Problem**: ICE candidates might arrive before peer connections are established
- **Location**: `frontend/src/services/websocket-client.ts:handlePOICallICECandidate()`
- **Current Behavior**: ICE candidates are ignored if no peer connection exists
- **Impact**: Connection establishment failures

### 3. **Inconsistent Group Call State Management**
- **Problem**: Group call state not properly synchronized between participants
- **Location**: `frontend/src/stores/videoCallStore.ts`
- **Impact**: UI inconsistencies, call state mismatches

### 4. **Missing Error Handling for WebRTC Failures**
- **Problem**: No retry mechanism for failed peer connections
- **Location**: `frontend/src/services/webrtc-service.ts:GroupWebRTCService`
- **Impact**: Permanent connection failures

### 5. **Display Name Resolution Issues**
- **Problem**: Inconsistent display name resolution across different code paths
- **Location**: Multiple files - websocket handlers, stores
- **Impact**: Users shown as "Unknown User" or UUID fragments

## Recommended Fixes

### 1. **Implement Proper Initialization Sequencing**
```typescript
// Add initialization lock to prevent race conditions
private _initializingGroupWebRTC: boolean = false;

async initializeGroupWebRTC() {
  if (this._initializingGroupWebRTC) {
    // Wait for ongoing initialization
    while (this._initializingGroupWebRTC) {
      await new Promise(resolve => setTimeout(resolve, 50));
    }
    return;
  }
  this._initializingGroupWebRTC = true;
  // ... initialization logic
  this._initializingGroupWebRTC = false;
}
```

### 2. **Add ICE Candidate Queuing**
```typescript
// Queue ICE candidates until peer connection is ready
private pendingIceCandidates: Map<string, RTCIceCandidateInit[]> = new Map();

handlePOICallICECandidate(data: any) {
  const peerConnection = this.peerConnections.get(fromUserId);
  if (!peerConnection) {
    // Queue candidate for later
    if (!this.pendingIceCandidates.has(fromUserId)) {
      this.pendingIceCandidates.set(fromUserId, []);
    }
    this.pendingIceCandidates.get(fromUserId)!.push(candidate);
    return;
  }
  // Process candidate immediately
  peerConnection.addIceCandidate(candidate);
}
```

### 3. **Add Connection State Synchronization**
```typescript
// Broadcast group call state changes
broadcastGroupCallState(poiId: string, state: GroupCallState) {
  this.manager.BroadcastToPOI(poiId, {
    type: 'group_call_state_update',
    data: { poiId, state }
  });
}
```

### 4. **Implement Retry Logic**
```typescript
// Add retry mechanism for failed connections
async addPeerWithRetry(userId: string, maxRetries: number = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      await this.addPeer(userId);
      return;
    } catch (error) {
      if (attempt === maxRetries) throw error;
      await new Promise(resolve => setTimeout(resolve, 1000 * attempt));
    }
  }
}
```

## Testing Strategy

1. **Unit Tests**: Test individual WebRTC service methods
2. **Integration Tests**: Test complete POI join → group call flow
3. **Race Condition Tests**: Simulate simultaneous POI joins
4. **Network Failure Tests**: Test connection recovery scenarios
5. **Load Tests**: Test with multiple participants (3+ users)

This analysis reveals that while the basic WebRTC infrastructure is in place, there are several timing and state management issues that could cause the group video calls to fail or behave inconsistently, especially when users join POIs in quick succession.