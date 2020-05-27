package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/derricw/siggo/signal"
	log "github.com/sirupsen/logrus"
)

var DeliveryStatus map[bool]string = map[bool]string{
	true:  "✓",
	false: "X",
}

var ReadStatus map[bool]string = map[bool]string{
	true:  "✓",
	false: "X",
}

// PhoneNumber is an alias for string not derived
type PhoneNumber = string

type Contact struct {
	Number PhoneNumber
	Name   string
	Index  int
}

type ContactList map[PhoneNumber]*Contact

// List returns a list of contacts (in random order)
func (cl ContactList) List() []*Contact {
	list := make([]*Contact, 0)
	for _, c := range cl {
		list = append(list, c)
	}
	return list
}

// SortedByNumber returns a slice of contacts sorted by phone number
// Idk why anyone would ever want to use this but here it is.
func (cl ContactList) SortedByNumber() []*Contact {
	list := cl.List()
	sort.Slice(list, func(i, j int) bool { return list[i].Number < list[j].Number })
	return list
}

// SortedByName returns a slice of contacts sorted alphabetically
func (cl ContactList) SortedByName() []*Contact {
	list := cl.List()
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// SortedByIndex returns a slice of contacts sorted by index provided by signal-cli
func (cl ContactList) SortedByIndex() []*Contact {
	list := cl.List()
	sort.Slice(list, func(i, j int) bool { return list[i].Index < list[j].Index })
	return list
}

type Message struct {
	Content     string               `json:"content"`
	From        string               `json:"from"`
	Timestamp   int64                `json:"timestamp"`
	IsDelivered bool                 `json:"is_delivered"`
	IsRead      bool                 `json:"is_read"`
	FromSelf    bool                 `json:"from_self"`
	Attachments []*signal.Attachment `json:"attachments"`
}

func (m *Message) String() string {
	var fromStr = m.From
	template := "%s|%s%s| %" + fmt.Sprintf("%dv", len(fromStr)) + ": %s\n"
	data := fmt.Sprintf(template,
		// lets come up with a way to avoid the *1000000
		// Magical Ref Data: Mon Jan 2 15:04:05 MST 2006
		time.Unix(0, m.Timestamp*1000000).Format("2006-01-02 15:04:05"),
		DeliveryStatus[m.IsDelivered],
		ReadStatus[m.IsRead],
		fromStr,
		m.Content,
	)
	if m.FromSelf == true {
		// dim messages from self (for now, until we support color for contacts)
		data = "[::d]" + data + "[::-]"
	} else if m.IsRead == false {
		// bold messages that haven't been read
		data = "[::b]" + data + "[::-]"
	}
	for _, a := range m.Attachments {
		aMsg := fmt.Sprintf(" 📎| %s | %s | %dB\n", a.Filename, a.ContentType, a.Size)
		data = fmt.Sprintf("%s%s", data, aMsg)
	}
	return data
}

type Conversation struct {
	Contact       *Contact
	Messages      map[int64]*Message
	MessageOrder  []int64
	HasNewMessage bool
	// hasNewData tracks whether new data has been added
	// since the last save to disk
	hasNewData bool
}

func (c *Conversation) String() string {
	out := ""
	for _, k := range c.MessageOrder {
		out += c.Messages[k].String()
	}
	return out
}

func (c *Conversation) AddMessage(message *Message) {
	c.addMessage(message)
}

func (c *Conversation) addMessage(message *Message) {
	_, ok := c.Messages[message.Timestamp]
	c.Messages[message.Timestamp] = message
	if !ok {
		// new messages
		c.MessageOrder = append(c.MessageOrder, message.Timestamp)
		c.HasNewMessage = true
		c.hasNewData = true
	}
}

// PopMessage just removes the last message, if it exists
func (c *Conversation) PopMessage() {
	nMessage := len(c.MessageOrder)
	if nMessage > 0 {
		lastMsg := c.MessageOrder[nMessage-1]
		c.MessageOrder = c.MessageOrder[:nMessage-1]
		delete(c.Messages, lastMsg)
	}
}

// CaughtUp iterates back through the messages of the conversation marking the un-read ones
// as read. We call this after we switch to this conversation.
func (c *Conversation) CaughtUp() {
	for i := len(c.MessageOrder) - 1; i >= 0; i-- {
		msg := c.Messages[c.MessageOrder[i]]
		if msg.IsRead && !msg.FromSelf {
			break
		}
		c.Messages[c.MessageOrder[i]].IsRead = true
	}
	c.HasNewMessage = false
}

// SaveAs writes the conversation to `path`.
func (c *Conversation) SaveAs(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, msgID := range c.MessageOrder {
		msg := c.Messages[msgID]
		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		f.Write(b)
		f.Write([]byte{'\n'})
	}
	return nil
}

// Save writes the conversation to the default location only if it has new data
func (c *Conversation) Save() error {
	if !c.hasNewData {
		return nil
	}
	folder := ConversationFolder()
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}
	path := filepath.Join(ConversationFolder(), c.Contact.Number)
	c.hasNewData = false // better to do this after successful save?
	return c.SaveAs(path)
}

