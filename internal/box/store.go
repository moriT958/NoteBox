package box

import "context"

type BoxFilterOp func(*filter)

type filter struct {
	id     *int
	active *bool
}

func WithID(id int) BoxFilterOp {
	return func(f *filter) {
		f.id = &id
	}
}

func WithActive(active bool) BoxFilterOp {
	return func(f *filter) {
		f.active = &active
	}
}

type BoxStore interface {
	Set(context.Context, Box) (Box, error)
	Get(context.Context, ...BoxFilterOp) ([]Box, error)
	Del(context.Context, ...BoxFilterOp) error
}
