# Full Account Authentication & Admin Panel - Requirements

## Overview

Add full account creation with email/password authentication alongside existing guest access. Implement JWT-based authentication, role-based access control, and an admin panel for user management.

## Business Goals

1. Enable users to create persistent accounts with email/password
2. Maintain existing guest access for quick onboarding
3. Provide super admin with user management capabilities
4. Prepare foundation for multi-map feature (full accounts can create maps)
5. Improve security with proper authentication

## User Stories

### As a New User
- I want to create a full account with email/password so I can have persistent access
- I want to login with my credentials so I can access my account from any device
- I want to continue as a guest so I can try the app without commitment
- I want to see clear options (Signup/Login/Guest) so I know my choices

### As a Full Account User
- I want my session to persist across browser sessions so I don't have to login repeatedly
- I want to logout when I'm done so my account is secure
- I want to reset my password if I forget it (future enhancement)
- I want to upgrade from guest to full account (future enhancement)

### As a Super Admin
- I want to view all users so I can monitor the system
- I want to delete users so I can remove spam/abuse accounts
- I want to upgrade users to admin so I can delegate management
- I want to view system statistics so I can understand usage
- I want to manage maps and POIs (future enhancement)

### As a Guest User
- I want to continue using the app as before so my experience isn't disrupted
- I want to see that full accounts have additional features so I'm motivated to upgrade

## Functional Requirements

### FR1: User Account Types
- **FR1.1**: System supports two account types: `guest` and `full`
- **FR1.2**: Guest accounts have no email/password (existing behavior)
- **FR1.3**: Full accounts require email and password
- **FR1.4**: Email must be unique across all full accounts
- **FR1.5**: Display names must be unique across all accounts (existing)

### FR2: User Roles
- **FR2.1**: System supports three roles: `user`, `admin`, `superadmin`
- **FR2.2**: Default role for new accounts is `user`
- **FR2.3**: Only superadmin can create other superadmins
- **FR2.4**: Admins can upgrade users to admin role
- **FR2.5**: Users cannot change their own role

### FR3: Signup Flow
- **FR3.1**: User can create full account from welcome screen
- **FR3.2**: Signup requires: email, password, display name
- **FR3.3**: Signup optionally accepts: about me, avatar image
- **FR3.4**: Password must be minimum 8 characters
- **FR3.5**: Password must contain: uppercase, lowercase, number, special character
- **FR3.6**: Email must be valid format
- **FR3.7**: Display name must be 3-50 characters
- **FR3.8**: System validates email uniqueness before account creation
- **FR3.9**: Password is hashed with bcrypt (cost factor 12)
- **FR3.10**: Successful signup returns JWT token and user profile

### FR4: Login Flow
- **FR4.1**: User can login from welcome screen
- **FR4.2**: Login requires: email, password
- **FR4.3**: System validates credentials against database
- **FR4.4**: Successful login returns JWT token and user profile
- **FR4.5**: Failed login returns clear error message
- **FR4.6**: Account locks after 5 failed login attempts
- **FR4.7**: Locked account unlocks after 30 minutes
- **FR4.8**: User can request password reset (future enhancement)

### FR5: JWT Authentication
- **FR5.1**: JWT token contains: user ID, email, role, expiry
- **FR5.2**: JWT token expires after 24 hours
- **FR5.3**: JWT token is signed with secret key from environment
- **FR5.4**: System validates JWT signature and expiry on protected routes
- **FR5.5**: Expired tokens return 401 Unauthorized
- **FR5.6**: Invalid tokens return 401 Unauthorized
- **FR5.7**: Refresh token mechanism extends session (future enhancement)

### FR6: Authentication Middleware
- **FR6.1**: Protected routes require valid JWT token
- **FR6.2**: JWT token passed in `Authorization: Bearer <token>` header
- **FR6.3**: Middleware extracts user ID, email, role from token
- **FR6.4**: Middleware stores user info in request context
- **FR6.5**: Missing/invalid token returns 401 Unauthorized
- **FR6.6**: Some routes require full account (not guest)
- **FR6.7**: Some routes require admin/superadmin role

### FR7: Guest Access (Backward Compatibility)
- **FR7.1**: Guest account creation continues to work as before
- **FR7.2**: Guest users can access all existing features
- **FR7.3**: Guest users cannot create maps (future restriction)
- **FR7.4**: Guest users see option to upgrade to full account (future)
- **FR7.5**: Guest sessions work without JWT authentication

