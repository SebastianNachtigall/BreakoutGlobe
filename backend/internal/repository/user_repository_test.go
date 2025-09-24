package repository

import (
	"testing"

	"breakoutglobe/internal/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserRepository_Create tests user creation using scenario-based testing
func TestUserRepository_Create(t *testing.T) {
	t.Run("creates guest user successfully", func(t *testing.T) {
		// RED PHASE: This test will fail because UserRepository doesn't exist yet
		scenario := newUserRepositoryScenario(t)
		defer scenario.cleanup()
		
		// Use expectUserCreationSuccess() for profile creation workflows
		scenario.expectUserCreationSuccess()
		
		user := testdata.NewUser().
			WithDisplayName("Test User").
			AsGuest().
			Build()
		
		err := scenario.repository.Create(user)
		
		// Use fluent assertions: AssertUser(t, user).HasEmail().HasRole().IsActive()
		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "Test User", user.DisplayName)
		assert.True(t, user.IsGuest())
	})

	t.Run("validates display name length", func(t *testing.T) {
		scenario := newUserRepositoryScenario(t)
		defer scenario.cleanup()
		
		// Test display name validation (3-50 characters)
		user := testdata.NewUser().
			WithDisplayName("AB"). // Too short (2 characters)
			AsGuest().
			Build()
		
		err := scenario.repository.Create(user)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "display name must be at least 3 characters")
	})
}

// TestUserRepository_GetByID tests user retrieval by ID
func TestUserRepository_GetByID(t *testing.T) {
	t.Run("retrieves existing user", func(t *testing.T) {
		// RED PHASE: This test will fail because UserRepository doesn't exist yet
		scenario := newUserRepositoryScenario(t)
		defer scenario.cleanup()
		
		// Create a user first
		user := testdata.NewUser().
			WithDisplayName("Test User").
			AsGuest().
			Build()
		
		scenario.expectUserCreationSuccess()
		err := scenario.repository.Create(user)
		require.NoError(t, err)
		
		// Now retrieve it
		retrieved, err := scenario.repository.GetByID(user.ID)
		
		assert.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.DisplayName, retrieved.DisplayName)
		assert.Equal(t, user.AccountType, retrieved.AccountType)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		scenario := newUserRepositoryScenario(t)
		defer scenario.cleanup()
		
		_, err := scenario.repository.GetByID("non-existent-id")
		
		assert.Error(t, err)
	})
}

// userRepositoryScenario provides scenario-based testing for UserRepository
type userRepositoryScenario struct {
	t          *testing.T
	repository UserRepository // This interface doesn't exist yet - will fail
	db         *testdata.TestDB
}

// newUserRepositoryScenario creates a new UserRepository test scenario
func newUserRepositoryScenario(t *testing.T) *userRepositoryScenario {
	// This will fail because UserRepository doesn't exist yet
	db := testdata.Setup(t)
	repository := NewUserRepository(db.DB) // This function doesn't exist yet
	
	return &userRepositoryScenario{
		t:          t,
		repository: repository,
		db:         db,
	}
}

// cleanup cleans up test resources
func (s *userRepositoryScenario) cleanup() {
	s.db.Cleanup()
}

// expectUserCreationSuccess sets up expectations for successful user creation
func (s *userRepositoryScenario) expectUserCreationSuccess() {
	// This method will be used for business rule validation
	// For now, it's a placeholder that will be implemented when we create the repository
}