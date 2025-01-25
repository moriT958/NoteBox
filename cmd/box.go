package cmd

import "notebox/box"

var (
	storagePath = "./storage"
	NoteBox     = box.NewNoteBox(storagePath)
)
