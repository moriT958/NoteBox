package box

import (
	"context"
	"errors"
	"testing"
)

type stubBoxStore struct {
	setFunc func(ctx context.Context, b Box) (*Box, error)
	getFunc func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error)
	delFunc func(ctx context.Context, opts ...BoxFilterOp) error
}

var _ BoxStore = (*stubBoxStore)(nil)

func (f *stubBoxStore) Set(ctx context.Context, b Box) (*Box, error) {
	return f.setFunc(ctx, b)
}

func (f *stubBoxStore) Get(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
	return f.getFunc(ctx, opts...)
}

func (f *stubBoxStore) Del(ctx context.Context, opts ...BoxFilterOp) error {
	return f.delFunc(ctx, opts...)
}

func TestBoxService_CreateBox(t *testing.T) {
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		return &Box{
			id:     1,
			title:  "New Test Box",
			path:   "/test/new/box/path",
			active: true,
		}, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	tests := []struct {
		name         string
		store        *stubBoxStore
		expectID     int
		expectTitle  string
		expectPath   string
		expectActive bool
		expectErr    bool
	}{
		{
			name:         "Successfully create new box.",
			store:        &stubBoxStore{stubSet_OK, nil, nil},
			expectID:     1,
			expectTitle:  "New Test Box",
			expectPath:   "/test/new/box/path",
			expectActive: true,
			expectErr:    false,
		},
		{
			name:      "Fail to create new box, when store returns error.",
			store:     &stubBoxStore{stubSet_Err, nil, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService(tt.store)

			got, err := s.CreateBox(context.Background(), "New Test Box", "/test/new/box/path")

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if got.id != tt.expectID {
					t.Errorf("id = %d, want %d", got.id, tt.expectID)
				}
				if got.title != tt.expectTitle {
					t.Errorf("title = %q, want %q", got.title, tt.expectTitle)
				}
				if got.path != tt.expectPath {
					t.Errorf("path = %q, want %q", got.path, tt.expectPath)
				}
				if got.active != tt.expectActive {
					t.Errorf("active = %t, want %t", got.active, tt.expectActive)
				}
			} else {
				if err == nil {
					t.Fatalf("error was expected, but not occured.")
				}
				if errors.Unwrap(err).Error() != "error at box store" {
					t.Errorf("unexpected error message: %s", err.Error())
				}
			}
		})
	}
}
