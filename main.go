package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"notebox/cli"
	"notebox/config"
	"notebox/note"
	"os"

	_ "modernc.org/sqlite"
)

func init() {
	// configファイルをロードする
	if err := config.LoadConfigFile(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config file: %v\n", err)
		os.Exit(1)
	}

	// ボリュームファイルの存在確認
	if _, err := os.Stat(config.Volume); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(config.Volume, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create volume file: %v\n", err)
			os.Exit(1)
		}
	}
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite", config.MetadataDir)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {

	// Noteリポジトリの初期化
	noteRepo, err := note.NewNoteRepository(initDB())
	if err != nil {
		log.Fatal(err)
	}
	cli.Nr = noteRepo

	// サブコマンドを登録
	ctx := context.Background()
	os.Exit(cli.InitCommands(ctx))
}
