package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkCmd)
}

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "link a device",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("linking...")
	},
}
