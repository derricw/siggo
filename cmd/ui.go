package cmd

import (
	"github.com/rivo/tview"
	"github.com/spf13/cobra"

	"github.com/derricw/siggo/model"
	"github.com/derricw/siggo/ui"
)

func init() {
	rootCmd.AddCommand(uiCmd)
}

var uiCmd = &cobra.Command{
	Use:   "uitest",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		cfg := &model.Config{
			UserNumber: User,
			UserName:   "me",
		}
		siggo := model.NewSiggo(cfg)

		chatWindow := ui.NewChatWindow(siggo)
		if err := tview.NewApplication().SetRoot(chatWindow, true).SetFocus(chatWindow).Run(); err != nil {
			panic(err)
		}
	},
}
