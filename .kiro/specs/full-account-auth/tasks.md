# Implementation Tasks - Full Account Authentication & Admin Panel

## Overview
This task list implements full account authentication with JWT, role-based access control, and admin panel for user management. All tasks follow TDD methodology with tests written before implementation.

---

## Phase 1: Backend Authentication Infrastructure

### 1. Install Dependencies and Setup Environment
- [x] 1.1 Install Go dependencies for authentication
  - Install `golang.org/x/crypto/bcrypt` for password hashing
  - Install `github.com/golang-jwt/jwt/v5` for JWT token handling
  - Update `go.mod` and `go.sum` files
  - _Requirements: FR5.1, FR5.3, FR3.9_

- [x] 1.2 Add environment variables for authentication
  - Add `JWT_SECRET` to `.env` file (generate secure random string)
  - Add `JWT_EXPIRY` to `.env` file (default: "24h")
  - Add `SUPERADMIN_EMAIL` to `.env` file
  - Add `SUPERADMIN_PASSWORD` to `.env` file
  - Update `backend/internal/config/config.go` to load new env vars
  - _Requirements: FR8.3, FR8.4, FR5.3_

### 2. Create Authentication Service
- [x] 2.1 Create auth service with password hashing
  - Create `backend/internal/services/auth_service.go`
  - Implement `HashPassword(password string) (string, error)` method
  - Implement `VerifyPassword(password, hash string) error` method
  - Use bcrypt with cost factor 12
  - _Requirements: FR3.9, FR4.3_

- [x] 2.3 Implement JWT token generation
  - Add `GenerateJWT(userID, email string, role UserRole) (string, time.Time, error)` to auth service
  - JWT claims include: user ID, email, role, issued at, expiry
  - Sign JWT with secret from environment
  - Return token and expiry time
  - _Requirements: FR5.1, FR5.2, FR5.3_

- [x] 2.5 Implement JWT token validation
  - Add `ValidateJWT(token string) (*JWTClaims, error)` to auth service
  - Verify JWT signature with secret
  - Check JWT expiry
  - Extract and return claims
  - _Requirements: FR5.4, FR5.5, FR5.6_

### 3. Update User Repository
- [x] 3.1 Add email lookup method to user repository
  - Add `GetByEmail(ctx context.Context, email string) (*models.User, error)` to `backend/internal/repository/user_repository.go`
  - Query users table by email field
  - Return user or error if not found
  - _Requirements: FR4.3, FR3.8_

### 4. Update User Service
- [x] 4.1 Implement full account creation
  - Add `CreateFullAccount(ctx, email, password, displayName, aboutMe string) (*models.User, error)` to `backend/internal/services/user_service.go`
  - Validate email format and uniqueness
  - Validate password strength (min 8 chars, uppercase, lowercase, number, special)
  - Hash password with auth service
  - Create user with `AccountTypeFull`
  - Save to repository
  - _Requirements: FR3.1, FR3.2, FR3.3, FR3.4, FR3.5, FR3.6, FR3.8, FR3.9_

- [x] 4.3 Add password validation helper
  - Create `ValidatePassword(password string) error` function
  - Check minimum 8 characters
  - Check contains uppercase letter
  - Check contains lowercase letter
  - Check contains number
  - Check contains special character
  - _Requirements: FR3.4, FR3.5_

### 5. Create Authentication Handler
- [x] 5.1 Create auth handler with signup endpoint
  - Create `backend/internal/handlers/auth_handler.go`
  - Implement `Signup(c *gin.Context)` handler
  - Parse signup request (email, password, displayName, aboutMe)
  - Validate request data
  - Call user service to create full account
  - Generate JWT token
  - Return token and user profile
  - _Requirements: FR3.1-FR3.10_

- [x] 5.3 Implement login endpoint
  - Implement `Login(c *gin.Context)` handler
  - Parse login request (email, password)
  - Get user by email from service
  - Verify password with auth service
  - Generate JWT token
  - Return token and user profile
  - _Requirements: FR4.1, FR4.2, FR4.3, FR4.4_

- [x] 5.5 Implement logout endpoint
  - Implement `Logout(c *gin.Context)` handler
  - Return success response (JWT is stateless, client clears token)
  - _Requirements: FR11.5_

