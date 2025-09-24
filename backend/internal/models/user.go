package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	UserRoleUser       UserRole = "user"
	UserRoleAdmin      UserRole = "admin"
	UserRoleSuperAdmin UserRole = "superadmin"
)

// AccountType represents the type of user account
type AccountType string

const (
	AccountTypeGuest AccountType = "guest"
	AccountTypeFull  AccountType = "full"
)

// User represents a user in the system
type User struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Email        string         `json:"email" gorm:"uniqueIndex;type:varchar(255)"`
	DisplayName  string         `json:"displayName" gorm:"type:varchar(100);not null"`
	AvatarURL    string         `json:"avatarUrl" gorm:"type:varchar(500)"`
	AboutMe      string         `json:"aboutMe" gorm:"type:text"`
	AccountType  AccountType    `json:"accountType" gorm:"type:varchar(20);not null;default:'full'"`
	Role         UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	PasswordHash string         `json:"-" gorm:"type:varchar(255)"` // Hidden from JSON
	IsActive     bool           `json:"isActive" gorm:"default:true"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete support
}

// NewUser creates a new User with default values
func NewUser(displayName string) (*User, error) {
	if displayName == "" {
		return nil, fmt.Errorf("display name is required")
	}

	user := &User{
		ID:          uuid.New().String(),
		DisplayName: displayName,
		AccountType: AccountTypeFull,
		Role:        UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return user, nil
}

// NewGuestUser creates a new guest user
func NewGuestUser(displayName string) (*User, error) {
	user, err := NewUser(displayName)
	if err != nil {
		return nil, err
	}

	user.AccountType = AccountTypeGuest
	return user, nil
}

// Validate validates the user model
func (u *User) Validate() error {
	if u.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}

	if len(u.DisplayName) < 2 {
		return fmt.Errorf("display name must be at least 2 characters")
	}

	if len(u.DisplayName) > 100 {
		return fmt.Errorf("display name must be less than 100 characters")
	}

	// Check for invalid characters in display name
	if strings.ContainsAny(u.DisplayName, `@#$%^&*()+={}[]|\:;"'<>?,./`) {
		return fmt.Errorf("display name contains invalid characters")
	}

	// Validate email for full accounts
	if u.AccountType == AccountTypeFull {
		if u.Email == "" {
			return fmt.Errorf("email is required for full accounts")
		}

		if err := u.validateEmail(); err != nil {
			return err
		}
	}

	// Validate role
	if err := u.validateRole(); err != nil {
		return err
	}

	// Validate account type
	if err := u.validateAccountType(); err != nil {
		return err
	}

	return nil
}

// validateEmail validates the email format
func (u *User) validateEmail() error {
	if u.Email == "" {
		return nil // Email is optional for guest accounts
	}

	if len(u.Email) > 255 {
		return fmt.Errorf("email must be less than 255 characters")
	}

	// Basic email regex validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// validateRole validates the user role
func (u *User) validateRole() error {
	validRoles := map[UserRole]bool{
		UserRoleUser:       true,
		UserRoleAdmin:      true,
		UserRoleSuperAdmin: true,
	}

	if !validRoles[u.Role] {
		return fmt.Errorf("invalid user role: %s", u.Role)
	}

	return nil
}

// validateAccountType validates the account type
func (u *User) validateAccountType() error {
	validAccountTypes := map[AccountType]bool{
		AccountTypeGuest: true,
		AccountTypeFull:  true,
	}

	if !validAccountTypes[u.AccountType] {
		return fmt.Errorf("invalid account type: %s", u.AccountType)
	}

	return nil
}

// HasPassword returns true if the user has a password set
func (u *User) HasPassword() bool {
	return u.PasswordHash != ""
}

// IsGuest returns true if the user is a guest account
func (u *User) IsGuest() bool {
	return u.AccountType == AccountTypeGuest
}

// IsFull returns true if the user is a full account
func (u *User) IsFull() bool {
	return u.AccountType == AccountTypeFull
}

// IsAdmin returns true if the user has admin privileges
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleSuperAdmin
}

// IsSuperAdmin returns true if the user is a super admin
func (u *User) IsSuperAdmin() bool {
	return u.Role == UserRoleSuperAdmin
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}