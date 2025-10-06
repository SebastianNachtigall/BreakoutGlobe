# Full Account Creation & Authentication Implementation Plan

## Executive Summary

This document outlines the requirements and implementation strategy for adding full account creation with email/password authentication to BreakoutGlobe. Currently, only guest accounts can be created. The goal is to add login/signup options alongside guest access, and create a super admin account for system management.

## Current State Analysis

### Existing Infrastructure (Already in Place!)

**User Model (`backend/internal/models/user.go`):**
- ✅ `AccountType` enum: `guest` and `full` already defined
- ✅ `UserRole` enum: `user`, `admin`, `superadmin` already defined
- ✅ `Email` field with unique index (nullable for guests)
- ✅ `PasswordHash` field (hidden from JSON, nullable)
- ✅ Helper methods: `IsGuest()`, `IsFull()`, `IsAdmin()`, `IsSuperAdmin()`, `HasPassword()`
- ✅ Validation: Email validation already implemented
- ✅ Permission checks: `CanBeModifiedBy()`, `CanBeAccessedBy()` patterns exist

**Current Flow:**
1. User visits app → WelcomeScreen → ProfileCreationModal
2. Only guest accounts can be created (no email/password)
3. User ID stored in localStorage
4. No authentication/session management
5. `X-User-ID` header used for API calls (insecure, temporary)

**What's Missing:**
- ❌ Password hashing (bcrypt)
- ❌ Email/password signup endpoint
- ❌ Login endpoint
- ❌ JWT token generation/validation
- ❌ Authentication middleware
- ❌ Session management (JWT-based)
- ❌ Password reset flow
- ❌ Super admin account creation
- ❌ Frontend login/signup UI

## Implementation Requirements

### Phase 1: Backend Authentication Infrastructure

#### 1.1 Password Hashing Service

**New File: `backend/internal/services/auth_service.go`**

```go
package services

import (
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    jwtSecret []byte
    tokenExpiry time.Duration
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error)

// VerifyPassword verifies a password against a hash
func (s *AuthService) VerifyPassword(password, hash string) error

// GenerateJWT generates a JWT token for a user
func (s *AuthService) GenerateJWT(userID string, email string, role UserRole) (string, error)

// ValidateJWT validates a JWT token and returns claims
func (s *AuthService) ValidateJWT(token string) (*JWTClaims, error)

// GenerateRefreshToken generates a refresh token
func (s *AuthService) GenerateRefreshToken(userID string) (string, error)
```

**Dependencies:**
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/golang-jwt/jwt/v5` - JWT token generation

**Business Rules:**
- Minimum password length: 8 characters
- Password must contain: uppercase, lowercase, number, special char
- Bcrypt cost factor: 12 (secure but not too slow)
- JWT expiry: 24 hours (configurable)
- Refresh token expiry: 30 days

#### 1.2 Authentication Endpoints

**Update: `backend/internal/handlers/auth_handler.go`** (new file)

```
POST   /api/auth/signup          - Create full account with email/password
POST   /api/auth/login           - Login with email/password
POST   /api/auth/refresh         - Refresh JWT token
POST   /api/auth/logout          - Logout (invalidate token)
POST   /api/auth/forgot-password - Request password reset
POST   /api/auth/reset-password  - Reset password with token
GET    /api/auth/me              - Get current authenticated user
```

**Request/Response DTOs:**

```go
// SignupRequest
type SignupRequest struct {
    Email       string `json:"email" binding:"required,email"`
    Password    string `json:"password" binding:"required,min=8"`
    DisplayName string `json:"displayName" binding:"required,min=3,max=50"`
    AboutMe     string `json:"aboutMe,omitempty"`
}

// LoginRequest
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// AuthResponse
type AuthResponse struct {
    Token        string      `json:"token"`
    RefreshToken string      `json:"refreshToken"`
    User         UserProfile `json:"user"`
    ExpiresAt    time.Time   `json:"expiresAt"`
}
```

**Business Rules:**
- Email must be unique (database constraint already exists)
- Password validation on signup
- Rate limiting on login attempts (prevent brute force)
- Failed login tracking (lock account after 5 failures)
- Email verification (optional for MVP, required for production)

#### 1.3 Authentication Middleware

**New File: `backend/internal/middleware/auth.go`**

```go
// RequireAuth middleware validates JWT token
func RequireAuth() gin.HandlerFunc

