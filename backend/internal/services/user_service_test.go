package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"breakoutglobe/internal/interfaces"
	"breakoutglobe/internal/models"
	"github.com/stretchr/testify/mock"
)

// UserServiceTestScenario provides scenario-based testing for UserService using established patterns
type UserServiceTestScenario struct {
	t            *testing.T
	service      *UserService
	mockUserRepo *MockUserRepository
}

// newUserServiceTestScenario creates a new UserService test scenario using established patterns
func newUserServiceTestScenario(t *testing.T) *UserServiceTestScenario {
	mockUserRepo := &MockUserRepository{}
	service := NewUserService(mockUserRepo)
	
	return &UserServiceTestScenario{
		t:            t,
		service:      service,
		mockUserRepo: mockUserRepo,
	}
}

// cleanup cleans up test resources
func (s *UserServiceTestScenario) cleanup() {
	s.mockUserRepo.AssertExpectations(s.t)
}

// expectGuestProfileCreationSuccess sets up the repository to successfully create a guest profile
func (s *UserServiceTestScenario) expectGuestProfileCreationSuccess(displayName string) *UserServiceTestScenario {
	s.mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.DisplayName == displayName && user.AccountType == models.AccountTypeGuest
	})).Return(nil)
	
	return s
}

// expectGuestProfileCreationError sets up the repository to return an error during guest profile creation
func (s *UserServiceTestScenario) expectGuestProfileCreationError(displayName string, err error) *UserServiceTestScenario {
	s.mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.DisplayName == displayName
	})).Return(err)
	
	return s
}

// expectUserRetrievalSuccess sets up the repository to successfully retrieve a user
func (s *UserServiceTestScenario) expectUserRetrievalSuccess(userID string, user *models.User) *UserServiceTestScenario {
	s.mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	
	return s
}

// expectUserRetrievalError sets up the repository to return an error during user retrieval
func (s *UserServiceTestScenario) expectUserRetrievalError(userID string, err error) *UserServiceTestScenario {
	s.mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, err)
	
	return s
}

// expectUserUpdateSuccess sets up the repository to successfully update a user
func (s *UserServiceTestScenario) expectUserUpdateSuccess() *UserServiceTestScenario {
	s.mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	
	return s
}

// expectUserUpdateError sets up the repository to return an error during user update
func (s *UserServiceTestScenario) expectUserUpdateError(err error) *UserServiceTestScenario {
	s.mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(err)
	
	return s
}

// createGuestProfile executes guest profile creation and returns the result
func (s *UserServiceTestScenario) createGuestProfile(displayName string) (*models.User, error) {
	return s.service.CreateGuestProfile(context.Background(), displayName)
}

// uploadAvatar executes avatar upload and returns the result
func (s *UserServiceTestScenario) uploadAvatar(userID string, filename string, fileData []byte) (*models.User, error) {
	return s.service.UploadAvatar(context.Background(), userID, filename, fileData)
}

// TestUserService demonstrates the established test infrastructure patterns

func TestUserService_CreateGuestProfile_Success(t *testing.T) {
	// Setup using established infrastructure patterns
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	displayName := "Test User"
	
	// Configure expectations - fluent and readable
	scenario.expectGuestProfileCreationSuccess(displayName)

	// Execute and verify - business intent is clear
	user, err := scenario.createGuestProfile(displayName)

	// Use assertions focused on business logic
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Errorf("Expected user to be created, got nil")
	}
	if user.DisplayName != displayName {
		t.Errorf("Expected display name '%s', got '%s'", displayName, user.DisplayName)
	}
	if user.AccountType != models.AccountTypeGuest {
		t.Errorf("Expected account type 'guest', got '%s'", user.AccountType)
	}
	if user.Role != models.UserRoleUser {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}
	if !user.IsActive {
		t.Errorf("Expected user to be active")
	}
	if user.ID == "" {
		t.Errorf("Expected user ID to be generated")
	}
}

