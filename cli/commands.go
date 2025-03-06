package cli

import (
	"context"
	"flag"
	"notebox/config"
	"notebox/store"

	"github.com/google/subcommands"
)

type storeMock struct{}

func (s *storeMock) Save(note store.Note) (int, error) {
	return 0, nil
}

func (s *storeMock) FindByID(id int) (store.Note, error) {
	n := store.Note{
		ID:    0,
		Title: "Sample Note",
		Path:  "data/test.md",
	}
	return n, nil
}

func (s *storeMock) FindAll() ([]store.Note, error) {
	nl := make([]store.Note, 3)
	for i := range 3 {
		nl[i] = store.Note{
			ID:    0,
			Title: "Sample Note",
			Path:  "data/test.md",
		}
	}
	return nl, nil
}

func (s *storeMock) DeleteByID(id int) error {
	return nil
}

func InitCommands(ctx context.Context, cfg *config.Config) int {
	//store := store.NewNoteStore(cfg.Volume)
	store := new(storeMock)

	subcommands.Register(&newCmd{cfg: cfg, store: store}, "")
	subcommands.Register(&lsCmd{cfg: cfg, store: store}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
