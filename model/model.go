package model

import (
	"fmt"
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

type Config struct {
	UserName   string
	UserNumber string
}

type Contact struct {
	Number string
	Name   string
}

type ContactList map[string]*Contact

// Sorted returns a sorted list of contacts
func (cl ContactList) Sorted() []*Contact {
	list := make([]*Contact, 0)
	for _, c := range cl {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Number < list[j].Number })
	return list
}

type Message struct {
	Content     string
	From        string
	Timestamp   int64
	IsDelivered bool
	IsRead      bool
	FromSelf    bool
}

func (m *Message) String() string {
	var fromStr = m.From
	if len(m.From) > 12 {
		fromStr = m.From[:12]
	}
	// Magical Ref Data: Mon Jan 2 15:04:05 MST 2006
	data := fmt.Sprintf("%s|%s%s %12v: %s\n",
		time.Unix(0, m.Timestamp*1000000).Format("2006-01-02 15:05:05"),
		DeliveryStatus[m.IsDelivered],
		ReadStatus[m.IsRead],
		fromStr,
		m.Content,
	)
	if m.FromSelf == true {
		data = "[::d]" + data + "[::-]"
	} else if m.IsRead == false {
		data = "[::b]" + data + "[::-]"
	}
	return data
}

type Conversation struct {
	Contact       *Contact
	Messages      map[int64]*Message
	MessageOrder  []int64
	HasNewMessage bool
}

func (c *Conversation) String() string {
	out := ""
	for _, k := range c.MessageOrder {
		out += c.Messages[k].String()
	}
	return out
}

func (c *Conversation) AddMessage(message *Message) {
	_, ok := c.Messages[message.Timestamp]
	c.Messages[message.Timestamp] = message
	if !ok {
		// new messages
		c.MessageOrder = append(c.MessageOrder, message.Timestamp)
		c.HasNewMessage = true
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
	Send(string, string) error
	Receive() error
	ReceiveUntil(chan struct{})
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
	return s.signal.Send(contact.Number, msg)
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

// ReceiveUntil
func (s *Siggo) ReceiveUntil(done chan struct{}) {
	s.signal.ReceiveUntil(done)
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
		log.Printf("New contact: %v", c)
		s.contacts[c.Number] = c
	}
	message := &Message{
		Content:     sentMsg.Message,
		From:        " ~ ",
		Timestamp:   sentMsg.Timestamp,
		IsDelivered: false,
		IsRead:      false,
		FromSelf:    true,
	}
	conv, ok := s.conversations[c]
	if !ok {
		log.Printf("new conversation for contact: %v", c)
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
		log.Printf("New contact: %v", c)
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
	}
	conv, ok := s.conversations[c]
	if !ok {
		log.Printf("new conversation for contact: %v", c)
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
			log.Warnf("read receipt for message we don't have: %+v", msg.Envelope.ReceiptMessage)
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
	return nil
}

func (s *Siggo) Conversations() map[*Contact]*Conversation {
	return s.conversations
}

func (s *Siggo) Contacts() ContactList {
	return s.contacts
}

// NewSiggo creates a new model
func NewSiggo(sig SignalAPI, config *Config) *Siggo {
	contacts := GetContacts(config.UserNumber)
	conversations := GetConversations(config.UserNumber, contacts)
	s := &Siggo{
		config:        config,
		contacts:      contacts,
		conversations: conversations,
		signal:        sig,

		NewInfo: func(*Conversation) {}, // noop
	}
	//sig.OnMessage(s.?)

	sig.OnSent(s.onSent)
	sig.OnReceived(s.onReceived)
	sig.OnReceipt(s.onReceipt)
	return s
}

// GetContacts reads the contact list from disk for a given user
func GetContacts(userNumber string) ContactList {
	list := make(ContactList)
	list[userNumber] = &Contact{
		Number: userNumber,
		Name:   "me",
	}
	return list
}

// GetConversations reads conversations from disk for a given user
// and contact list
func GetConversations(userNumber string, contacts ContactList) map[*Contact]*Conversation {
	conversations := make(map[*Contact]*Conversation)
	for _, contact := range contacts {
		log.Printf("Adding conversation for: %+v\n", contact)
		conv := NewConversation(contact)
		conversations[contact] = conv
	}
	return conversations
}
