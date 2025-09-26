# WebVideo POC Design Document

## Overview

This document outlines the design for a WebVideo Proof of Concept (POC) that enables peer-to-peer video calling between users by clicking on avatars in BreakoutGlobe. The design focuses on simplicity and learning rather than production-ready implementation.

## Architecture

### High-Level Architecture

```
┌─────────────────┐    WebSocket     ┌─────────────────┐
│   Browser A     │◄────Signaling───►│   Browser B     │
│                 │                  │                 │
│ ┌─────────────┐ │                  │ ┌─────────────┐ │
│ │ Video Call  │ │                  │ │ Video Call  │ │
│ │ Component   │ │                  │ │ Component   │ │
│ └─────────────┘ │                  │ └─────────────┘ │
│        │        │                  │        │        │
│ ┌─────────────┐ │   WebRTC P2P     │ ┌─────────────┐ │
│ │ WebRTC Peer │ │◄────Video/Audio──►│ │ WebRTC Peer │ │
│ │ Connection  │ │                  │ │ Connection  │ │
│ └─────────────┘ │                  │ └─────────────┘ │
└─────────────────┘                  └─────────────────┘
         │                                    │
         └────────────────────────────────────┘
                    STUN Server
                 (Public Internet)
```

### Component Architecture

```
Frontend Components:
├── VideoCallModal (UI for video calls)
├── VideoCallStore (State management)
├── WebRTCService (WebRTC connection handling)
└── App.tsx (Integration point)

Backend Extensions:
├── WebSocket Handler (Call signaling)
├── Message Types (call_request, call_accept, etc.)
└── Call State Tracking (Active calls registry)
```

## Components and Interfaces

### Frontend Components

#### VideoCallModal Component
```typescript
interface VideoCallModalProps {
  isOpen: boolean;
  callState: 'idle' | 'calling' | 'ringing' | 'connecting' | 'connected' | 'ended';
  targetUser: {
    id: string;
    displayName: string;
    avatarURL?: string;
  };
  onAccept: () => void;
  onReject: () => void;
  onEnd: () => void;
  onClose: () => void;
}
```

#### VideoCallStore (Zustand)
```typescript
interface VideoCallState {
  callState: CallState;
  currentCall: CallInfo | null;
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  peerConnection: RTCPeerConnection | null;
  
  // Actions
  initiateCall: (targetUserId: string) => void;
  acceptCall: () => void;
  rejectCall: () => void;
  endCall: () => void;
}
```

#### WebRTCService
```typescript
class WebRTCService {
  private peerConnection: RTCPeerConnection;
  private localStream: MediaStream | null = null;
  
  async initializeLocalStream(): Promise<MediaStream>;
  async createOffer(): Promise<RTCSessionDescriptionInit>;
  async createAnswer(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit>;
  async handleAnswer(answer: RTCSessionDescriptionInit): Promise<void>;
  async addIceCandidate(candidate: RTCIceCandidateInit): Promise<void>;
  cleanup(): void;
}
```

### Backend Extensions

#### WebSocket Message Types
```go
// New message types for video calling
const (
    MessageTypeCallRequest  = "call_request"
    MessageTypeCallAccept   = "call_accept"
    MessageTypeCallReject   = "call_reject"
    MessageTypeCallEnd      = "call_end"
    MessageTypeWebRTCOffer  = "webrtc_offer"
    MessageTypeWebRTCAnswer = "webrtc_answer"
    MessageTypeICECandidate = "ice_candidate"
)
```

#### Call State Management
```go
type CallManager struct {
    activeCalls map[string]*Call
    mutex       sync.RWMutex
}

type Call struct {
    ID          string
    CallerID    string
    CalleeID    string
    State       CallState
    CreatedAt   time.Time
}
```

## Data Models

### Call State Flow
```
idle → calling → ringing → connecting → connected → ended
  ↓      ↓         ↓          ↓           ↓
reject reject   reject    error       end_call
  ↓      ↓         ↓          ↓           ↓
ended  ended    ended     ended       ended
```

