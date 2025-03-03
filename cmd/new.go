package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create New Note.",
	Long:  `Writing...`,
	Args:  validateArgNum,
	Run: func(cmd *cobra.Command, args []string) {
		var title string

		if len(args) == 0 {
			fmt.Print("Enter Title: ")
			fmt.Scan(&title)
		} else {
			title = args[0]
		}

		// Create Markdown file
		noteDir := "./data" // TODO: 存在しない場合は作成
		topHeader := "# " + title + "\n\n"
		// TODO: ファイル名を時刻+タイトルにする
		// とりあえずuuidでいっとく
		notefile := filepath.Join(noteDir, uuid.NewString()+".md")

		// TODO:
		// すでに同じタイトルのノートが存在する場合はエラー
		f, err := os.Create(notefile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create file: %v", err)
		}
		defer f.Close()

		fmt.Fprint(f, topHeader)

		// TODO: lipglossでUI強化
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5F87")).
			Background(lipgloss.Color("#1E1E1E")).
			Padding(1, 2).
			Margin(1).
			Border(lipgloss.RoundedBorder())
		fmt.Println(titleStyle.Render(fmt.Sprintf("✨ ID:%d | %s Created ✨\n", 1, title)))
	},
}

func validateArgNum(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("only one argument is needed")
	}
	return nil
}
