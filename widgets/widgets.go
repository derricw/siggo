package widgets

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

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
	searchPanel       tview.Primitive
	app               *tview.Application
}

func (c *ChatWindow) InsertMode() {
	log.Debug("INSERT MODE")
	c.app.SetFocus(c.sendPanel)
	c.sendPanel.SetBorderColor(tcell.ColorOrange)
	c.mode = InsertMode
}

func (c *ChatWindow) YankMode() {
	log.Debug("YANK MODE")
	c.conversationPanel.SetBorderColor(tcell.ColorOrange)
	c.mode = YankMode
}

func (c *ChatWindow) NormalMode() {
	log.Debug("NORMAL MODE")
	c.app.SetFocus(c)
	// clear our highlights
	c.conversationPanel.SetBorderColor(tcell.ColorWhite)
	c.sendPanel.SetBorderColor(tcell.ColorWhite)
	c.mode = NormalMode
}

func (c *ChatWindow) ShowContactSearch() {
	log.Debug("SHOWING CONTACT SEARCH")
	p := NewContactSearch(c)
	c.searchPanel = p
	c.SetRows(0, 3, p.maxHeight)
	c.AddItem(p, 2, 0, 1, 2, 0, 0, false)
	c.app.SetFocus(p)
}

func (c *ChatWindow) HideSearch() {
	log.Debug("HIDING SEARCH")
	c.RemoveItem(c.searchPanel)
	c.SetRows(0, 3)
	c.app.SetFocus(c)
}

// TODO: remove code duplication with ContactDown()
func (c *ChatWindow) ContactUp() {
	log.Debug("PREVIOUS CONVERSATION")
	prevContact := c.contactsPanel.Previous()
	if prevContact != c.currentContact {
		c.currentContact = prevContact
		c.contactsPanel.Update()
		currentConv, ok := c.siggo.Conversations()[c.currentContact]
		if ok {
			c.conversationPanel.Update(currentConv)
			currentConv.CaughtUp()
		}
		c.conversationPanel.ScrollToEnd()
	}
}

func (c *ChatWindow) ContactDown() {
	log.Debug("NEXT CONVERSATION")
	nextContact := c.contactsPanel.Next()
	if nextContact != c.currentContact {
		c.currentContact = nextContact
		c.contactsPanel.Update()
		currentConv, ok := c.siggo.Conversations()[c.currentContact]
		if ok {
			c.conversationPanel.Update(currentConv)
			currentConv.CaughtUp()
		}
		c.conversationPanel.ScrollToEnd()
	}
}

// Compose opens an EDITOR to compose a command. If any text is saved in the buffer,
// we send it as a message to the current conversation.
func (c *ChatWindow) Compose() {
	msg := ""
	success := c.app.Suspend(func() {
		msg = FancyCompose()
	})
	// need to sleep because there seems to be a race condition in tview
	// https://github.com/rivo/tview/issues/244
	time.Sleep(100 * time.Millisecond)
	if !success {
		log.Error("failed to suspend siggo")
		return
	}
	if msg != "" {
		contact := c.currentContact
		c.ShowTempSentMsg(msg)
		go c.siggo.Send(msg, contact)
		log.Infof("sending message: %s to contact: %s", msg, contact)
	}
}

// ShowTempSentMsg shows a temporary message when a message is sent but before delivery.
// Only displayed for the second or two after a message is sent.
func (c *ChatWindow) ShowTempSentMsg(msg string) {
	tmpMsg := &model.Message{
		Content:     msg,
		From:        " ~ ",
		Timestamp:   time.Now().Unix() * 1000,
		IsDelivered: false,
		IsRead:      false,
		FromSelf:    true,
	}
	// write directly to conv panel but don't add to conversation
	// no color since its from us
	c.conversationPanel.Write([]byte(tmpMsg.String("")))
}

