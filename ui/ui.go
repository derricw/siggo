package ui

import (
	"fmt"

	"github.com/derricw/siggo/model"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
	YankMode
)

type ConvInfo map[*model.Contact]*model.Conversation

type ChatWindow struct {
	// todo: maybe use Flex instead?
	*tview.Grid
	siggo          *model.Siggo
	currentContact *model.Contact
	mode           Mode

	sendPanel         *SendPanel
	contactsPanel     *ContactListPanel
	conversationPanel *ConversationPanel
	app               *tview.Application
}

func (c *ChatWindow) InsertMode() {
	log.Info("INSERT MODE")
	c.app.SetFocus(c.sendPanel)
	c.sendPanel.SetBorderColor(tcell.ColorOrange)
	c.mode = InsertMode
}

func (c *ChatWindow) YankMode() {
	log.Info("YANK MODE")
	c.conversationPanel.SetBorderColor(tcell.ColorOrange)
	c.mode = YankMode
}

func (c *ChatWindow) NormalMode() {
	log.Info("NORMAL MODE")
	c.app.SetFocus(c)
	// clear our highlights
	c.conversationPanel.SetBorderColor(tcell.ColorWhite)
	c.sendPanel.SetBorderColor(tcell.ColorWhite)
	c.mode = NormalMode
}

// TODO: remove code duplication with ContactDown()
func (c *ChatWindow) ContactUp() {
	log.Info("PREVIOUS CONVERSATION")
	prevContact := c.contactsPanel.Previous()
	if prevContact != c.currentContact {
		c.currentContact = prevContact
		c.contactsPanel.Update()
		currentConv, ok := c.siggo.Conversations()[c.currentContact]
		if ok {
			c.conversationPanel.Update(currentConv)
			currentConv.CaughtUp()
		}
	}
}

func (c *ChatWindow) ContactDown() {
	log.Info("NEXT CONVERSATION")
	nextContact := c.contactsPanel.Next()
	if nextContact != c.currentContact {
		c.currentContact = nextContact
		c.contactsPanel.Update()
		currentConv, ok := c.siggo.Conversations()[c.currentContact]
		if ok {
			c.conversationPanel.Update(currentConv)
			currentConv.CaughtUp()
		}
	}
}

func (c *ChatWindow) update() {
	convs := c.siggo.Conversations()
	if convs != nil {
		c.contactsPanel.Update()
		currentConv, ok := convs[c.currentContact]
		if ok {
			c.conversationPanel.Update(currentConv)
		} else {
			panic("no conversation for current contact")
		}
	}
}

type SendPanel struct {
	*tview.InputField
	parent *ChatWindow
	siggo  *model.Siggo
}

func (s *SendPanel) Send() {
	msg := s.GetText()
	contact := s.parent.currentContact
	go s.siggo.Send(msg, contact)
	log.Infof("sent message: %s to contact: %s", msg, contact)
	s.SetText("")
}

func (s *SendPanel) Defocus() {
	s.parent.NormalMode()
}

func NewSendPanel(parent *ChatWindow, siggo *model.Siggo) *SendPanel {
	s := &SendPanel{
		InputField: tview.NewInputField(),
		siggo:      siggo,
		parent:     parent,
	}
	s.SetTitle(" send: ")
	s.SetTitleAlign(0)
	s.SetBorder(true)
	//s.SetFieldBackgroundColor(tcell.ColorDefault)
	s.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			s.Defocus()
			return nil
		case tcell.KeyEnter:
			s.Send()
			return nil
		}
		return event
	})
	return s
}

type ContactListPanel struct {
	*tview.TextView
	siggo          *model.Siggo
	parent         *ChatWindow
	sortedContacts []*model.Contact
	currentIndex   int
}

func (cl *ContactListPanel) Next() *model.Contact {
	if cl.currentIndex < len(cl.sortedContacts)-1 {
		cl.currentIndex++
	}
	return cl.sortedContacts[cl.currentIndex]
}

