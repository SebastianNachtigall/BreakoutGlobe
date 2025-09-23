package testdata

import (
	"fmt"
	"os"
	"time"

	"breakoutglobe/internal/database"
	"breakoutglobe/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TestDB provides database integration testing utilities
type TestDB struct {
	DB     *gorm.DB
	dbName string
	t      TestingT
}

// TestDBConfig holds configuration for test database setup
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	SSLMode  string
}

// DefaultTestDBConfig returns default configuration for test database
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
	}
}

// Setup creates a new isolated test database for integration testing
func Setup(t TestingT) *TestDB {
	t.Helper()
	
	config := DefaultTestDBConfig()
	
	// Generate unique database name for this test
	dbName := fmt.Sprintf("test_%s_%d", 
		sanitizeTestName(getTestName(t)), 
		time.Now().UnixNano())
	
	// Create the test database
	if err := createTestDatabase(config, dbName); err != nil {
		t.Errorf("Failed to create test database: %v", err)
		return nil
	}
	
	// Connect to the test database
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, dbName, config.SSLMode)
	
	db, err := database.NewConnection(databaseURL)
	if err != nil {
		// Cleanup the database if connection fails
		dropTestDatabase(config, dbName)
		t.Errorf("Failed to connect to test database: %v", err)
		return nil
	}
	
	testDB := &TestDB{
		DB:     db,
		dbName: dbName,
		t:      t,
	}
	
	// Run migrations
	if err := testDB.RunMigrations(); err != nil {
		testDB.Cleanup()
		t.Errorf("Failed to run migrations: %v", err)
		return nil
	}
	
	// Register cleanup
	if cleaner, ok := t.(interface{ Cleanup(func()) }); ok {
		cleaner.Cleanup(testDB.Cleanup)
	}
	
	return testDB
}

// RunMigrations executes database migrations for test database
func (tdb *TestDB) RunMigrations() error {
	if tdb.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Auto-migrate core models
	return tdb.DB.AutoMigrate(
		&models.Session{},
		&models.POI{},
		// Add other models as needed
	)
}

// Cleanup closes the database connection and drops the test database
func (tdb *TestDB) Cleanup() {
	if tdb.DB != nil {
		database.CloseConnection(tdb.DB)
	}
	
	if tdb.dbName != "" {
		config := DefaultTestDBConfig()
		if err := dropTestDatabase(config, tdb.dbName); err != nil {
			// Log error but don't fail the test
			fmt.Printf("Warning: Failed to cleanup test database %s: %v\n", tdb.dbName, err)
		}
	}
}

// SeedFixtures loads test data into the database
func (tdb *TestDB) SeedFixtures(fixtures ...interface{}) error {
	if tdb.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	for _, fixture := range fixtures {
		if err := tdb.DB.Create(fixture).Error; err != nil {
			return fmt.Errorf("failed to seed fixture: %w", err)
		}
	}
	
	return nil
}

// Clear removes all data from specified tables
func (tdb *TestDB) Clear(models ...interface{}) error {
	if tdb.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Clear in reverse order to handle foreign key constraints
	for i := len(models) - 1; i >= 0; i-- {
		if err := tdb.DB.Unscoped().Delete(models[i], "1=1").Error; err != nil {
			return fmt.Errorf("failed to clear table: %w", err)
		}
	}
	
	return nil
}

// Transaction executes a function within a database transaction
func (tdb *TestDB) Transaction(fn func(*gorm.DB) error) error {
	if tdb.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	return tdb.DB.Transaction(fn)
}

// Helper functions

func createTestDatabase(config *TestDBConfig, dbName string) error {
	// Connect to postgres database to create the test database
	adminURL := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.SSLMode)
	
	adminDB, err := database.NewConnection(adminURL)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer database.CloseConnection(adminDB)
	
	// Create the test database (quote the name to handle special characters)
	sql := fmt.Sprintf("CREATE DATABASE %q", dbName)
	if err := adminDB.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}
	
	// Verify database was created
	var exists bool
	checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s')", dbName)
	if err := adminDB.Raw(checkSQL).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to verify database creation: %w", err)
	}
	
	if !exists {
		return fmt.Errorf("database %s was not created successfully", dbName)
	}
	
	// Small delay to ensure database is ready
	time.Sleep(200 * time.Millisecond)
	
	return nil
}

