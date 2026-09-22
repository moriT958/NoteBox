package box

import "context"

type BoxFilterOp func(*filter)

type filter struct {
	path   *string
	active *bool
}

func WithPath(path string) BoxFilterOp {
	return func(f *filter) {
		f.path = &path
	}
}

func WithActive(active bool) BoxFilterOp {
	return func(f *filter) {
		f.active = &active
	}
}

type BoxStore interface {
	Set(ctx context.Context, box Box) (*Box, error)
	GetByID(ctx context.Context, id string) (*Box, error)
	Get(ctx context.Context, opts ...BoxFilterOp) ([]Box, error)
	Del(ctx context.Context, id string) error
}
