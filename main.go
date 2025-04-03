package main

import (
	"context"
	"database/sql"
	"log"
	"notebox/cli"
	"notebox/config"
	"notebox/models"
	"os"

	_ "modernc.org/sqlite"
)

func init() {
	config.GetConfig()
}

func main() {
	db, err := sql.Open("sqlite", config.MetaDataDir())
	if err != nil {
		log.Fatal(err)
	}

	// Noteリポジトリの初期化
	noteRepo, err := models.NewNoteRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	cli.Nr = noteRepo

	// サブコマンドを登録
	ctx := context.Background()
	os.Exit(cli.InitCommands(ctx))
}
