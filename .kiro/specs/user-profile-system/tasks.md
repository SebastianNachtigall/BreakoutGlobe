# User Profile & Account System Implementation Plan

## Overview

This implementation plan transforms the current anonymous session system into a comprehensive user profile and account system with guest profiles and full accounts, including role-based permissions and real-time avatar display.

## Implementation Tasks

- [x] 1. Backend User Model and Database Schema
  - [x] Create User model with profile fields (displayName, email, avatarURL, aboutMe, accountType, role)
  - [x] Write tests for User model validation and constraints
  - [x] Create Map model for multi-map support with proper relationships
  - [x] Create database migration for users and maps tables with proper indexes
  - [x] Update Session model to reference both User and Map for proper isolation
  - [x] Update POI model to reference Map and Creator (User) for ownership tracking
  - [x] Write tests for User-Session-Map relationships and foreign key constraints
  - [x] Update all services and handlers to work with UUID-based models
  - [ ] Fix all existing test files to work with UUID-based models
  - _Requirements: 1.1, 2.1, 7.1, 8.1, 9.1_

- [ ] 2. User Repository and Service Layer
  - [ ] 2.1 Implement User Repository with CRUD operations
    - Write tests for user creation, retrieval, update, and deletion
    - Implement UserRepository interface with database operations
    - Write tests for email uniqueness validation and conflict handling
    - Add user search and filtering capabilities (by role, account type, active status)
    - _Requirements: 1.1, 2.1, 6.1_

  - [ ] 2.2 Create User Service with business logic
    - Write tests for guest profile creation workflow
    - Implement guest profile creation with localStorage backup sync
    - Write tests for full account upgrade workflow with email verification
    - Implement password hashing and validation using bcrypt
    - Write tests for profile update operations with role-based permissions
    - _Requirements: 1.1, 2.1, 5.1, 8.1_

  - [ ] 2.3 Implement Avatar Upload and Management
    - Write tests for avatar image upload with file validation
    - Implement avatar storage (filesystem for dev, S3 for production)
    - Write tests for image processing (resize, format conversion, optimization)
    - Add avatar URL generation and CDN integration
    - Write tests for avatar deletion and cleanup
    - _Requirements: 1.1, 3.1, 8.3_

- [ ] 3. Authentication and Authorization System
  - [ ] 3.1 Implement Authentication Middleware
    - Write tests for JWT token generation and validation
    - Implement login/logout endpoints with session management
    - Write tests for password reset workflow with secure tokens
    - Add email verification system with token-based confirmation
    - Write tests for session expiration and refresh token handling
    - _Requirements: 2.1, 4.1, 8.1_

  - [ ] 3.2 Create Role-Based Access Control
    - Write tests for role-based permission checking (user, admin, superadmin)
    - Implement authorization middleware for API endpoints
    - Write tests for hierarchical permission enforcement
    - Add role assignment and validation logic
    - Write tests for permission boundary enforcement (admins cannot modify superadmins)
    - _Requirements: 4.1, 6.1, 8.2_

- [ ] 4. User Management API Endpoints
  - [ ] 4.1 Profile Management Endpoints
    - Write tests for POST /api/users/profile (guest profile creation)
    - Implement guest profile creation with validation and localStorage sync
    - Write tests for POST /api/users/account (upgrade to full account)
    - Implement full account upgrade with email verification
    - Write tests for GET /api/users/profile (get current user profile)
    - Write tests for PUT /api/users/profile (update profile with role restrictions)
    - _Requirements: 1.1, 2.1, 5.1_

  - [ ] 4.4 Map Management Endpoints
    - Write tests for GET /api/maps (list available maps)
    - Implement map listing with user access control
    - Write tests for GET /api/maps/:id (get map details)
    - Write tests for POST /api/maps (create map - admin/superadmin only)
    - Implement map creation with proper ownership and permissions
    - Write tests for GET /api/maps/:id/users (get users active on specific map)
    - _Requirements: 8.1_

  - [ ] 4.2 Authentication Endpoints
    - Write tests for POST /api/auth/login with email/password validation
    - Implement login endpoint with rate limiting and brute force protection
    - Write tests for POST /api/auth/logout with session invalidation
    - Write tests for POST /api/users/verify-email with token validation
    - Implement password reset endpoints with secure token generation
    - _Requirements: 2.1, 4.1, 8.1_

  - [ ] 4.3 Avatar Management Endpoints
    - Write tests for POST /api/users/avatar with file upload validation
    - Implement avatar upload with image processing and storage
    - Write tests for DELETE /api/users/avatar with cleanup
    - Add avatar serving endpoint with CDN integration
    - Write tests for avatar URL generation and access control
    - _Requirements: 1.1, 3.1, 8.3_