- [x] 5.6 Implement get current user endpoint
  - Implement `GetCurrentUser(c *gin.Context)` handler
  - Extract user ID from JWT claims (set by middleware)
  - Get user from service
  - Return user profile
  - _Requirements: FR6.4_

### 6. Create Authentication Middleware
- [x] 6.1 Create RequireAuth middleware
  - Create `backend/internal/middleware/auth.go`
  - Implement `RequireAuth() gin.HandlerFunc`
  - Extract JWT from `Authorization: Bearer <token>` header
  - Validate JWT with auth service
  - Extract user ID, email, role from claims
  - Store in Gin context: `c.Set("userID", userID)`, etc.
  - Return 401 if token missing or invalid
  - _Requirements: FR6.1, FR6.2, FR6.3, FR6.4, FR6.5_

- [x] 6.3 Create RequireFullAccount middleware
  - Implement `RequireFullAccount() gin.HandlerFunc`
  - Use RequireAuth first
  - Check account type is `full` (not `guest`)
  - Return 403 if guest account
  - _Requirements: FR6.6_

- [x] 6.5 Create RequireAdmin middleware
  - Implement `RequireAdmin() gin.HandlerFunc`
  - Use RequireAuth first
  - Check role is `admin` or `superadmin`
  - Return 403 if not admin
  - _Requirements: FR6.7, FR9.2_

- [x] 6.7 Create RequireSuperAdmin middleware
  - Implement `RequireSuperAdmin() gin.HandlerFunc`
  - Use RequireAuth first
  - Check role is `superadmin`
  - Return 403 if not superadmin
  - _Requirements: FR6.7, FR9.10_

### 7. Register Authentication Routes
- [x] 7.1 Register auth routes in server
  - Update `backend/internal/server/server.go`
  - Create auth handler instance
  - Register POST `/api/auth/signup`
  - Register POST `/api/auth/login`
  - Register POST `/api/auth/logout`
  - Register GET `/api/auth/me` with RequireAuth middleware
  - _Requirements: FR3.1, FR4.1_

- [x] 7.2 Apply authentication middleware to existing routes
  - Keep guest profile creation public (no middleware)
  - Keep POI read operations public
  - Apply RequireAuth to POI create/update/delete
  - Apply RequireAuth to session operations
  - Apply RequireAuth to profile updates
  - _Requirements: FR6.1, FR7.1, FR7.2_

### 8. Create Super Admin Account
- [x] 8.1 Implement super admin creation in migrations
  - Update `backend/internal/database/migrations.go`
  - Add `CreateSuperAdminIfNotExists(db *gorm.DB) error` function
  - Check if superadmin exists (query by role)
  - If not exists, create from environment variables
  - Hash password with bcrypt
  - Set role to `superadmin`, account type to `full`
  - Call in `RunMigrations()` after other migrations
  - _Requirements: FR8.1, FR8.2, FR8.3, FR8.4, FR8.5, FR8.6_


---

## Phase 2: Frontend Authentication UI ✅ COMPLETE

**Progress Summary:** Frontend authentication complete with 50 passing tests
- WelcomeScreen: 4 tests
- SignupModal: 19 tests (13 validation + 6 submission)
- LoginModal: 19 tests (13 validation + 6 submission)
- App Integration: 8 tests (modal state + auth flow)

### 9. Create Authentication Store
- [x] 9.1 Create auth store with Zustand
  - Create `frontend/src/stores/authStore.ts`
  - State: `token`, `user`, `isAuthenticated`, `isLoading`
  - Action: `login(email, password)`
  - Action: `signup(signupData)`
  - Action: `logout()`
  - Action: `setUser(user)`
  - Action: `loadAuthFromStorage()`
  - Store token and user in localStorage
  - _Requirements: FR11.1, FR11.2, FR11.3_

### 10. Update API Service
- [x] 10.1 Add authentication API functions
  - Update `frontend/src/services/api.ts`
  - Add `signup(data: SignupRequest): Promise<AuthResponse>`
  - Add `login(email: string, password: string): Promise<AuthResponse>`
  - Add `logout(): Promise<void>`
  - Add `getCurrentUser(): Promise<UserProfile>`
  - _Requirements: FR3.10, FR4.4_

