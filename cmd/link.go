package cmd

import (
	"fmt"

	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/signal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkCmd)
}

var linkCmd = &cobra.Command{
	Use:   "link <phone number> <device name>",
	Short: "link a device",
	Long: `Generates a QR code that can be scanned by an existing signal device to link to.
	Example:
	$ siggo link +1234567890 work_laptop`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("linking...")
		fmt.Println("In the mobile app, go to Settings -> Linked Devices -> Add")
		sig := signal.NewSignal(args[0])
		sig.Link(args[1])

		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatalf("failed to read config @ %s", model.ConfigPath())
		}
		cfg.UserNumber = args[0]
		cfg.Save()
	},
}
