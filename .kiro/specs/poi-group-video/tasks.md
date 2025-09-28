# Implementation Plan

- [x] 1. Implement basic POI group call detection and UI trigger
  - Add simple group call state to videoCallStore (currentPOI, isGroupCallActive)
  - Modify POI join flow to trigger group call modal when joining POI with existing participants
  - Create basic group call modal that shows "Group call active in this POI" message
  - Wire POI join/leave to show/hide group call modal
  - _Requirements: 1.1_

- [x] 2. Add dual peer WebRTC support for 2-person POI calls
  - Extend WebRTCService to handle one additional peer connection
  - Modify videoCallStore to track second participant in POI
  - Update VideoCallModal to show two video streams side by side
  - Test with two users joining same POI and establishing video call
  - _Requirements: 1.3, 2.1, 2.2_

- [x] 3. Implement POI-based call signaling through websocket
  - Add group call websocket messages (poi_call_join, poi_call_offer, poi_call_answer)
  - Extend backend websocket handler to route POI call messages between participants
  - Connect POI join/leave events to automatically trigger group call signaling
  - Test complete signaling flow for POI-based calls
  - _Requirements: 1.1, 1.2, 2.3_

- [x] 4. Add third participant support and dynamic video grid
  - Extend WebRTC service to support 3 peer connections using Map structure
  - Implement dynamic video grid layout in VideoCallModal (2x2 grid)
  - Add participant management for adding/removing streams dynamically
  - Test with three users in same POI establishing group video call
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 5. Fix session persistence for display name resolution
  - Ensure WebSocket connections create and persist sessions to database
  - Fix session lookup in GetByIDWithUser to properly resolve display names
  - Test that group call participants show correct display names instead of user IDs
  - Verify session data is available for POI call signaling
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 5. Fix session persistence for display name resolution
  - Ensure WebSocket connections create and persist sessions to database
  - Fix session lookup in GetByIDWithUser to properly resolve display names
  - Test that group call participants show correct display names instead of user IDs
  - Verify session data is available for POI call signaling
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 6. Polish group call experience and cleanup
  - Add proper error handling for failed peer connections
  - Implement automatic cleanup when users leave POI
  - Add visual indicators and participant names in video grid
  - Test complete user journey: join POI → auto group call → leave POI → cleanup
  - _Requirements: 1.2, 2.3_

- [ ] 7. Audit codebase for mock implementations
  - Review server.go for other simple/mock endpoint implementations
  - Replace any remaining mock handlers with proper service-backed handlers
  - Ensure all API endpoints use proper validation, error handling, and persistence
  - Document any intentional mock implementations for testing purposes
  - _Requirements: Infrastructure improvement_