func TestUserService_CreateGuestProfile_ValidationError(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	// No repository expectations needed - validation happens before repository call

	// Execute with invalid display name
	user, err := scenario.createGuestProfile("AB") // Too short

	// Verify validation error
	if err == nil {
		t.Errorf("Expected validation error, got nil")
	}
	if user != nil {
		t.Errorf("Expected no user to be created, got %v", user)
	}
	if err != nil && !contains(err.Error(), "display name must be at least 3 characters") {
		t.Errorf("Expected validation error message, got: %s", err.Error())
	}
}

func TestUserService_CreateGuestProfile_RepositoryError(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	displayName := "Test User"
	repositoryError := errors.New("database connection failed")
	
	// Repository error scenario
	scenario.expectGuestProfileCreationError(displayName, repositoryError)

	// Execute request expecting repository error
	user, err := scenario.createGuestProfile(displayName)

	// Verify error handling
	if err == nil {
		t.Errorf("Expected repository error, got nil")
	}
	if user != nil {
		t.Errorf("Expected no user to be created, got %v", user)
	}
	if err != nil && !contains(err.Error(), "failed to save user") {
		t.Errorf("Expected wrapped error message, got: %s", err.Error())
	}
}

func TestUserService_UploadAvatar_Success(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	userID := "user-123"
	filename := "avatar.jpg"
	fileData := []byte("fake-jpeg-data")
	
	// Create existing user
	existingUser := &models.User{
		ID:          userID,
		DisplayName: "Test User",
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Configure expectations
	scenario.expectUserRetrievalSuccess(userID, existingUser).
		expectUserUpdateSuccess()

	// Execute avatar upload
	user, err := scenario.uploadAvatar(userID, filename, fileData)

	// Verify success
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Errorf("Expected user to be returned, got nil")
	}
	if user.AvatarURL == "" {
		t.Errorf("Expected avatar URL to be set")
	}
	if !contains(user.AvatarURL, "/api/users/avatar/") {
		t.Errorf("Expected avatar URL to contain '/api/users/avatar/', got: %s", user.AvatarURL)
	}
	if !contains(user.AvatarURL, ".jpg") {
		t.Errorf("Expected avatar URL to contain file extension, got: %s", user.AvatarURL)
	}
}

func TestUserService_UploadAvatar_UserNotFound(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	userID := "non-existent-user"
	filename := "avatar.jpg"
	fileData := []byte("fake-jpeg-data")
	
	// User not found scenario
	scenario.expectUserRetrievalError(userID, errors.New("user not found"))

	// Execute request expecting user not found error
	user, err := scenario.uploadAvatar(userID, filename, fileData)

	// Verify error handling
	if err == nil {
		t.Errorf("Expected user not found error, got nil")
	}
	if user != nil {
		t.Errorf("Expected no user to be returned, got %v", user)
	}
	if err != nil && !contains(err.Error(), "failed to get user") {
		t.Errorf("Expected wrapped error message, got: %s", err.Error())
	}
}

func TestUserService_UploadAvatar_UpdateError(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	userID := "user-123"
	filename := "avatar.jpg"
	fileData := []byte("fake-jpeg-data")
	
	// Create existing user
	existingUser := &models.User{
		ID:          userID,
		DisplayName: "Test User",
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	updateError := errors.New("database update failed")
	
	// Configure expectations
	scenario.expectUserRetrievalSuccess(userID, existingUser).
		expectUserUpdateError(updateError)

	// Execute request expecting update error
	user, err := scenario.uploadAvatar(userID, filename, fileData)

	// Verify error handling
	if err == nil {
		t.Errorf("Expected update error, got nil")
	}
	if user != nil {
		t.Errorf("Expected no user to be returned, got %v", user)
	}
	if err != nil && !contains(err.Error(), "failed to update user") {
		t.Errorf("Expected wrapped error message, got: %s", err.Error())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(substr) == 0 || 
		indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// MockUserRepository for testing - follows established patterns
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements the interface
var _ interfaces.UserRepositoryInterface = (*MockUserRepository)(nil)

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}