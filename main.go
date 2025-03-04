package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

const configFile = "./config.json"

func main() {
	// configファイルをロードする
	cfg, err := NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config file: %v\n", err)
		os.Exit(1)
	}

	// TODO:
	// ここでdata.jsonをロードし、リストを作成

	// サブコマンドを登録
	subcommands.Register(&newCmd{cfg: cfg}, "")
	subcommands.Register(&lsCmd{cfg: cfg}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
