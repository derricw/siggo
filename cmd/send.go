package cmd

import (
	"log"

	"github.com/derricw/siggo/signal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sendCmd)
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send a single message",
	Long:  `$siggo -u +1098765432 send +1234567890 "hello good sir"`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sig := signal.NewSignal(User)
		err := sig.Send(args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}
	},
}