// Quit shuts down gracefully
func (c *ChatWindow) Quit() {
	c.app.Stop()
	// do we need to do anything else?
	c.siggo.Quit()
	os.Exit(0)
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
	s.parent.ShowTempSentMsg(msg)
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
		case tcell.KeyRune:
			switch event.Rune() {
			case 113: // q
				if event.Modifiers() == 4 { // ALT+q
					s.parent.Quit()
				}
			}
		case tcell.KeyCtrlL:
			s.SetText("")
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
	log.Debug("updating contact panel...")
	// this is dumb, we re-sort every update
	// TODO: don't
	sorted := cl.siggo.Contacts().SortedByIndex()
	convs := cl.siggo.Conversations()
	log.Debug("sorted contacts: %v", sorted)
	for i, c := range sorted {
		id := ""
		if c.Name != "" {
			id = c.Name
		} else {
			id = c.Number
		}
		line := fmt.Sprintf("%s\n", id)
		color := convs[c].Color()
		if cl.currentIndex == i {
			line = fmt.Sprintf("[%s::r]%s[-::-]", color, line)
			cl.currentIndex = i
		} else if convs[c].HasNewMessage {
			line = fmt.Sprintf("[%s::b]*%s[-::-]", color, line)
		} else {
			line = fmt.Sprintf("[%s::]%s[-::]", color, line)
		}
		data += line
	}
	cl.sortedContacts = sorted
	cl.SetText(data)
}

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

type SearchPanel struct {
	*tview.Grid
	list      *tview.TextView
	input     *SearchInput
	parent    *ChatWindow
	maxHeight int
}

func (p *SearchPanel) Close() {
	p.parent.HideSearch()
}

func NewContactSearch(parent *ChatWindow) *SearchPanel {
	maxHeight := 10
	p := &SearchPanel{
		Grid:      tview.NewGrid().SetRows(maxHeight-3, 1),
		list:      tview.NewTextView(),
		parent:    parent,
		maxHeight: maxHeight,
	}
	//contactList := parent.siggo.Contacts().SortedByName()
	p.input = NewSearchInput(p)
	p.AddItem(p.list, 0, 0, 1, 1, 0, 0, false)
	p.AddItem(p.input, 1, 0, 1, 1, 0, 0, true)
	p.SetBorder(true)
	p.SetTitle("search contacts...")
	return p
}

type SearchInput struct {
	*tview.InputField
	parent *SearchPanel
}

func NewSearchInput(parent *SearchPanel) *SearchInput {
	si := &SearchInput{
		InputField: tview.NewInputField(),
		parent:     parent,
	}
	si.SetLabel("> ")
	si.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Debug("Key Event <SEARCH>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyESC:
			si.parent.Close()
			return nil
		}
		return event
	})
	return si
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
		log.Debug("Key Event <MAIN>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 106: // j
				w.ContactDown()
				return nil
			case 107: // k
				w.ContactUp()
				return nil
			case 105: // i
				if event.Modifiers() == 4 { // ALT+i
					w.Compose()
				} else {
					w.InsertMode()
				}
				return nil
			case 113: // q
				if event.Modifiers() == 4 { // ALT+q
					w.Quit()
				}
				return nil
			case 121:
				w.YankMode() // y
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
		case tcell.KeyCtrlT:
			w.ShowContactSearch()
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
	w.conversationPanel.ScrollToEnd()
	siggo.NewInfo = func(conv *model.Conversation) {
		app.QueueUpdateDraw(func() {
			w.update()
		})
	}

	return w
}

// FancyCompose opens up EDITOR and composes a big fancy message.
func FancyCompose() string {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "siggo-compose-")
	if err != nil {
		log.Error("failed to create temp file for compose: %v", err)
		return ""
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		log.Error("cannot compose: no EDITOR set in environment")
		return ""
	}
	fname := tmpFile.Name()
	defer os.Remove(fname)
	cmd := exec.Command(editor, fname)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		log.Error("failed to start editor: %v", err)
		return ""
	}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Error("failed to read temp file: %v", err)
		return ""
	}
	return string(b)
}