// Load will load a conversation saved @ `path`
// TODO: load only the last N messages based on config
func (c *Conversation) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		msg := &Message{}
		err := json.Unmarshal(s.Bytes(), msg)
		if err != nil {
			return err
		}
		c.addMessage(msg)
	}
	return nil
}

func NewConversation(contact *Contact) *Conversation {
	return &Conversation{
		Contact:       contact,
		Messages:      make(map[int64]*Message),
		MessageOrder:  make([]int64, 0),
		HasNewMessage: false,
	}
}

type SignalAPI interface {
	Send(string, string) (int64, error)
	SendDbus(string, string) (int64, error)
	Receive() error
	ReceiveForever()
	OnReceived(signal.ReceivedCallback)
	OnReceipt(signal.ReceiptCallback)
	OnSent(signal.SentCallback)
}

type Siggo struct {
	config        *Config
	contacts      ContactList
	conversations map[*Contact]*Conversation
	contactOrder  []*Contact
	signal        SignalAPI

	NewInfo func(*Conversation)
}

// Send sends a message to a contact.
func (s *Siggo) Send(msg string, contact *Contact) error {
	message := &Message{
		Content:     msg,
		From:        " ~ ",
		Timestamp:   time.Now().Unix() * 1000,
		IsDelivered: false,
		IsRead:      false,
		FromSelf:    true,
	}
	conv, ok := s.conversations[contact]
	if !ok {
		log.Infof("new conversation for contact: %v", contact)
		conv = s.newConversation(contact)
	}
	// finally send the message
	ID, err := s.signal.SendDbus(contact.Number, msg)
	if err != nil {
		message.Content = fmt.Sprintf("FAILED TO SEND: %s ERROR: %v", message.Content, err)
		s.NewInfo(conv)
		return err
	}
	// use the official timestamp on success
	message.Timestamp = ID
	conv.CaughtUp()
	conv.AddMessage(message)
	s.NewInfo(conv)
	log.Infof("successfully sent message %s with timestamp: %d", message.Content, message.Timestamp)
	return nil
}

func (s *Siggo) newConversation(contact *Contact) *Conversation {
	conv := NewConversation(contact)
	s.conversations[contact] = conv
	return conv
}

func (s *Siggo) newContact(number string) *Contact {
	contact := &Contact{
		Number: number,
	}
	s.contacts[number] = contact
	return contact
}

// Receive
func (s *Siggo) Receive() error {
	return s.signal.Receive()
}

// ReceiveForever
func (s *Siggo) ReceiveForever() {
	s.signal.ReceiveForever()
}

func (s *Siggo) onSent(msg *signal.Message) error {
	// add new message to conversation
	sentMsg := msg.Envelope.SyncMessage.SentMessage
	contactNumber := sentMsg.Destination
	// if we have a name for this contact, use it
	// otherwise it will be the phone number
	c, ok := s.contacts[contactNumber]
	if !ok {
		c = &Contact{
			Number: contactNumber,
		}
		log.Infof("New contact: %v", c)
		s.contacts[c.Number] = c
	}
	message := &Message{
		Content:     sentMsg.Message,
		From:        " ~ ",
		Timestamp:   sentMsg.Timestamp,
		IsDelivered: false,
		IsRead:      false,
		FromSelf:    true,
		Attachments: sentMsg.Attachments,
	}
	conv, ok := s.conversations[c]
	if !ok {
		log.Infof("new conversation for contact: %v", c)
		conv = s.newConversation(c)
	}
	conv.AddMessage(message)
	s.NewInfo(conv)
	return nil
}

