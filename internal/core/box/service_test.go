package box

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type stubBoxStore struct {
	setFunc     func(ctx context.Context, b Box) (*Box, error)
	getByIDFunc func(ctx context.Context, id string) (*Box, error)
	getFunc     func(ctx context.Context, opts Filter) ([]Box, error)
	delFunc     func(ctx context.Context, id string) error
}

var _ BoxStore = (*stubBoxStore)(nil)

func (f *stubBoxStore) Set(ctx context.Context, b Box) (*Box, error) {
	return f.setFunc(ctx, b)
}

func (f *stubBoxStore) GetByID(ctx context.Context, id string) (*Box, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *stubBoxStore) Get(ctx context.Context, opts Filter) ([]Box, error) {
	return f.getFunc(ctx, opts)
}

func (f *stubBoxStore) Del(ctx context.Context, id string) error {
	return f.delFunc(ctx, id)
}

// realTempDir returns a temp dir with symlinks resolved (e.g. /var ->
// /private/var on macOS), so it matches paths resolved from the working
// directory.
func realTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("failed to resolve temp dir: %v", err)
	}
	return dir
}

func TestBoxService_CreateBox(t *testing.T) {
	stubGet_Empty := func(ctx context.Context, opts Filter) ([]Box, error) {
		return []Box{}, nil
	}
	stubGet_Err := func(ctx context.Context, opts Filter) ([]Box, error) {
		return nil, errors.New("error at box store")
	}
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		return &b, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	t.Run("Successfully create new box at the given path.", func(t *testing.T) {
		configDir := t.TempDir()
		base := t.TempDir()

		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		got, err := s.CreateBox(context.Background(), "New Box", &base)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got.ID == "" {
			t.Errorf("id should be generated, got empty string")
		}
		wantPath := filepath.Join(base, "New Box")
		if got.Path != wantPath {
			t.Errorf("path = %q, want %q", got.Path, wantPath)
		}
		if info, statErr := os.Stat(wantPath); statErr != nil || !info.IsDir() {
			t.Errorf("expected directory to be created at %q: %v", wantPath, statErr)
		}
	})

	t.Run("Successfully create new box under the default dir, when path is nil.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		got, err := s.CreateBox(context.Background(), "My Box", nil)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		wantPath := filepath.Join(configDir, "My Box")
		if got.Path != wantPath {
			t.Errorf("path = %q, want %q", got.Path, wantPath)
		}
		if info, statErr := os.Stat(wantPath); statErr != nil || !info.IsDir() {
			t.Errorf("expected directory to be created at %q: %v", wantPath, statErr)
		}
	})

	t.Run("Fail to create new box, when title is empty.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		_, err := s.CreateBox(context.Background(), "", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to create new box, when the path is already used by an active box.", func(t *testing.T) {
		configDir := t.TempDir()
		stubGet_Dup := func(ctx context.Context, opts Filter) ([]Box, error) {
			return []Box{{ID: "existing-id", Title: "Existing", Path: filepath.Join(configDir, "My Box"), Active: true}}, nil
		}
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, nil, stubGet_Dup, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Successfully revive an inactive box at the same path.", func(t *testing.T) {
		configDir := t.TempDir()
		stubGet_Inactive := func(ctx context.Context, opts Filter) ([]Box, error) {
			return []Box{{ID: "existing-id", Title: "Old Title", Path: filepath.Join(configDir, "My Box"), Active: false}}, nil
		}
		var gotSet Box
		stubSet_Capture := func(ctx context.Context, b Box) (*Box, error) {
			gotSet = b
			return &b, nil
		}
		s := NewBoxService(configDir, &stubBoxStore{stubSet_Capture, nil, stubGet_Inactive, nil})

		got, err := s.CreateBox(context.Background(), "New Title", nil)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got.ID != "existing-id" {
			t.Errorf("id = %q, want %q (revived box should keep its id)", got.ID, "existing-id")
		}
		if gotSet.Title != "New Title" {
			t.Errorf("title = %q, want %q", gotSet.Title, "New Title")
		}
		if !gotSet.Active {
			t.Errorf("active = %t, want true", gotSet.Active)
		}
	})

	t.Run("Fail to create new box, when store.Get returns error.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_OK, nil, stubGet_Err, nil})

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
		s := NewBoxService(blocker, &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to create new box, when store.Set returns error.", func(t *testing.T) {
		configDir := t.TempDir()
		s := NewBoxService(configDir, &stubBoxStore{stubSet_Err, nil, stubGet_Empty, nil})

		_, err := s.CreateBox(context.Background(), "My Box", nil)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}

func TestBoxService_OpenFolderAsBox(t *testing.T) {
	stubGet_Empty := func(ctx context.Context, opts Filter) ([]Box, error) {
		return []Box{}, nil
	}
	stubSet_OK := func(ctx context.Context, b Box) (*Box, error) {
		return &b, nil
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	t.Run("Successfully open an existing folder as a box.", func(t *testing.T) {
		dir := t.TempDir()
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		got, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got.ID == "" {
			t.Errorf("id should be generated, got empty string")
		}
		if got.Title != "My Folder" {
			t.Errorf("title = %q, want %q", got.Title, "My Folder")
		}
		if got.Path != dir {
			t.Errorf("path = %q, want %q", got.Path, dir)
		}
	})

	t.Run("Fail to open folder as box, when title is empty.", func(t *testing.T) {
		dir := t.TempDir()
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to open folder as box, when the path does not exist.", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "does-not-exist")
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

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
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", file)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to open folder as box, when the path is already used by an active box.", func(t *testing.T) {
		dir := t.TempDir()
		stubGet_Dup := func(ctx context.Context, opts Filter) ([]Box, error) {
			return []Box{{ID: "existing-id", Title: "Existing", Path: dir, Active: true}}, nil
		}
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_Dup, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Successfully store a relative path as an absolute path.", func(t *testing.T) {
		dir := realTempDir(t)
		t.Chdir(filepath.Dir(dir))
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_Empty, nil})

		got, err := s.OpenFolderAsBox(context.Background(), "My Folder", filepath.Base(dir))
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got.Path != dir {
			t.Errorf("path = %q, want %q", got.Path, dir)
		}
	})

	t.Run("Fail to open folder as box, when a differently written path is already used by an active box.", func(t *testing.T) {
		dir := realTempDir(t)
		t.Chdir(filepath.Dir(dir))
		stubGet_ByPath := func(ctx context.Context, opts Filter) ([]Box, error) {
			if opts.Path != nil && *opts.Path == dir {
				return []Box{{ID: "existing-id", Title: "Existing", Path: dir, Active: true}}, nil
			}
			return []Box{}, nil
		}
		s := NewBoxService("", &stubBoxStore{stubSet_OK, nil, stubGet_ByPath, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", filepath.Base(dir)+"/")
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Successfully revive an inactive box at the same path.", func(t *testing.T) {
		dir := t.TempDir()
		stubGet_Inactive := func(ctx context.Context, opts Filter) ([]Box, error) {
			return []Box{{ID: "existing-id", Title: "Old Title", Path: dir, Active: false}}, nil
		}
		var gotSet Box
		stubSet_Capture := func(ctx context.Context, b Box) (*Box, error) {
			gotSet = b
			return &b, nil
		}
		s := NewBoxService("", &stubBoxStore{stubSet_Capture, nil, stubGet_Inactive, nil})

		got, err := s.OpenFolderAsBox(context.Background(), "New Title", dir)
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got.ID != "existing-id" {
			t.Errorf("id = %q, want %q (revived box should keep its id)", got.ID, "existing-id")
		}
		if gotSet.Title != "New Title" {
			t.Errorf("title = %q, want %q", gotSet.Title, "New Title")
		}
		if !gotSet.Active {
			t.Errorf("active = %t, want true", gotSet.Active)
		}
	})

	t.Run("Fail to open folder as box, when store.Set returns error.", func(t *testing.T) {
		dir := t.TempDir()
		s := NewBoxService("", &stubBoxStore{stubSet_Err, nil, stubGet_Empty, nil})

		_, err := s.OpenFolderAsBox(context.Background(), "My Folder", dir)
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}

func TestBoxService_GetBox(t *testing.T) {
	t.Run("Successfully get a box by id.", func(t *testing.T) {
		stubGetByID := func(ctx context.Context, id string) (*Box, error) {
			return &Box{ID: id, Title: "Box A", Path: "/a", Active: true}, nil
		}
		s := NewBoxService("", &stubBoxStore{nil, stubGetByID, nil, nil})

		got, err := s.GetBox(context.Background(), "1")
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got == nil || got.ID != "1" {
			t.Errorf("got = %+v, want box with id %q", got, "1")
		}
	})

	t.Run("Returns nil, when no box has the id.", func(t *testing.T) {
		stubGetByID := func(ctx context.Context, id string) (*Box, error) {
			return nil, nil
		}
		s := NewBoxService("", &stubBoxStore{nil, stubGetByID, nil, nil})

		got, err := s.GetBox(context.Background(), "missing")
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got != nil {
			t.Errorf("got = %+v, want nil", got)
		}
	})

	t.Run("Fail to get box, when store.GetByID returns error.", func(t *testing.T) {
		stubGetByID := func(ctx context.Context, id string) (*Box, error) {
			return nil, errors.New("error at box store")
		}
		s := NewBoxService("", &stubBoxStore{nil, stubGetByID, nil, nil})

		if _, err := s.GetBox(context.Background(), "1"); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}

func TestBoxService_GetBoxes(t *testing.T) {
	stubGet_OK := func(ctx context.Context, opts Filter) ([]Box, error) {
		return []Box{
			{ID: "1", Title: "Box A", Path: "/a", Active: true},
			{ID: "2", Title: "Box B", Path: "/b", Active: false},
		}, nil
	}
	stubGet_Err := func(ctx context.Context, opts Filter) ([]Box, error) {
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
			store:     &stubBoxStore{nil, nil, stubGet_OK, nil},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "Fail to get boxes, when store returns error.",
			store:     &stubBoxStore{nil, nil, stubGet_Err, nil},
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
	stubGet_OK := func(ctx context.Context, opts Filter) ([]Box, error) {
		return []Box{
			{ID: "1", Title: "Box A", Path: "/a", Active: true},
			{ID: "2", Title: "Box B", Path: "/b", Active: true},
		}, nil
	}
	stubGet_Err := func(ctx context.Context, opts Filter) ([]Box, error) {
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
			store:     &stubBoxStore{nil, nil, stubGet_OK, nil},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "Fail to get active boxes, when store returns error.",
			store:     &stubBoxStore{nil, nil, stubGet_Err, nil},
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
					if !b.Active {
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
	stubGet_OK := func(ctx context.Context, opts Filter) ([]Box, error) {
		return []Box{
			{ID: "1", Title: "Box A", Path: "/a", Active: false},
			{ID: "2", Title: "Box B", Path: "/b", Active: false},
		}, nil
	}
	stubGet_Err := func(ctx context.Context, opts Filter) ([]Box, error) {
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
			store:     &stubBoxStore{nil, nil, stubGet_OK, nil},
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "Fail to get inactive boxes, when store returns error.",
			store:     &stubBoxStore{nil, nil, stubGet_Err, nil},
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
					if b.Active {
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
	stubGetByID_Found := func(ctx context.Context, id string) (*Box, error) {
		return &Box{ID: "1", Title: "Old Title", Path: "/a", Active: true}, nil
	}
	stubGetByID_NotFound := func(ctx context.Context, id string) (*Box, error) {
		return nil, nil
	}
	stubGetByID_Err := func(ctx context.Context, id string) (*Box, error) {
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
			store:       &stubBoxStore{stubSet_OK, stubGetByID_Found, nil, nil},
			expectTitle: "New Title",
			expectErr:   false,
		},
		{
			name:      "Fail to rename box, when box not found.",
			store:     &stubBoxStore{stubSet_OK, stubGetByID_NotFound, nil, nil},
			expectErr: true,
		},
		{
			name:      "Fail to rename box, when store.GetByID returns error.",
			store:     &stubBoxStore{stubSet_OK, stubGetByID_Err, nil, nil},
			expectErr: true,
		},
		{
			name:      "Fail to rename box, when store.Set returns error.",
			store:     &stubBoxStore{stubSet_Err, stubGetByID_Found, nil, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.RenameBox(context.Background(), "1", "New Title")

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if got.Title != tt.expectTitle {
					t.Errorf("title = %q, want %q", got.Title, tt.expectTitle)
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
	stubGetByID_Found := func(ctx context.Context, id string) (*Box, error) {
		return &Box{ID: "1", Title: "Box A", Path: "/old/path", Active: true}, nil
	}
	stubGetByID_NotFound := func(ctx context.Context, id string) (*Box, error) {
		return nil, nil
	}
	stubGetByID_Err := func(ctx context.Context, id string) (*Box, error) {
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
			store:      &stubBoxStore{stubSet_OK, stubGetByID_Found, nil, nil},
			expectPath: "/new/path",
			expectErr:  false,
		},
		{
			name:      "Fail to change box path, when box not found.",
			store:     &stubBoxStore{stubSet_OK, stubGetByID_NotFound, nil, nil},
			expectErr: true,
		},
		{
			name:      "Fail to change box path, when store.GetByID returns error.",
			store:     &stubBoxStore{stubSet_OK, stubGetByID_Err, nil, nil},
			expectErr: true,
		},
		{
			name:      "Fail to change box path, when store.Set returns error.",
			store:     &stubBoxStore{stubSet_Err, stubGetByID_Found, nil, nil},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBoxService("", tt.store)

			got, err := s.ChangeBoxPath(context.Background(), "1", "/new/path")

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("unexpected err occurred: %v", err)
				}
				if got.Path != tt.expectPath {
					t.Errorf("path = %q, want %q", got.Path, tt.expectPath)
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
	stubGetByID_Found := func(ctx context.Context, id string) (*Box, error) {
		return &Box{ID: "1", Title: "Box A", Path: "/a", Active: true}, nil
	}
	stubGetByID_NotFound := func(ctx context.Context, id string) (*Box, error) {
		return nil, nil
	}
	stubGetByID_Err := func(ctx context.Context, id string) (*Box, error) {
		return nil, errors.New("error at box store")
	}
	stubSet_Err := func(ctx context.Context, b Box) (*Box, error) {
		return nil, errors.New("error at box store")
	}

	t.Run("Successfully remove (deactivate) box.", func(t *testing.T) {
		var gotSet Box
		stubSet_Capture := func(ctx context.Context, b Box) (*Box, error) {
			gotSet = b
			return &b, nil
		}
		s := NewBoxService("", &stubBoxStore{stubSet_Capture, stubGetByID_Found, nil, nil})

		if err := s.RemoveBox(context.Background(), "1"); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if gotSet.Active {
			t.Errorf("active = %t, want false", gotSet.Active)
		}
	})

	t.Run("Fail to remove box, when box not found.", func(t *testing.T) {
		s := NewBoxService("", &stubBoxStore{nil, stubGetByID_NotFound, nil, nil})

		if err := s.RemoveBox(context.Background(), "1"); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to remove box, when store.GetByID returns error.", func(t *testing.T) {
		s := NewBoxService("", &stubBoxStore{nil, stubGetByID_Err, nil, nil})

		if err := s.RemoveBox(context.Background(), "1"); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to remove box, when store.Set returns error.", func(t *testing.T) {
		s := NewBoxService("", &stubBoxStore{stubSet_Err, stubGetByID_Found, nil, nil})

		if err := s.RemoveBox(context.Background(), "1"); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}

func TestBoxService_PruneBoxes(t *testing.T) {
	t.Run("Successfully remove directories and delete inactive boxes.", func(t *testing.T) {
		dirA := t.TempDir()
		dirB := t.TempDir()

		stubGet_Inactive := func(ctx context.Context, opts Filter) ([]Box, error) {
			return []Box{
				{ID: "1", Title: "Box A", Path: dirA, Active: false},
				{ID: "2", Title: "Box B", Path: dirB, Active: false},
			}, nil
		}
		var deletedIDs []string
		stubDel_Capture := func(ctx context.Context, id string) error {
			deletedIDs = append(deletedIDs, id)
			return nil
		}

		s := NewBoxService("", &stubBoxStore{nil, nil, stubGet_Inactive, stubDel_Capture})

		if err := s.PruneBoxes(context.Background()); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		if _, err := os.Stat(dirA); !os.IsNotExist(err) {
			t.Errorf("expected %q to be removed, stat err = %v", dirA, err)
		}
		if _, err := os.Stat(dirB); !os.IsNotExist(err) {
			t.Errorf("expected %q to be removed, stat err = %v", dirB, err)
		}
		if len(deletedIDs) != 2 {
			t.Errorf("len(deletedIDs) = %d, want 2", len(deletedIDs))
		}
	})

	t.Run("Fail to prune boxes, when store.Get returns error.", func(t *testing.T) {
		stubGet_Err := func(ctx context.Context, opts Filter) ([]Box, error) {
			return nil, errors.New("error at box store")
		}
		s := NewBoxService("", &stubBoxStore{nil, nil, stubGet_Err, nil})

		if err := s.PruneBoxes(context.Background()); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})

	t.Run("Fail to prune boxes, when store.Del returns error.", func(t *testing.T) {
		dir := t.TempDir()
		stubGet_Inactive := func(ctx context.Context, opts Filter) ([]Box, error) {
			return []Box{{ID: "1", Title: "Box A", Path: dir, Active: false}}, nil
		}
		stubDel_Err := func(ctx context.Context, id string) error {
			return errors.New("error at box store")
		}
		s := NewBoxService("", &stubBoxStore{nil, nil, stubGet_Inactive, stubDel_Err})

		if err := s.PruneBoxes(context.Background()); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}
