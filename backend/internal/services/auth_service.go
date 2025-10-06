package services

import (
	"fmt"
	"time"

	"breakoutglobe/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims represents the claims stored in a JWT token
type JWTClaims struct {
	UserID string           `json:"userId"`
	Email  string           `json:"email"`
	Role   models.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret   []byte
	jwtExpiry   time.Duration
}

// NewAuthService creates a new AuthService instance
func NewAuthService(jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:   []byte(jwtSecret),
		jwtExpiry:   jwtExpiry,
	}
}

// HashPassword hashes a password using bcrypt with cost factor 12
func (s *AuthService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Use bcrypt cost factor 12 for security
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a bcrypt hash
func (s *AuthService) VerifyPassword(password, hash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}


// GenerateJWT generates a JWT token for a user
func (s *AuthService) GenerateJWT(userID, email string, role models.UserRole) (string, time.Time, error) {
	if userID == "" {
		return "", time.Time{}, fmt.Errorf("user ID cannot be empty")
	}
	if email == "" {
		return "", time.Time{}, fmt.Errorf("email cannot be empty")
	}
	if len(s.jwtSecret) == 0 {
		return "", time.Time{}, fmt.Errorf("JWT secret not configured")
	}

	// Calculate expiry time
	expiryTime := time.Now().Add(s.jwtExpiry)

	// Create claims
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiryTime, nil
}

// ValidateJWT validates a JWT token and returns the claims
func (s *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	if len(s.jwtSecret) == 0 {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check expiry
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}
