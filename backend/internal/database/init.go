package database

import (
	"gorm.io/gorm"
)

// Initialize sets up the database connection and runs migrations
func Initialize(databaseURL string) (*gorm.DB, error) {
	// Create connection
	db, err := NewConnection(databaseURL)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		CloseConnection(db) // Clean up on error
		return nil, err
	}

	return db, nil
}

// InitializeWithoutMigrations sets up the database connection without running migrations
func InitializeWithoutMigrations(databaseURL string) (*gorm.DB, error) {
	return NewConnection(databaseURL)
}

// MustInitialize initializes the database and panics on error (useful for main.go)
func MustInitialize(databaseURL string) *gorm.DB {
	db, err := Initialize(databaseURL)
	if err != nil {
		panic("Failed to initialize database: " + err.Error())
	}
	return db
}