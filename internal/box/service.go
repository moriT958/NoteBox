package box

import (
	"context"
	"fmt"
)

type BoxService struct {
	store BoxStore
}

func NewBoxService(store BoxStore) *BoxService {
	return &BoxService{store}
}

func (s *BoxService) CreateBox(ctx context.Context, title, path string) (Box, error) {
	return s.store.Set(ctx, Box{
		title:  title,
		path:   path,
		active: true,
	})
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
	return &b, nil
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
	return &b, nil
}

func (s *BoxService) RemoveBox(ctx context.Context, id int) error {
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
