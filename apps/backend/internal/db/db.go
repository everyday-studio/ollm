package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/everyday-studio/ollm/internal/config"
)

func NewDBConnection(cfg *config.Config) (*sql.DB, error) {
	//DSN Create
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, cfg.DB.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database (host: %s, db: %s): %w", cfg.DB.Host, cfg.DB.DBName, err)
	}

	return db, nil
}
