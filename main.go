package main

import (
	"database/sql"
	"fmt"
	"log"
	"notebox/cli"
	"notebox/config"
	"notebox/models"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	// ctx := context.Background()
	// os.Exit(cli.InitCommands(ctx))

	// BubbleTea用
	notes, err := noteRepo.FindAll()
	if err != nil {
		log.Fatal(err)
	}

	items := make([]list.Item, len(notes))
	for i, n := range notes {
		items[i] = n
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "My Fave Things"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
