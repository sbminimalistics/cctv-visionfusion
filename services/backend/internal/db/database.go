package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBWrapper encapsulates the GORM instance for clean dependency injection
type DBWrapper struct {
	ORM *gorm.DB
}

// Connect initializes and returns the GORM database engine
func Connect(host string, port int, user, password, dbname string) (*DBWrapper, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)

	fmt.Println("connecting to postgres via GORM")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres via GORM: %w", err)
	}

	fmt.Println("successfully connected to postgres via GORM")

	// Access the underlying standard sql.DB pool to tweak performance limits
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return &DBWrapper{ORM: db}, nil
}
