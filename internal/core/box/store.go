package box

import "context"

type Filter struct {
	Path   *string
	Active *bool
}

type BoxStore interface {
	Set(ctx context.Context, box Box) (*Box, error)
	GetByID(ctx context.Context, id string) (*Box, error)
	Get(ctx context.Context, opts Filter) ([]Box, error)
	Del(ctx context.Context, id string) error
}
