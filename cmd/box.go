package cmd

import "notebox/box"

var (
	storagePath = "./storage"
	nb          = box.NewNoteBox(storagePath)
)
