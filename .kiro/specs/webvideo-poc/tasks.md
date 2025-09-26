# WebVideo POC Implementation Plan

## Phase 1: Basic UI Foundation (No WebRTC)

- [x] 1. Create VideoCallModal component with basic UI
  - Create modal component with call states (calling, ringing, connected, ended)
  - Add basic styling with Tailwind CSS using emoji icons for simplicity
  - Include caller information display (name, avatar)
  - Add accept/reject/end call buttons
  - _Requirements: 1.1, 2.2, 5.1, 5.2, 5.4, 5.5_

- [x] 2. Create VideoCallStore for state management
  - Implement Zustand store for call state management
  - Add actions for initiateCall, acceptCall, rejectCall, endCall
  - Include call information (callId, targetUser, isIncoming)
  - Add state transitions and validation
  - _Requirements: 2.1, 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 3. Integrate avatar click detection
  - Add onAvatarClick handler to MapContainer component
  - Implement click detection that excludes self-clicks
  - Connect avatar clicks to VideoCallStore.initiateCall
  - Add VideoCallModal to App.tsx with proper state binding
  - _Requirements: 1.1, 1.2, 1.3, 7.1, 7.2_

- [x] 4. Test basic UI flow with mock data
  - Test call initiation from avatar clicks
  - Verify modal displays with correct user information
  - Test accept/reject/end call button functionality
  - Ensure proper state transitions without WebRTC
  - _Requirements: 1.1, 2.2, 2.3, 2.4, 5.1, 5.2, 5.4, 5.5_

## Phase 2: WebSocket Signaling Infrastructure

- [x] 5. Extend WebSocket message types in backend
  - Add call_request, call_accept, call_reject, call_end message types
  - Implement message validation for new call message types
  - Add message routing in WebSocket handler
  - Create CallManager for tracking active calls
  - _Requirements: 2.1, 2.5, 6.4, 7.3_

- [x] 6. Implement call signaling message handlers
  - Handle call_request messages and route to target user
  - Implement call_accept/call_reject message processing
  - Add call timeout handling and automatic cleanup
  - Include user availability checking
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 6.4_

- [x] 7. Connect frontend to WebSocket signaling
  - Extend WebSocketClient to handle call messages
  - Connect VideoCallStore actions to WebSocket message sending
  - Add incoming call message handlers
  - Implement call state synchronization between users
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 5.1, 5.2, 5.3_

- [x] 8. Test two-browser call signaling
  - Test call initiation between two browser tabs
  - Verify incoming call notifications work correctly
  - Test accept/reject flows and state synchronization
  - Ensure proper cleanup when calls end or timeout
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 7.5_

## Phase 3: WebRTC Integration

- [x] 9. Create WebRTCService for peer connection management
  - Implement WebRTC peer connection setup
  - Add local media stream initialization (getUserMedia)
  - Create SDP offer/answer generation methods
  - Implement ICE candidate handling
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 6.1, 6.2_

- [x] 10. Add WebRTC signaling message types
  - Extend backend with webrtc_offer, webrtc_answer, ice_candidate messages
  - Add SDP and ICE candidate message routing
  - Implement WebRTC message validation
  - Connect WebRTCService to WebSocket signaling
  - _Requirements: 3.1, 6.2, 6.3_

- [x] 11. Implement video stream display
  - Add local video element to VideoCallModal
  - Add remote video element with proper sizing
  - Connect WebRTC streams to video elements
  - Implement picture-in-picture layout for local video
  - _Requirements: 3.2, 3.3, 7.1_

- [x] 12. Test end-to-end WebRTC connection
  - Test complete call flow from click to video connection
  - Verify both local and remote video streams work
  - Test audio transmission between browsers
  - Ensure proper connection cleanup on call end
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.4_

## Phase 4: Call Controls and Polish

- [x] 13. Implement call controls
  - Add mute/unmute microphone functionality
  - Add camera on/off toggle
  - Implement visual indicators for muted/camera-off states
  - Add keyboard shortcuts for common actions
  - _Requirements: 4.1, 4.2, 4.3, 5.3_

- [ ] 14. Add comprehensive error handling
  - Handle camera/microphone permission denied errors
  - Add WebRTC connection failure error handling
  - Implement network error detection and user feedback
  - Add graceful degradation for unsupported browsers
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 15. Implement proper resource cleanup
  - Ensure media streams are stopped on call end
  - Clean up peer connections and event listeners
  - Handle browser tab close/refresh during calls
  - Add memory leak prevention measures
  - _Requirements: 4.4, 6.3, 7.2_

- [ ] 16. Test cross-browser compatibility
  - Test in Chrome, Firefox, and Safari
  - Verify WebRTC feature support across browsers
  - Test mobile browser compatibility (basic)
  - Document browser-specific issues and workarounds
  - _Requirements: 6.1, 6.2, 6.5_

## Phase 5: Integration and Learning Documentation

- [ ] 17. Ensure seamless integration with existing features
  - Test video calls while map interactions continue
  - Verify POI functionality works during calls
  - Ensure avatar movements sync properly during calls
  - Test multiple users scenario (non-participants)
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 18. Performance testing and optimization
  - Test with slow network connections
  - Monitor memory usage during extended calls
  - Test multiple simultaneous calls (if applicable)
  - Optimize video quality based on connection
  - _Requirements: 6.3, 7.5_

- [ ] 19. Document learnings and recommendations
  - Document WebRTC implementation challenges
  - Record browser compatibility findings
  - Note performance and scalability considerations
  - Create recommendations for production implementation
  - Document security and privacy considerations
  - _Requirements: All requirements - learning outcomes_

- [ ] 20. Create demo and presentation materials
  - Prepare demo scenarios for stakeholder review
  - Document key technical decisions and trade-offs
  - Create comparison with alternative solutions
  - Outline next steps for production implementation
  - _Requirements: All requirements - demonstration_