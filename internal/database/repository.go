package database

import (
	"context"
	"database/sql"

	"notebox/internal/note"
)

type BoxRepository struct {
	q *Queries
}

func NewBoxRepository(db *sql.DB) *BoxRepository {
	return &BoxRepository{q: New(db)}
}

func (r *BoxRepository) FindAll(ctx context.Context) ([]note.Box, error) {
	boxes, err := r.q.ListBoxes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]note.Box, len(boxes))
	for i, b := range boxes {
		result[i] = note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path}
	}
	return result, nil
}

func (r *BoxRepository) CreateBox(ctx context.Context, box note.Box) (note.Box, error) {
	b, err := r.q.CreateBox(ctx, CreateBoxParams{Title: box.Title, Path: box.Path})
	if err != nil {
		return note.Box{}, err
	}
	return note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path}, nil
}

func (r *BoxRepository) UpdateBox(ctx context.Context, box note.Box) (note.Box, error) {
	b, err := r.q.UpdateBox(ctx, UpdateBoxParams{ID: int64(box.ID), Title: box.Title, Path: box.Path})
	if err != nil {
		return note.Box{}, err
	}
	return note.Box{ID: int(b.ID), Title: b.Title, Path: b.Path}, nil
}

func (r *BoxRepository) DeleteBox(ctx context.Context, box note.Box) error {
	return r.q.DeleteBox(ctx, int64(box.ID))
}