- [x] 10.2 Update API calls to include JWT token
  - Create helper to get token from auth store
  - Add `Authorization: Bearer <token>` header to all API calls
  - Handle 401 responses (token expired, redirect to login)
  - Remove `X-User-ID` header usage
  - _Requirements: FR6.2, FR11.6_

### 11. Update Welcome Screen
- [x] 11.1 Add three action buttons to WelcomeScreen
  - Update `frontend/src/components/WelcomeScreen.tsx`
  - Add "Create Full Account" button (calls `onSignup` prop)
  - Add "Login" button (calls `onLogin` prop)
  - Keep "Continue as Guest" button (calls `onCreateProfile`)
  - Style buttons to be visually distinct
  - Tests: 4 tests passing
  - _Requirements: FR12.1, FR12.2, FR12.3, FR12.4, FR12.5_

### 12. Create Signup Modal
- [x] 12.1 Create signup modal component
  - Create `frontend/src/components/SignupModal.tsx`
  - Form fields: email, password, confirm password, display name, about me
  - Show/hide password toggle
  - "Already have an account? Login" link
  - Tests: 13 tests passing (5 rendering, 4 validation, 2 submission, 2 interactions)
  - _Requirements: FR3.2, FR3.3, NFR3.1, NFR3.3_

- [x] 12.2 Implement signup form validation
  - Email format validation (real-time)
  - Password strength validation (real-time)
  - Confirm password match validation
  - Display name length validation (3-50 chars)
  - About me length validation (max 500 chars)
  - Show validation errors inline
  - _Requirements: FR3.4, FR3.5, FR3.6, FR3.7, NFR3.2_

- [x] 12.3 Implement signup submission
  - Call auth store `signup()` action
  - Show loading state during submission
  - Handle success: close modal
  - Handle errors: display error message
  - Handle network errors gracefully
  - Tests: 6 tests passing (2 loading, 3 error handling, 1 success)
  - _Requirements: FR3.10, NFR3.2_

### 13. Create Login Modal
- [x] 13.1 Create login modal component
  - Create `frontend/src/components/LoginModal.tsx`
  - Form fields: email, password
  - Show/hide password toggle
  - "Don't have an account? Sign up" link
  - Tests: 13 tests passing (6 rendering, 2 validation, 2 submission, 3 interactions)
  - _Requirements: FR4.1, FR4.2, NFR3.3_

- [x] 13.2 Implement login form validation
  - Email format validation
  - Password required validation
  - Show validation errors inline
  - _Requirements: FR4.2, NFR3.2_

- [x] 13.3 Implement login submission
  - Call auth store `login()` action
  - Show loading state during submission
  - Handle success: close modal
  - Handle errors: display clear error message
  - Tests: 6 tests passing (2 loading, 3 error handling, 1 success)
  - _Requirements: FR4.3, FR4.4, FR4.5, NFR3.2_

### 14. Update App Component
- [x] 14.1 Add authentication state management to App
  - Update `frontend/src/App.tsx`
  - Add state for `showSignup`, `showLogin`
  - Import authStore
  - Tests: 8 integration tests passing
  - _Requirements: FR11.3, FR11.4_

- [x] 14.2 Integrate signup and login modals
  - Render SignupModal with `isOpen={showSignup}`
  - Render LoginModal with `isOpen={showLogin}`
  - Handle modal open/close state
  - Handle switching between signup and login
  - Handle successful auth (close modals, proceed to map)
  - Tests: Covered by integration tests
  - _Requirements: FR12.2, FR12.3_

- [x] 14.3 Update welcome screen integration
  - Pass `onSignup` prop to open signup modal
  - Pass `onLogin` prop to open login modal
  - Keep existing `onCreateProfile` for guest flow
  - Tests: Covered by WelcomeScreen tests
  - _Requirements: FR12.1, FR12.2, FR12.3, FR12.4_

---

## Phase 3: Admin Panel

