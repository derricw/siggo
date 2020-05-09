package cmd

import (
	"fmt"

	"github.com/derricw/siggo/signal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkCmd)
}

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "link a device",
	Long: `Generates a QR code that can be scanned by an existing signal device to link to.
	Example:
	$ siggo link work_laptop`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("linking...")
		sig := signal.NewSignal(User)
		sig.Link(args[0])
	},
}
