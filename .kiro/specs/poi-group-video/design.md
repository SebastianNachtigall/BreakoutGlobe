# Design Document

## Overview

This design extends the existing peer-to-peer video calling system to support group video calls organized by POIs. The solution leverages the current WebRTC infrastructure and websocket signaling while adding group call coordination. Users automatically join group video calls when they join a POI, creating location-based video chat rooms.

## Architecture

### High-Level Flow
1. User connects to WebSocket and session is persisted to database
2. User joins a POI via existing `poi_join` websocket message
3. Backend checks if POI has active group video call
4. If active call exists, user is automatically added to the group call
5. WebRTC mesh network is established between all participants in the POI
6. Display names are resolved using persisted session and user data
7. When user leaves POI, they are automatically removed from group call

### Component Integration
- **Existing POI System**: Leverage current POI join/leave mechanics
- **Existing Video System**: Extend current WebRTC service for multiple peers
- **Websocket Signaling**: Add group call coordination messages
- **State Management**: Extend video call store for group calls

## Components and Interfaces

### Frontend Components

#### Enhanced WebRTC Service
```typescript
interface GroupWebRTCService extends WebRTCService {
  // Multiple peer connection management
  peerConnections: Map<string, RTCPeerConnection>;
  remoteStreams: Map<string, MediaStream>;
  
  // Group call methods
  addPeer(userId: string): Promise<void>;
  removePeer(userId: string): void;
  createOfferForPeer(userId: string): Promise<RTCSessionDescriptionInit>;
  handleAnswerFromPeer(userId: string, answer: RTCSessionDescriptionInit): Promise<void>;
}
```

#### Enhanced Video Call Store
```typescript
interface GroupVideoCallState extends VideoCallState {
  // Group call state
  currentPOI: string | null;
  groupCallParticipants: Map<string, UserInfo>;
  remoteStreams: Map<string, MediaStream>;
  
  // Group call actions
  joinPOICall(poiId: string): void;
  leavePOICall(): void;
  addParticipant(userId: string, userInfo: UserInfo): void;
  removeParticipant(userId: string): void;
}
```

#### Enhanced Video Call Modal
- Support multiple video streams in grid layout
- Show participant names and avatars
- Handle dynamic participant addition/removal

### Backend Components

#### Enhanced WebSocket Handler
```go
// New message types for group video calls
type GroupCallMessage struct {
    Type string `json:"type"` // "group_call_join", "group_call_leave", "group_call_offer", etc.
    POIId string `json:"poiId"`
    ParticipantId string `json:"participantId"`
    Data interface{} `json:"data"`
}

// Session persistence on WebSocket connection
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
    // Ensure session exists in database before allowing WebSocket connection
    // Create session if it doesn't exist
    // Associate session with user for display name resolution
}
```

#### POI Group Call Manager
```go
type POIGroupCallManager struct {
    // Track active group calls by POI ID
    activeGroupCalls map[string]*GroupCall
    
    // Methods
    JoinGroupCall(poiId, userId string) error
    LeaveGroupCall(poiId, userId string) error
    GetGroupCallParticipants(poiId string) []string
    BroadcastToGroupCall(poiId string, message Message) error
}

type GroupCall struct {
    POIId string
    Participants map[string]*Client
    CreatedAt time.Time
}
```

## Data Models

### Group Call State
```typescript
interface GroupCallState {
  poiId: string;
  participants: UserInfo[];
  isActive: boolean;
  createdAt: Date;
}

interface UserInfo {
  userId: string;
  displayName: string;
  avatarURL?: string;
}
```

### WebSocket Message Extensions
```typescript
// New message types
interface GroupCallJoinMessage {
  type: 'group_call_join';
  data: {
    poiId: string;
    participants: UserInfo[];
  };
}

interface GroupCallOfferMessage {
  type: 'group_call_offer';
  data: {
    poiId: string;
    fromUserId: string;
    toUserId: string;
    offer: RTCSessionDescriptionInit;
  };
}
```

## Error Handling

### Connection Failures
- Graceful degradation when WebRTC connections fail
- Automatic retry for failed peer connections
- Fallback to audio-only if video fails

### POI State Synchronization
- Handle race conditions when multiple users join/leave simultaneously
- Ensure consistent group call state across all participants
- Recovery mechanisms for desynchronized state

### Resource Management
- Automatic cleanup of abandoned group calls
- Memory management for multiple peer connections
- Bandwidth optimization for multiple video streams

## Testing Strategy

### Unit Tests
- Group call state management
- Multiple peer connection handling
- Message routing and validation

### Integration Tests
- POI join/leave with group call coordination
- Multi-user WebRTC establishment
- Websocket message flow for group calls

### Manual Testing
- Multiple browser windows simulating different users
- Network interruption scenarios
- POI switching with active calls

## Implementation Phases

### Phase 1: Basic Group Call Infrastructure
- Extend WebRTC service for multiple peers
- Add group call state management
- Implement basic POI-based call joining

### Phase 2: Enhanced UI and UX
- Multi-participant video grid layout
- Participant management controls
- Visual indicators for active group calls

### Phase 3: Optimization and Polish
- Bandwidth optimization
- Connection quality indicators
- Advanced error handling and recovery