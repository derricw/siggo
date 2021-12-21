package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

// NewFilterInput is a command input that lets you filter the current conversation
func NewFilterInput(parent *ChatWindow) *CommandInput {
	ci := &CommandInput{
		InputField: tview.NewInputField(),
		parent:     parent,
	}
	ci.SetLabel("/")
	ci.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	ci.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Debugf("Key Event <FILTER>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyESC:
			ci.parent.conversationPanel.Filter("")
			ci.parent.HideCommandInput()
			return nil
		case tcell.KeyEnter:
			s := ci.GetText()
			if s != "" {
				ci.parent.conversationPanel.Filter(s)
			}
			ci.parent.NormalMode()
			ci.parent.update()
			return nil
		}
		return event
	})
	return ci
}
