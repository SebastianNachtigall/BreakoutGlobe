package services

import (
	"context"
	"errors"
	"fmt"
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
	mockStorage  *MockFileStorage
}

// newUserServiceTestScenario creates a new UserService test scenario using established patterns
func newUserServiceTestScenario(t *testing.T) *UserServiceTestScenario {
	mockUserRepo := &MockUserRepository{}
	mockStorage := &MockFileStorage{}
	service := NewUserService(mockUserRepo, mockStorage)
	
	return &UserServiceTestScenario{
		t:            t,
		service:      service,
		mockUserRepo: mockUserRepo,
		mockStorage:  mockStorage,
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

// expectStorageUploadSuccess sets up the storage to successfully upload a file
func (s *UserServiceTestScenario) expectStorageUploadSuccess(userID string, filename string, fileData []byte) *UserServiceTestScenario {
	fileKey := fmt.Sprintf("avatars/%s_%d.jpg", userID, time.Now().Unix())
	avatarURL := fmt.Sprintf("http://localhost:8080/uploads/%s", fileKey)
	
	s.mockStorage.On("GenerateUniqueKey", "avatars", userID, filename).Return(fileKey)
	s.mockStorage.On("UploadFile", mock.Anything, fileKey, fileData, "image/jpeg").Return(avatarURL, nil)
	
	return s
}

// expectStorageUploadError sets up the storage to return an error during upload
func (s *UserServiceTestScenario) expectStorageUploadError(userID string, filename string, fileData []byte, err error) *UserServiceTestScenario {
	fileKey := fmt.Sprintf("avatars/%s_%d.jpg", userID, time.Now().Unix())
	
	s.mockStorage.On("GenerateUniqueKey", "avatars", userID, filename).Return(fileKey)
	s.mockStorage.On("UploadFile", mock.Anything, fileKey, fileData, "image/jpeg").Return("", err)
	
	return s
}

// expectStorageDeleteOnFailure sets up the storage to handle cleanup on failure
func (s *UserServiceTestScenario) expectStorageDeleteOnFailure() *UserServiceTestScenario {
	s.mockStorage.On("DeleteFile", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	
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
		expectStorageUploadSuccess(userID, filename, fileData).
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
	if user.AvatarURL == nil || *user.AvatarURL == "" {
		t.Errorf("Expected avatar URL to be set")
	}
	if user.AvatarURL != nil && !contains(*user.AvatarURL, "/uploads/") {
		t.Errorf("Expected avatar URL to contain '/uploads/', got: %s", *user.AvatarURL)
	}
	if user.AvatarURL != nil && !contains(*user.AvatarURL, ".jpg") {
		t.Errorf("Expected avatar URL to contain file extension, got: %s", *user.AvatarURL)
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
		expectStorageUploadSuccess(userID, filename, fileData).
		expectStorageDeleteOnFailure().
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

// GetUser tests - Task 7

func TestUserService_GetUser_Success(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()
	
	userID := "test-user-123"
	expectedUser := &models.User{
		ID:          userID,
		DisplayName: "Test User",
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	
	// Setup expectations using established patterns
	scenario.expectUserRetrievalSuccess(userID, expectedUser)
	
	// Execute operation
	user, err := scenario.service.GetUser(context.Background(), userID)
	
	// Verify results using established patterns
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	
	if user == nil {
		t.Error("Expected user, got nil")
		return
	}
	
	if user.ID != expectedUser.ID {
		t.Errorf("Expected ID %s, got %s", expectedUser.ID, user.ID)
	}
	if user.DisplayName != expectedUser.DisplayName {
		t.Errorf("Expected DisplayName %s, got %s", expectedUser.DisplayName, user.DisplayName)
	}
}

func TestUserService_GetUser_UserNotFound(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()
	
	userID := "non-existent-user"
	
	// Setup expectations for user not found
	scenario.expectUserRetrievalError(userID, errors.New("record not found"))
	
	// Execute operation
	user, err := scenario.service.GetUser(context.Background(), userID)
	
	// Verify error response
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}
	
	if user != nil {
		t.Error("Expected nil user, got user")
	}
	
	if err.Error() != "user not found" {
		t.Errorf("Expected 'user not found' error, got %s", err.Error())
	}
}

func TestUserService_GetUser_EmptyUserID(t *testing.T) {
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()
	
	// Execute operation with empty user ID
	user, err := scenario.service.GetUser(context.Background(), "")
	
	// Verify error response
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}
	
	if user != nil {
		t.Error("Expected nil user, got user")
	}
	
	if err.Error() != "user ID cannot be empty" {
		t.Errorf("Expected 'user ID cannot be empty' error, got %s", err.Error())
	}
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

func (m *MockUserRepository) ClearAllUsers(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestUserService_CreateGuestProfile_WithAboutMe(t *testing.T) {
	// This test demonstrates the current limitation - aboutMe is not supported
	// Following TDD, this test should fail first, then we implement the feature
	scenario := newUserServiceTestScenario(t)
	defer scenario.cleanup()

	displayName := "Test User"
	aboutMe := "Hello, I am a test user!"
	
	// Configure expectations for aboutMe support
	scenario.mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.DisplayName == displayName && 
			user.AccountType == models.AccountTypeGuest &&
			user.AboutMe != nil && *user.AboutMe == aboutMe
	})).Return(nil)

	// This should create a guest profile with aboutMe
	// Currently this will fail because CreateGuestProfile doesn't accept aboutMe
	user, err := scenario.service.CreateGuestProfileWithAboutMe(context.Background(), displayName, aboutMe)

	// Verify aboutMe is set correctly
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Errorf("Expected user to be created, got nil")
	}
	if user.AboutMe == nil || *user.AboutMe != aboutMe {
		t.Errorf("Expected aboutMe '%s', got %v", aboutMe, user.AboutMe)
	}
}

// MockFileStorage provides a mock implementation of FileStorage for testing
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, key, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) DeleteFile(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockFileStorage) GetFileURL(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockFileStorage) FileExists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockFileStorage) GenerateUniqueKey(prefix, userID, originalFilename string) string {
	args := m.Called(prefix, userID, originalFilename)
	return args.String(0)
}