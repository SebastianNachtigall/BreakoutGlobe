# User Profile & Account System Implementation Plan

## Overview

This implementation plan uses a **vertical slice approach** to deliver working functionality early and enable browser testing from the first few tasks. Each slice builds a complete user journey from database to UI, allowing for immediate testing and user feedback.

## Implementation Tasks

### Slice 1: Basic Guest Profile Creation & Display (Browser Testable)

- [x] 1. Create minimal User model and database foundation
  - Write tests using NewUser() builder for basic model validation (displayName, accountType)
  - Create User model with minimal fields: ID, DisplayName, AccountType, CreatedAt
  - Create database migration using established migration patterns
  - Write tests for display name validation (3-50 characters) and guest account creation
  - _Requirements: 1.1, 1.2_

- [x] 2. Implement basic User repository for profile creation
  - Write tests using newUserRepositoryScenario(t) for Create and GetByID operations
  - Implement UserRepository interface with Create() and GetByID() methods only
  - Use fluent assertions: AssertUser(t, user).HasDisplayName().HasAccountType()
  - Write tests using expectUserCreationSuccess() for guest profile workflows
  - _Requirements: 1.1, 1.6_

- [x] 3. Create profile creation API endpoint
  - Write tests using newUserHandlerScenario(t) for POST /api/users/profile
  - Use expectProfileCreationSuccess() for guest profile creation workflow
  - Implement minimal User service with CreateGuestProfile() method
  - Write tests using expectRateLimitSuccess() for profile creation rate limiting
  - _Requirements: 1.1, 1.6_

- [x] 4. Build profile creation modal (frontend)
  - Write tests using component test patterns for ProfileCreationModal
  - Use expectFormValidation() for display name validation (3-50 characters)
  - Create basic modal with display name input and create button
  - Write tests using expectGuestProfileCreation() for API integration
  - _Requirements: 1.1, 1.7_

- [x] 5. Display user avatars on map with names
  - Write tests using enhanced AvatarData interface for display name rendering
  - Update existing avatar display to show display names instead of session IDs
  - Use expectAvatarTooltipDisplay() for user name hover functionality
  - Write tests using expectInitialsFallback() for default avatar generation from names
  - _Requirements: 3.1, 3.3_

**Result after Slice 1**: Users can create guest profiles with display names and see each other's names on the map. Fully browser testable!

### Slice 2: Avatar Upload & Profile Persistence (Complete Avatar Experience)

- [x] 6. Add basic avatar upload functionality
  - Extend User model with AvatarURL field and update migration
  - Write tests using newAvatarScenario(t) for basic upload workflows
  - Implement POST /api/users/avatar endpoint with file validation
  - Use expectFileUploadValidation() for size and type checking (max 2MB, JPG/PNG)
  - _Requirements: 1.3, 3.1_

- [x] 6.1. Fix import cycle in repository tests (Tech Debt)
  - Resolve services ↔ testdata ↔ repository import cycle
  - Re-enable user repository tests that were disabled
  - Refactor testdata package to avoid circular dependencies
  - Ensure all repository tests pass with proper isolation
  - _Technical Debt: Critical for test infrastructure integrity_

- [ ] 7. Add profile retrieval endpoint for persistence
  - Write tests using newUserHandlerScenario(t) for GET /api/users/profile
  - Implement GET /api/users/profile endpoint to retrieve existing profiles
  - Use expectProfileRetrievalSuccess() for profile lookup workflows
  - Add session-based user identification for profile persistence
  - _Requirements: 7.1, 7.2_

- [ ] 8. Add localStorage sync for guest profiles
  - Write tests using expectLocalStorageSync() for profile persistence patterns
  - Create userProfileStore with localStorage backup for guest profiles
  - Use expectBackendSynchronization() for profile data sync workflows
  - Write tests using expectOfflineProfileAccess() for localStorage fallback
  - _Requirements: 7.1, 7.2_

- [ ] 9. Add avatar file serving and storage infrastructure
  - Write tests using expectAvatarFileServing() for uploaded image access
  - Create file storage directory structure and serving endpoint
  - Implement GET /api/users/avatar/:filename endpoint for image serving
  - Use expectAvatarFileValidation() for secure file access and MIME type validation
  - _Requirements: 3.1, 3.2_

