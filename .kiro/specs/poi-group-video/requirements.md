# Requirements Document

## Introduction

This feature extends the existing user-to-user video functionality to support group video calls organized by Points of Interest (POIs). Users can join video calls by joining a POI, creating location-based video chat rooms. Each POI maintains its own separate video call session, allowing multiple concurrent group calls across different locations on the map.

## Requirements

### Requirement 1

**User Story:** As a user, I want to join a group video call when I join a POI, so that I can have video conversations with other users at the same location.

#### Acceptance Criteria

1. WHEN a user joins a POI THEN the system SHALL automatically connect them to the POI's group video call
2. WHEN a user leaves a POI THEN the system SHALL automatically disconnect them from the POI's group video call
3. WHEN multiple users are in the same POI THEN the system SHALL establish a multi-party video connection between all participants

### Requirement 2

**User Story:** As a user, I want to see video feeds from all other users in my current POI, so that I can have face-to-face conversations with the group.

#### Acceptance Criteria

1. WHEN I am in a POI with other users THEN the system SHALL display video feeds from all other participants
2. WHEN a new user joins my POI THEN the system SHALL add their video feed to my call interface
3. WHEN a user leaves my POI THEN the system SHALL remove their video feed from my call interface
4. WHEN no other users are in my POI THEN the system SHALL show only my own video feed

### Requirement 3

**User Story:** As a user, I want to see proper display names for other participants in group video calls, so that I can identify who I'm talking to.

#### Acceptance Criteria

1. WHEN I join a POI group call THEN the system SHALL display the correct display names for all participants
2. WHEN a user connects to the system THEN the system SHALL persist their session information to the database
3. WHEN resolving display names for video calls THEN the system SHALL successfully retrieve user and session information from the database
4. WHEN session lookup fails THEN the system SHALL gracefully fallback to using the user ID as display name

