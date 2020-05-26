package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var User string
var Mock string
var Debug bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&User, "user", "u", "", "user (currently phone number)")
	rootCmd.PersistentFlags().StringVarP(&Mock, "mock", "m", "", "mock mode (uses example data)")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug logging")
}

var rootCmd = &cobra.Command{
	Use:   "siggo",
	Short: "siggo is a terminal gui for signal-cli",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
