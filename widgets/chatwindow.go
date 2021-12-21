package widgets

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derricw/siggo/model"
	"github.com/gdamore/tcell/v2"
	"github.com/kyokomi/emoji"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
	YankMode
	OpenMode
	LinkMode
)

// stolen from suckoverflow
var urlRegex = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

// ChatWindow is the main panel for the UI.
type ChatWindow struct {
	// todo: maybe use Flex instead of Grid?
	*tview.Grid
	siggo          *model.Siggo
	currentContact *model.Contact
	mode           Mode

	sendPanel         *SendPanel
	contactsPanel     *ContactListPanel
	conversationPanel *ConversationPanel
	searchPanel       tview.Primitive
	commandPanel      tview.Primitive
	statusBar         *StatusBar
	app               *tview.Application
	normalKeybinds    func(*tcell.EventKey) *tcell.EventKey
	yankKeybinds      func(*tcell.EventKey) *tcell.EventKey
	openKeybinds      func(*tcell.EventKey) *tcell.EventKey
	linkKeybinds      func(*tcell.EventKey) *tcell.EventKey
	goKeybinds        func(*tcell.EventKey) *tcell.EventKey
}

// InsertMode enters insert mode
func (c *ChatWindow) InsertMode() {
	log.Debug("INSERT MODE")
	c.app.SetFocus(c.sendPanel)
	c.sendPanel.SetBorderColor(tcell.ColorOrange)
	c.mode = InsertMode
}

// YankMode enters yank mode
func (c *ChatWindow) YankMode() {
	log.Debug("YANK MODE")
	c.conversationPanel.SetBorderColor(tcell.ColorOrange)
	c.mode = YankMode
	c.SetInputCapture(c.yankKeybinds)
}

// OpenMode enters open mode which lets us select an attachment to open
func (c *ChatWindow) OpenMode() {
	log.Debug("OPEN MODE")
	c.mode = OpenMode
	oi := NewOpenInput(c)
	c.HideConversation(oi)
	c.app.SetFocus(oi)
}

// LinkMode enters link mode
func (c *ChatWindow) LinkMode() {
	log.Debug("LINK MODE")
	c.mode = LinkMode
	li := NewLinksInput(c)
	c.HideConversation(li)
	c.app.SetFocus(li)
}

// NormalMode enters normal mode
func (c *ChatWindow) NormalMode() {
	log.Debug("NORMAL MODE")
	c.app.SetFocus(c)

	// clear our highlights
	c.conversationPanel.SetBorderColor(tcell.ColorWhite)
	c.sendPanel.SetBorderColor(tcell.ColorWhite)
	c.mode = NormalMode
	c.SetInputCapture(c.normalKeybinds)
	// save draft
	conv, err := c.currentConversation()
	if err != nil {
		c.SetErrorStatus(err)
		return
	}
	conv.StagedMessage = c.sendPanel.GetText()
}

// ShowConversation ensures that the conversation panel is showing. This should be called when
// any widget is done hiding the conversation panel
func (c *ChatWindow) ShowConversation() {
	c.Grid.AddItem(c.conversationPanel, 0, 1, 1, 1, 0, 0, false)
}

// HideConversation temporarily replaces the conversation panel with another widget
func (c *ChatWindow) HideConversation(replacement tview.Primitive) {
	c.Grid.RemoveItem(c.conversationPanel)
	c.Grid.AddItem(replacement, 0, 1, 1, 1, 0, 0, false)
}

// YankLastMsg copies the last message of a conversation to the clipboard.
func (c *ChatWindow) YankLastMsg() {
	c.NormalMode()
	conv, err := c.currentConversation()
	if err != nil {
		c.SetErrorStatus(err)
		return
	}
	if conv == nil {
		c.SetErrorStatus(fmt.Errorf("<NO CONVERSATION>")) // this shouldn't happen
		return
	}
	var lastMsg *model.Message
	if lastMsg = conv.LastMessage(); lastMsg == nil {
		c.SetStatus("📋<NO MESSAGES>") // this is fine
		return
	}
	content := strings.TrimSpace(lastMsg.Content)
	err = clipboard.WriteAll(content)
	if err != nil {
		c.SetErrorStatus(err)
		return
	}
	c.SetStatus(fmt.Sprintf("📋%s", content))
}

func (c *ChatWindow) getLinks() []string {
	toSearch := c.conversationPanel.GetText(true)
	return urlRegex.FindAllString(toSearch, -1)
}

func (c *ChatWindow) getAttachments() []*model.Attachment {
	a := make([]*model.Attachment, 0)
	conv, err := c.currentConversation()
	if err != nil {
		return a
	}
	// TODO: make siggo.Conversation keep a list of attachments
	// so that we don't have to search for them like this
	for _, ID := range conv.MessageOrder {
		msg := conv.Messages[ID]
		if len(msg.Attachments) > 0 {
			a = append(a, msg.Attachments...)
		}
	}
	return a
}