**Result after Slice 2**: Complete avatar experience - users can upload profile pictures, see them on the map, and profiles persist across browser sessions. Frontend avatar rendering is already implemented and ready! Fully end-to-end testable!

### Slice 3: Profile Management & Settings

- [ ] 10. Implement profile update functionality
  - Extend User model with AboutMe field and update migration
  - Write tests using expectProfileUpdateAuthorization() for guest profile updates
  - Add PUT /api/users/profile endpoint with display name and about me updates
  - Use expectGuestProfileUpdateRestrictions() for limited editing capabilities
  - _Requirements: 5.1, 5.2_

- [ ] 11. Create basic profile settings UI
  - Write tests using component test patterns for ProfileSettingsModal
  - Use expectProfileSettingsAccess() for guest profile editing restrictions
  - Create settings modal with display name (read-only) and about me (editable)
  - Write tests using expectProfileUpdateSuccess() for save functionality
  - _Requirements: 5.1, 5.4_

**Result after Slice 3**: Users can edit their profiles and manage their information through a settings interface.

### Slice 4: Full Account Upgrade & Authentication

- [ ] 12. Add email and password fields to User model
  - Extend User model with Email, PasswordHash, EmailVerified fields
  - Write tests using NewUser() builder for full account validation
  - Create migration to add new fields with proper constraints
  - Write tests for email uniqueness and password hashing validation
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 13. Implement authentication endpoints
  - Write tests using newAuthScenario(t) for login/logout workflows
  - Create POST /api/auth/login and POST /api/auth/logout endpoints
  - Use expectLoginSuccess() and expectLogoutCleanup() for session management
  - Implement JWT token generation and validation middleware
  - _Requirements: 2.1, 4.6_

- [ ] 14. Create account upgrade flow (frontend)
  - Write tests using component test patterns for AccountUpgradeModal
  - Use expectAccountUpgradeFlow() for guest to full account conversion
  - Create upgrade modal with email, password, and confirm password fields
  - Write tests using expectEmailVerificationUI() for verification workflow
  - _Requirements: 2.1, 2.4, 2.5_

- [ ] 15. Add email verification system
  - Write tests using expectEmailVerificationSuccess() for token-based confirmation
  - Implement POST /api/users/verify-email endpoint with secure token handling
  - Create email verification service with token generation and validation
  - Use expectEmailVerificationSecurity() for token expiration and security
  - _Requirements: 2.5, 2.6_

**Result after Slice 4**: Users can upgrade to full accounts with email/password and login from any device.

### Slice 5: POI Creation Permissions & Ownership

- [ ] 16. Update POI model with creator relationship
  - Extend POI model with CreatedBy field referencing User ID
  - Write tests using existing newPOIScenario(t) enhanced with user ownership
  - Create migration to add CreatedBy field to existing POIs
  - Use expectPOIOwnershipTracking() for creator assignment validation
  - _Requirements: 9.1, 9.6_

- [ ] 17. Implement POI creation permissions
  - Write tests using expectPOICreationPermissions() for full account requirement
  - Update POI creation endpoints to require full account authentication
  - Use expectPermissionDenied() for guest profile POI creation attempts
  - Add permission checks to existing POI creation workflows
  - _Requirements: 9.1, 9.2_

- [ ] 18. Update POI creation UI with permission checks
  - Write tests using expectPOICreationUI() for account type restrictions
  - Update POI creation modal to show upgrade prompt for guest profiles
  - Use expectAccountUpgradePrompt() for POI creation permission messaging
  - Add visual indicators for POI ownership in the UI
  - _Requirements: 9.1, 9.7_

- [ ] 19. Add POI edit/delete permissions
  - Write tests using expectPOIEditPermissions() for owner authorization
  - Implement PUT /api/pois/:id and DELETE /api/pois/:id with ownership checks
  - Use expectPOIDeletionAuthorization() for removal permissions
  - Add edit/delete buttons to POI details for owners
  - _Requirements: 9.6, 9.7_

**Result after Slice 5**: Full accounts can create, edit, and delete POIs. Guest profiles see upgrade prompts.

### Slice 6: Multi-Map Support & User Isolation

