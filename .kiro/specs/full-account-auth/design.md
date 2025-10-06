# Design Document: Full Account Authentication & Admin Panel

## Overview

This document outlines the technical design for implementing full account authentication with email/password, JWT-based session management, and a basic admin panel for user management. The design prioritizes simplicity, security, and backward compatibility with existing guest accounts.

**Architecture Pattern:** Layered architecture with clear separation between authentication, authorization, and business logic.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (React)                      │
├─────────────────────────────────────────────────────────────┤
│  WelcomeScreen  │  SignupModal  │  LoginModal  │  AdminPanel│
│  AuthStore      │  API Service  │  Protected Routes         │
└─────────────────────────────────────────────────────────────┘
                              │
                         JWT Token
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Backend (Go/Gin)                        │
├─────────────────────────────────────────────────────────────┤
│  Auth Middleware  │  Auth Handler  │  Admin Handler         │
│  Auth Service     │  User Service  │  Rate Limiter          │
│  User Repository  │  Database      │  Redis (rate limiting) │
└─────────────────────────────────────────────────────────────┘
```

### Authentication Flow

```
User → WelcomeScreen → [Signup/Login/Guest]
                              │
                              ↓
                    [Signup] → SignupModal → POST /api/auth/signup
                              │                      │
                              │                      ↓
                              │              AuthService.CreateFullAccount
                              │                      │
                              │                      ↓
                              │              Generate JWT Token
                              │                      │
                              │                      ↓
                              │              Return {token, user}
                              │                      │
                              ↓                      ↓
                    Store token in localStorage
                              │
                              ↓
                    Redirect to Map (authenticated)
```

## Components and Interfaces

### Backend Components

#### 1. Authentication Service (`backend/internal/services/auth_service.go`)

**Purpose:** Handle password hashing, JWT generation/validation, and authentication logic.

**Interface:**
```go
type AuthService struct {
    jwtSecret     []byte
    jwtExpiry     time.Duration
    userRepo      interfaces.UserRepositoryInterface
    rateLimiter   RateLimiterInterface
}

// Password operations
func (s *AuthService) HashPassword(password string) (string, error)
func (s *AuthService) VerifyPassword(password, hash string) error
func (s *AuthService) ValidatePasswordComplexity(password string) error

// JWT operations
func (s *AuthService) GenerateJWT(user *models.User) (string, time.Time, error)
func (s *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error)

// Authentication operations
func (s *AuthService) Authenticate(email, password string) (*models.User, string, error)
func (s *AuthService) CheckAccountLockout(email string) error
func (s *AuthService) RecordFailedLogin(email string) error
```

**JWT Claims Structure:**
```go
type JWTClaims struct {
    UserID      string `json:"userId"`
    Email       string `json:"email"`
    Role        string `json:"role"`
    AccountType string `json:"accountType"`
    jwt.RegisteredClaims
}
```


**Password Complexity Rules:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character from: `!@#$%^&*()_+-=[]{}|;:,.<>?`

**Account Lockout Logic:**
- Track failed login attempts in Redis with key: `login_attempts:{email}`
- Increment counter on failed login
- Set TTL of 15 minutes on first failed attempt
- Lock account after 5 failed attempts
- Store lockout in Redis with key: `account_locked:{email}` with TTL of 30 minutes

#### 2. Authentication Handler (`backend/internal/handlers/auth_handler.go`)

**Purpose:** Handle HTTP requests for authentication endpoints.

**Endpoints:**
```go
POST   /api/auth/signup          - Create full account
POST   /api/auth/login           - Login with email/password
POST   /api/auth/logout          - Logout (client-side token removal)
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
    Token     string      `json:"token"`
    ExpiresAt time.Time   `json:"expiresAt"`
    User      UserProfile `json:"user"`
}

// UserProfile (response)
type UserProfile struct {
    ID          string `json:"id"`
    Email       string `json:"email"`
    DisplayName string `json:"displayName"`
    AvatarURL   string `json:"avatarUrl,omitempty"`
    AboutMe     string `json:"aboutMe,omitempty"`
    AccountType string `json:"accountType"`
    Role        string `json:"role"`
}
```