// YankLastLink copies the last link in a converstaion to the clipboard
func (c *ChatWindow) YankLastLink() {
	c.NormalMode()
	links := c.getLinks()
	if len(links) > 0 {
		last := links[len(links)-1]
		if err := clipboard.WriteAll(last); err != nil {
			c.SetErrorStatus(err)
			return
		}
		c.SetStatus(fmt.Sprintf("📋%s", last))
	} else {
		c.SetStatus(fmt.Sprintf("📋<NO MATCHES>"))
	}
}

// FocusMe gives focus to the chat window
func (c *ChatWindow) FocusMe() {
	c.app.SetFocus(c)
}

// ShowAttachInput opens a commandPanel to choose a file to attach
func (c *ChatWindow) ShowAttachInput() {
	c.HideCommandInput() // only one at a time
	log.Debug("SHOWING ATTACH INPUT")
	p := NewAttachInput(c)
	c.commandPanel = p
	c.SetRows(0, 3, 1)
	c.AddItem(p, 2, 0, 1, 2, 0, 0, false)
	c.app.SetFocus(p)
}

// ShowFilterInput opens a commandPanel to filter the conversation
func (c *ChatWindow) ShowFilterInput() {
	c.HideCommandInput() // only one at a time
	log.Debug("SHOWING FILTER INPUT")
	p := NewFilterInput(c)
	c.commandPanel = p
	c.SetRows(0, 3, 1)
	c.AddItem(p, 2, 0, 1, 2, 0, 0, false)
	c.app.SetFocus(p)
}

// HideCommandInput hides any current CommandInput panel
func (c *ChatWindow) HideCommandInput() {
	log.Debug("HIDING COMMAND INPUT")
	c.RemoveItem(c.commandPanel)
	// TODO: this should happen automatically when i clear a FilterInput
	// maybe command panel should be an interface with a "Close" method
	// that does stuff like this
	c.conversationPanel.Filter("")
	c.SetRows(0, 3)
	c.update()
	c.FocusMe()
}

// ShowStatusBar shows the bottom status bar
func (c *ChatWindow) ShowStatusBar() {
	c.SetRows(0, 3, 1)
	c.AddItem(c.statusBar, 2, 0, 1, 2, 0, 0, false)
}

// HideStatusBar stops showing the status bar
func (c *ChatWindow) HideStatusBar() {
	c.RemoveItem(c.statusBar)
	c.SetRows(0, 3)
}

// SetStatus shows a status message on the status bar
func (c *ChatWindow) SetStatus(statusMsg string) {
	log.Info(statusMsg)
	c.statusBar.SetText(statusMsg)
	c.ShowStatusBar()
}

// SetErrorStatus shows an error status in the status bar
func (c *ChatWindow) SetErrorStatus(err error) {
	log.Errorf("%s", err)
	c.statusBar.SetText(fmt.Sprintf("🔥%s", err))
	c.ShowStatusBar()
}

func (c *ChatWindow) currentConversation() (*model.Conversation, error) {
	currentConv, ok := c.siggo.Conversations()[c.currentContact]
	if ok {
		return currentConv, nil
	} else {
		return nil, fmt.Errorf("no conversation for current contact: %v", c.currentContact)
	}
}

func (c *ChatWindow) currentContactName() string {
	if c.currentContact != nil {
		return c.currentContact.String()
	}
	return ""
}

// SetCurrentContact sets the active contact
func (c *ChatWindow) SetCurrentContact(contact *model.Contact) error {
	log.Debugf("setting current contact to: %v", contact)
	c.currentContact = contact
	c.contactsPanel.GotoContact(contact)
	c.contactsPanel.Render()
	conv, err := c.currentConversation()
	if err != nil {
		return err
	}
	c.conversationPanel.Update(conv)
	conv.CaughtUp()
	c.sendPanel.Clear()
	c.sendPanel.Update()
	c.conversationPanel.ScrollToEnd()
	return nil
}

// SetCurrentContactByString sets the active contact to the first contact whose
// string representation matches
func (c *ChatWindow) SetCurrentContactByString(contactName string) error {
	log.Debugf("setting current contact to: %v", contactName)
	contact := c.siggo.Contacts().FindContact(contactName)
	if contact != nil {
		return c.SetCurrentContact(contact)
	}
	return fmt.Errorf("couldn't find contact: %s", contactName)
}

