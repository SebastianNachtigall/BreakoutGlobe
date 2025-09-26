package testdata

import (
	"time"

	"breakoutglobe/internal/models"

	"github.com/google/uuid"
)

// TestFixtures provides common test data for integration tests
type TestFixtures struct {
	db *TestDB
}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures(db *TestDB) *TestFixtures {
	return &TestFixtures{db: db}
}

// CreateTestUser creates a test user in the database
func (f *TestFixtures) CreateTestUser(id, displayName string) *models.User {
	if id == "" {
		id = uuid.New().String()
	}
	if displayName == "" {
		displayName = "Test User"
	}

	user := &models.User{
		ID:          id,
		DisplayName: displayName,
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := f.db.DB.Create(user).Error
	if err != nil {
		panic("Failed to create test user: " + err.Error())
	}

	return user
}

// CreateTestMap creates a test map in the database
func (f *TestFixtures) CreateTestMap(id, name, createdBy string) *models.Map {
	if id == "" {
		id = uuid.New().String()
	}
	if name == "" {
		name = "Test Map"
	}
	if createdBy == "" {
		// Create a default user if none provided
		user := f.CreateTestUser("", "Map Creator")
		createdBy = user.ID
	}

	testMap := &models.Map{
		ID:          id,
		Name:        name,
		Description: "Test map for integration tests",
		CreatedBy:   createdBy,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := f.db.DB.Create(testMap).Error
	if err != nil {
		panic("Failed to create test map: " + err.Error())
	}

	return testMap
}

// CreateTestPOI creates a test POI in the database
func (f *TestFixtures) CreateTestPOI(id, name, mapID, createdBy string) *models.POI {
	if id == "" {
		id = uuid.New().String()
	}
	if name == "" {
		name = "Test POI"
	}
	if mapID == "" {
		// Create a default map if none provided
		testMap := f.CreateTestMap("", "Default Map", "")
		mapID = testMap.ID
	}
	if createdBy == "" {
		// Create a default user if none provided
		user := f.CreateTestUser("", "POI Creator")
		createdBy = user.ID
	}

	poi := &models.POI{
		ID:              id,
		MapID:           mapID,
		Name:            name,
		Description:     "Test POI for integration tests",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       createdBy,
		MaxParticipants: 10,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := f.db.DB.Create(poi).Error
	if err != nil {
		panic("Failed to create test POI: " + err.Error())
	}

	return poi
}

// CreateTestSession creates a test session in the database
func (f *TestFixtures) CreateTestSession(id, userID, mapID string) *models.Session {
	if id == "" {
		id = uuid.New().String()
	}
	if userID == "" {
		// Create a default user if none provided
		user := f.CreateTestUser("", "Session User")
		userID = user.ID
	}
	if mapID == "" {
		// Create a default map if none provided
		testMap := f.CreateTestMap("", "Session Map", "")
		mapID = testMap.ID
	}

	session := &models.Session{
		ID:         id,
		UserID:     userID,
		MapID:      mapID,
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
	}

	err := f.db.DB.Create(session).Error
	if err != nil {
		panic("Failed to create test session: " + err.Error())
	}

	return session
}

// SetupBasicTestData creates a basic set of test data for integration tests
func (f *TestFixtures) SetupBasicTestData() *BasicTestData {
	// Create test users
	user1 := f.CreateTestUser("test-user-1", "Test User 1")
	user2 := f.CreateTestUser("test-user-2", "Test User 2")
	mapCreator := f.CreateTestUser("map-creator", "Map Creator")

	// Create test maps
	testMap := f.CreateTestMap("map-test", "Test Map", mapCreator.ID)
	otherMap := f.CreateTestMap("map-other", "Other Map", mapCreator.ID)

	return &BasicTestData{
		Users: []*models.User{user1, user2, mapCreator},
		Maps:  []*models.Map{testMap, otherMap},
	}
}

// BasicTestData contains basic test data for integration tests
type BasicTestData struct {
	Users []*models.User
	Maps  []*models.Map
}

// GetUser returns a user by index
func (btd *BasicTestData) GetUser(index int) *models.User {
	if index < 0 || index >= len(btd.Users) {
		return btd.Users[0] // Return first user as default
	}
	return btd.Users[index]
}

// GetMap returns a map by index
func (btd *BasicTestData) GetMap(index int) *models.Map {
	if index < 0 || index >= len(btd.Maps) {
		return btd.Maps[0] // Return first map as default
	}
	return btd.Maps[index]
}

// GetTestMap returns the main test map
func (btd *BasicTestData) GetTestMap() *models.Map {
	for _, m := range btd.Maps {
		if m.ID == "map-test" {
			return m
		}
	}
	return btd.Maps[0]
}

// GetOtherMap returns the other test map
func (btd *BasicTestData) GetOtherMap() *models.Map {
	for _, m := range btd.Maps {
		if m.ID == "map-other" {
			return m
		}
	}
	return btd.Maps[1]
}