### FR8: Super Admin Account
- **FR8.1**: Super admin account created on first application startup
- **FR8.2**: Super admin credentials loaded from environment variables
- **FR8.3**: Super admin email: `SUPERADMIN_EMAIL` env var
- **FR8.4**: Super admin password: `SUPERADMIN_PASSWORD` env var
- **FR8.5**: Super admin cannot be deleted
- **FR8.6**: Super admin role cannot be changed
- **FR8.7**: Multiple super admins can exist (created by existing superadmin)

### FR9: Admin Panel - User Management
- **FR9.1**: Admin panel accessible at `/admin` route
- **FR9.2**: Only admin/superadmin can access admin panel
- **FR9.3**: Admin panel shows list of all users
- **FR9.4**: User list shows: display name, email, account type, role, created date
- **FR9.5**: User list supports pagination (20 users per page)
- **FR9.6**: User list supports search by name or email
- **FR9.7**: Admin can view user details (profile, sessions, POIs created)
- **FR9.8**: Admin can delete users (except superadmin)
- **FR9.9**: Admin can upgrade user role to admin
- **FR9.10**: Superadmin can upgrade user role to superadmin
- **FR9.11**: Deleting user also deletes their sessions and POIs

### FR10: Admin Panel - System Statistics
- **FR10.1**: Admin panel shows total user count
- **FR10.2**: Admin panel shows guest vs full account breakdown
- **FR10.3**: Admin panel shows total map count (future)
- **FR10.4**: Admin panel shows total POI count
- **FR10.5**: Admin panel shows active sessions count
- **FR10.6**: Statistics update in real-time

### FR11: Frontend State Management
- **FR11.1**: JWT token stored in localStorage
- **FR11.2**: User profile stored in localStorage
- **FR11.3**: Token validated on app load
- **FR11.4**: Expired token triggers re-login
- **FR11.5**: Logout clears token and profile from localStorage
- **FR11.6**: All API calls include JWT token in header (if authenticated)

### FR12: Welcome Screen Updates
- **FR12.1**: Welcome screen shows three options: Signup, Login, Guest
- **FR12.2**: "Create Full Account" button opens signup modal
- **FR12.3**: "Login" button opens login modal
- **FR12.4**: "Continue as Guest" button opens guest profile modal (existing)
- **FR12.5**: Options are clearly labeled and visually distinct

## Non-Functional Requirements

### NFR1: Security
- **NFR1.1**: Passwords hashed with bcrypt (cost factor 12)
- **NFR1.2**: JWT tokens signed with secure secret key (min 32 characters)
- **NFR1.3**: JWT secret stored in environment variable, never in code
- **NFR1.4**: Passwords never logged or exposed in API responses
- **NFR1.5**: Rate limiting on login attempts (5 per 15 minutes per IP)
- **NFR1.6**: Rate limiting on signup attempts (3 per hour per IP)
- **NFR1.7**: HTTPS required in production
- **NFR1.8**: SQL injection prevention via parameterized queries
- **NFR1.9**: XSS prevention via input sanitization

### NFR2: Performance
- **NFR2.1**: Login response time < 500ms
- **NFR2.2**: Signup response time < 1000ms
- **NFR2.3**: JWT validation < 50ms
- **NFR2.4**: Admin panel loads < 2 seconds
- **NFR2.5**: User list pagination loads < 500ms

### NFR3: Usability
- **NFR3.1**: Password strength indicator on signup
- **NFR3.2**: Clear error messages for validation failures
- **NFR3.3**: Show/hide password toggle
- **NFR3.4**: Remember me option (future enhancement)
- **NFR3.5**: Email verification (future enhancement)
- **NFR3.6**: Password reset flow (future enhancement)

### NFR4: Reliability
- **NFR4.1**: System handles database connection failures gracefully
- **NFR4.2**: System handles JWT validation errors gracefully
- **NFR4.3**: Failed operations don't leave partial data
- **NFR4.4**: Transaction rollback on errors

### NFR5: Maintainability
- **NFR5.1**: Authentication logic isolated in auth service
- **NFR5.2**: Middleware reusable across routes
- **NFR5.3**: Clear separation of concerns (service/handler/repository)
- **NFR5.4**: Comprehensive unit tests for auth logic
- **NFR5.5**: Integration tests for auth flows

### NFR6: Scalability
- **NFR6.1**: JWT stateless authentication (no server-side sessions)
- **NFR6.2**: Token validation doesn't require database lookup
- **NFR6.3**: Admin panel pagination supports large user counts
- **NFR6.4**: Rate limiting uses Redis for distributed systems

