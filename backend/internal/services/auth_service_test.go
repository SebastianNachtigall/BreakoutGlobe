package services

import (
	"strings"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	testJWTSecret = "test-secret-key-for-jwt-signing-minimum-32-chars"
	testJWTExpiry = 24 * time.Hour
)

// TestHashPassword_Success tests successful password hashing
func TestHashPassword_Success(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	password := "TestPassword123!"
	hash, err := authService.HashPassword(password)
	
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "Hash should not equal plain password")
	assert.True(t, strings.HasPrefix(hash, "$2a$"), "Should be bcrypt hash")
}

// TestHashPassword_EmptyPassword tests hashing empty password
func TestHashPassword_EmptyPassword(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	hash, err := authService.HashPassword("")
	
	assert.Error(t, err)
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

// TestHashPassword_DifferentPasswordsDifferentHashes tests that same password produces different hashes
func TestHashPassword_DifferentPasswordsDifferentHashes(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	password := "TestPassword123!"
	hash1, err1 := authService.HashPassword(password)
	hash2, err2 := authService.HashPassword(password)
	
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes (bcrypt salt)")
}

// TestVerifyPassword_Success tests successful password verification
func TestVerifyPassword_Success(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	password := "TestPassword123!"
	hash, err := authService.HashPassword(password)
	require.NoError(t, err)
	
	err = authService.VerifyPassword(password, hash)
	assert.NoError(t, err)
}

// TestVerifyPassword_WrongPassword tests verification with wrong password
func TestVerifyPassword_WrongPassword(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	password := "TestPassword123!"
	wrongPassword := "WrongPassword456!"
	hash, err := authService.HashPassword(password)
	require.NoError(t, err)
	
	err = authService.VerifyPassword(wrongPassword, hash)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

// TestVerifyPassword_EmptyPassword tests verification with empty password
func TestVerifyPassword_EmptyPassword(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	err := authService.VerifyPassword("", "somehash")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

// TestVerifyPassword_EmptyHash tests verification with empty hash
func TestVerifyPassword_EmptyHash(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	err := authService.VerifyPassword("password", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash cannot be empty")
}

// TestVerifyPassword_InvalidHash tests verification with invalid hash format
func TestVerifyPassword_InvalidHash(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	err := authService.VerifyPassword("password", "not-a-valid-bcrypt-hash")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

// TestGenerateJWT_Success tests successful JWT generation
func TestGenerateJWT_Success(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	userID := "user-123"
	email := "test@example.com"
	role := models.UserRoleUser
	
	tokenString, expiryTime, err := authService.GenerateJWT(userID, email, role)
	
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	assert.False(t, expiryTime.IsZero())
	assert.True(t, expiryTime.After(time.Now()))
	
	// Verify token structure (JWT has 3 parts separated by dots)
	parts := strings.Split(tokenString, ".")
	assert.Equal(t, 3, len(parts), "JWT should have 3 parts")
}

// TestGenerateJWT_ValidClaims tests that generated JWT contains correct claims
func TestGenerateJWT_ValidClaims(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	userID := "user-123"
	email := "test@example.com"
	role := models.UserRoleAdmin
	
	tokenString, _, err := authService.GenerateJWT(userID, email, role)
	require.NoError(t, err)
	
	// Validate and extract claims
	claims, err := authService.ValidateJWT(tokenString)
	require.NoError(t, err)
	
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

// TestGenerateJWT_EmptyUserID tests JWT generation with empty user ID
func TestGenerateJWT_EmptyUserID(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	tokenString, expiryTime, err := authService.GenerateJWT("", "test@example.com", models.UserRoleUser)
	
	assert.Error(t, err)
	assert.Empty(t, tokenString)
	assert.True(t, expiryTime.IsZero())
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

// TestGenerateJWT_EmptyEmail tests JWT generation with empty email
func TestGenerateJWT_EmptyEmail(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	tokenString, expiryTime, err := authService.GenerateJWT("user-123", "", models.UserRoleUser)
	
	assert.Error(t, err)
	assert.Empty(t, tokenString)
	assert.True(t, expiryTime.IsZero())
	assert.Contains(t, err.Error(), "email cannot be empty")
}

// TestGenerateJWT_NoSecret tests JWT generation without secret
func TestGenerateJWT_NoSecret(t *testing.T) {
	authService := NewAuthService("", testJWTExpiry)
	
	tokenString, expiryTime, err := authService.GenerateJWT("user-123", "test@example.com", models.UserRoleUser)
	
	assert.Error(t, err)
	assert.Empty(t, tokenString)
	assert.True(t, expiryTime.IsZero())
	assert.Contains(t, err.Error(), "JWT secret not configured")
}

// TestValidateJWT_Success tests successful JWT validation
func TestValidateJWT_Success(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	userID := "user-123"
	email := "test@example.com"
	role := models.UserRoleSuperAdmin
	
	tokenString, _, err := authService.GenerateJWT(userID, email, role)
	require.NoError(t, err)
	
	claims, err := authService.ValidateJWT(tokenString)
	
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
}

// TestValidateJWT_ExpiredToken tests validation of expired token
func TestValidateJWT_ExpiredToken(t *testing.T) {
	// Create service with very short expiry
	authService := NewAuthService(testJWTSecret, 1*time.Millisecond)
	
	tokenString, _, err := authService.GenerateJWT("user-123", "test@example.com", models.UserRoleUser)
	require.NoError(t, err)
	
	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)
	
	claims, err := authService.ValidateJWT(tokenString)
	
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

// TestValidateJWT_InvalidSignature tests validation with wrong secret
func TestValidateJWT_InvalidSignature(t *testing.T) {
	authService1 := NewAuthService("secret1", testJWTExpiry)
	authService2 := NewAuthService("secret2-different", testJWTExpiry)
	
	// Generate token with first service
	tokenString, _, err := authService1.GenerateJWT("user-123", "test@example.com", models.UserRoleUser)
	require.NoError(t, err)
	
	// Try to validate with second service (different secret)
	claims, err := authService2.ValidateJWT(tokenString)
	
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateJWT_MalformedToken tests validation of malformed token
func TestValidateJWT_MalformedToken(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"random string", "not-a-jwt-token"},
		{"incomplete JWT", "header.payload"},
		{"invalid base64", "!!!.!!!.!!!"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := authService.ValidateJWT(tc.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

// TestValidateJWT_TamperedToken tests validation of tampered token
func TestValidateJWT_TamperedToken(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	tokenString, _, err := authService.GenerateJWT("user-123", "test@example.com", models.UserRoleUser)
	require.NoError(t, err)
	
	// Tamper with the token by changing a character
	tamperedToken := tokenString[:len(tokenString)-5] + "XXXXX"
	
	claims, err := authService.ValidateJWT(tamperedToken)
	
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// TestValidateJWT_NoSecret tests validation without secret
func TestValidateJWT_NoSecret(t *testing.T) {
	authService := NewAuthService("", testJWTExpiry)
	
	claims, err := authService.ValidateJWT("some.jwt.token")
	
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "JWT secret not configured")
}

// TestValidateJWT_WrongSigningMethod tests validation with wrong signing method
func TestValidateJWT_WrongSigningMethod(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	// Create a token with RS256 instead of HS256 (will fail validation)
	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   models.UserRoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(testJWTExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	// Try to create token with None algorithm (unsigned)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	
	// Validation should fail
	validatedClaims, err := authService.ValidateJWT(tokenString)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

// TestJWT_RoundTrip tests complete JWT lifecycle
func TestJWT_RoundTrip(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	testCases := []struct {
		name   string
		userID string
		email  string
		role   models.UserRole
	}{
		{"regular user", "user-1", "user@example.com", models.UserRoleUser},
		{"admin user", "admin-1", "admin@example.com", models.UserRoleAdmin},
		{"superadmin user", "superadmin-1", "superadmin@example.com", models.UserRoleSuperAdmin},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate token
			tokenString, expiryTime, err := authService.GenerateJWT(tc.userID, tc.email, tc.role)
			require.NoError(t, err)
			assert.NotEmpty(t, tokenString)
			assert.True(t, expiryTime.After(time.Now()))
			
			// Validate token
			claims, err := authService.ValidateJWT(tokenString)
			require.NoError(t, err)
			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.email, claims.Email)
			assert.Equal(t, tc.role, claims.Role)
		})
	}
}

// TestPasswordHashVerify_RoundTrip tests complete password hash/verify lifecycle
func TestPasswordHashVerify_RoundTrip(t *testing.T) {
	authService := NewAuthService(testJWTSecret, testJWTExpiry)
	
	testPasswords := []string{
		"SimplePassword123!",
		"C0mpl3x!P@ssw0rd#2024",
		"Short1!",
		"VeryLongPasswordWithManyCharacters123!@#$%^&*()",
		"Пароль123!", // Unicode password
	}
	
	for _, password := range testPasswords {
		t.Run(password, func(t *testing.T) {
			// Hash password
			hash, err := authService.HashPassword(password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			
			// Verify correct password
			err = authService.VerifyPassword(password, hash)
			assert.NoError(t, err)
			
			// Verify wrong password fails
			err = authService.VerifyPassword(password+"wrong", hash)
			assert.Error(t, err)
		})
	}
}
