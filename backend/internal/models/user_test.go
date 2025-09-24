package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser_Validate tests User model validation following existing patterns
func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid full account user",
			user: User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "Test User",
				AccountType: AccountTypeFull,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid guest account user",
			user: User{
				ID:          "user-123",
				DisplayName: "Guest User",
				AccountType: AccountTypeGuest,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty display name",
			user: User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "",
				AccountType: AccountTypeFull,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "display name is required",
		},
		{
			name: "display name too short",
			user: User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "A",
				AccountType: AccountTypeFull,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "display name must be at least 2 characters",
		},
		{
			name: "display name too long",
			user: User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "This is a very long display name that exceeds the maximum allowed length of one hundred characters which should fail validation",
				AccountType: AccountTypeFull,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "display name must be less than 100 characters",
		},
		{
			name: "full account without email",
			user: User{
				ID:          "user-123",
				DisplayName: "Test User",
				AccountType: AccountTypeFull,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "email is required for full accounts",
		},
		{
			name: "invalid email format",
			user: User{
				ID:          "user-123",
				Email:       "invalid-email",
				DisplayName: "Test User",
				AccountType: AccountTypeFull,
				Role:        UserRoleUser,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewUser tests User creation function
func TestNewUser(t *testing.T) {
	t.Run("creates user with valid display name", func(t *testing.T) {
		user, err := NewUser("Test User")
		require.NoError(t, err)
		
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "Test User", user.DisplayName)
		assert.Equal(t, AccountTypeFull, user.AccountType)
		assert.Equal(t, UserRoleUser, user.Role)
		assert.True(t, user.IsActive)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("fails with empty display name", func(t *testing.T) {
		user, err := NewUser("")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "display name is required")
	})
}

// TestNewGuestUser tests guest user creation
func TestNewGuestUser(t *testing.T) {
	t.Run("creates guest user", func(t *testing.T) {
		user, err := NewGuestUser("Guest User")
		require.NoError(t, err)
		
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "Guest User", user.DisplayName)
		assert.Equal(t, AccountTypeGuest, user.AccountType)
		assert.Equal(t, UserRoleUser, user.Role)
		assert.True(t, user.IsActive)
		assert.Empty(t, user.Email) // Guests don't have email by default
	})
}

// TestUser_BusinessMethods tests User business logic methods
func TestUser_BusinessMethods(t *testing.T) {
	t.Run("HasPassword", func(t *testing.T) {
		user := User{PasswordHash: ""}
		assert.False(t, user.HasPassword())
		
		user.PasswordHash = "hashed-password"
		assert.True(t, user.HasPassword())
	})

	t.Run("IsGuest", func(t *testing.T) {
		user := User{AccountType: AccountTypeGuest}
		assert.True(t, user.IsGuest())
		
		user.AccountType = AccountTypeFull
		assert.False(t, user.IsGuest())
	})

	t.Run("IsFull", func(t *testing.T) {
		user := User{AccountType: AccountTypeFull}
		assert.True(t, user.IsFull())
		
		user.AccountType = AccountTypeGuest
		assert.False(t, user.IsFull())
	})

	t.Run("IsAdmin", func(t *testing.T) {
		user := User{Role: UserRoleUser}
		assert.False(t, user.IsAdmin())
		
		user.Role = UserRoleAdmin
		assert.True(t, user.IsAdmin())
		
		user.Role = UserRoleSuperAdmin
		assert.True(t, user.IsAdmin())
	})

	t.Run("IsSuperAdmin", func(t *testing.T) {
		user := User{Role: UserRoleUser}
		assert.False(t, user.IsSuperAdmin())
		
		user.Role = UserRoleAdmin
		assert.False(t, user.IsSuperAdmin())
		
		user.Role = UserRoleSuperAdmin
		assert.True(t, user.IsSuperAdmin())
	})
}