func (cl *ContactListPanel) Previous() *model.Contact {
	if cl.currentIndex > 0 {
		cl.currentIndex--
	}
	return cl.sortedContacts[cl.currentIndex]
}

func (cl *ContactListPanel) Update() {
	data := ""
	log.Printf("updating contact panel...")
	// this is dumb, we re-sort every update
	// TODO: don't
	sorted := cl.siggo.Contacts().SortedByIndex()
	convs := cl.siggo.Conversations()
	log.Printf("sorted contacts: %v", sorted)
	//log.Printf("current contact idx: %v", cl.currentIndex)
	for i, c := range sorted {
		id := ""
		if c.Name != "" {
			id = c.Name
		} else {
			id = c.Number
		}
		line := fmt.Sprintf("%s\n", id)
		if cl.currentIndex == i {
			line = "[::r]" + line + "[::-]"
			cl.currentIndex = i
		} else if convs[c].HasNewMessage {
			line = "[::b]*" + line + "[::-]"
		}
		data += line
	}
	//log.Printf("data: %s", data)
	cl.sortedContacts = sorted
	cl.SetText(data)
}

//func (cl *ContactListPanel)

// NewContactListPanel creates a new contact list widget
func NewContactListPanel(parent *ChatWindow, siggo *model.Siggo) *ContactListPanel {
	c := &ContactListPanel{
		TextView: tview.NewTextView(),
		siggo:    siggo,
		parent:   parent,
	}
	c.SetDynamicColors(true)
	c.SetTitle("contacts")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	return c
}

type ConversationPanel struct {
	*tview.TextView
}

func (p *ConversationPanel) Update(conv *model.Conversation) {
	p.Clear()
	p.SetText(conv.String())
	p.SetTitle(fmt.Sprintf("%s <%s>", conv.Contact.Name, conv.Contact.Number))
}

func (p *ConversationPanel) Clear() {
	p.SetText("")
}

func NewConversationPanel(siggo *model.Siggo) *ConversationPanel {
	c := &ConversationPanel{
		TextView: tview.NewTextView(),
	}
	c.SetDynamicColors(true)
	c.SetTitle("<name of contact>")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	return c
}

func NewChatWindow(siggo *model.Siggo, app *tview.Application) *ChatWindow {
	layout := tview.NewGrid().
		SetRows(0, 3).
		SetColumns(20, 0)
	w := &ChatWindow{
		Grid:  layout,
		siggo: siggo,
		app:   app,
	}

	w.conversationPanel = NewConversationPanel(siggo)
	convInputHandler := w.conversationPanel.InputHandler()
	w.contactsPanel = NewContactListPanel(w, siggo)
	w.sendPanel = NewSendPanel(w, siggo)
	w.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Printf("Key Event <MAIN>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 106:
				w.ContactDown()
				return nil
			case 107:
				w.ContactUp()
				return nil
			case 105:
				w.InsertMode()
				return nil
			case 121:
				w.YankMode()
				return nil
			}
		// pass some events on to the conversation panel
		case tcell.KeyPgUp:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyPgDn:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyEnd:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyHome:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyESC:
			w.NormalMode()
			return nil
		}
		return event
	})

	// primitiv, row, col, rowSpan, colSpan, minGridHeight, maxGridHeight, focus)
	// TODO: lets make some of the spans confiurable?
	w.AddItem(w.contactsPanel, 0, 0, 2, 1, 0, 0, false)
	w.AddItem(w.conversationPanel, 0, 1, 1, 1, 0, 0, false)
	w.AddItem(w.sendPanel, 1, 1, 1, 1, 0, 0, false)

	w.siggo = siggo
	contacts := siggo.Contacts().SortedByIndex()
	if len(contacts) > 0 {
		w.currentContact = contacts[0]
	}
	// update gui when events happen in siggo
	w.update()
	siggo.NewInfo = func(conv *model.Conversation) {
		app.QueueUpdateDraw(func() {
			w.update()
		})
	}

	return w
}
