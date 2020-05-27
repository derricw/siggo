package cmd

import (
	"io/ioutil"
	"os"
	//ossig "os/signal"
	//"syscall"

	//"github.com/gdamore/tcell"
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
		if Debug {
			log.SetLevel(log.DebugLevel)
		}
		log.SetOutput(outputFile)

		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatal("failed to read config @ %s", model.DefaultConfigPath())
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

		s.ReceiveForever()
		//tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
		app := tview.NewApplication()
		chatWindow := ui.NewChatWindow(s, app)

		// finally, start the tview app
		if err := app.SetRoot(chatWindow, true).SetFocus(chatWindow).Run(); err != nil {
			panic(err)
		}
		// clean up when we're done
		chatWindow.Quit()
	},
}
