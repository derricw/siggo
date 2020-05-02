package cmd

import (
	"io/ioutil"
	"os"

	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/signal"
	"github.com/derricw/siggo/ui"
)

func init() {
	rootCmd.AddCommand(uiCmd)
}

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, err := os.Create("/tmp/siggo.log")
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(outputFile)

		cfg := &model.Config{
			UserNumber: User,
			UserName:   "me",
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

		//err := s.Receive()
		//if err != nil {
		//log.Fatal(err)
		//}
		s.ReceiveUntil(make(chan struct{}))
		//<-make(chan struct{})

		//log.Printf("contacts: %v", s.Contacts())

		chatWindow := ui.NewChatWindow(s)
		if err := tview.NewApplication().SetRoot(chatWindow, true).SetFocus(chatWindow).Run(); err != nil {
			panic(err)
		}
	},
}
