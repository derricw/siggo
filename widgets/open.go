package widgets

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"

	"github.com/derricw/siggo/model"
)

// OpenInput is a widget that allows us to select an attachment to open.
type OpenInput struct {
	*tview.List
	parent      *ChatWindow
	attachments []*model.Attachment
}

func (oi *OpenInput) Close() {
	oi.parent.Grid.RemoveItem(oi)
	oi.parent.ShowConversation()
	oi.parent.FocusMe()
}

// init populates the list with attachments
func (oi *OpenInput) init() {
	oi.Clear()
	a := oi.parent.getAttachments() // should be sorted by date
	oi.attachments = a
	for _, item := range a {
		var text string
		if item.FromSelf {
			text = fmt.Sprintf(" <- %s", item.String())
		} else {
			text = fmt.Sprintf(" -> %s", item.String())
		}
		oi.AddItem(text, "", 0, nil)
	}
}

func (oi *OpenInput) Previous() {
	current := oi.GetCurrentItem()
	oi.SetCurrentItem(current - 1)
}

func (oi *OpenInput) Next() {
	current := oi.GetCurrentItem()
	oi.SetCurrentItem(current + 1)
}

// OpenLastAttachment opens the last attachment that it finds in the conversation
func (oi *OpenInput) OpenLast() {
	oi.Close()
	attachments := oi.attachments
	if len(attachments) > 0 {
		last := attachments[len(attachments)-1]
		oi.OpenAttachment(last)
	} else {
		oi.parent.SetStatus(fmt.Sprintf("📎<NO MATCHES>"))
	}
}

// OpenSelected opens whichever attachment is selected
func (oi *OpenInput) OpenSelected() {
	nattach := len(oi.attachments)
	selected := oi.GetCurrentItem()
	if nattach == 0 || selected >= nattach {
		return
	}
	oi.OpenAttachment(oi.attachments[selected])
}

// OpenAttachment opens a `*signal.Attachment`
func (oi *OpenInput) OpenAttachment(attachment *model.Attachment) {
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
		List:   tview.NewList(),
		parent: parent,
	}
	inputHandler := oi.List.InputHandler()
	oi.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Debugf("Key Event <OPEN>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyESC:
			oi.Close()
			oi.parent.NormalMode()
			return nil
		//case tcell.KeyUp:
		//oi.Next()
		//return nil
		//case tcell.KeyDown:
		//oi.Previous()
		//return nil
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
				oi.Next()
				return nil
			case 107: // k
				oi.Previous()
				return nil
			}
		case tcell.KeyEnter:
			oi.OpenSelected()
			return nil
		}

		return event
	})

	//oi.SetDynamicColors(true)
	oi.SetHighlightFullLine(true)
	oi.ShowSecondaryText(false)
	oi.SetBorder(true)
	oi.SetTitle(fmt.Sprintf("attachments: %s", parent.currentContactName()))
	oi.SetTitleAlign(0)
	oi.init()
	oi.SetCurrentItem(-1)

	return oi
}
