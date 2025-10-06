# Avatar Upload Authentication Fix

## Problem
Full account users were unable to upload avatar images. The console showed:
```
POST http://localhost:8080/api/users/avatar 401 (Unauthorized)
Avatar upload failed: APIError: User ID required
```

## Root Cause
The backend's `UploadAvatar` handler was only checking for the `X-User-ID` header, which is used for guest users. However, full account users authenticate using JWT tokens sent in the `Authorization` header.

**Backend Issue (`user_handler.go`):**
```go
// Old code - only checked header
userID := c.GetHeader("X-User-ID")
if userID == "" {
    return 401 Unauthorized
}
```

**Frontend Behavior (`api.ts`):**
```typescript
// Full accounts send JWT token
if (authToken) {
  headers['Authorization'] = `Bearer ${authToken}`;
} else if (userId) {
  // Guest users send X-User-ID header
  headers['X-User-ID'] = userId;
}
```

The auth middleware (`OptionalAuth`) validates the JWT and sets `userID` in the Gin context, but the handler wasn't checking the context.

## Solution
Updated both `UploadAvatar` and `UpdateProfile` handlers to check the context first (for authenticated users) before falling back to the header (for guest users):

```go
// Get user ID from context (set by auth middleware) or header (for guest users)
var userID string
if contextUserID, exists := c.Get("userID"); exists {
    userID = contextUserID.(string)
} else {
    userID = c.GetHeader("X-User-ID")
}

if userID == "" {
    c.JSON(http.StatusUnauthorized, ErrorResponse{
        Code:    "UNAUTHORIZED",
        Message: "User ID required",
    })
    return
}
```

## Changes Made
1. **`backend/internal/handlers/user_handler.go`**:
   - Updated `UploadAvatar` handler to check context first (for JWT auth)
   - Updated `UpdateProfile` handler to check context first (for JWT auth)
   - Fixed `UploadAvatar` response to include `AboutMe` field
   - Both handlers now support both authentication methods:
     - JWT token (full accounts) → user ID from context
     - X-User-ID header (guest accounts) → user ID from header

2. **`frontend/src/services/api.ts`**:
   - Added debug logging to track avatar upload responses

3. **`frontend/src/components/ProfileSettingsModal.tsx`**:
   - Added debug logging to track avatar upload flow

## Testing
- ✅ Existing tests pass
- ✅ Backend compiles without errors
- ✅ Changes are backward compatible with guest users
- ✅ Avatar file is saved to disk
- ✅ Avatar URL is saved to database
- ⏳ Testing avatar display in UI

## How to Verify
1. Log in with a full account
2. Open Profile Settings modal
3. Upload an avatar image
4. Should succeed without 401 error
5. Avatar should be displayed in the modal
6. Check browser console for debug logs showing the response

## Database Verification
```sql
SELECT id, display_name, avatar_url FROM users WHERE account_type = 'full';
```

The fix maintains backward compatibility with guest users while properly supporting full account authentication.
