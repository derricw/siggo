package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/derricw/siggo/signal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(processCmd)
}

func printMsg(msg *signal.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", b)
	return nil
}

func printSent(msg *signal.Message) error {
	sentMsg := msg.Envelope.SyncMessage.SentMessage
	fmt.Printf("MESSAGE SENT:\n")
	fmt.Printf("  TO: %s\n", sentMsg.Destination)
	fmt.Printf("  Message: %s\n", sentMsg.Message)
	fmt.Printf("  Timestamp: %d\n", sentMsg.Timestamp)
	return nil
}

func printReceived(msg *signal.Message) error {
	receiveMsg := msg.Envelope.DataMessage
	fmt.Printf("MESSAGE Received:\n")
	fmt.Printf("  From: %s\n", msg.Envelope.Source)
	fmt.Printf("  Message: %s\n", receiveMsg.Message)
	fmt.Printf("  Timestamp: %d\n", receiveMsg.Timestamp)
	return nil
}

func printReceipt(msg *signal.Message) error {
	receiptMsg := msg.Envelope.ReceiptMessage
	fmt.Printf("RECEIPT Received:\n")
	fmt.Printf("  From: %s\n", msg.Envelope.Source)
	fmt.Printf("  Delivered: %t\n", receiptMsg.IsDelivery)
	fmt.Printf("  Read: %t\n", receiptMsg.IsRead)
	fmt.Printf("  Timestamps: %v\n", receiptMsg.Timestamps)
	return nil
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "process a stream of messages from stdin",
	Long: `example:
	signal-cli -u +12067902360 receive --json | siggo process`,
	Run: func(cmd *cobra.Command, args []string) {
		sig := signal.NewSignal("")
		sig.OnMessage(printMsg)
		sig.OnSent(printSent)
		sig.OnReceived(printReceived)
		sig.OnReceipt(printReceipt)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			wire := scanner.Bytes()
			err := sig.ProcessWire(wire)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}
