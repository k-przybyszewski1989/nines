package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connect opens a connection to MariaDB and pings it.
func Connect(dsn string) (*gorm.DB, error) {
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Retry a few times to allow MariaDB to start up in Docker.
	for i := 0; i < 10; i++ {
		if err = sqlDB.Ping(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return database, nil
}

// Migrate runs schema migrations using GORM AutoMigrate.
func Migrate(database *gorm.DB) error {
	return database.AutoMigrate(&GameRow{}, &MoveRow{})
}