#### 3. Authentication Middleware (`backend/internal/middleware/auth.go`)

**Purpose:** Validate JWT tokens and enforce authorization rules.

**Middleware Functions:**
```go
// RequireAuth validates JWT token and sets user context
func RequireAuth() gin.HandlerFunc

// RequireFullAccount ensures user has full account (not guest)
func RequireFullAccount() gin.HandlerFunc

// RequireAdmin ensures user is admin or superadmin
func RequireAdmin() gin.HandlerFunc

// RequireSuperAdmin ensures user is superadmin
func RequireSuperAdmin() gin.HandlerFunc

// OptionalAuth validates token if present but doesn't require it
func OptionalAuth() gin.HandlerFunc
```

**Middleware Behavior:**
1. Extract JWT from `Authorization: Bearer <token>` header
2. Validate token signature and expiry
3. Extract claims (userID, email, role, accountType)
4. Store in Gin context: `c.Set("userID", userID)`, `c.Set("userRole", role)`
5. Return 401 Unauthorized if invalid/missing (for required auth)
6. Continue to handler if valid

**Context Keys:**
- `userID` (string) - User's unique identifier
- `userEmail` (string) - User's email address
- `userRole` (string) - User's role (user/admin/superadmin)
- `accountType` (string) - Account type (guest/full)

#### 4. Admin Handler (`backend/internal/handlers/admin_handler.go`)

**Purpose:** Handle admin panel operations for user management.

**Endpoints:**
```go
GET    /api/admin/users              - List all users (paginated)
GET    /api/admin/users/:id          - Get user details
PUT    /api/admin/users/:id/role     - Update user role
DELETE /api/admin/users/:id          - Delete user (soft delete)
```

**Request/Response DTOs:**
```go
// ListUsersRequest (query params)
type ListUsersRequest struct {
    Page     int    `form:"page" binding:"min=1"`
    PageSize int    `form:"pageSize" binding:"min=1,max=100"`
    Search   string `form:"search"`
    Role     string `form:"role"`
}

// ListUsersResponse
type ListUsersResponse struct {
    Users      []AdminUserSummary `json:"users"`
    TotalCount int                `json:"totalCount"`
    Page       int                `json:"page"`
    PageSize   int                `json:"pageSize"`
}

// AdminUserSummary
type AdminUserSummary struct {
    ID          string    `json:"id"`
    Email       string    `json:"email,omitempty"`
    DisplayName string    `json:"displayName"`
    AccountType string    `json:"accountType"`
    Role        string    `json:"role"`
    IsActive    bool      `json:"isActive"`
    CreatedAt   time.Time `json:"createdAt"`
}

// AdminUserDetails
type AdminUserDetails struct {
    AdminUserSummary
    AvatarURL      string    `json:"avatarUrl,omitempty"`
    AboutMe        string    `json:"aboutMe,omitempty"`
    UpdatedAt      time.Time `json:"updatedAt"`
    MapsCreated    int       `json:"mapsCreated"`
    POIsCreated    int       `json:"poisCreated"`
    ActiveSessions int       `json:"activeSessions"`
}

// UpdateRoleRequest
type UpdateRoleRequest struct {
    Role string `json:"role" binding:"required,oneof=user admin superadmin"`
}
```

#### 5. User Service Updates (`backend/internal/services/user_service.go`)

**New Methods:**
```go
// CreateFullAccount creates a full account with email/password
func (s *UserService) CreateFullAccount(
    ctx context.Context,
    email, password, displayName, aboutMe string,
) (*models.User, error)

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(
    ctx context.Context,
    email string,
) (*models.User, error)

// UpdateUserRole updates a user's role (admin operation)
func (s *UserService) UpdateUserRole(
    ctx context.Context,
    userID string,
    newRole models.UserRole,
    adminUserID string,
) error

// GetUserStats returns statistics for admin panel
func (s *UserService) GetUserStats(
    ctx context.Context,
    userID string,
) (*UserStats, error)
```