- [ ] 5. Enhanced POI Management with Permissions
  - [ ] 5.1 POI Creation and Ownership System
    - Write tests for POI creation permissions (full account required)
    - Implement POI creation with user ownership tracking and map association
    - Write tests for POI ownership validation and edit permissions
    - Update existing POI endpoints to include map context and creator information
    - Write tests for POI deletion permissions (owner or admin)
    - Add POI moderation capabilities for admin users
    - _Requirements: 9.1_

  - [ ] 5.2 Map-Specific POI Management
    - Write tests for GET /api/maps/:mapId/pois (POIs filtered by map)
    - Update POI listing to be map-specific with proper isolation
    - Write tests for POST /api/maps/:mapId/pois (create POI on specific map)
    - Implement map-scoped POI operations with permission checking
    - Write tests for POI visibility and access control per map
    - Add POI transfer capabilities between maps for admins
    - _Requirements: 8.1, 9.1_

- [ ] 6. Admin User Management System
  - [ ] 6.1 User Management Endpoints (Admin/Superadmin)
    - Write tests for GET /api/admin/users with pagination and filtering (optionally by map)
    - Implement user listing with role-based access control and map context
    - Write tests for GET /api/admin/users/:id with detailed user information
    - Write tests for PUT /api/admin/users/:id/role with permission validation
    - Implement role assignment with hierarchical permission checks
    - Write tests for PUT /api/admin/users/:id/status (enable/disable users)
    - _Requirements: 6.1, 8.1_

  - [ ] 6.2 Admin User Interface Components
    - Write tests for UserManagementPanel component with user listing
    - Implement user management UI with search, filter, and pagination
    - Write tests for UserDetailsModal with profile information display
    - Implement user role assignment interface with permission validation
    - Write tests for user status management (enable/disable accounts)
    - Add audit logging display for user management actions
    - _Requirements: 6.1_

- [ ] 7. Frontend Profile Management System
  - [ ] 7.1 Profile Creation and Onboarding
    - Write tests for ProfileCreationModal with form validation
    - Implement guest profile creation form with name, avatar, and about me fields
    - Write tests for AccountUpgradeModal with email/password validation
    - Implement full account upgrade flow with email verification
    - Write tests for profile creation flow with localStorage and backend sync
    - Add onboarding wizard for new users accessing maps
    - _Requirements: 1.1, 2.1, 4.1_

  - [ ] 7.2 Profile Management Interface
    - Write tests for ProfileSettingsPanel with account type-specific options
    - Implement profile editing with role-based field restrictions
    - Write tests for AvatarUpload component with image preview and validation
    - Implement avatar upload with progress indication and error handling
    - Write tests for password change functionality for full accounts
    - Add email change workflow with verification for full accounts
    - _Requirements: 5.1, 8.1_

- [ ] 8. Enhanced Avatar Display System
  - [ ] 8.1 Real-time Avatar Rendering
    - Write tests for enhanced AvatarData interface with user profile information
    - Update MapContainer to display user avatars with profile images
    - Write tests for avatar image loading with fallback to initials
    - Implement avatar hover tooltips with user name and role
    - Write tests for avatar click interaction showing user profile card
    - Add real-time avatar updates when users change their profile images
    - _Requirements: 3.1, 3.2_

  - [ ] 8.2 Multi-User Avatar Management with Map Isolation
    - Write tests for multiple user avatar display and positioning with map isolation
    - Implement real-time avatar synchronization via WebSocket filtered by current map
    - Write tests for avatar collision detection and positioning optimization
    - Add user presence indicators (online/offline status) scoped to current map
    - Write tests for avatar animation and smooth movement transitions
    - Implement avatar clustering for crowded areas with map-specific grouping
    - _Requirements: 3.1, 7.1, 8.1_