- [ ] 20. Create Map model and basic relationships
  - Write tests using NewMap() builder for map creation and validation
  - Create Map model with ID, Name, Description, CreatedBy, IsActive fields
  - Create migration for maps table with proper foreign key constraints
  - Use expectMapCreationSuccess() for basic map management workflows
  - _Requirements: 8.1, 8.2_

- [ ] 21. Update Session and POI models for map relationships
  - Add MapID field to Session and POI models with foreign key constraints
  - Write tests using expectMapIsolatedSessions() for user isolation per map
  - Create migrations to add MapID to existing sessions and POIs
  - Use expectMapPOIIsolation() for cross-map visibility validation
  - _Requirements: 8.1, 8.3, 8.4_

- [ ] 22. Implement map selection and user isolation
  - Write tests using newMapScenario(t) for map-scoped operations
  - Create GET /api/maps endpoint for available maps listing
  - Update session creation to associate with specific map
  - Use expectMapUserIsolation() for user visibility per map
  - _Requirements: 8.1, 8.2, 8.5_

- [ ] 23. Add map selection UI and context switching
  - Write tests using component test patterns for MapSelector component
  - Create map selection dropdown in the main UI
  - Use expectMapContextSwitching() for session transfer between maps
  - Update avatar display to show only users on current map
  - _Requirements: 8.1, 8.2_

**Result after Slice 6**: Multiple independent maps with isolated users and POIs.

### Slice 7: Admin Features & User Management

- [ ] 24. Add role-based permissions system
  - Extend User model with Role field (user, admin, superadmin)
  - Write tests using newAuthorizationScenario(t) for permission checking
  - Create migration to add role field with default 'user' value
  - Use expectRolePermissionSuccess() for hierarchical access validation
  - _Requirements: 4.1, 4.4, 6.1_

- [ ] 25. Create user management API endpoints
  - Write tests using newAdminScenario(t) for user management workflows
  - Implement GET /api/admin/users and PUT /api/admin/users/:id/role endpoints
  - Use expectAdminUserListingAuthorization() for admin-only access
  - Add role assignment and user status management functionality
  - _Requirements: 6.1, 6.2, 6.4, 6.5_

- [ ] 26. Build admin user management UI
  - Write tests using component test patterns for UserManagementPanel
  - Create admin panel with user listing, search, and role assignment
  - Use expectUserListingDisplay() and expectRoleAssignmentInterface() patterns
  - Add user status toggle and role management controls
  - _Requirements: 6.1, 6.3, 6.4_

- [ ] 27. Add admin POI moderation capabilities
  - Write tests using expectPOIModerationPermissions() for admin POI management
  - Allow admins to edit/delete any POI on maps they manage
  - Use expectAdminBoundaryEnforcement() for superadmin protection
  - Add admin indicators and moderation controls to POI UI
  - _Requirements: 9.7, 6.1_

**Result after Slice 7**: Complete admin system with user management and POI moderation.

### Slice 8: Enhanced Real-time Features & Polish

- [ ] 28. Implement real-time profile updates via WebSocket
  - Write tests using current TestWebSocket infrastructure for profile events
  - Add profile update broadcasting to existing WebSocket system
  - Use expectProfileUpdateBroadcast() and expectAvatarChangeBroadcast() patterns
  - Integrate with existing WebSocket message handling infrastructure
  - _Requirements: 3.1, 7.1_

- [ ] 29. Add advanced avatar features and animations
  - Write tests using expectAvatarAnimationTransitions() for smooth movement
  - Implement avatar clustering and collision detection for crowded areas
  - Use expectPresenceIndicatorDisplay() for online/offline status
  - Add profile card display on avatar click with user information
  - _Requirements: 3.1, 3.2, 3.4_

- [ ] 30. Implement security hardening and rate limiting
  - Write tests using newSecurityScenario(t) for comprehensive security validation
  - Add brute force protection for login attempts
  - Use expectRateLimitingEnforcement() for abuse prevention across all endpoints
  - Implement input sanitization and XSS prevention
  - _Requirements: 10.1, 10.2, 10.6_

- [ ] 31. Add data export and privacy controls
  - Write tests using expectUserDataExport() for GDPR compliance
  - Implement user data export and account deletion functionality
  - Use expectDataDeletionCompliance() for right to be forgotten
  - Add privacy settings and data retention policies
  - _Requirements: 10.4, 10.7_

**Result after Slice 8**: Production-ready user profile system with advanced features and security.