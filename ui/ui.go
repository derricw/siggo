package ui

import (
	"fmt"

	"github.com/derricw/siggo/model"
	"github.com/gdamore/tcell"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type ConvInfo map[*model.Contact]*model.Conversation

type ChatWindow struct {
	// todo: maybe use Flex instead?
	*tview.Grid
	siggo          *model.Siggo
	currentContact *model.Contact

	sendPanel         *SendPanel
	contactsPanel     *ContactListPanel
	conversationPanel *ConversationPanel
}

func (c *ChatWindow) ContactUp()   {}
func (c *ChatWindow) ContactDown() {}

func (c *ChatWindow) send(msg string) {
	// send message to the current contact
	c.siggo.Send(msg, c.currentContact)
}

func (c *ChatWindow) update() {
	convs := c.siggo.Conversations()
	if convs != nil {
		c.contactsPanel.Update(convs)
		currentConv, ok := convs[c.currentContact]
		if ok {
			//log.Printf("updating convs: %v", convs)
			c.conversationPanel.Update(currentConv)
		} else {
			panic("no conversation for current contact")
		}
	}
}

type SendPanel struct {
	//*tview.TextView
	//*tview.InputField
	*femto.View
	sendCallbacks []func(string)
	KeyEvent      func(*tcell.EventKey) *tcell.EventKey
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
		// first let anyone hooked in handle the event
		if s.KeyEvent != nil {
			log.Printf("key handling...")
			e := s.KeyEvent(event)
			if e == nil {
				return nil
			}
		}
		// then handle it ourselves
		switch event.Key() {
		case tcell.KeyEnter:
			if event.Modifiers() == 4 {
				s.Send()
				return nil
			}
			return event
		case tcell.KeyCtrlQ:
			return nil
		}
		return event
	})
	return s
}

type ContactListPanel struct {
	*tview.TextView
}

func (p *ContactListPanel) Update(convs ConvInfo) {
	data := ""
	log.Printf("updating contact panel...")
	for c, _ := range convs {
		id := ""
		if c.Name != "" {
			id = c.Name
		} else {
			id = c.Number
		}
		data += fmt.Sprintf("%s\n", id)
	}
	p.SetText(data)
}

func NewContactListPanel() *ContactListPanel {
	c := &ContactListPanel{
		TextView: tview.NewTextView(),
	}
	c.SetTitle("contacts")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	return c
}

type ConversationPanel struct {
	*tview.TextView
}

func (p *ConversationPanel) Update(conv *model.Conversation) {
	p.SetText(conv.String())
	p.SetTitle(fmt.Sprintf("%s <%s>", conv.Contact.Name, conv.Contact.Number))
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
	w := &ChatWindow{
		Grid:  layout,
		siggo: siggo,
	}

	w.conversationPanel = NewConversationPanel()
	w.contactsPanel = NewContactListPanel()
	w.sendPanel = NewSendPanel()
	// Setup keys
	// TODO: lets move this somewhere better so we can merge all of our keybinds
	w.sendPanel.KeyEvent = func(event *tcell.EventKey) *tcell.EventKey {
		log.Printf("Key Event: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyDown:
			log.Printf("key Down...")
			if event.Modifiers() == 4 {
				w.ContactDown()
				return nil
			}
			return event
		case tcell.KeyUp:
			log.Printf("key Up...")
			if event.Modifiers() == 4 {
				w.ContactUp()
				return nil
			}
			return event
		case tcell.KeyCtrlQ:
			return nil
		}
		return event
	}

	// primitiv, row, col, rowSpan, colSpan, minGridHeight, maxGridHeight, focus)
	// TODO: lets make some of the spans confiurable?
	w.AddItem(w.contactsPanel, 0, 0, 2, 1, 0, 0, false)
	w.AddItem(w.conversationPanel, 0, 1, 1, 1, 0, 0, false)
	w.AddItem(w.sendPanel, 1, 1, 1, 1, 0, 0, true)

	w.siggo = siggo
	contacts := siggo.Contacts()
	// get initial contact here by saving the last one
	// for now we just get one at random
	for _, c := range contacts {
		w.currentContact = c
		break
	}
	// update gui when events happen in siggo
	w.update()
	siggo.NewInfo = func(conv *model.Conversation) {
		w.update()
	}

	return w
}
