package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"notebox/internal/note"
)

type BoxRepository struct {
	q *Queries
}

func NewBoxRepository(db *sql.DB) *BoxRepository {
	return &BoxRepository{q: New(db)}
}

func (r *BoxRepository) FindAll(ctx context.Context) ([]note.Box, error) {
	boxes, err := r.q.ListAllBoxes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]note.Box, len(boxes))
	for i, b := range boxes {
		result[i] = note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path, Active: !b.DeletedAt.Valid}
	}
	return result, nil
}

func (r *BoxRepository) FindAllActive(ctx context.Context) ([]note.Box, error) {
	boxes, err := r.q.ListActiveBoxes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]note.Box, len(boxes))
	for i, b := range boxes {
		result[i] = note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path, Active: true}
	}
	return result, nil
}

func (r *BoxRepository) FindInactiveBoxes(ctx context.Context) ([]note.Box, error) {
	boxes, err := r.q.ListInactiveBoxes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]note.Box, len(boxes))
	for i, b := range boxes {
		result[i] = note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path, Active: false}
	}
	return result, nil
}

func (r *BoxRepository) PruneBoxes(ctx context.Context) error {
	return r.q.PruneBoxes(ctx)
}

func (r *BoxRepository) CreateBox(ctx context.Context, box note.Box) (note.Box, error) {
	b, err := r.q.CreateBox(ctx, CreateBoxParams{Title: box.Title, Path: box.Path})
	if err != nil {
		// the upsert returns no row when the path conflicts with an active box
		if errors.Is(err, sql.ErrNoRows) {
			return note.Box{}, fmt.Errorf("box already exists: %s", box.Path)
		}
		return note.Box{}, err
	}
	return note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path, Active: true}, nil
}

func (r *BoxRepository) UpdateBox(ctx context.Context, box note.Box) (note.Box, error) {
	b, err := r.q.UpdateBox(ctx, UpdateBoxParams{ID: int64(box.ID), Title: box.Title, Path: box.Path})
	if err != nil {
		return note.Box{}, err
	}
	return note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path, Active: !b.DeletedAt.Valid}, nil
}

func (r *BoxRepository) DeleteBox(ctx context.Context, box note.Box) error {
	return r.q.DeleteBox(ctx, int64(box.ID))
}
