package box

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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
	stubGet_Empty := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		b.id = 1
		return &b, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	t.Run("Successfully create new box at the given path.", func(t *testing.T) {
		configDir := t.TempDir()
		base := t.TempDir()

		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		got, err := s.CreateBox(context.Background(), "New Box", &base)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		wantPath := filepath.Join(base, "New Box")
		if got.path != wantPath {
			t.Errorf("path = %q, want %q", got.path, wantPath)
		}
		if info, statErr := os.Stat(wantPath); statErr != nil || !info.IsDir() {
			t.Errorf("expected directory to be created at %q: %v", wantPath, statErr)
		}
	})

	t.Run("Successfully create new box under the default dir, when path is nil.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		got, err := s.CreateBox(context.Background(), "My Box", nil)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		wantPath := filepath.Join(configDir, "My Box")
		if got.path != wantPath {
			t.Errorf("path = %q, want %q", got.path, wantPath)
		}
		if info, statErr := os.Stat(wantPath); statErr != nil || !info.IsDir() {
			t.Errorf("expected directory to be created at %q: %v", wantPath, statErr)
		}
	})

	t.Run("Fail to create new box, when title is empty.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		_, err := s.CreateBox(context.Background(), "", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to create new box, when the path is already used by another box.", func(t *testing.T) {
		configDir := t.TempDir()
		stubGet_Dup := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
			return []Box{{id: 1, title: "Existing", path: filepath.Join(configDir, "My Box"), active: true}}, nil
		}
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, stubGet_Dup, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to create new box, when store.Get returns error.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, stubGet_Err, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to create new box, when the directory cannot be created.", func(t *testing.T) {
		configDir := t.TempDir()
		// A regular file cannot have subdirectories created under it.
		blocker := filepath.Join(configDir, "blocker")
		if err := os.WriteFile(blocker, []byte(""), 0644); err != nil {
			t.Fatalf("failed to set up test: %v", err)
		}
		s := NewBoxService(blocker, &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to create new box, when store.Set returns error.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_Err, stubGet_Empty, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}

func TestBoxService_OpenFolderAsBox(t *testing.T) {
	stubGet_Empty := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{}, nil
	}
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		b.id = 1
		return &b, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	t.Run("Successfully open an existing folder as a box.", func(t *testing.T) {
		dir := t.TempDir()
		s := NewBoxService("", &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		got, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got.title != "My Folder" {
			t.Errorf("title = %q, want %q", got.title, "My Folder")
		}
		if got.path != dir {
			t.Errorf("path = %q, want %q", got.path, dir)
		}
	})

	t.Run("Fail to open folder as box, when title is empty.", func(t *testing.T) {
		dir := t.TempDir()
		s := NewBoxService("", &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to open folder as box, when the path does not exist.", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		s := NewBoxService("", &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to open folder as box, when the path is a file, not a directory.", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "note.md")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatalf("failed to set up test: %v", err)
		}
		s := NewBoxService("", &stubBoxStore{stubSet_OK, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", file)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to open folder as box, when the path is already used by another box.", func(t *testing.T) {
		dir := t.TempDir()
		stubGet_Dup := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
			return []Box{{id: 1, title: "Existing", path: dir, active: true}}, nil
		}
		s := NewBoxService("", &stubBoxStore{stubSet_OK, stubGet_Dup, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to open folder as box, when store.Set returns error.", func(t *testing.T) {
		dir := t.TempDir()
		s := NewBoxService("", &stubBoxStore{stubSet_Err, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}

func TestBoxService_GetBoxes(t *testing.T) {
	stubGet_OK := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{
			{id: 1, title: "Box A", path: "/a", active: true},
			{id: 2, title: "Box B", path: "/b", active: false},
		}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}

	tests := []struct {
		name      string
		store     *stubBoxStore
		expectLen int
		expectErr bool
	}{
		{
			name:      "Successfully get all boxes.",
			store:     &stubBoxStore{nil, stubGet_OK, nil},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "Fail to get boxes, when store returns error.",
			store:     &stubBoxStore{nil, stubGet_Err, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.GetBoxes(context.Background())

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if len(got) != tt.expectLen {
					t.Errorf("len(got) = %d, want %d", len(got), tt.expectLen)
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

func TestBoxService_GetActiveBoxes(t *testing.T) {
	stubGetOk_returns_only_active_boxes := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{
			{id: 1, title: "Box A", path: "/a", active: true},
			{id: 2, title: "Box B", path: "/b", active: true},
		}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}

	tests := []struct {
		name      string
		store     *stubBoxStore
		expectLen int
		expectErr bool
	}{
		{
			name:      "Successfully get only active boxes.",
			store:     &stubBoxStore{nil, stubGetOk_returns_only_active_boxes, nil},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "Fail to get active boxes, when store returns error.",
			store:     &stubBoxStore{nil, stubGet_Err, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.GetActiveBoxes(context.Background())

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if len(got) != tt.expectLen {
					t.Errorf("len(got) = %d, want %d", len(got), tt.expectLen)
				}
				for _, b := range got {
					if !b.active {
						t.Errorf("got inactive box in active boxes: %+v", b)
					}
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

func TestBoxService_GetInactiveBoxes(t *testing.T) {
	stubGet_Ok_returns_only_inactive_boxes := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{
			{id: 1, title: "Box A", path: "/a", active: false},
			{id: 2, title: "Box B", path: "/b", active: false},
		}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}

	tests := []struct {
		name      string
		store     *stubBoxStore
		expectLen int
		expectErr bool
	}{
		{
			name:      "Successfully get only inactive boxes.",
			store:     &stubBoxStore{nil, stubGet_Ok_returns_only_inactive_boxes, nil},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "Fail to get inactive boxes, when store returns error.",
			store:     &stubBoxStore{nil, stubGet_Err, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.GetInactiveBoxes(context.Background())

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if len(got) != tt.expectLen {
					t.Errorf("len(got) = %d, want %d", len(got), tt.expectLen)
				}
				for _, b := range got {
					if b.active {
						t.Errorf("got active box in inactive boxes: %+v", b)
					}
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

func TestBoxService_RenameBox(t *testing.T) {
	stubGet_Found := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{{id: 1, title: "Old Title", path: "/a", active: true}}, nil
	}
	stubGet_NotFound := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		return &b, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	tests := []struct {
		name        string
		store       *stubBoxStore
		expectTitle string
		expectErr   bool
	}{
		{
			name:        "Successfully rename box.",
			store:       &stubBoxStore{stubSet_OK, stubGet_Found, nil},
			expectTitle: "New Title",
			expectErr:   false,
		},
		{
			name:      "Fail to rename box, when box not found.",
			store:     &stubBoxStore{stubSet_OK, stubGet_NotFound, nil},
			expectErr: true,
		},
		{
			name:      "Fail to rename box, when store.Get returns error.",
			store:     &stubBoxStore{stubSet_OK, stubGet_Err, nil},
			expectErr: true,
		},
		{
			name:      "Fail to rename box, when store.Set returns error.",
			store:     &stubBoxStore{stubSet_Err, stubGet_Found, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.RenameBox(context.Background(), 1, "New Title")

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if got.title != tt.expectTitle {
					t.Errorf("title = %q, want %q", got.title, tt.expectTitle)
				}
			} else {
				if err == nil {
					t.Fatalf("error was expected, but not occured.")
				}
			}
		})
	}
}

func TestBoxService_ChangeBoxPath(t *testing.T) {
	stubGet_Found := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{{id: 1, title: "Box A", path: "/old/path", active: true}}, nil
	}
	stubGet_NotFound := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		return &b, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	tests := []struct {
		name       string
		store      *stubBoxStore
		expectPath string
		expectErr  bool
	}{
		{
			name:       "Successfully change box path.",
			store:      &stubBoxStore{stubSet_OK, stubGet_Found, nil},
			expectPath: "/new/path",
			expectErr:  false,
		},
		{
			name:      "Fail to change box path, when box not found.",
			store:     &stubBoxStore{stubSet_OK, stubGet_NotFound, nil},
			expectErr: true,
		},
		{
			name:      "Fail to change box path, when store.Get returns error.",
			store:     &stubBoxStore{stubSet_OK, stubGet_Err, nil},
			expectErr: true,
		},
		{
			name:      "Fail to change box path, when store.Set returns error.",
			store:     &stubBoxStore{stubSet_Err, stubGet_Found, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.ChangeBoxPath(context.Background(), 1, "/new/path")

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if got.path != tt.expectPath {
					t.Errorf("path = %q, want %q", got.path, tt.expectPath)
				}
			} else {
				if err == nil {
					t.Fatalf("error was expected, but not occured.")
				}
			}
		})
	}
}

func TestBoxService_RemoveBox(t *testing.T) {
	stubGet_Found := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{{id: 1, title: "Box A", path: "/a", active: true}}, nil
	}
	stubGet_NotFound := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return []Box{}, nil
	}
	stubGet_Err := func(ctx context.Context, opts ...BoxFilterOp) ([]Box, error) {
		return nil, errors.New("error at box store")
	}
	stubDel_OK := func(ctx context.Context, opts ...BoxFilterOp) error {
		return nil
	}
	stubDel_Err := func(ctx context.Context, opts ...BoxFilterOp) error {
		return errors.New("error at box store")
	}

	tests := []struct {
		name      string
		store     *stubBoxStore
		expectErr bool
	}{
		{
			name:      "Successfully remove box.",
			store:     &stubBoxStore{nil, stubGet_Found, stubDel_OK},
			expectErr: false,
		},
		{
			name:      "Fail to remove box, when box not found.",
			store:     &stubBoxStore{nil, stubGet_NotFound, stubDel_OK},
			expectErr: true,
		},
		{
			name:      "Fail to remove box, when store.Get returns error.",
			store:     &stubBoxStore{nil, stubGet_Err, stubDel_OK},
			expectErr: true,
		},
		{
			name:      "Fail to remove box, when store.Del returns error.",
			store:     &stubBoxStore{nil, stubGet_Found, stubDel_Err},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			err := s.RemoveBox(context.Background(), 1)

			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}
			if tt.expectErr && err == nil {
				t.Fatalf("error was expected, but not occured.")
			}
		})
	}
}

func TestBoxService_PruneBoxes(t *testing.T) {
	stubDel_OK := func(ctx context.Context, opts ...BoxFilterOp) error {
		return nil
	}
	stubDel_Err := func(ctx context.Context, opts ...BoxFilterOp) error {
		return errors.New("error at box store")
	}

	tests := []struct {
		name      string
		store     *stubBoxStore
		expectErr bool
	}{
		{
			name:      "Successfully prune inactive boxes.",
			store:     &stubBoxStore{nil, nil, stubDel_OK},
			expectErr: false,
		},
		{
			name:      "Fail to prune boxes, when store.Del returns error.",
			store:     &stubBoxStore{nil, nil, stubDel_Err},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			err := s.PruneBoxes(context.Background())

			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}
			if tt.expectErr && err == nil {
				t.Fatalf("error was expected, but not occured.")
			}
		})
	}
}
