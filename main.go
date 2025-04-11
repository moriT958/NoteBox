package main

import (
	"context"
	"database/sql"
	"log"
	"notebox/cli"
	"notebox/config"
	"notebox/models"
	"notebox/tui"
	"os"

	_ "modernc.org/sqlite"
)

func init() {
	config.LoadConfig()
}

func main() {
	db, err := sql.Open("sqlite", config.MetaDataDir())
	if err != nil {
		log.Fatal(err)
	}

	// Noteリポジトリの初期化
	if err := models.NewNoteRepository(db); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		// サブコマンドを登録
		ctx := context.Background()
		os.Exit(cli.InitCommands(ctx))
	} else {
		log.Fatal(tui.StartApp())
	}
}
