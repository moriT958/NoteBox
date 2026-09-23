package box

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"uuid"
)

type BoxService struct {
	// TODO:
	// Replace to AppConfig service later.
	// note box base dir path.
	config string
	store  BoxStore
}

func NewBoxService(config string, store BoxStore) *BoxService {
	return &BoxService{
		config: config,
		store:  store,
	}
}

func (s *BoxService) findByPath(ctx context.Context, path string) (*Box, error) {
	boxes, err := s.store.Get(ctx, Filter{Path: &path})
	if err != nil {
		return nil, err
	}
	if len(boxes) == 0 {
		return nil, nil
	}
	return &boxes[0], nil
}

func (s *BoxService) CreateBox(ctx context.Context, title string, path *string) (*Box, error) {
	if title == "" {
		return nil, fmt.Errorf("failed to create box: title is required")
	}

	base := s.config
	if path != nil {
		p, err := normalizePath(*path)
		if err != nil {
			return nil, fmt.Errorf("failed to create box: %w", err)
		}
		base = p
	}
	boxPath := filepath.Join(base, encodeDirName(title))

	existing, err := s.findByPath(ctx, boxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}
	if existing != nil && existing.Active {
		return nil, fmt.Errorf("failed to create box: a box with this path already exists")
	}

	if err := os.MkdirAll(boxPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}

	newBox := Box{ID: uuid.New().String(), Title: title, Path: boxPath, Active: true}
	if existing != nil {
		newBox.ID = existing.ID // revive a soft-deleted box at the same path
	}

	box, err := s.store.Set(ctx, newBox)
	if err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}
	if box == nil {
		return nil, fmt.Errorf("failed to create box: box is nil")
	}

	return box, nil
}

func (s *BoxService) OpenFolderAsBox(ctx context.Context, title, path string) (*Box, error) {
	if title == "" {
		return nil, fmt.Errorf("failed to open folder as box: title is required")
	}

	path, err := normalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open folder as box: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("failed to open folder as box: path must be an existing directory")
	}

	existing, err := s.findByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open folder as box: %w", err)
	}
	if existing != nil && existing.Active {
		return nil, fmt.Errorf("failed to open folder as box: a box with this path already exists")
	}

	newBox := Box{ID: uuid.New().String(), Title: title, Path: path, Active: true}
	if existing != nil {
		newBox.ID = existing.ID
	}

	box, err := s.store.Set(ctx, newBox)
	if err != nil {
		return nil, fmt.Errorf("failed to open folder as box: %w", err)
	}
	if box == nil {
		return nil, fmt.Errorf("failed to open folder as box: box is nil")
	}

	return box, nil
}

// GetBox returns the box with the given ID, or nil if there is none.
// Inactive (removed) boxes are returned too.
func (s *BoxService) GetBox(ctx context.Context, id string) (*Box, error) {
	box, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get box: %w", err)
	}
	return box, nil
}

func (s *BoxService) GetBoxes(ctx context.Context) ([]Box, error) {
	boxes, err := s.store.Get(ctx, Filter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get boxes: %w", err)
	}
	return boxes, nil
}

func (s *BoxService) GetActiveBoxes(ctx context.Context) ([]Box, error) {
	active := true
	boxes, err := s.store.Get(ctx, Filter{Active: &active})
	if err != nil {
		return nil, fmt.Errorf("failed to get active boxes: %w", err)
	}
	return boxes, nil
}

func (s *BoxService) GetInactiveBoxes(ctx context.Context) ([]Box, error) {
	active := false
	boxes, err := s.store.Get(ctx, Filter{Active: &active})
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive boxes: %w", err)
	}
	return boxes, nil
}

func (s *BoxService) RenameBox(ctx context.Context, id string, title string) (*Box, error) {
	box, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to rename box: %w", err)
	}
	if box == nil {
		return nil, fmt.Errorf("failed to rename box: box not found")
	}

	box.Title = title
	b, err := s.store.Set(ctx, *box)
	if err != nil {
		return nil, fmt.Errorf("failed to rename box: %w", err)
	}
	return b, nil
}

func (s *BoxService) ChangeBoxPath(ctx context.Context, id string, path string) (*Box, error) {
	path, err := normalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to change box path: %w", err)
	}

	box, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to change box path: %w", err)
	}
	if box == nil {
		return nil, fmt.Errorf("failed to change box path: box not found")
	}

	box.Path = path
	b, err := s.store.Set(ctx, *box)
	if err != nil {
		return nil, fmt.Errorf("failed to change box path: %w", err)
	}
	return b, nil
}

func (s *BoxService) RemoveBox(ctx context.Context, id string) error {
	box, err := s.store.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove box: %w", err)
	}
	if box == nil {
		return fmt.Errorf("failed to remove box: box not found")
	}

	box.Active = false
	if _, err := s.store.Set(ctx, *box); err != nil {
		return fmt.Errorf("failed to remove box: %w", err)
	}
	return nil
}

func (s *BoxService) PruneBoxes(ctx context.Context) error {
	active := false
	boxes, err := s.store.Get(ctx, Filter{Active: &active})
	if err != nil {
		return fmt.Errorf("failed to prune boxes: %w", err)
	}

	for _, b := range boxes {
		if err := os.RemoveAll(b.Path); err != nil {
			return fmt.Errorf("failed to prune boxes: %w", err)
		}
		if err := s.store.Del(ctx, b.ID); err != nil {
			return fmt.Errorf("failed to prune boxes: %w", err)
		}
	}

	return nil
}