// RequireFullAccount middleware ensures user has full account
func RequireFullAccount() gin.HandlerFunc

// RequireAdmin middleware ensures user is admin or superadmin
func RequireAdmin() gin.HandlerFunc

// RequireSuperAdmin middleware ensures user is superadmin
func RequireSuperAdmin() gin.HandlerFunc

// OptionalAuth middleware validates token if present, but doesn't require it
func OptionalAuth() gin.HandlerFunc
```

**Usage:**
```go
// Protected routes
api.Use(RequireAuth())
api.POST("/maps", RequireFullAccount(), mapHandler.CreateMap)
api.DELETE("/users/:id", RequireAdmin(), userHandler.DeleteUser)

// Public routes (no middleware)
api.POST("/auth/login", authHandler.Login)
api.POST("/auth/signup", authHandler.Signup)
```

**Middleware Behavior:**
- Extract JWT from `Authorization: Bearer <token>` header
- Validate token signature and expiry
- Extract user ID, email, role from claims
- Store in Gin context: `c.Set("userID", userID)`
- Return 401 Unauthorized if invalid/missing

#### 1.4 User Service Updates

**Update: `backend/internal/services/user_service.go`**

Add new methods:
```go
// CreateFullAccount creates a full account with email/password
func (s *UserService) CreateFullAccount(ctx context.Context, email, password, displayName, aboutMe string) (*models.User, error)

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error)

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(ctx context.Context, userID, newPassword string) error

// VerifyPassword verifies a user's password
func (s *UserService) VerifyPassword(ctx context.Context, userID, password string) error
```

**Update: `backend/internal/repository/user_repository.go`**

Add new methods:
```go
// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error)

// UpdatePassword updates a user's password hash
func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error
```

#### 1.5 Super Admin Account Creation

**Update: `backend/internal/database/migrations.go`**

Add function to create super admin on first run:
```go
// CreateSuperAdminIfNotExists creates the super admin account if it doesn't exist
func CreateSuperAdminIfNotExists(db *gorm.DB) error {
    // Check if super admin exists
    var count int64
    if err := db.Model(&models.User{}).Where("role = ?", "superadmin").Count(&count).Error; err != nil {
        return err
    }
    
    if count == 0 {
        // Get credentials from environment variables
        email := os.Getenv("SUPERADMIN_EMAIL")
        password := os.Getenv("SUPERADMIN_PASSWORD")
        
        if email == "" || password == "" {
            return fmt.Errorf("SUPERADMIN_EMAIL and SUPERADMIN_PASSWORD must be set")
        }
        
        // Hash password
        passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
        if err != nil {
            return err
        }
        
        // Create super admin user
        superAdmin := &models.User{
            ID:           uuid.New().String(),
            Email:        &email,
            DisplayName:  "Super Admin",
            AccountType:  models.AccountTypeFull,
            Role:         models.UserRoleSuperAdmin,
            PasswordHash: stringPtr(string(passwordHash)),
            IsActive:     true,
            CreatedAt:    time.Now(),
            UpdatedAt:    time.Now(),
        }
        
        if err := db.Create(superAdmin).Error; err != nil {
            return err
        }
        
        log.Printf("✅ Created super admin account: %s", email)
    }
    
    return nil
}
```

**Environment Variables:**
```bash
SUPERADMIN_EMAIL=admin@breakoutglobe.com
SUPERADMIN_PASSWORD=<secure-password>
JWT_SECRET=<random-secret-key>
JWT_EXPIRY=24h
```

**Call in `RunMigrations()`:**
```go
func RunMigrations(db *gorm.DB) error {
    // ... existing migrations ...
    
    // Create super admin account
    if err := CreateSuperAdminIfNotExists(db); err != nil {
        return fmt.Errorf("failed to create super admin: %w", err)
    }
    
    return nil
}
```

### Phase 2: Frontend Authentication UI

#### 2.1 Authentication State Management

**New File: `frontend/src/stores/authStore.ts`**

```typescript
interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: UserProfile | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  signup: (data: SignupData) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<void>;
  setUser: (user: UserProfile) => void;
}

