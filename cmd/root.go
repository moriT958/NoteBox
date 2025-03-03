package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "note",
	Short: "NoteBox is ...",
	Long: `A Fast and Flexible Static Site Generator built with
love by spf13 and friends in Go.
Complete documentation is available at https://gohugo.io/documentation/`,
	Run: func(cmd *cobra.Command, args []string) {
		println("NoteBox")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
