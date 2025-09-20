// Package database provides database connection, migration, and utility functions
package database

import "gorm.io/gorm"

// DB is a type alias for gorm.DB for convenience
type DB = gorm.DB