#### 6. User Repository Updates (`backend/internal/repository/user_repository.go`)

**New Methods:**
```go
// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(
    ctx context.Context,
    email string,
) (*models.User, error)

// ListUsers returns paginated list of users
func (r *userRepository) ListUsers(
    ctx context.Context,
    page, pageSize int,
    search, roleFilter string,
) ([]*models.User, int, error)

// CountUserMaps counts maps created by user
func (r *userRepository) CountUserMaps(
    ctx context.Context,
    userID string,
) (int, error)

// CountUserPOIs counts POIs created by user
func (r *userRepository) CountUserPOIs(
    ctx context.Context,
    userID string,
) (int, error)

// CountActiveSessions counts active sessions for user
func (r *userRepository) CountActiveSessions(
    ctx context.Context,
    userID string,
) (int, error)
```

### Frontend Components

#### 1. Authentication Store (`frontend/src/stores/authStore.ts`)

**Purpose:** Manage authentication state and token lifecycle.

**State:**
```typescript
interface AuthState {
  token: string | null;
  user: UserProfile | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  signup: (data: SignupData) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
  clearError: () => void;
}
```

**Token Management:**
- Store JWT in localStorage key: `authToken`
- Parse JWT to extract expiry time
- Check token expiry on app load
- Auto-logout when token expires
- Clear token on logout

#### 2. Welcome Screen Updates (`frontend/src/components/WelcomeScreen.tsx`)

**Changes:**
- Add three buttons: "Create Full Account", "Login", "Continue as Guest"
- Add state for which modal to show
- Pass callbacks to parent component

**Props:**
```typescript
interface WelcomeScreenProps {
  isOpen: boolean;
  onSignup: () => void;
  onLogin: () => void;
  onGuestAccess: () => void;
}
```

#### 3. Signup Modal (`frontend/src/components/SignupModal.tsx`)

**Purpose:** Collect user information for full account creation.

**Form Fields:**
- Email (required, email validation)
- Password (required, min 8 chars, complexity validation)
- Confirm Password (required, must match password)
- Display Name (required, 3-50 chars)
- About Me (optional, max 500 chars)

**Features:**
- Real-time password strength indicator
- Show/hide password toggle
- Client-side validation before submission
- "Already have an account? Login" link
- Error display for API errors

**Props:**
```typescript
interface SignupModalProps {
  isOpen: boolean;
  onSignupSuccess: (user: UserProfile) => void;
  onClose: () => void;
  onSwitchToLogin: () => void;
}
```

#### 4. Login Modal (`frontend/src/components/LoginModal.tsx`)

**Purpose:** Authenticate existing users.

**Form Fields:**
- Email (required)
- Password (required)

**Features:**
- Error display for invalid credentials
- Account lockout message display
- "Don't have an account? Sign up" link
- Loading state during authentication

**Props:**
```typescript
interface LoginModalProps {
  isOpen: boolean;
  onLoginSuccess: (user: UserProfile) => void;
  onClose: () => void;
  onSwitchToSignup: () => void;
}
```

#### 5. Admin Dashboard (`frontend/src/components/AdminDashboard.tsx`)

**Purpose:** Provide user management interface for admins.

**Sections:**
- User list table with columns: Display Name, Email, Account Type, Role, Created Date, Actions
- Search bar (filter by name/email)
- Role filter dropdown
- Pagination controls
- User details modal
- Delete confirmation dialog
- Role change dropdown

**Features:**
- Real-time search (debounced)
- Sortable columns
- Bulk actions (future enhancement)
- Export users (future enhancement)

**Props:**
```typescript
interface AdminDashboardProps {
  // No props - uses auth store for current user
}
```

#### 6. API Service Updates (`frontend/src/services/api.ts`)

