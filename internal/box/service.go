package box

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

func (s *BoxService) isPathUsed(ctx context.Context, path string) (bool, error) {
	boxes, err := s.store.Get(ctx)
	if err != nil {
		return false, err
	}
	for _, b := range boxes {
		if b.path == path {
			return true, nil
		}
	}
	return false, nil
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

	used, err := s.isPathUsed(ctx, boxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}
	if used {
		return nil, fmt.Errorf("failed to create box: a box with this path already exists")
	}

	if err := os.MkdirAll(boxPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}

	box, err := s.store.Set(ctx, Box{
		title:  title,
		path:   boxPath,
		active: true,
	})
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

	used, err := s.isPathUsed(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open folder as box: %w", err)
	}
	if used {
		return nil, fmt.Errorf("failed to open folder as box: a box with this path already exists")
	}

	box, err := s.store.Set(ctx, Box{
		title:  title,
		path:   path,
		active: true,
	})
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

func (s *BoxService) RenameBox(ctx context.Context, id int, title string) (*Box, error) {
	box, err := s.store.Get(ctx, WithID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to rename box: %w", err)
	}

	if len(box) == 0 {
		return nil, fmt.Errorf("failed to rename box: box not found")
	}

	box[0].title = title
	b, err := s.store.Set(ctx, box[0])
	if err != nil {
		return nil, fmt.Errorf("failed to rename box: %w", err)
	}
	return b, nil
}

func (s *BoxService) ChangeBoxPath(ctx context.Context, id int, path string) (*Box, error) {
	box, err := s.store.Get(ctx, WithID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to change box path: %w", err)
	}

	if len(box) == 0 {
		return nil, fmt.Errorf("failed to change box path: box not found")
	}

	box[0].path = path
	b, err := s.store.Set(ctx, box[0])
	if err != nil {
		return nil, fmt.Errorf("failed to change box path: %w", err)
	}
	return b, nil
}

func (s *BoxService) RemoveBox(ctx context.Context, id int) error {
	box, err := s.store.Get(ctx, WithID(id))
	if err != nil {
		return fmt.Errorf("failed to remove box path: %w", err)
	}

	if len(box) == 0 {
		return fmt.Errorf("failed to remove box path: box not found")
	}

	if err := s.store.Del(ctx, WithID(id)); err != nil {
		return fmt.Errorf("failed to remove box: %w", err)
	}
	return nil
}

func (s *BoxService) PruneBoxes(ctx context.Context) error {
	if err := s.store.Del(ctx, WithActive(false)); err != nil {
		return fmt.Errorf("failed to prune boxes: %w", err)
	}
	return nil
}
