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
	mock  string
	debug bool
)

const defaultLogPath = "/tmp/siggo.log"

func init() {
	rootCmd.PersistentFlags().StringVarP(&mock, "mock", "m", "", "mock mode (uses example data)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug logging")
}

func initLogging(cfg *model.Config) {
	if cfg.LogFilePath == "" {
		cfg.LogFilePath = defaultLogPath
	}
	logFile, err := os.Create(cfg.LogFilePath)
	if err != nil {
		log.Fatalf("error creating log file: %v %v", cfg.LogFilePath, err)
	}
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(logFile)
}

func setupMock(mockFileName string, cfg *model.Config) *signal.MockSignal {
	b, err := ioutil.ReadFile(mock)
	if err != nil {
		log.Fatalf("couldn't open mock data: %v %v", mock, err)
	}
	return signal.NewMockSignal(cfg.UserNumber, b)
}

var rootCmd = &cobra.Command{
	Use:   "siggo",
	Short: "siggo is a terminal gui for signal-cli",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatalf("failed to read config @ %s", model.DefaultConfigPath())
		}

		initLogging(cfg)

		if cfg.UserNumber == "" {
			log.Fatalf("no user phone number configured @ %s", model.DefaultConfigPath())
		}

		var signalAPI model.SignalAPI = signal.NewSignal(cfg.UserNumber)
		if mock != "" {
			signalAPI = setupMock(mock, cfg)
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
