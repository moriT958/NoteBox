/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"notebox/filesystem"
	"notebox/model"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new note",
	Long: fmt.Sprintf(
		"new command creates new note. Arg Title is needed. \nTitle will be the headline of markdonw file.\n",
	),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		newNote := model.Note{
			ID:      uuid.NewString(),
			Title:   title,
			Content: "",
		}

		if err := filesystem.Store.Save(newNote); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
