package ui

import (
	//"log"

	"github.com/derricw/siggo/model"
	"github.com/gdamore/tcell"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
)

type ChatWindow struct {
	// todo: maybe use Flex instead?
	*tview.Grid
	siggo          *model.Siggo
	currentContact *model.Contact

	sendPanel         *SendPanel
	contactsPanel     *ContactListPanel
	conversationPanel *ConversationPanel
}

func (c *ChatWindow) send(msg string) {
	// send message to the current contact
	c.siggo.Send(msg, c.currentContact)
}

type SendPanel struct {
	//*tview.TextView
	//*tview.InputField
	*femto.View
	sendCallbacks []func(string)
}

func (s *SendPanel) Send() {
	// publish message to anyone listening
	msg := s.Buf.LineArray.String()
	if len(s.sendCallbacks) > 0 {
		for _, cb := range s.sendCallbacks {
			cb(msg)
		}
	}
	// clear input buffer
	s.OpenBuffer(femto.NewBufferFromString("", ""))
}

func NewSendPanel() *SendPanel {
	s := &SendPanel{
		//TextView: tview.NewTextView(),
		//InputField: tview.NewInputField(),
		View: femto.NewView(femto.NewBufferFromString("", "")),
	}
	s.SetTitle(" send: ")
	s.SetTitleAlign(0)
	s.SetBorder(true)
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if event.Modifiers() == 4 {
				s.Send()
				return nil
			}
			return event
		case tcell.KeyCtrlS:
			return nil
		case tcell.KeyCtrlQ:
			return nil
		}
		return event
	})
	return s
}

type ContactListPanel struct {
	*tview.List
}

func NewContactListPanel() *ContactListPanel {
	c := &ContactListPanel{
		List: tview.NewList(),
	}
	c.SetTitle("contacts")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	return c
}

type ConversationPanel struct {
	*tview.TextView
}

func NewConversationPanel() *ConversationPanel {
	c := &ConversationPanel{
		TextView: tview.NewTextView(),
	}
	c.SetTitle("<name of contact>")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	return c
}

func NewChatWindow(siggo *model.Siggo) *ChatWindow {
	layout := tview.NewGrid().
		SetRows(0, 8).
		SetColumns(20, 0)
	window := &ChatWindow{
		Grid:  layout,
		siggo: siggo,
	}

	conversation := NewConversationPanel()
	contacts := NewContactListPanel()
	send := NewSendPanel()

	// primitiv, row, col, rowSpan, colSpan, minGridHeight, maxGridHeight, focus)
	// TODO: lets make some of the spans confiurable?
	window.AddItem(contacts, 0, 0, 2, 1, 0, 0, false)
	window.AddItem(conversation, 0, 1, 1, 1, 0, 0, false)
	window.AddItem(send, 1, 1, 1, 1, 0, 0, true)

	window.siggo = siggo

	return window
}