func (s *Siggo) onReceived(msg *signal.Message) error {
	// add new message to conversation
	receiveMsg := msg.Envelope.DataMessage
	contactNumber := msg.Envelope.Source
	// if we have a name for this contact, use it
	// otherwise it will be the phone number
	c, ok := s.contacts[contactNumber]

	var fromStr string
	// TODO: fix this when i can load contact names from
	// somewhere
	if !ok {
		fromStr = contactNumber
		c = &Contact{
			Number: contactNumber,
		}
		log.Infof("New contact: %v", c)
		s.contacts[c.Number] = c
	} else if c.Name == "" {
		fromStr = contactNumber
	} else {
		fromStr = c.Name
	}
	message := &Message{
		Content:     receiveMsg.Message,
		From:        fromStr,
		Timestamp:   receiveMsg.Timestamp,
		IsDelivered: true,
		IsRead:      false,
		Attachments: receiveMsg.Attachments,
	}
	conv, ok := s.conversations[c]
	if !ok {
		log.Infof("new conversation for contact: %v", c)
		conv = s.newConversation(c)
	}
	conv.AddMessage(message)
	s.NewInfo(conv)
	return nil
}

func (s *Siggo) onReceipt(msg *signal.Message) error {
	receiptMsg := msg.Envelope.ReceiptMessage
	// if the message exists, edit it with new data
	contactNumber := msg.Envelope.Source
	// if we have a name for this contact, use it
	// otherwise it will be the phone number
	c, ok := s.contacts[contactNumber]
	if !ok {
		c = s.newContact(contactNumber)
	}
	conv, ok := s.conversations[c]
	if !ok {
		conv = s.newConversation(c)
	}
	for _, ts := range receiptMsg.Timestamps {
		message, ok := conv.Messages[ts]
		if !ok {
			// TODO: handle case where we get a read receipt for
			// a message that we don't have
			b, err := json.Marshal(msg)
			if err != nil {
				log.Warnf("couldn't marshal receipt for message we don't have: %v", err)
			}
			log.Warnf("read receipt for message we don't have: %s", b)
			continue
		}
		if receiptMsg.IsRead {
			// for whatever reason messages can be marked as
			// read but not delivered, so we go ahead and assume any
			// message that has been read has also been delivered
			message.IsDelivered = true
			message.IsRead = true
		} else {
			message.IsDelivered = receiptMsg.IsDelivery
			message.IsRead = receiptMsg.IsRead
		}
	}
	conv.hasNewData = true
	return nil
}

// Conversations returns the current converstation book
func (s *Siggo) Conversations() map[*Contact]*Conversation {
	return s.conversations
}

// Contacts returns the current contact list
func (s *Siggo) Contacts() ContactList {
	return s.contacts
}

// SaveConversations saves all conversations to disk
func (s *Siggo) SaveConversations() {
	for _, conv := range s.conversations {
		err := conv.Save()
		if err != nil {
			log.Errorf("failed to save conversation: %v", err)
		}
	}
}

// Quit does any cleanup we want to do at exit.
func (s *Siggo) Quit() {
	if s.config.SaveMessages {
		s.SaveConversations()
	}
}

// NewSiggo creates a new model
func NewSiggo(sig SignalAPI, config *Config) *Siggo {
	s := &Siggo{
		config: config,
		signal: sig,

		NewInfo: func(*Conversation) {}, // noop
	}
	s.init()
	//sig.OnMessage(s.?)

	sig.OnSent(s.onSent)
	sig.OnReceived(s.onReceived)
	sig.OnReceipt(s.onReceipt)
	return s
}

func (s *Siggo) init() {
	//load contacts and conversations for the first time
	s.contacts = s.getContacts()
	if self, ok := s.contacts[s.config.UserNumber]; ok {
		self.Name = s.config.UserName
	}
	s.conversations = s.getConversations()
}

// getContacts reads a fresh contact list from disk for the configured user
func (s *Siggo) getContacts() ContactList {
	list := make(ContactList)
	sig := signal.NewSignal(s.config.UserNumber)
	contacts, err := sig.GetContactList()
	if err != nil {
		log.Warnf("failed to read contacts from disk: %v", err)
		return list
	}
	for _, c := range contacts {
		if c.InboxPosition != nil {
			list[c.Number] = &Contact{
				Number: c.Number,
				Name:   c.Name,
				Index:  *c.InboxPosition,
			}
		}
	}
	return list
}

// getConversations reads conversations from disk for the configured user's contact list
func (s *Siggo) getConversations() map[*Contact]*Conversation {
	conversations := make(map[*Contact]*Conversation)
	for _, contact := range s.contacts {
		log.Debugf("Adding conversation for: %+v\n", contact)
		conv := NewConversation(contact)
		// check if we have a conversation file for this user
		if s.config.SaveMessages {
			convPath := filepath.Join(ConversationFolder(), contact.Number)
			err := conv.Load(convPath) // if we fail to load, oh well
			if err == nil {
				log.Infof("loaded conversation from: %s", contact.Name)
			}
		}
		conv.CaughtUp()
		conversations[contact] = conv
	}
	return conversations
}
