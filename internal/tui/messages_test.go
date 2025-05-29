package tui

import (
	"notebox/internal/config"
	"testing"
)

func TestRenderPreviewCmd(t *testing.T) {
	t.Run("model has a item", func(t *testing.T) {
		model := model{
			listPanel: listPanel{
				items: []note{{"hello", "testdata/notes/hello-2025-05-02.md"}},
			}}

		want := renderPreviewMsg("# hello\n\n")

		cmd := model.renderPreviewCmd(model.listPanel.items[0].path)
		got := cmd()

		if err, ok := got.(errMsg); ok {
			t.Fatal(err.Error())
		}

		if content, ok := got.(renderPreviewMsg); ok {
			if want != got {
				t.Errorf("want %q, but got %q\n", want, content)
			}
		}
	})

	t.Run("model has no item", func(t *testing.T) {
		model := model{
			cfg: &config.Config{DummyNoteDir: "testdata/dummy.md"},
			listPanel: listPanel{
				items: []note{},
			}}

		want := renderPreviewMsg("(( No Note Selected ))\n")

		cmd := model.renderPreviewCmd("")
		got := cmd()

		if err, ok := got.(errMsg); ok {
			t.Fatal(err.Error())
		}

		if content, ok := got.(renderPreviewMsg); ok {
			if want != got {
				t.Errorf("want %q, but got %q\n", want, content)
			}
		}
	})
}
