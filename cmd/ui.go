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

		siggo := &model.Siggo{}

		chatWindow := ui.NewChatWindow(siggo)
		if err := tview.NewApplication().SetRoot(chatWindow, true).SetFocus(chatWindow).Run(); err != nil {
			panic(err)
		}
	},
}

//newPrimitive := func(text string) tview.Primitive {
//return tview.NewTextView().
//SetTextAlign(tview.AlignCenter).
//SetText(text)
//}
//menu := newPrimitive("Menu")
//main := newPrimitive("Main content")
//sideBar := newPrimitive("Side Bar")

//grid := tview.NewGrid().
//SetRows(3, 0, 3).
//SetColumns(30, 0, 30).
//SetBorders(true).
//AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
//AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

//// Layout for screens narrower than 100 cells (menu and side bar are hidden).
//grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
//AddItem(main, 1, 0, 1, 3, 0, 0, false).
//AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

//// Layout for screens wider than 100 cells.
//grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
//AddItem(main, 1, 1, 1, 1, 0, 100, false).
//AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

//if err := tview.NewApplication().SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
//panic(err)
//}
