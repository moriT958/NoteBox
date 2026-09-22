package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"notebox/internal/box"
	"notebox/internal/config"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

const dbFileName = "notebox.sqlite"

type SQLiteBoxStore struct {
	q *Queries
}

var _ box.BoxStore = (*SQLiteBoxStore)(nil)

func NewSQLiteBoxStore() (*SQLiteBoxStore, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite3: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteBoxStore{q: New(db)}, nil
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

func toBox(b Box) box.Box {
	return box.Box{ID: b.ID, Title: b.Title, Path: b.Path, Active: b.Active}
}

func (s *SQLiteBoxStore) Set(ctx context.Context, b box.Box) (*box.Box, error) {
	row, err := s.q.UpsertBox(ctx, UpsertBoxParams{
		ID:     b.ID,
		Title:  b.Title,
		Path:   b.Path,
		Active: b.Active,
	})
	if err != nil {
		return nil, err
	}
	result := toBox(row)
	return &result, nil
}

func (s *SQLiteBoxStore) GetByID(ctx context.Context, id string) (*box.Box, error) {
	row, err := s.q.GetBoxByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	result := toBox(row)
	return &result, nil
}

func (s *SQLiteBoxStore) Get(ctx context.Context, opts box.Filter) ([]box.Box, error) {
	var (
		rows []Box
		err  error
	)
	switch {
	case opts.Path != nil:
		rows, err = s.q.ListBoxesByPath(ctx, *opts.Path)
	case opts.Active != nil:
		rows, err = s.q.ListBoxesByActive(ctx, *opts.Active)
	default:
		rows, err = s.q.ListBoxes(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]box.Box, len(rows))
	for i, r := range rows {
		result[i] = toBox(r)
	}
	return result, nil
}

func (s *SQLiteBoxStore) Del(ctx context.Context, id string) error {
	return s.q.DeleteBox(ctx, id)
}
