package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	ossig "os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
		cfg.LogFilePath = model.LogPath()
	}

	dir := filepath.Dir(cfg.LogFilePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create folder: %s", dir)
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

func hasSignalCLI() bool {
	_, err := exec.LookPath("signal-cli")
	return err == nil
}

var rootCmd = &cobra.Command{
	Use:   "siggo",
	Short: "siggo is a terminal gui for signal-cli",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if !hasSignalCLI() {
			log.Fatalf("failed to find signal-cli in PATH")
		}

		cfg, err := model.GetConfig()
		if err != nil {
			log.Fatalf("failed to read config @ %s", model.ConfigPath())
		}

		if cfg.UserNumber == "" {
			log.Fatalf("no user phone number configured @ %s", model.ConfigPath())
		}

		if !strings.HasPrefix(cfg.UserNumber, "+") {
			cfg.UserNumber = fmt.Sprintf("+%s", cfg.UserNumber)
		}

		if len(cfg.UserNumber) < 12 {
			log.Fatalf("user phone number: %s is too short. did you forget a country code?", cfg.UserNumber)
		}

		initLogging(cfg)

		var signalAPI model.SignalAPI = signal.NewSignal(cfg.UserNumber)
		if mock != "" {
			signalAPI = setupMock(mock, cfg)
		}
		defer signalAPI.Close()

		s := model.NewSiggo(signalAPI, cfg)

		s.ReceiveForever()
		//tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
		app := tview.NewApplication()
		chatWindow := widgets.NewChatWindow(s, app)

		// also want to make sure to handle signals
		sigChan := make(chan os.Signal, 1)
		ossig.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGIOT,
			syscall.SIGQUIT, syscall.SIGSEGV) // doesn't catch syscall.SIGKILL but might as well include it
		go func() {
			s := <-sigChan
			log.Infof("caught signal: %s", s)
			chatWindow.Quit()
		}()

		// finally, start the tview app
		if err := app.SetRoot(chatWindow, true).SetFocus(chatWindow).Run(); err != nil {
			signalAPI.Close() // redundant?
			panic(err)
		}
		// clean up when we're done
		chatWindow.Quit()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
