package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"notebox/internal/config"
)

//go:embed migrations/*.sql
var migrations embed.FS

const dbFileName = "notebox.sqlite"

func NewSQLiteDB() (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite3: %w", err)
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func dbPath() (string, error) {
	if os.Getenv("APP_ENV") == "development" {
		return dbFileName, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}

	dir := filepath.Join(home, config.AppDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config dir: %w", err)
	}

	return filepath.Join(dir, dbFileName), nil
}
