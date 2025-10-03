# Requirements Document

## Introduction

This spec addresses the critical race condition issue in POI group video modal triggering. Currently, there are two separate code paths that can trigger group video calls, leading to race conditions, duplicate initializations, and inconsistent modal display. This refactoring will centralize all group call decision logic into a single, reliable path.

## Requirements

### Requirement 1: Centralize Group Call Decision Logic

**User Story:** As a developer, I want all group call decisions to flow through a single method, so that race conditions and duplicate logic are eliminated.

#### Acceptance Criteria

1. WHEN any event occurs that might trigger a group call THEN the system SHALL call a single centralized method `checkAndStartGroupCall()`
2. WHEN `checkAndStartGroupCall()` is called THEN it SHALL be the only method that can set `isGroupCallActive = true`
3. WHEN multiple events try to trigger group calls simultaneously THEN only one SHALL succeed due to initialization locking
4. WHEN the centralized method determines a group call should start THEN it SHALL set both `isGroupCallActive` and `currentPOI` atomically

### Requirement 2: Remove Duplicate Group Call Logic

**User Story:** As a developer, I want to eliminate all duplicate group call triggering code paths, so that there is only one source of truth.

#### Acceptance Criteria

1. WHEN reviewing the codebase THEN there SHALL be only one method that can start group calls
2. WHEN a user joins a POI via App.tsx THEN it SHALL NOT contain group call logic directly
3. WHEN a WebSocket poi_joined event is received THEN it SHALL NOT contain group call logic directly
4. WHEN any POI-related event occurs THEN it SHALL delegate to the centralized method

### Requirement 3: Fix State Synchronization Issues

**User Story:** As a user, I want the group video modal to show reliably when I join a POI with other participants, regardless of network timing.

#### Acceptance Criteria

1. WHEN a user joins a POI optimistically THEN `currentUserPOI` SHALL be set immediately
2. WHEN the centralized method checks if a user is in a POI THEN it SHALL use the immediately-set state
3. WHEN API responses are delayed THEN the group call decision SHALL NOT be affected
4. WHEN WebSocket events arrive before API responses THEN the group call SHALL still trigger correctly

### Requirement 4: Implement Initialization Locking

**User Story:** As a developer, I want to prevent race conditions during group call initialization, so that duplicate WebRTC services are not created.

#### Acceptance Criteria

1. WHEN group call initialization starts THEN an initialization lock SHALL be set
2. WHEN another initialization attempt occurs while locked THEN it SHALL be ignored
3. WHEN initialization completes successfully THEN the lock SHALL be released
4. WHEN initialization fails THEN the lock SHALL be released and state cleaned up

### Requirement 5: Remove Dynamic Imports

**User Story:** As a developer, I want to eliminate async import issues in WebSocket handlers, so that group call triggering is not delayed or failed due to import timing.

#### Acceptance Criteria

1. WHEN WebSocket handlers need to trigger group calls THEN they SHALL use direct imports
2. WHEN the websocket-client.ts file is loaded THEN all required stores SHALL be imported at the top
3. WHEN a WebSocket event handler executes THEN it SHALL NOT use dynamic imports for core functionality
4. WHEN store methods are called from WebSocket handlers THEN they SHALL execute synchronously

### Requirement 6: Clean Up All Old Code Paths

**User Story:** As a developer, I want all artifacts of the old dual-path system removed, so that there is no confusion or potential for regression.

#### Acceptance Criteria

1. WHEN the refactoring is complete THEN no group call logic SHALL remain in App.tsx handleJoinPOI method
2. WHEN the refactoring is complete THEN no group call logic SHALL remain in websocket-client.ts handlePOIJoined method
3. WHEN searching the codebase for group call triggers THEN only the centralized method SHALL be found
4. WHEN reviewing commit history THEN all old group call triggering code SHALL be removed, not just commented out