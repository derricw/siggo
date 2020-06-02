package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/signal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(convCmd)
}

var convCmd = &cobra.Command{
	Use:   "conv <contact>",
	Short: "prints the saved conversation for a given contact",
	Long: `example:
	$ siggo conv +1234567890
	$ siggo conv "John Smith"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatalf("failed to read config @ %s", model.DefaultConfigPath())
		}
		if cfg.UserNumber == "" {
			log.Fatalf("no user phone number configured @ %s", model.DefaultConfigPath())
		}

		var signalAPI model.SignalAPI = signal.NewSignal(cfg.UserNumber)
		if Mock != "" {
			b, err := ioutil.ReadFile(Mock)
			if err != nil {
				log.Fatalf("couldn't open mock data")
			}
			signalAPI = signal.NewMockSignal(cfg.UserNumber, b)
		}
		s := model.NewSiggo(signalAPI, cfg)
		if Mock != "" {
			s.Receive()
		}

		var conv *model.Conversation
		if contact, ok := s.Contacts()[args[0]]; ok {
			// arg is a number, get conversation directly
			conv = s.Conversations()[contact]
		} else {
			// maybe a name, have to scan list
			for _, contact := range s.Contacts() {
				if contact.Name == args[0] {
					conv = s.Conversations()[contact]
				}
			}
		}
		if conv != nil {
			fmt.Printf("%s", conv.String())
		} else {
			log.Fatalf("failed to find conversation")
		}
	},
}
