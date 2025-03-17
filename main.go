package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"notebox/cli"
	"notebox/config"
	"notebox/note"
	"os"
)

const (
	configFile = "./config.json"
	sqliteDsn  = "./db.sqlite"
)

func initDB() *sql.DB {

	db, err := sql.Open("sqlite", sqliteDsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func main() {
	// configファイルをロードする
	if err := config.LoadConfigFile(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config file: %v\n", err)
		os.Exit(1)
	}

	noteRepo, err := note.NewNoteRepository(initDB())
	if err != nil {
		log.Fatal(err)
	}
	cli.Nr = noteRepo

	// サブコマンドを登録
	ctx := context.Background()
	os.Exit(cli.InitCommands(ctx))
}