- [ ] 9. Frontend State Management Enhancement
  - [ ] 9.1 User Profile Store
    - Write tests for userProfileStore with profile state management
    - Implement profile store with localStorage and backend synchronization
    - Write tests for profile update operations with optimistic updates
    - Add authentication state management (login/logout/session)
    - Write tests for role-based UI state management
    - Implement profile data persistence and rehydration
    - _Requirements: 1.1, 2.1, 7.1_

  - [ ] 9.2 Enhanced Session Store with Map Context
    - Write tests for updated sessionStore with user profile and map context integration
    - Update session management to work with user profiles and map isolation
    - Write tests for multi-user session tracking and avatar display filtered by map
    - Implement real-time user presence and activity tracking scoped to current map
    - Write tests for session synchronization across browser tabs with map context
    - Add offline profile management with sync queue and map state preservation
    - _Requirements: 3.1, 4.1, 7.1, 8.1_

- [ ] 10. WebSocket Real-time Profile Updates
  - [ ] 10.1 Profile Update Broadcasting
    - Write tests for profile update WebSocket events (profile_updated, avatar_changed)
    - Implement real-time profile change broadcasting to connected users
    - Write tests for user join/leave events with profile information
    - Add real-time avatar image updates across all connected clients
    - Write tests for role change notifications and UI updates
    - Implement user status change broadcasting (online/offline, active/inactive)
    - _Requirements: 3.1, 7.1_

  - [ ] 10.2 Enhanced WebSocket Client
    - Write tests for WebSocket client handling of profile-related events
    - Update WebSocket client to process user profile updates
    - Write tests for avatar synchronization and real-time display updates
    - Implement connection state management with user authentication
    - Write tests for reconnection handling with profile state restoration
    - Add WebSocket message queuing for offline profile changes
    - _Requirements: 3.1, 4.1, 7.1_

- [ ] 11. Security and Data Protection
  - [ ] 11.1 Input Validation and Sanitization
    - Write tests for comprehensive input validation on all profile fields
    - Implement server-side validation for display names, emails, and about me text
    - Write tests for avatar upload security (file type, size, malware scanning)
    - Add XSS protection for user-generated content display
    - Write tests for SQL injection prevention in user queries
    - Implement rate limiting for profile creation and update operations
    - _Requirements: 8.1, 8.3_

  - [ ] 11.2 Data Privacy and GDPR Compliance
    - Write tests for user data export functionality
    - Implement complete user data deletion (right to be forgotten)
    - Write tests for data retention policies and automated cleanup
    - Add privacy controls for profile visibility and data sharing
    - Write tests for audit logging of sensitive operations
    - Implement consent management for data processing
    - _Requirements: 8.1, 8.4_

- [ ] 12. Integration Testing and System Validation
  - [ ] 12.1 End-to-End User Flows
    - Write E2E tests for complete new user onboarding (guest profile creation)
    - Test full account upgrade workflow with email verification
    - Write E2E tests for profile management across different account types
    - Test multi-user avatar display and real-time synchronization
    - Write E2E tests for admin user management workflows
    - Test cross-device profile access and synchronization for full accounts
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 6.1_

  - [ ] 12.2 Performance and Load Testing
    - Write load tests for concurrent user profile creation and updates
    - Test avatar upload and serving performance under load
    - Write performance tests for real-time avatar synchronization with many users
    - Test database performance with large numbers of user profiles
    - Write tests for WebSocket performance with profile update broadcasting
    - Add monitoring for profile-related API response times and resource usage
    - _Requirements: 3.1, 7.1_

- [ ] 13. Production Deployment and Monitoring
  - [ ] 13.1 Production Configuration
    - Configure avatar storage with cloud provider (AWS S3, CloudFront CDN)
    - Set up email service for verification and password reset (SendGrid, SES)
    - Write tests for production environment configuration and health checks
    - Configure database indexes and performance optimization for user queries
    - Set up monitoring and alerting for user management system
    - Add backup and disaster recovery procedures for user data
    - _Requirements: 8.1, 8.4_

  - [ ] 13.2 Security Hardening
    - Implement comprehensive security headers and HTTPS enforcement
    - Configure rate limiting and DDoS protection for user endpoints
    - Write security tests for authentication and authorization systems
    - Set up intrusion detection and security monitoring
    - Implement secure session management and token handling
    - Add security audit logging and compliance reporting
    - _Requirements: 8.1, 8.2, 8.3_