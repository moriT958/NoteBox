package box

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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
	boxes, err := s.store.Get(ctx, WithPath(path))
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
		base = *path
	}
	boxPath := filepath.Join(base, encodeDirName(title))

	existing, err := s.findByPath(ctx, boxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}
	if existing != nil && existing.active {
		return nil, fmt.Errorf("failed to create box: a box with this path already exists")
	}

	if err := os.MkdirAll(boxPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}

	newBox := Box{id: uuid.NewString(), title: title, path: boxPath, active: true}
	if existing != nil {
		newBox.id = existing.id // revive a soft-deleted box at the same path
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

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("failed to open folder as box: path must be an existing directory")
	}

	existing, err := s.findByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open folder as box: %w", err)
	}
	if existing != nil && existing.active {
		return nil, fmt.Errorf("failed to open folder as box: a box with this path already exists")
	}

	newBox := Box{id: uuid.NewString(), title: title, path: path, active: true}
	if existing != nil {
		newBox.id = existing.id
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

func (s *BoxService) GetBoxes(ctx context.Context) ([]Box, error) {
	boxes, err := s.store.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get boxes: %w", err)
	}
	return boxes, nil
}

func (s *BoxService) GetActiveBoxes(ctx context.Context) ([]Box, error) {
	boxes, err := s.store.Get(ctx, WithActive(true))
	if err != nil {
		return nil, fmt.Errorf("failed to get active boxes: %w", err)
	}
	return boxes, nil
}

func (s *BoxService) GetInactiveBoxes(ctx context.Context) ([]Box, error) {
	boxes, err := s.store.Get(ctx, WithActive(false))
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

	box.title = title
	b, err := s.store.Set(ctx, *box)
	if err != nil {
		return nil, fmt.Errorf("failed to rename box: %w", err)
	}
	return b, nil
}

func (s *BoxService) ChangeBoxPath(ctx context.Context, id string, path string) (*Box, error) {
	box, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to change box path: %w", err)
	}
	if box == nil {
		return nil, fmt.Errorf("failed to change box path: box not found")
	}

	box.path = path
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

	box.active = false
	if _, err := s.store.Set(ctx, *box); err != nil {
		return fmt.Errorf("failed to remove box: %w", err)
	}
	return nil
}

func (s *BoxService) PruneBoxes(ctx context.Context) error {
	boxes, err := s.store.Get(ctx, WithActive(false))
	if err != nil {
		return fmt.Errorf("failed to prune boxes: %w", err)
	}

	for _, b := range boxes {
		if err := os.RemoveAll(b.path); err != nil {
			return fmt.Errorf("failed to prune boxes: %w", err)
		}
		if err := s.store.Del(ctx, b.id); err != nil {
			return fmt.Errorf("failed to prune boxes: %w", err)
		}
	}

	return nil
}