func dropTestDatabase(config *TestDBConfig, dbName string) error {
	// Connect to postgres database to drop the test database
	adminURL := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.SSLMode)
	
	adminDB, err := database.NewConnection(adminURL)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer database.CloseConnection(adminDB)
	
	// Terminate connections to the database
	terminateSQL := fmt.Sprintf(`
		SELECT pg_terminate_backend(pid) 
		FROM pg_stat_activity 
		WHERE datname = '%s' AND pid <> pg_backend_pid()`, dbName)
	adminDB.Exec(terminateSQL)
	
	// Drop the test database
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	if err := adminDB.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}
	
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func sanitizeTestName(name string) string {
	// Replace invalid characters for database names
	sanitized := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sanitized += string(r)
		} else {
			sanitized += "_"
		}
	}
	
	// Ensure it starts with a letter
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "t_" + sanitized
	}
	
	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	
	return sanitized
}

func getTestName(t TestingT) string {
	// Try to get test name using reflection
	if tester, ok := t.(interface{ Name() string }); ok {
		return tester.Name()
	}
	
	// Fallback to UUID
	return uuid.New().String()[:8]
}

// Database fixture builders for integration tests

// DatabasePOIFixture creates POI test data for database integration tests
type DatabasePOIFixture struct {
	poi *models.POI
}

// NewDatabasePOI creates a new POI fixture builder for database tests
func NewDatabasePOI() *DatabasePOIFixture {
	return &DatabasePOIFixture{
		poi: &models.POI{
			ID:              uuid.New().String(),
			MapID:           uuid.New().String(),
			Name:            "Test POI",
			Description:     "Test POI Description",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       uuid.New().String(),
			MaxParticipants: 10,
			CreatedAt:       time.Now(),
		},
	}
}

// WithID sets the POI ID
func (f *DatabasePOIFixture) WithID(id string) *DatabasePOIFixture {
	f.poi.ID = id
	return f
}

// WithMapID sets the map ID
func (f *DatabasePOIFixture) WithMapID(mapID string) *DatabasePOIFixture {
	f.poi.MapID = mapID
	return f
}

// WithName sets the POI name
func (f *DatabasePOIFixture) WithName(name string) *DatabasePOIFixture {
	f.poi.Name = name
	return f
}

// WithPosition sets the POI position
func (f *DatabasePOIFixture) WithPosition(lat, lng float64) *DatabasePOIFixture {
	f.poi.Position = models.LatLng{Lat: lat, Lng: lng}
	return f
}

// WithCreator sets the creator ID
func (f *DatabasePOIFixture) WithCreator(creatorID string) *DatabasePOIFixture {
	f.poi.CreatedBy = creatorID
	return f
}

// Build returns the POI model
func (f *DatabasePOIFixture) Build() *models.POI {
	return f.poi
}

// DatabaseSessionFixture creates Session test data for database integration tests
type DatabaseSessionFixture struct {
	session *models.Session
}

// NewDatabaseSession creates a new Session fixture builder for database tests
func NewDatabaseSession() *DatabaseSessionFixture {
	now := time.Now()
	return &DatabaseSessionFixture{
		session: &models.Session{
			ID:         uuid.New().String(),
			UserID:     uuid.New().String(),
			MapID:      uuid.New().String(),
			AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedAt:  now,
			LastActive: now,
			IsActive:   true,
		},
	}
}

// WithID sets the session ID
func (f *DatabaseSessionFixture) WithID(id string) *DatabaseSessionFixture {
	f.session.ID = id
	return f
}

// WithUserID sets the user ID
func (f *DatabaseSessionFixture) WithUserID(userID string) *DatabaseSessionFixture {
	f.session.UserID = userID
	return f
}

// WithMapID sets the map ID
func (f *DatabaseSessionFixture) WithMapID(mapID string) *DatabaseSessionFixture {
	f.session.MapID = mapID
	return f
}

// WithPosition sets the avatar position
func (f *DatabaseSessionFixture) WithPosition(lat, lng float64) *DatabaseSessionFixture {
	f.session.AvatarPos = models.LatLng{Lat: lat, Lng: lng}
	return f
}

// WithActive sets the active status
func (f *DatabaseSessionFixture) WithActive(active bool) *DatabaseSessionFixture {
	f.session.IsActive = active
	return f
}

// Build returns the Session model
func (f *DatabaseSessionFixture) Build() *models.Session {
	return f.session
}