**New Functions:**
```typescript
// Authentication
export async function signup(data: SignupRequest): Promise<AuthResponse>
export async function login(email: string, password: string): Promise<AuthResponse>
export async function logout(): Promise<void>
export async function getCurrentAuthUser(): Promise<UserProfile>

// Admin
export async function listUsers(params: ListUsersParams): Promise<ListUsersResponse>
export async function getUserDetails(userId: string): Promise<AdminUserDetails>
export async function updateUserRole(userId: string, role: string): Promise<void>
export async function deleteUser(userId: string): Promise<void>
```

**Update All API Calls:**
```typescript
// Add JWT token to all requests
const headers: Record<string, string> = {
  'Content-Type': 'application/json',
};

const token = authStore.getState().token;
if (token) {
  headers['Authorization'] = `Bearer ${token}`;
}
```

## Data Models

### User Model (Existing - No Changes Needed)

```go
type User struct {
    ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Email        *string        `json:"email" gorm:"uniqueIndex;type:varchar(255)"`
    DisplayName  string         `json:"displayName" gorm:"type:varchar(50);not null"`
    AvatarURL    *string        `json:"avatarUrl" gorm:"type:varchar(500)"`
    AboutMe      *string        `json:"aboutMe" gorm:"type:text"`
    AccountType  AccountType    `json:"accountType" gorm:"type:varchar(20);not null;default:'full'"`
    Role         UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
    PasswordHash *string        `json:"-" gorm:"type:varchar(255)"`
    IsActive     bool           `json:"isActive" gorm:"default:true"`
    CreatedAt    time.Time      `json:"createdAt"`
    UpdatedAt    time.Time      `json:"updatedAt"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
