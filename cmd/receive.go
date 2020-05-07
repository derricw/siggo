package cmd

import (
	"io/ioutil"

	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/signal"
	log "github.com/sirupsen/logrus"
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

		var signalAPI model.SignalAPI = signal.NewSignal(User)
		if Mock != "" {
			b, err := ioutil.ReadFile(Mock)
			if err != nil {
				log.Fatalf("couldn't open mock data")
			}
			signalAPI = signal.NewMockSignal(User, b)
		}

		s := model.NewSiggo(signalAPI, cfg)

		s.NewInfo = func(conv *model.Conversation) {
			log.Printf("From: %v | Conv: \n%s", conv.Contact, conv.String())
		}
		s.ReceiveForever()
		<-make(chan struct{})
	},
}