// Store tokens in localStorage
// Auto-refresh token before expiry
// Clear tokens on logout
```

**Token Management:**
- Store JWT in localStorage: `authToken`
- Store refresh token in localStorage: `refreshToken`
- Auto-refresh token 5 minutes before expiry
- Clear tokens on logout
- Validate token on app load

#### 2.2 Updated Welcome Flow

**Update: `frontend/src/components/WelcomeScreen.tsx`**

Add three options:
```tsx
<div className="space-y-4">
  <button onClick={onSignup} className="w-full btn-primary">
    Create Full Account
  </button>
  
  <button onClick={onLogin} className="w-full btn-secondary">
    Login
  </button>
  
  <button onClick={onGuestAccess} className="w-full btn-tertiary">
    Continue as Guest
  </button>
</div>
```

**Flow:**
1. User sees WelcomeScreen with 3 options
2. **Create Full Account** → SignupModal
3. **Login** → LoginModal
4. **Continue as Guest** → ProfileCreationModal (existing)

#### 2.3 New Authentication Modals

**New File: `frontend/src/components/SignupModal.tsx`**

```tsx
interface SignupModalProps {
  isOpen: boolean;
  onSignupSuccess: (user: UserProfile) => void;
  onClose: () => void;
  onSwitchToLogin: () => void;
}

// Form fields:
// - Email (required, validated)
// - Password (required, min 8 chars, strength indicator)
// - Confirm Password (must match)
// - Display Name (required, 3-50 chars)
// - About Me (optional, 500 chars max)
// - Avatar Upload (optional)

// Features:
// - Real-time validation
// - Password strength indicator
// - Show/hide password toggle
// - "Already have an account? Login" link
```

**New File: `frontend/src/components/LoginModal.tsx`**

```tsx
interface LoginModalProps {
  isOpen: boolean;
  onLoginSuccess: (user: UserProfile) => void;
  onClose: () => void;
  onSwitchToSignup: () => void;
  onForgotPassword: () => void;
}

// Form fields:
// - Email (required)
// - Password (required)
// - Remember Me (checkbox)

// Features:
// - "Forgot Password?" link
// - "Don't have an account? Sign up" link
// - Error handling (invalid credentials, account locked)
```

**New File: `frontend/src/components/ForgotPasswordModal.tsx`**

```tsx
interface ForgotPasswordModalProps {
  isOpen: boolean;
  onClose: () => void;
  onBackToLogin: () => void;
}

// Form fields:
// - Email (required)

// Features:
// - Send password reset email
// - Success message
// - "Back to Login" link
```

#### 2.4 API Service Updates

**Update: `frontend/src/services/api.ts`**

Add authentication functions:
```typescript
export async function signup(data: SignupRequest): Promise<AuthResponse>

export async function login(email: string, password: string): Promise<AuthResponse>

export async function refreshToken(refreshToken: string): Promise<AuthResponse>

export async function logout(): Promise<void>

export async function forgotPassword(email: string): Promise<void>

export async function resetPassword(token: string, newPassword: string): Promise<void>

export async function getCurrentUser(): Promise<UserProfile>
```

**Update all API calls to include JWT:**
```typescript
const headers: Record<string, string> = {
  'Content-Type': 'application/json',
};

// Add JWT token if authenticated
const token = authStore.getState().token;
if (token) {
  headers['Authorization'] = `Bearer ${token}`;
}
```

**Remove `X-User-ID` header** (replaced by JWT authentication)

#### 2.5 Protected Routes

**Update: `frontend/src/App.tsx`**

Add route protection:
```tsx
// Redirect to login if not authenticated
useEffect(() => {
  if (!authState.isAuthenticated && !authState.isLoading) {
    // Show welcome screen
    setShowWelcome(true);
  }
}, [authState.isAuthenticated, authState.isLoading]);

