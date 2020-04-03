package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/derricw/siggo/signal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(receiveCmd)
}

func printMsg(msg *signal.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", b)
	return nil
}

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "receive all outstanding messages",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		sig := signal.NewSignal(User)
		sig.OnMessage(printMsg)
		err := sig.Receive()
		if err != nil {
			log.Fatal(err)
		}
	},
}