```

**No database migrations needed!** All required fields already exist.

### Redis Data Structures

**Failed Login Attempts:**
```
Key: login_attempts:{email}
Type: String (counter)
TTL: 15 minutes
Value: Number of failed attempts
```

**Account Lockout:**
```
Key: account_locked:{email}
Type: String (timestamp)
TTL: 30 minutes
Value: Lockout timestamp
```

## Error Handling

### Backend Error Responses

**Standard Error Format:**
```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": "Additional error details (optional)"
}
```

**Error Codes:**
- `INVALID_REQUEST` - Malformed request body
- `VALIDATION_ERROR` - Field validation failed
- `UNAUTHORIZED` - Missing or invalid JWT token
- `FORBIDDEN` - Insufficient permissions
- `EMAIL_IN_USE` - Email already registered
- `INVALID_CREDENTIALS` - Wrong email/password
- `ACCOUNT_LOCKED` - Too many failed login attempts
- `USER_NOT_FOUND` - User doesn't exist
- `INTERNAL_ERROR` - Server error

### Frontend Error Handling

**Error Display Strategy:**
- Form validation errors: Show inline below field
- API errors: Show in modal/toast notification
- Authentication errors: Show in login/signup modal
- Authorization errors: Redirect to welcome screen with message

## Testing Strategy

### Backend Tests

**Unit Tests:**
- `auth_service_test.go` - Password hashing, JWT generation/validation
- `auth_handler_test.go` - Signup/login endpoint logic
- `admin_handler_test.go` - Admin operations
- `auth_middleware_test.go` - Token validation, role checks

**Integration Tests:**
- Full signup flow (API → Service → Repository → Database)
- Full login flow with JWT validation
- Account lockout after failed attempts
- Admin user management operations
- Protected route access control

**Test Scenarios:**
- Valid signup with all fields
- Signup with duplicate email
- Signup with weak password
- Valid login
- Login with invalid credentials
- Login with locked account
- JWT token expiry
- Admin accessing user list
- Non-admin accessing admin endpoints
- Super admin changing user roles
- User attempting to change own role

### Frontend Tests

**Component Tests:**
- SignupModal form validation
- LoginModal form validation
- AdminDashboard user list rendering
- Password strength indicator

**Integration Tests:**
- Complete signup flow
- Complete login flow
- Token storage and retrieval
- Auto-logout on token expiry
- Protected route redirection

## Security Considerations

### Password Security
- Bcrypt hashing with cost factor 12
- Minimum 8 characters with complexity requirements
- Password not logged or exposed in errors
- Password hash never returned in API responses

### JWT Security
- Signed with secret key from environment variables
- Short expiry (24 hours)
- Claims include only necessary data (no sensitive info)
- Validated on every protected request
- Stored in localStorage (XSS risk mitigated by CSP headers)

### Rate Limiting
- Login attempts: 5 per 15 minutes per email
- Signup attempts: 3 per hour per IP (existing rate limiter)
- Admin operations: 30 per minute per user (existing rate limiter)

### Account Lockout
- Temporary lockout after 5 failed login attempts
- 30-minute lockout duration
- Lockout data stored in Redis (auto-expires)

### Admin Security
- Super admin credentials in environment variables only
- Admin endpoints protected by middleware
- Cannot change own role
- All admin actions logged (future enhancement)

## Deployment Strategy

### Environment Variables

**Required:**
```bash
SUPERADMIN_EMAIL=admin@breakoutglobe.com
SUPERADMIN_PASSWORD=<secure-random-password>
JWT_SECRET=<secure-random-key-min-32-chars>
JWT_EXPIRY=24h
```

**Optional:**
```bash
ACCOUNT_LOCKOUT_ATTEMPTS=5
ACCOUNT_LOCKOUT_DURATION=30m
LOGIN_ATTEMPT_WINDOW=15m
```

### Database Migration

**No schema changes needed!** The user table already has all required fields.

**Super Admin Creation:**
- Add `CreateSuperAdminIfNotExists()` to `RunMigrations()`
- Check if any user with `role='superadmin'` exists
- If not, create from environment variables
- Log success/failure

### Backward Compatibility

**Transition Period (2 weeks):**
1. Deploy backend with both JWT and `X-User-ID` header support
2. Frontend continues using `X-User-ID` for existing guest users
3. New signups/logins use JWT
4. After 2 weeks, deprecate `X-User-ID` header

**Guest User Migration:**
- Existing guest users continue working with cached user ID
- No forced migration to full accounts
- Guest users can create new accounts if desired

## Performance Considerations

### Backend Optimizations
- JWT validation is stateless (no database lookup)
- Failed login attempts cached in Redis (fast lookups)
- User list pagination (max 100 per page)
- Database indexes on email (already exists)

### Frontend Optimizations
- Token stored in localStorage (no repeated API calls)
- Auth state cached in Zustand store
- Admin dashboard uses virtual scrolling for large lists (future)
- Debounced search (300ms delay)

### Caching Strategy
- JWT tokens cached in localStorage
- User profile cached in auth store
- Admin user list cached for 30 seconds
- Failed login attempts cached in Redis (15 min TTL)

## Monitoring and Logging

### Backend Logging
- Log all authentication attempts (success/failure)
- Log account lockouts
- Log admin operations (user deletion, role changes)
- Log JWT validation failures
- Do NOT log passwords or tokens

### Frontend Logging
- Log authentication state changes
- Log API errors
- Log token expiry events
- Do NOT log passwords or tokens

### Metrics to Track
- Signup conversion rate
- Login success rate
- Failed login attempts per user
- Account lockout frequency
- Admin panel usage
- JWT token expiry rate

## Future Enhancements (Post-MVP)

1. **Password Reset Flow** - Email-based password reset
2. **Refresh Tokens** - Long-lived tokens for seamless re-authentication
3. **Email Verification** - Verify email addresses on signup
4. **Account Upgrade** - Convert guest accounts to full accounts
5. **Multi-Factor Authentication** - TOTP-based 2FA
6. **Social Login** - OAuth with Google/Facebook
7. **Admin Activity Logs** - Audit trail of all admin actions
8. **Session Management** - View and terminate active sessions
9. **Account Deactivation** - Temporary suspension without deletion
10. **Map/POI Management** - Admin panel for maps and POIs (after multi-map feature)
