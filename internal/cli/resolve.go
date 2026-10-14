package cli

import (
	"context"
	"fmt"
	"notebox/internal/core/box"
)

// resolveBox returns the active box registered at path. An empty path means
// the default box.
func resolveBox(ctx context.Context, boxes *box.BoxService, path string) (box.Box, error) {
	if path == "" {
		b, err := boxes.DefaultBox(ctx)
		if err != nil {
			return box.Box{}, err
		}
		return *b, nil
	}

	b, err := boxes.GetBoxByPath(ctx, path)
	if err != nil {
		return box.Box{}, err
	}
	if b == nil || !b.Active {
		return box.Box{}, fmt.Errorf("box not found: %s (register it with 'nbx box add')", path)
	}
	return *b, nil
}