// Auto-refresh token on app load
useEffect(() => {
  authStore.getState().refreshAuth();
}, []);
```

### Phase 3: Admin Panel (Super Admin Features)

#### 3.1 Admin Dashboard

**New File: `frontend/src/components/AdminDashboard.tsx`**

Features:
- User management (list, view, edit, delete, upgrade to admin)
- Map management (list, view, delete)
- POI management (list, view, delete)
- System statistics (user count, map count, POI count)
- Activity logs

**Access Control:**
- Only visible to admin/superadmin users
- Accessed via `/admin` route
- Protected by `RequireAdmin()` middleware

#### 3.2 Admin API Endpoints

**New File: `backend/internal/handlers/admin_handler.go`**

```
GET    /api/admin/users              - List all users (paginated)
GET    /api/admin/users/:id          - Get user details
PUT    /api/admin/users/:id/role     - Update user role
DELETE /api/admin/users/:id          - Delete user
GET    /api/admin/maps               - List all maps
DELETE /api/admin/maps/:id           - Delete map
GET    /api/admin/pois               - List all POIs
DELETE /api/admin/pois/:id           - Delete POI
GET    /api/admin/stats              - System statistics
```

**Middleware:**
All admin routes use `RequireAdmin()` or `RequireSuperAdmin()` middleware

#### 3.3 User Role Management

**Upgrade User to Admin:**
```go
// PUT /api/admin/users/:id/role
type UpdateRoleRequest struct {
    Role string `json:"role"` // "user", "admin", "superadmin"
}

// Business Rules:
// - Only superadmin can create other superadmins
// - Admin can upgrade users to admin
// - Cannot downgrade yourself
// - Cannot change superadmin role (except by another superadmin)
```

## Migration Strategy

### Database Changes
**No schema changes needed!** The user table already has all required fields:
- `email` (nullable, unique index)
- `password_hash` (nullable)
- `account_type` (guest/full)
- `role` (user/admin/superadmin)

### Backward Compatibility

**Guest Accounts:**
- Existing guest accounts continue to work
- Guest users can upgrade to full account later (optional feature)
- Guest users have limited permissions (cannot create maps)

**API Compatibility:**
- Keep existing `X-User-ID` header support temporarily
- Add JWT authentication alongside
- Deprecate `X-User-ID` after migration period
- Frontend can use either method during transition

**Upgrade Path for Guests:**
```
POST /api/auth/upgrade-guest
Body: { email, password }

