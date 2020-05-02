package cmd

import (
	"log"

	"github.com/derricw/siggo/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(receiveCmd)
}

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "receive all outstanding messages",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &model.Config{
			UserNumber: User,
		}
		s := model.NewSiggo(cfg)
		s.NewInfo = func(conv *model.Conversation) {
			log.Printf("From: %v\nConv: %s", conv.Contact, conv.String())
		}
		err := s.Receive()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%+v", s)
	},
}
