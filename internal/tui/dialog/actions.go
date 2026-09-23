package dialog

import (
	"notebox/internal/core/box"
	"notebox/internal/core/note"

	tea "charm.land/bubbletea/v2"
)

// Action is what a dialog returns from HandleMsg; nil means nothing for the
// root to do.
type Action any

type (
	// ActionClose closes the front dialog.
	ActionClose struct{}
	// ActionOpen opens another dialog on top of the current one.
	ActionOpen struct{ Dialog Dialog }
	// ActionCmd hands a command (e.g. from a text input) to the program.
	ActionCmd struct{ Cmd tea.Cmd }

	ActionCreateNote struct{ Title string }
	ActionDeleteNote struct{ Note note.Note }
	ActionSelectNote struct{ Note note.Note }

	ActionSwitchBox struct{ Box box.Box }
	// ActionCreateBox creates a box directory under Path, or under the
	// default location when Path is empty.
	ActionCreateBox struct{ Title, Path string }
	// ActionOpenFolder registers the existing directory Path as a box.
	ActionOpenFolder struct{ Title, Path string }
	ActionRenameBox  struct {
		Box   box.Box
		Title string
	}
	ActionDeleteBox struct{ Box box.Box }
)

func cmdAction(cmd tea.Cmd) Action {
	if cmd == nil {
		return nil
	}
	return ActionCmd{Cmd: cmd}
}
