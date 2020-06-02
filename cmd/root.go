package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/signal"
	"github.com/derricw/siggo/widgets"
)

var (
	Mock  string
	Debug bool
)

const defaultLogPath = "/tmp/siggo.log"

func init() {
	rootCmd.PersistentFlags().StringVarP(&Mock, "mock", "m", "", "mock mode (uses example data)")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug logging")
}

func initLogging(cfg *model.Config) {
	if cfg.LogFilePath == "" {
		cfg.LogFilePath = defaultLogPath
	}
	logFile, err := os.Create(cfg.LogFilePath)
	if err != nil {
		log.Fatal("error creating log file: %v %v", cfg.LogFilePath, err)
	}
	if Debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(logFile)
}

var rootCmd = &cobra.Command{
	Use:   "siggo",
	Short: "siggo is a terminal gui for signal-cli",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatal("failed to read config @ %s", model.DefaultConfigPath())
		}

		initLogging(cfg)

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
		chatWindow := widgets.NewChatWindow(s, app)

		// finally, start the tview app
		if err := app.SetRoot(chatWindow, true).SetFocus(chatWindow).Run(); err != nil {
			panic(err)
		}
		// clean up when we're done
		chatWindow.Quit()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