// NextUnreadMessage searches for the next conversation with unread messages and makes that the
// active conversation.
func (c *ChatWindow) NextUnreadMessage() error {
	for contact, conv := range c.siggo.Conversations() {
		if conv.HasNewMessage {
			err := c.SetCurrentContact(contact)
			if err != nil {
				c.SetErrorStatus(err)
			}
			return nil
		}
	}
	return nil
}

func (c *ChatWindow) Paste() error {
	err := AttachFromClipboard(c)
	if err != nil {
		c.SetErrorStatus(err)
	}

	return err
}

// TODO: remove code duplication with ContactDown()
func (c *ChatWindow) ContactUp() {
	log.Debug("PREVIOUS CONVERSATION")
	prevContact := c.contactsPanel.Previous()
	if prevContact != c.currentContact {
		err := c.SetCurrentContact(prevContact)
		if err != nil {
			c.SetErrorStatus(err)
		}
	}
}

func (c *ChatWindow) ContactDown() {
	log.Debug("NEXT CONVERSATION")
	nextContact := c.contactsPanel.Next()
	if nextContact != c.currentContact {
		err := c.SetCurrentContact(nextContact)
		if err != nil {
			c.SetErrorStatus(err)
		}
	}
}

// Compose opens an EDITOR to compose a command. If any text is saved in the buffer,
// we send it as a message to the current conversation.
func (c *ChatWindow) Compose() {
	msg := ""
	var err error

	success := c.app.Suspend(func() {
		msg, err = FancyCompose()
	})
	// need to sleep because there seems to be a race condition in tview
	// https://github.com/rivo/tview/issues/244
	//time.Sleep(100 * time.Millisecond)
	if !success {
		c.SetErrorStatus(fmt.Errorf("failed to suspend siggo"))
		return
	}
	if err != nil {
		c.SetErrorStatus(err)
		return
	}
	if msg != "" {
		msg = emoji.Sprint(msg)
		contact := c.currentContact
		c.ShowTempSentMsg(msg)
		go c.siggo.Send(msg, contact)
		log.Infof("sending message: %s to contact: %s", msg, contact)
	}
}

// FancyAttach opens FZF and selects a file to attach
func (c *ChatWindow) FancyAttach() {
	path := ""
	var err error
	_, err = exec.LookPath("fzf")
	if err != nil {
		c.SetErrorStatus(fmt.Errorf("failed to find in PATH: fzf"))
		return
	}
	success := c.app.Suspend(func() {
		path, err = FZFFile()
	})
	//time.Sleep(100 * time.Millisecond)
	if !success {
		c.SetErrorStatus(fmt.Errorf("failed to suspend siggo"))
		return
	}
	if err != nil {
		c.SetErrorStatus(err)
	}
	if path != "" {
		log.Errorf("attaching path: %s", path)
		conv, err := c.currentConversation()
		if err != nil {
			c.SetErrorStatus(err)
			return
		}
		err = conv.AddAttachment(path)
		if err != nil {
			c.SetErrorStatus(err)
		}
		c.sendPanel.Update()
		c.InsertMode()
	}
}