// Converts guest account to full account
// Preserves user ID, display name, avatar, etc.
// Adds email and password
```

## Security Considerations

### Password Security
- ✅ Bcrypt with cost factor 12
- ✅ Minimum 8 characters
- ✅ Complexity requirements (uppercase, lowercase, number, special)
- ✅ Password strength indicator on signup
- ⚠️ Consider: Password history (prevent reuse)
- ⚠️ Consider: Password expiry (force change after X days)

### JWT Security
- ✅ Short expiry (24 hours)
- ✅ Refresh token mechanism
- ✅ Signed with secret key (from environment)
- ⚠️ Consider: Token blacklist for logout
- ⚠️ Consider: Rotate JWT secret periodically

### Rate Limiting
- ✅ Login attempts: 5 per 15 minutes per IP
- ✅ Signup attempts: 3 per hour per IP
- ✅ Password reset: 3 per hour per email
- ⚠️ Consider: CAPTCHA after failed attempts

### Email Verification
- ⚠️ MVP: Optional (allow unverified accounts)
- ⚠️ Production: Required (send verification email)
- ⚠️ Consider: Resend verification email

### Account Lockout
- ✅ Lock account after 5 failed login attempts
- ✅ Unlock after 30 minutes
- ✅ Email notification on lockout
- ⚠️ Consider: Admin can manually unlock

### Super Admin Protection
- ✅ Super admin credentials in environment variables
- ✅ Cannot delete super admin account
- ✅ Cannot downgrade super admin role
- ⚠️ Consider: Multi-factor authentication for super admin

## Implementation Checklist

### Phase 1: Backend (3-4 days)

**Authentication Service:**
- [ ] Install dependencies: `bcrypt`, `jwt`
- [ ] Create `auth_service.go`
  - [ ] `HashPassword()` method
  - [ ] `VerifyPassword()` method
  - [ ] `GenerateJWT()` method
  - [ ] `ValidateJWT()` method
  - [ ] `GenerateRefreshToken()` method
- [ ] Add JWT secret to environment variables
- [ ] Unit tests for auth service

**Authentication Endpoints:**
- [ ] Create `auth_handler.go`
  - [ ] `Signup` handler
  - [ ] `Login` handler
  - [ ] `RefreshToken` handler
  - [ ] `Logout` handler
  - [ ] `ForgotPassword` handler (optional for MVP)
  - [ ] `ResetPassword` handler (optional for MVP)
  - [ ] `GetCurrentUser` handler
- [ ] Register auth routes in `server.go`
- [ ] Integration tests for auth endpoints

**Authentication Middleware:**
- [ ] Create `middleware/auth.go`
  - [ ] `RequireAuth()` middleware
  - [ ] `RequireFullAccount()` middleware
  - [ ] `RequireAdmin()` middleware
  - [ ] `RequireSuperAdmin()` middleware
  - [ ] `OptionalAuth()` middleware
- [ ] Unit tests for middleware

**User Service Updates:**
- [ ] Add `CreateFullAccount()` method
- [ ] Add `GetUserByEmail()` method
- [ ] Add `UpdatePassword()` method
- [ ] Add `VerifyPassword()` method
- [ ] Unit tests for new methods

**User Repository Updates:**
- [ ] Add `GetByEmail()` method
- [ ] Add `UpdatePassword()` method
- [ ] Unit tests for new methods

**Super Admin Setup:**
- [ ] Add `CreateSuperAdminIfNotExists()` to migrations
- [ ] Add environment variables for super admin
- [ ] Test super admin creation on startup
- [ ] Document super admin credentials management

**Apply Middleware to Routes:**
- [ ] Protect map creation: `RequireFullAccount()`
- [ ] Protect map modification: `RequireAuth()`
- [ ] Protect POI operations: `RequireAuth()`
- [ ] Keep guest endpoints public

**Testing:**
- [ ] Test signup flow
- [ ] Test login flow
- [ ] Test JWT validation
- [ ] Test middleware protection
- [ ] Test super admin creation
- [ ] Test role-based access control

### Phase 2: Frontend (3-4 days)

**Authentication Store:**
- [ ] Create `authStore.ts`
  - [ ] Token management
  - [ ] User state
  - [ ] Login/signup/logout actions
  - [ ] Auto-refresh token logic
- [ ] Persist tokens in localStorage
- [ ] Clear tokens on logout

**Welcome Screen Updates:**
- [ ] Add "Create Full Account" button
- [ ] Add "Login" button
- [ ] Keep "Continue as Guest" button
- [ ] Update flow logic

**Authentication Modals:**
- [ ] Create `SignupModal.tsx`
  - [ ] Email/password form
  - [ ] Password strength indicator
  - [ ] Display name field
  - [ ] About me field
  - [ ] Avatar upload
  - [ ] Validation
  - [ ] Error handling
  - [ ] "Switch to Login" link

- [ ] Create `LoginModal.tsx`
  - [ ] Email/password form
  - [ ] Remember me checkbox
  - [ ] Error handling
  - [ ] "Switch to Signup" link
  - [ ] "Forgot Password" link

- [ ] Create `ForgotPasswordModal.tsx` (optional for MVP)
  - [ ] Email input
  - [ ] Success message
  - [ ] "Back to Login" link

**API Service Updates:**
- [ ] Add `signup()` function
- [ ] Add `login()` function
- [ ] Add `refreshToken()` function
- [ ] Add `logout()` function
- [ ] Add `forgotPassword()` function (optional)
- [ ] Add `resetPassword()` function (optional)
- [ ] Update all API calls to include JWT
- [ ] Remove `X-User-ID` header usage

**App Updates:**
- [ ] Add authentication check on load
- [ ] Auto-refresh token logic
- [ ] Redirect to welcome if not authenticated
- [ ] Handle token expiry
- [ ] Handle logout

**Testing:**
- [ ] Test signup flow
- [ ] Test login flow
- [ ] Test logout flow
- [ ] Test token refresh
- [ ] Test protected routes
- [ ] Test guest access still works

### Phase 3: Admin Panel (2-3 days)

**Admin Dashboard:**
- [ ] Create `AdminDashboard.tsx`
  - [ ] User list with pagination
  - [ ] Map list
  - [ ] POI list
  - [ ] System statistics
  - [ ] Role management UI

**Admin API Endpoints:**
- [ ] Create `admin_handler.go`
  - [ ] List users endpoint
  - [ ] Get user details endpoint
  - [ ] Update user role endpoint
  - [ ] Delete user endpoint
  - [ ] List maps endpoint
  - [ ] Delete map endpoint
  - [ ] List POIs endpoint
  - [ ] Delete POI endpoint
  - [ ] System stats endpoint
- [ ] Apply `RequireAdmin()` middleware
- [ ] Integration tests

**Admin Routes:**
- [ ] Add `/admin` route
- [ ] Protect with admin check
- [ ] Add navigation link (conditional on role)

**Testing:**
- [ ] Test admin access control
- [ ] Test user management
- [ ] Test map management
- [ ] Test POI management
- [ ] Test role upgrades
- [ ] Test non-admin cannot access

### Documentation

- [ ] Update README with authentication setup
- [ ] Document environment variables
- [ ] Document super admin setup
- [ ] API documentation for auth endpoints
- [ ] User guide for signup/login
- [ ] Admin guide for user management

## Estimated Effort

### Phase 1: Backend Authentication
- Auth service: 1 day
- Auth endpoints: 1 day
- Middleware: 0.5 day
- User service updates: 0.5 day
- Super admin setup: 0.5 day
- Testing: 0.5 day
**Total: 4 days**

### Phase 2: Frontend Authentication
- Auth store: 0.5 day
- Signup/Login modals: 1.5 days
- API updates: 0.5 day
- App integration: 0.5 day
- Testing: 0.5 day
**Total: 3.5 days**

### Phase 3: Admin Panel
- Admin dashboard: 1 day
- Admin API: 1 day
- Testing: 0.5 day
**Total: 2.5 days**

**Grand Total: 10 days**

## Success Criteria

### Phase 1: Backend
- [ ] Users can create full accounts with email/password
- [ ] Users can login with email/password
- [ ] JWT tokens are generated and validated
- [ ] Middleware protects routes based on authentication
- [ ] Super admin account is created on first run
- [ ] All tests pass

### Phase 2: Frontend
- [ ] Users can signup from welcome screen
- [ ] Users can login from welcome screen
- [ ] Users can continue as guest (existing flow)
- [ ] JWT tokens are stored and auto-refreshed
- [ ] Protected routes require authentication
- [ ] Guest access still works

### Phase 3: Admin Panel
- [ ] Super admin can access admin dashboard
- [ ] Super admin can view all users/maps/POIs
- [ ] Super admin can delete users/maps/POIs
- [ ] Super admin can upgrade users to admin
- [ ] Regular users cannot access admin panel

## Risk Assessment

### Low Risk
- ✅ User model already supports full accounts
- ✅ Database schema ready (no migrations)
- ✅ Guest accounts continue to work

### Medium Risk
- ⚠️ JWT implementation complexity
- ⚠️ Token refresh logic
- ⚠️ Frontend state management for auth
- ⚠️ Testing authentication flows

### High Risk
- ❌ Security vulnerabilities if not implemented correctly
- ❌ Super admin credentials management
- ❌ Password reset flow (if email service fails)

## Conclusion

The BreakoutGlobe codebase is **well-prepared** for full account authentication! The user model already has all the necessary fields (`email`, `password_hash`, `account_type`, `role`). The main work involves:

1. **Adding authentication service** (password hashing, JWT generation)
2. **Creating auth endpoints** (signup, login, refresh)
3. **Implementing middleware** (protect routes based on auth)
4. **Building frontend UI** (signup/login modals)
5. **Creating super admin account** (for system management)
6. **Building admin panel** (user/map/POI management)

The implementation is straightforward with an estimated effort of **10 days** for all three phases. Security is the main concern, so careful implementation and testing of authentication flows is critical.