## Out of Scope (Future Enhancements)

1. Email verification on signup
2. Password reset via email
3. Two-factor authentication (2FA)
4. OAuth/Social login (Google, GitHub, etc.)
5. Upgrade guest to full account
6. Session management (view/revoke active sessions)
7. Audit logs (track admin actions)
8. User activity tracking
9. Account suspension (temporary disable)
10. Bulk user operations (bulk delete, bulk role change)
11. Admin panel - Map management
12. Admin panel - POI management
13. User profile privacy settings
14. Account deletion by user (self-service)

## Success Criteria

1. ✅ Users can create full accounts with email/password
2. ✅ Users can login with credentials and receive JWT token
3. ✅ Guest access continues to work without disruption
4. ✅ JWT authentication protects API routes
5. ✅ Super admin account created on first startup
6. ✅ Admin panel accessible to admin/superadmin only
7. ✅ Admin can view, delete, and upgrade users
8. ✅ All tests pass (unit + integration)
9. ✅ No security vulnerabilities in authentication flow
10. ✅ Performance meets NFR requirements

## Acceptance Criteria

### AC1: Signup Flow
- Given I am on the welcome screen
- When I click "Create Full Account"
- Then I see a signup modal with email, password, display name fields
- When I fill valid data and submit
- Then my account is created and I'm logged in with JWT token
- And I can access the map

### AC2: Login Flow
- Given I have a full account
- When I click "Login" on welcome screen
- Then I see a login modal with email and password fields
- When I enter correct credentials and submit
- Then I receive JWT token and am logged in
- And I can access the map

### AC3: Guest Flow
- Given I am on the welcome screen
- When I click "Continue as Guest"
- Then I see the existing guest profile modal
- When I create guest profile
- Then I can access the map without authentication

### AC4: Admin Panel Access
- Given I am logged in as super admin
- When I navigate to `/admin`
- Then I see the admin dashboard with user list
- And I can view, delete, and upgrade users

### AC5: Protected Routes
- Given I am not authenticated
- When I try to access a protected route
- Then I receive 401 Unauthorized
- And I'm redirected to welcome screen

### AC6: Token Expiry
- Given I am logged in with JWT token
- When my token expires (24 hours)
- Then my next API call returns 401 Unauthorized
- And I'm prompted to login again

## Dependencies

### Backend Dependencies
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/golang-jwt/jwt/v5` - JWT token generation/validation

### Frontend Dependencies
- None (use existing React, Zustand, etc.)

### Environment Variables
- `JWT_SECRET` - Secret key for JWT signing (required)
- `JWT_EXPIRY` - Token expiry duration (default: 24h)
- `SUPERADMIN_EMAIL` - Super admin email (required)
- `SUPERADMIN_PASSWORD` - Super admin password (required)

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| JWT secret leaked | High | Low | Store in env var, rotate regularly, never commit to git |
| Password brute force | High | Medium | Rate limiting, account lockout, strong password requirements |
| XSS attack | High | Low | Input sanitization, CSP headers, React auto-escaping |
| SQL injection | High | Low | Parameterized queries, ORM usage (GORM) |
| Token theft | Medium | Low | HTTPS only, short expiry, refresh token rotation |
| Admin abuse | Medium | Low | Audit logs (future), limit admin count, monitor actions |
| Guest to full migration | Low | Medium | Design upgrade path early, test thoroughly |

## Assumptions

1. Users have valid email addresses
2. Users can remember passwords (no reset flow in MVP)
3. Super admin credentials are securely managed
4. HTTPS is available in production
5. Redis is available for rate limiting
6. Database supports transactions
7. Frontend supports localStorage
8. Users understand difference between guest and full accounts

## Constraints

1. Must maintain backward compatibility with guest accounts
2. Must not break existing features
3. Must follow TDD methodology
4. Must pass all existing tests
5. Must meet security NFRs
6. Must be deployable to Railway
7. Must work with existing PostgreSQL database
8. Must work with existing Redis instance

## Open Questions

1. Should we require email verification on signup? (Decision: No for MVP)
2. Should we implement password reset flow? (Decision: No for MVP)
3. Should guests be able to upgrade to full accounts? (Decision: Future enhancement)
4. Should we implement refresh tokens? (Decision: Future enhancement)
5. Should we track login history? (Decision: Future enhancement)
6. Should we implement 2FA? (Decision: Future enhancement)
7. How long should account lockout last? (Decision: 30 minutes)
8. Should we implement "remember me"? (Decision: Future enhancement)
