package cmd

import (
	"fmt"
	"github.com/derricw/siggo/signal"
	"github.com/derricw/siggo/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of siggo",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Build Date:", version.BuildDate)
		fmt.Println("Git Commit:", version.GitCommit)
		fmt.Println("Version:", version.Version)
		fmt.Println("Go Version:", version.GoVersion)
		fmt.Println("OS/Arch:", version.OsArch)
		fmt.Printf("signal-cli Version: ")

		sig := &signal.Signal{}
		signalVersion, err := sig.Version()
		if err != nil {
			fmt.Printf("Unknown\b")
		}
		fmt.Printf("%s", signalVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