### 15. Create Admin Handler
- [ ] 15.1 Create admin handler with user list endpoint
  - Create `backend/internal/handlers/admin_handler.go`
  - Implement `ListUsers(c *gin.Context)` handler
  - Support pagination (page, pageSize query params)
  - Support search (query param for name/email)
  - Return user list with total count
  - Apply RequireAdmin middleware
  - _Requirements: FR9.3, FR9.4, FR9.5, FR9.6_

- [ ] 15.3 Implement delete user endpoint
  - Implement `DeleteUser(c *gin.Context)` handler
  - Get user ID from URL param
  - Prevent deleting superadmin
  - Delete user's sessions and POIs (cascade)
  - Delete user from database
  - Apply RequireAdmin middleware
  - _Requirements: FR9.8, FR9.11, FR8.5_

- [ ] 15.5 Implement update user role endpoint
  - Implement `UpdateUserRole(c *gin.Context)` handler
  - Get user ID from URL param
  - Get new role from request body
  - Validate role change permissions (admin can't create superadmin)
  - Prevent changing own role
  - Update user role in database
  - Apply RequireAdmin middleware
  - _Requirements: FR9.9, FR9.10, FR2.3, FR2.5_

- [ ] 15.7 Implement system statistics endpoint
  - Implement `GetSystemStats(c *gin.Context)` handler
  - Count total users
  - Count guest vs full accounts
  - Count total POIs
  - Count active sessions
  - Return statistics object
  - Apply RequireAdmin middleware
  - _Requirements: FR10.1, FR10.2, FR10.4, FR10.5_

### 16. Register Admin Routes
- [ ] 16.1 Register admin routes in server
  - Update `backend/internal/server/server.go`
  - Create admin handler instance
  - Register GET `/api/admin/users` with RequireAdmin
  - Register DELETE `/api/admin/users/:id` with RequireAdmin
  - Register PUT `/api/admin/users/:id/role` with RequireAdmin
  - Register GET `/api/admin/stats` with RequireAdmin
  - _Requirements: FR9.1, FR9.2_

### 17. Create Admin Dashboard Component
- [ ] 17.1 Create admin dashboard layout
  - Create `frontend/src/components/AdminDashboard.tsx`
  - Add navigation tabs: Users, Statistics
  - Add header with "Admin Panel" title
  - Add logout button
  - Only render if user is admin/superadmin
  - _Requirements: FR9.1, FR9.2_

- [ ] 17.2 Implement user list view
  - Display user table with columns: name, email, account type, role, created date
  - Add pagination controls (prev/next, page numbers)
  - Add search input (filter by name or email)
  - Add action buttons per user: Delete, Change Role
  - Disable delete for superadmin users
  - _Requirements: FR9.3, FR9.4, FR9.5, FR9.6, FR9.8, FR9.9_

- [ ] 17.3 Implement delete user functionality
  - Add delete button with confirmation dialog
  - Call admin API to delete user
  - Refresh user list after deletion
  - Show success/error toast
  - Disable for superadmin users
  - _Requirements: FR9.8, FR8.5_

- [ ] 17.4 Implement change role functionality
  - Add "Change Role" dropdown per user
  - Options: user, admin, superadmin (if current user is superadmin)
  - Call admin API to update role
  - Refresh user list after update
  - Show success/error toast
  - _Requirements: FR9.9, FR9.10_

- [ ] 17.5 Implement statistics view
  - Display system statistics cards
  - Show total users count
  - Show guest vs full account breakdown (pie chart or bars)
  - Show total POIs count
  - Show active sessions count
  - Auto-refresh statistics every 30 seconds
  - _Requirements: FR10.1, FR10.2, FR10.4, FR10.5, FR10.6_

### 18. Add Admin Panel Routing
- [ ] 18.1 Add admin route to App
  - Update `frontend/src/App.tsx` or create router
  - Add route `/admin` that renders AdminDashboard
  - Protect route: redirect to home if not admin
  - Add link to admin panel in profile menu (if admin)
  - _Requirements: FR9.1, FR9.2_

---

## Notes

- All tasks marked with `*` are optional testing tasks that can be skipped for MVP
- Each task should be completed in order within its section
- Run all tests after completing each major section
- Follow TDD: write tests before implementation for all non-optional tasks
- Ensure backward compatibility with guest accounts throughout
- Test on both development and production environments before deployment
