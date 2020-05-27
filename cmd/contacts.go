package cmd

import (
	"fmt"

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
	$ siggo contacts`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatal("failed to read config @ %s", model.DefaultConfigPath())
		}
		if cfg.UserNumber == "" {
			log.Fatalf("no user phone number configured @ %s", model.DefaultConfigPath())
		}
		signalAPI := signal.NewSignal(cfg.UserNumber)
		s := model.NewSiggo(signalAPI, cfg)

		for _, c := range s.Contacts().SortedByName() {
			fmt.Printf("%s - %s\n", c.Name, c.Number)
		}
	},
}
