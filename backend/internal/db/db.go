package db

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Connect opens a connection to MariaDB and pings it.
func Connect(dsn string) (*sqlx.DB, error) {
	database, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	// Retry a few times to allow MariaDB to start up in Docker.
	for i := 0; i < 10; i++ {
		if err = database.Ping(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return database, nil
}

// Migrate runs the schema DDL (idempotent).
func Migrate(database *sqlx.DB) error {
	_, err := database.Exec(`
CREATE TABLE IF NOT EXISTS games (
  id         CHAR(36) PRIMARY KEY,
  mode       ENUM('singleplayer','multiplayer') NOT NULL,
  status     ENUM('waiting','in_progress','finished') NOT NULL DEFAULT 'waiting',
  room_code  VARCHAR(8) UNIQUE,
  white_nick VARCHAR(50) NOT NULL,
  black_nick VARCHAR(50),
  ai_level   ENUM('easy','medium','hard'),
  turn       ENUM('white','black') NOT NULL DEFAULT 'white',
  winner     ENUM('white','black'),
  board      JSON NOT NULL,
  move_num   INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)`)
	if err != nil {
		return fmt.Errorf("create games table: %w", err)
	}

	_, err = database.Exec(`
CREATE TABLE IF NOT EXISTS moves (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  game_id    CHAR(36) NOT NULL,
  player     ENUM('white','black') NOT NULL,
  move_num   INT NOT NULL,
  from_pos   VARCHAR(2) NOT NULL,
  path       JSON NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (game_id) REFERENCES games(id)
)`)
	if err != nil {
		return fmt.Errorf("create moves table: %w", err)
	}

	_, err = database.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS white_score INT NOT NULL DEFAULT 0`)
	if err != nil {
		return fmt.Errorf("alter games add white_score: %w", err)
	}
	_, err = database.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS black_score INT NOT NULL DEFAULT 0`)
	if err != nil {
		return fmt.Errorf("alter games add black_score: %w", err)
	}
	return nil
}