### WebSocket Message Formats

#### Call Request
```json
{
  "type": "call_request",
  "data": {
    "callId": "call-123",
    "targetUserId": "user-456",
    "callerInfo": {
      "userId": "user-789",
      "displayName": "John Doe",
      "avatarURL": "https://..."
    }
  }
}
```

#### WebRTC Signaling
```json
{
  "type": "webrtc_offer",
  "data": {
    "callId": "call-123",
    "sdp": "v=0\r\no=...",
    "type": "offer"
  }
}
```

## Error Handling

### Error Categories

1. **Media Access Errors**
   - Camera/microphone permission denied
   - Hardware not available
   - Media constraints not supported

2. **WebRTC Connection Errors**
   - ICE connection failed
   - DTLS handshake failed
   - Peer connection timeout

3. **Signaling Errors**
   - WebSocket disconnection during call
   - Invalid message format
   - User not found or offline

4. **Network Errors**
   - NAT traversal failure
   - Firewall blocking
   - Bandwidth insufficient

### Error Recovery Strategies

- **Graceful Degradation**: Fall back to audio-only if video fails
- **Retry Logic**: Automatic reconnection attempts with exponential backoff
- **User Feedback**: Clear error messages with suggested actions
- **Cleanup**: Proper resource cleanup on any error

## Testing Strategy

### Unit Tests
- VideoCallStore state transitions
- WebRTCService connection handling
- Message validation and parsing
- Error handling scenarios

### Integration Tests
- End-to-end call flow between two browser instances
- WebSocket signaling message exchange
- WebRTC connection establishment
- Error scenarios and recovery

### Manual Testing
- Cross-browser compatibility (Chrome, Firefox, Safari)
- Network condition testing (slow, unstable connections)
- Multiple simultaneous calls
- User experience and UI responsiveness

## Security Considerations

### WebRTC Security
- **Encryption**: WebRTC provides built-in DTLS encryption
- **Origin Validation**: Ensure calls only between authenticated users
- **Rate Limiting**: Prevent call spam or DoS attacks

### Privacy Considerations
- **Consent**: Clear user consent for camera/microphone access
- **Indicators**: Visual indicators when camera/microphone are active
- **Data Retention**: No recording or storage of video/audio streams

## Performance Considerations

### Optimization Strategies
- **Lazy Loading**: Load WebRTC components only when needed
- **Resource Cleanup**: Proper cleanup of media streams and peer connections
- **Bandwidth Management**: Adaptive bitrate based on connection quality
- **UI Responsiveness**: Non-blocking operations for call setup

### Monitoring
- **Connection Quality**: Monitor RTCStats for connection health
- **Error Tracking**: Log WebRTC errors for debugging
- **Performance Metrics**: Track call setup time and success rates

## Implementation Phases

### Phase 1: Basic UI (No WebRTC)
- VideoCallModal component with mock states
- Avatar click detection and call initiation
- Basic state management with VideoCallStore
- WebSocket message structure (without WebRTC)

### Phase 2: WebSocket Signaling
- Backend WebSocket handler extensions
- Call state management and message routing
- Two-browser testing with mock video streams
- Error handling for signaling failures

### Phase 3: WebRTC Integration
- WebRTCService implementation
- SDP offer/answer exchange via WebSocket
- ICE candidate handling
- Local and remote video stream display

### Phase 4: Polish and Learning
- Error handling and user feedback
- Connection quality monitoring
- Performance optimization
- Documentation of learnings and recommendations

## Future Considerations

### Production Readiness Requirements
- **TURN Server**: For NAT traversal in corporate networks
- **Scalability**: Handle multiple simultaneous calls
- **Recording**: Optional call recording functionality
- **Screen Sharing**: Extend to screen sharing capabilities
- **Mobile Support**: Responsive design for mobile devices

### Integration Opportunities
- **POI Integration**: Video calls within POI discussions
- **Breakout Rooms**: Multiple participants in video calls
- **Presentation Mode**: One-to-many video broadcasting
- **Chat Integration**: Text chat during video calls