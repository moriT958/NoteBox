package database

import (
	"context"
	"database/sql"
	"notebox/internal/core/box"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

func newTestBoxStore(t *testing.T) *SQLiteBoxStore {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &SQLiteBoxStore{q: New(db)}
}

func TestSQLiteBoxStore_Set(t *testing.T) {
	t.Run("Successfully creates a new box.", func(t *testing.T) {
		store := newTestBoxStore(t)
		ctx := context.Background()

		b := box.Box{ID: "id-1", Title: "My Box", Path: "/tmp/my-box", Active: true}
		created, err := store.Set(ctx, b)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if *created != b {
			t.Errorf("created = %+v, want %+v", *created, b)
		}
	})

	t.Run("Successfully updates an existing box, when the id already exists.", func(t *testing.T) {
		store := newTestBoxStore(t)
		ctx := context.Background()

		b := box.Box{ID: "id-1", Title: "Old Title", Path: "/tmp/box", Active: true}
		if _, err := store.Set(ctx, b); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		b.Title = "New Title"
		b.Active = false
		updated, err := store.Set(ctx, b)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if updated.Title != "New Title" || updated.Active {
			t.Errorf("updated = %+v, want title=%q active=false", *updated, "New Title")
		}

		all, err := store.Get(ctx, box.Filter{})
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("expected 1 box after update (not a duplicate insert), got %d", len(all))
		}
	})
}

func TestSQLiteBoxStore_GetByID(t *testing.T) {
	t.Run("Successfully finds a box by id.", func(t *testing.T) {
		store := newTestBoxStore(t)
		ctx := context.Background()

		b := box.Box{ID: "id-1", Title: "My Box", Path: "/tmp/my-box", Active: true}
		if _, err := store.Set(ctx, b); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		got, err := store.GetByID(ctx, "id-1")
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got == nil {
			t.Fatalf("expected box to be found")
		}
		if *got != b {
			t.Errorf("got = %+v, want %+v", *got, b)
		}
	})

	t.Run("Returns nil, when no box with the given id exists.", func(t *testing.T) {
		store := newTestBoxStore(t)

		got, err := store.GetByID(context.Background(), "does-not-exist")
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestSQLiteBoxStore_Get(t *testing.T) {
	setup := func(t *testing.T) *SQLiteBoxStore {
		store := newTestBoxStore(t)
		ctx := context.Background()

		boxes := []box.Box{
			{ID: "1", Title: "A", Path: "/a", Active: true},
			{ID: "2", Title: "B", Path: "/b", Active: false},
			{ID: "3", Title: "C", Path: "/c", Active: true},
		}
		for _, b := range boxes {
			if _, err := store.Set(ctx, b); err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}
		}
		return store
	}

	t.Run("Successfully returns every box, when no filter is given.", func(t *testing.T) {
		store := setup(t)

		got, err := store.Get(context.Background(), box.Filter{})
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("len(got) = %d, want 3", len(got))
		}
	})

	t.Run("Successfully filters by active.", func(t *testing.T) {
		store := setup(t)

		active := true
		got, err := store.Get(context.Background(), box.Filter{Active: &active})
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("len(got) = %d, want 2", len(got))
		}
	})

	t.Run("Successfully filters by path.", func(t *testing.T) {
		store := setup(t)

		path := "/b"
		got, err := store.Get(context.Background(), box.Filter{Path: &path})
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(got) != 1 || got[0].ID != "2" {
			t.Errorf("got = %+v, want single box with id 2", got)
		}
	})
}

func TestSQLiteBoxStore_Del(t *testing.T) {
	t.Run("Successfully deletes a box by id.", func(t *testing.T) {
		store := newTestBoxStore(t)
		ctx := context.Background()

		b := box.Box{ID: "id-1", Title: "Box", Path: "/tmp/box", Active: false}
		if _, err := store.Set(ctx, b); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		if err := store.Del(ctx, "id-1"); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		got, err := store.GetByID(ctx, "id-1")
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got != nil {
			t.Errorf("expected box to be deleted, got %+v", got)
		}
	})
}
