package listpanel

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getAllNoteFiles(volume string) []Note {

	files, err := os.ReadDir(volume)
	if err != nil {
		log.Fatal(err)
	}

	var notes []Note
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		// set desc content of file mod time
		info, err := f.Info()
		if err != nil {
			log.Fatal(err)
		}

		notePath := filepath.Join(volume, f.Name())
		parts := strings.Split(f.Name(), "-")

		notes = append(notes, Note{
			title: parts[0],
			desc:  info.ModTime().Format(`2006-01-02`),
			path:  notePath,
		})
	}

	return notes
}