// FuzzyGoTo goes to a contact or group with fuzzy matching
func (c *ChatWindow) FuzzyGoTo() {
	contactName := ""
	var err error
	_, err = exec.LookPath("fzf")
	if err != nil {
		c.SetErrorStatus(fmt.Errorf("failed to find in PATH: fzf"))
		return
	}
	contactList := c.siggo.Contacts().AllNames()
	success := c.app.Suspend(func() {
		contactName, err = FZFList(contactList)
	})
	//time.Sleep(100 * time.Millisecond)
	if !success {
		c.SetErrorStatus(fmt.Errorf("failed to suspend siggo"))
		return
	}
	if err != nil {
		c.SetErrorStatus(err)
	}
	if contactName != "" {
		log.Infof("going to contact: %s", contactName)
		c.SetCurrentContactByString(contactName)
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
	c.conversationPanel.Write([]byte(tmpMsg.String()))
}

// Quit shuts down gracefully
func (c *ChatWindow) Quit() {
	c.app.Stop()
	c.siggo.Quit()
	os.Exit(0)
}

func (c *ChatWindow) update() {
	convs := c.siggo.Conversations()
	if convs != nil && len(convs) > 0 {
		if c.currentContact == nil {
			// there is a conversation but we haven't set a current contact yet
			// just grab the first one we find
			for _, contact := range c.siggo.Contacts() {
				c.currentContact = contact
				break
			}
		}
		c.contactsPanel.Render()
		currentConv, ok := convs[c.currentContact]
		if ok {
			c.conversationPanel.Update(currentConv)
		} else {
			// this is a panic because it shouldn't be possible?
			log.Panicf("no conversation for current contact: %s", c.currentContact)
		}
	}
}

type StatusBar struct {
	*tview.TextView
	parent *ChatWindow
}

func NewStatusBar(parent *ChatWindow) *StatusBar {
	sb := &StatusBar{
		TextView: tview.NewTextView(),
		parent:   parent,
	}
	return sb
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
	w.statusBar = NewStatusBar(w)
	// NORMAL MODE KEYBINDINGS
	w.normalKeybinds = func(event *tcell.EventKey) *tcell.EventKey {
		log.Debugf("Key Event <NORMAL>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 106: // j
				convInputHandler(event, func(p tview.Primitive) {})
				return nil
			case 107: // k
				convInputHandler(event, func(p tview.Primitive) {})
				return nil
			case 74: // J
				w.ContactDown()
				return nil
			case 75: // K
				w.ContactUp()
				return nil
			case 105: // i
				w.InsertMode()
				return nil
			case 73: // I
				w.Compose()
				return nil
			case 121: // y
				w.YankMode()
				return nil
			case 111: // o
				w.OpenMode()
				return nil
			case 108: // l
				w.LinkMode()
				return nil
			case 97: // a
				w.ShowAttachInput()
				return nil
			case 47: // /
				w.ShowFilterInput()
				return nil
			case 65: // A
				w.FancyAttach()
				return nil
			case 116: // t
				w.FuzzyGoTo()
				return nil
			case 112: // p
				w.Paste()
				return nil
			case 110: // n
				w.NextUnreadMessage()
				return nil
			}
			// pass some events on to the conversation panel
		case tcell.KeyCtrlQ:
			w.Quit()
		case tcell.KeyPgUp:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyPgDn:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyUp:
			convInputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyDown:
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
			w.HideStatusBar()
			w.HideCommandInput()
			return nil
		case tcell.KeyCtrlN:
			w.NextUnreadMessage()
			return nil
		case tcell.KeyCtrlV:
			w.Paste()
			return nil
		}
		return event
	}
	w.yankKeybinds = func(event *tcell.EventKey) *tcell.EventKey {
		log.Debugf("Key Event <YANK>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 121: // y
				w.YankLastMsg()
				return nil
			case 108: // l
				w.YankLastLink()
				return nil
			}
		case tcell.KeyCtrlQ:
			w.Quit()
		case tcell.KeyESC:
			w.NormalMode()
			return nil
		}
		return event
	}
	w.SetInputCapture(w.normalKeybinds)

	// primitiv, row, col, rowSpan, colSpan, minGridHeight, maxGridHeight, focus)
	// TODO: lets make some of the spans confiurable?
	w.AddItem(w.contactsPanel, 0, 0, 2, 1, 0, 0, false)
	w.ShowConversation()
	w.AddItem(w.sendPanel, 1, 1, 1, 1, 0, 0, false)

	if w.siggo.Config().HidePanelTitles {
		w.contactsPanel.SetTitle("")
		w.sendPanel.SetTitle("")
		w.conversationPanel.SetTitle("")
		w.conversationPanel.hideTitle = true
	}
	if w.siggo.Config().HidePhoneNumbers {
		w.conversationPanel.hidePhoneNumber = true
	}

	w.siggo = siggo
	contacts := siggo.Contacts().SortedByIndex()
	log.Debugf("contacts found: %v", contacts)
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
	siggo.ErrorEvent = w.SetErrorStatus
	return w
}

// FancyCompose opens up EDITOR and composes a big fancy message.
func FancyCompose() (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "siggo-compose-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for compose: %v", err)
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "", fmt.Errorf("cannot compose: no $EDITOR set in environment")
	}
	fname := tmpFile.Name()
	defer os.Remove(fname)
	cmd := exec.Command(editor, fname)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to start editor: %v", err)
	}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %v", err)
	}
	return string(b), nil
}

// CompletePath autocompletes a path stub
func CompletePath(path string) string {
	if path == "" {
		return ""
	}
	if path[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			return ""
		}
		path = usr.HomeDir + path[1:]
	}
	matches, err := filepath.Glob(path + "*")
	if err != nil || matches == nil || len(matches) == 0 {
		return path
	}
	if len(matches) == 1 {
		path = matches[0]
	} else if !strings.HasSuffix(path, "/") {
		path = GetSharedPrefix(matches...)
	}
	stat, err := os.Stat(path)
	if err != nil {
		return path
	}
	if stat.IsDir() {
		if !strings.HasSuffix(path, "/") {
			return path + "/"
		}
	}
	return path
}

// GetSharedPrefix finds the prefix shared by any number of strings
// Is there a more efficient way to do this?
func GetSharedPrefix(s ...string) string {
	var out strings.Builder
	for i := 0; i < len(s[0]); i++ {
		c := s[0][i]
		for _, str := range s {
			if str[i] != c {
				return out.String()
			}
		}
		out.WriteByte(c)
	}
	return out.String()
}
