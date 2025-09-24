package database

import (
	"fmt"

	"breakoutglobe/internal/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.User{}, // Must be first since other models reference it
		&models.Map{},
		&models.Session{},
		&models.POI{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	// Create custom indexes
	if err := CreateIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// CreateIndexes creates custom database indexes for performance
func CreateIndexes(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Session indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id)").Error; err != nil {
		return fmt.Errorf("failed to create sessions user_id index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_map_id ON sessions (map_id)").Error; err != nil {
		return fmt.Errorf("failed to create sessions map_id index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_last_active ON sessions (last_active)").Error; err != nil {
		return fmt.Errorf("failed to create sessions last_active index: %w", err)
	}

	// POI indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pois_map_id ON pois (map_id)").Error; err != nil {
		return fmt.Errorf("failed to create pois map_id index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pois_created_by ON pois (created_by)").Error; err != nil {
		return fmt.Errorf("failed to create pois created_by index: %w", err)
	}

	// Spatial index for POI positions (for efficient location-based queries)
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pois_position ON pois (position_lat, position_lng)").Error; err != nil {
		return fmt.Errorf("failed to create pois position index: %w", err)
	}

	// Map indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_maps_created_by ON maps (created_by)").Error; err != nil {
		return fmt.Errorf("failed to create maps created_by index: %w", err)
	}

	return nil
}

// DropAllTables drops all application tables (useful for testing)
func DropAllTables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Drop tables in reverse dependency order
	tables := []interface{}{
		&models.POI{},     // Has foreign key to maps and users
		&models.Session{}, // Has foreign key to maps and users
		&models.Map{},     // Has foreign key to users
		&models.User{},    // Base table
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	return nil
}

// RollbackMigrations rolls back all migrations (drops all tables)
func RollbackMigrations(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	return DropAllTables(db)
}

// GetMigrationStatus returns information about the current migration status
func GetMigrationStatus(db *gorm.DB) (map[string]bool, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	status := make(map[string]bool)

	// Check if tables exist
	status["users"] = db.Migrator().HasTable(&models.User{})
	status["maps"] = db.Migrator().HasTable(&models.Map{})
	status["sessions"] = db.Migrator().HasTable(&models.Session{})
	status["pois"] = db.Migrator().HasTable(&models.POI{})

	// Check if indexes exist
	status["idx_sessions_user_id"] = db.Migrator().HasIndex(&models.Session{}, "idx_sessions_user_id")
	status["idx_sessions_map_id"] = db.Migrator().HasIndex(&models.Session{}, "idx_sessions_map_id")
	status["idx_sessions_last_active"] = db.Migrator().HasIndex(&models.Session{}, "idx_sessions_last_active")
	status["idx_pois_map_id"] = db.Migrator().HasIndex(&models.POI{}, "idx_pois_map_id")
	status["idx_pois_created_by"] = db.Migrator().HasIndex(&models.POI{}, "idx_pois_created_by")
	status["idx_pois_position"] = db.Migrator().HasIndex(&models.POI{}, "idx_pois_position")
	status["idx_maps_created_by"] = db.Migrator().HasIndex(&models.Map{}, "idx_maps_created_by")

	return status, nil
}