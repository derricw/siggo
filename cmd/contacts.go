package cmd

import (
	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/signal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(contactsCmd)
}

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "list contacts for a given user",
	Long: `example:
	$ siggo contacts +1234567890`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		user := args[0]

		cfg := &model.Config{
			UserNumber: user,
		}

		signalAPI := signal.NewSignal(user)
		s := model.NewSiggo(signalAPI, cfg)

		log.Info("%s", s.Contacts())
	},
}
