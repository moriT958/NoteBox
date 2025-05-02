package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

/* Note Entity */

type note struct {
	title   string
	content string
}

// sample data
// var notes []note = []note{
// 	{title: "sample", content: "# sample0\n\n## hello\n\nthis is example0."},
// 	{title: "sample", content: "# sample1\n\n## hello\n\nthis is example1."},
// 	{title: "sample", content: "# sample2\n\n## hello\n\nthis is example2."},
// 	{title: "sample", content: "# sample3\n\n## hello\n\nthis is example3."},
// 	{title: "sample", content: "# sample4\n\n## hello\n\nthis is example4."},
// 	{title: "sample", content: "# sample5\n\n## hello\n\nthis is example5."},
// }

const baseDir string = "./notes"

func loadNoteFiles(baseDir string) ([]note, error) {
	notes := make([]note, 0)

	if err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return filepath.SkipDir
		}

		_, filename := filepath.Split(path)
		filename = strings.TrimSuffix(filename, ".md")
		parts := strings.Split(filename, "-")
		if len(parts) < 4 {
			return errors.New(fmt.Sprintf("unexpected filename format: %s", filename))
		}

		title := strings.Join(parts[:len(parts)-3], "-")
		var content []byte
		fp, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fp.Close()
		fp.Read(content)

		notes = append(notes, note{title, string(content)})

		return nil
	}); err != nil {
		return nil, err
	}

	return notes, nil
}

func createNewNoteFile(title string) tea.Cmd {
	timeStr := time.Now().Format(time.DateOnly)

	return func() tea.Msg {
		filename := filepath.Join(baseDir,
			title+"-"+timeStr+".md")

		fp, err := os.Create(filename)
		if err != nil {
			return errMsg{err}
		}
		defer fp.Close()

		content := fmt.Sprintf("# %s\n\n", title)
		fmt.Fprint(fp, content)

		return nil
	}
}
