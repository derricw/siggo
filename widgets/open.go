package widgets

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"

	"github.com/derricw/siggo/signal"
)

// OpenInput is a widget that allows us to select an attachment to open.
type OpenInput struct {
	*tview.TextView
	parent          *ChatWindow
	selected        int
	numAttachements int
}

func (oi *OpenInput) Close() {
	oi.parent.Grid.RemoveItem(oi)
	oi.parent.ShowConversation()
	oi.parent.FocusMe()
}

func (oi *OpenInput) Render() {
	a := oi.parent.getAttachments() // should be sorted by date
	oi.numAttachements = len(a)
	if oi.numAttachements == 0 {
		return
	}

	text := ""
	for i := oi.numAttachements - 1; i >= 0; i-- {
		item := fmt.Sprintf("%s", a[i])
		if i == oi.selected {
			item = fmt.Sprintf("[::r]%s[::-]", item)
		}
		text += item
	}
	oi.SetText(text)
}

func (oi *OpenInput) Previous() {
	oi.selected--
	if oi.selected < 0 {
		oi.selected = 0
	}
	oi.Render()
}

func (oi *OpenInput) Next() {
	oi.selected++
	if oi.selected >= oi.numAttachements {
		oi.selected = oi.numAttachements - 1
	}
	oi.Render()
}

// OpenLastAttachment opens the last attachment that it finds in the conversation
func (oi *OpenInput) OpenLast() {
	oi.Close()
	attachments := oi.parent.getAttachments()
	if len(attachments) > 0 {
		last := attachments[len(attachments)-1]
		oi.OpenAttachment(last)
	} else {
		oi.parent.SetStatus(fmt.Sprintf("📎<NO MATCHES>"))
	}
}

// OpenSelected opens whichever attachment is selected
// TODO: should this also `Close()` and return to normal mode?
func (oi *OpenInput) OpenSelected() {
	attachments := oi.parent.getAttachments()
	if len(attachments) == 0 || oi.selected >= len(attachments) {
		return
	}
	oi.OpenAttachment(attachments[oi.selected])
}

// OpenAttachment opens a `*signal.Attachment`
func (oi *OpenInput) OpenAttachment(attachment *signal.Attachment) {
	path, err := attachment.Path()
	if err != nil {
		oi.parent.SetErrorStatus(fmt.Errorf("📎failed to find attachment: %v", err))
		return
	}
	oi.OpenPath(path)
}

// OpenAttachment opens any file @ path using xdg-open
func (oi *OpenInput) OpenPath(path string) {
	go func() {
		err := open.Run(path)
		if err != nil {
			oi.parent.SetErrorStatus(fmt.Errorf("📎<OPEN FAILED: %v>", err))
		} else {
			oi.parent.SetStatus(fmt.Sprintf("📎%s", path))
		}
	}()
}

func NewOpenInput(parent *ChatWindow) *OpenInput {
	oi := &OpenInput{
		TextView: tview.NewTextView(),
		parent:   parent,
	}
	inputHandler := oi.TextView.InputHandler()
	oi.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Debugf("Key Event <OPEN>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyESC:
			oi.Close()
			oi.parent.NormalMode()
			return nil
		case tcell.KeyUp:
			oi.Next()
			return nil
		case tcell.KeyDown:
			oi.Previous()
			return nil
		case tcell.KeyPgUp:
			inputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyPgDn:
			inputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyEnd:
			inputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyHome:
			inputHandler(event, func(p tview.Primitive) {})
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case 111: // o
				oi.OpenLast()
				return nil
			case 106: // j
				oi.Previous()
				return nil
			case 107: // k
				oi.Next()
				return nil
			}

		case tcell.KeyEnter:
			oi.OpenSelected()
			return nil
		}

		return event
	})

	oi.SetDynamicColors(true)
	oi.SetBorder(true)

	oi.Render()

	return oi
}
