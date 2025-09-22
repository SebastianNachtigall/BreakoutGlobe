# Enhanced POI Interaction - Implementation Tasks (Simplified)

## Minimal Implementation Plan

- [x] 1. Fix POI join button functionality
  - Add working join button click handler to POIDetailsPanel
  - Connect to existing `/api/pois/:id/join` endpoint
  - Show basic loading state during API call
  - Update POI participant count after successful join
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 2. Add participant list display
  - Modify backend to return participant usernames in POI API responses
  - Update POIDetailsPanel to show list of participant usernames
  - Display "No participants" when empty
  - _Requirements: 1.3_

- [x] 3. Implement auto-leave on map/POI clicks
  - Add map click handler to leave current POI and close details panel
  - Add POI click handler to auto-leave current POI when clicking different POI
  - Track which POI user is currently in (simple state)
  - _Requirements: 3.1, 3.2_

- [x] 4. Add basic discussion timer
  - Show simple timer when 2+ people are in POI: "Discussion active for: X minutes"
  - Start timer when participant count reaches 2
  - Reset timer when participant count drops to 0
  - Update timer every minute (not every second initially)
  - _Requirements: 4.1, 4.2, 4.5_

## Success Criteria Validation

After completing all tasks, verify:
- ✅ POI details panel shows participant usernames and discussion timer
- ✅ Join button works and updates all displays in real-time
- ✅ Auto-leave works when clicking map or other POIs
- ✅ Discussion timer accurately tracks active discussion time
- ✅ Real-time synchronization works across multiple sessions
- ✅ POI capacity limits are enforced and displayed
- ✅ Session persistence maintains POI state across refreshes
- ✅ User experience is intuitive and responsive