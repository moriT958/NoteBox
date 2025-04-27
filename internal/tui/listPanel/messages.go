package listpanel

type NoteMsg struct {
	Note
}

type editorFinishedMsg struct{ err error }
