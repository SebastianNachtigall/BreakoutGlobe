package middleware

import (
	"net/http"
	"strings"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthService interface for JWT validation
type AuthService interface {
	ValidateJWT(token string) (*services.JWTClaims, error)
}

// UserService interface for user operations
type UserService interface {
	GetUser(c *gin.Context, userID string) (*models.User, error)
}

// RequireAuth middleware validates JWT token and sets user info in context
func RequireAuth(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "MISSING_TOKEN",
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN_FORMAT",
				"message": "Authorization header must be in format: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := authService.ValidateJWT(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    "TOKEN_EXPIRED",
					"message": "Authentication token has expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid authentication token",
				})
			}
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireFullAccount middleware ensures user has a full account (not guest)
// NOTE: This middleware assumes RequireAuth has already been called in the middleware chain
// Usage: router.GET("/path", RequireAuth(authService), RequireFullAccount(userService), handler)
func RequireFullAccount(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by RequireAuth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user from service
		user, err := userService.GetUser(c, userID.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user has full account
		if user.AccountType != models.AccountTypeFull {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "FULL_ACCOUNT_REQUIRED",
				"message": "This action requires a full account",
			})
			c.Abort()
			return
		}

		// Store account type in context
		c.Set("accountType", user.AccountType)

		c.Next()
	}
}

// RequireAdmin middleware ensures user has admin or superadmin role
// NOTE: This middleware assumes RequireAuth has already been called in the middleware chain
// Usage: router.GET("/path", RequireAuth(authService), RequireAdmin(), handler)
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by RequireAuth middleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user is admin or superadmin
		userRole := role.(models.UserRole)
		if userRole != models.UserRoleAdmin && userRole != models.UserRoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "ADMIN_REQUIRED",
				"message": "This action requires admin privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin middleware ensures user has superadmin role
// NOTE: This middleware assumes RequireAuth has already been called in the middleware chain
// Usage: router.GET("/path", RequireAuth(authService), RequireSuperAdmin(), handler)
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by RequireAuth middleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user is superadmin
		userRole := role.(models.UserRole)
		if userRole != models.UserRoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "SUPERADMIN_REQUIRED",
				"message": "This action requires superadmin privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware validates token if present, but doesn't require it
func OptionalAuth(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := authService.ValidateJWT(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Store user info in context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
