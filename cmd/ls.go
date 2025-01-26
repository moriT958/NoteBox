/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	_  = iota
	kb = 1 << (10 * iota)
	mb
	gb
	tb
)

func fileSizeString(s int64) string {
	var sizeStr string
	if s >= tb {
		sizeStr = fmt.Sprintf("%.2f TB", float64(s)/float64(tb))
	} else if s >= gb {
		sizeStr = fmt.Sprintf("%.2f GB", float64(s)/float64(gb))
	} else if s >= mb {
		sizeStr = fmt.Sprintf("%.2f MB", float64(s)/float64(mb))
	} else if s >= kb {
		sizeStr = fmt.Sprintf("%.2f KB", float64(s)/float64(kb))
	} else {
		sizeStr = fmt.Sprintf("%d B", s)
	}
	return sizeStr
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		notes, err := nb.FindAll()
		if err != nil {
			return err
		}

		fmt.Printf("notes_total: %d\n", len(notes))
		fmt.Println("==============================")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "title\tsize\tcreated_at")
		for i := range len(notes) {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				notes[i].Title,
				fileSizeString(notes[i].Size),
				notes[i].CreatedAt.Format("2006/01/02 15:04"),
			)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
