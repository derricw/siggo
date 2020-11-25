package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/derricw/siggo/signal"
	"github.com/gen2brain/beeep"
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
	Number  PhoneNumber
	Name    string
	Index   int
	alias   string
	color   string
	isGroup bool
}

// String returns a string to display for this contact. Priority is Alias > Name > Number.
// Groups get a cute little # indicator.
func (c *Contact) String() string {
	if c.alias != "" {
		return c.alias
	}
	if c.Name != "" {
		if c.isGroup {
			return fmt.Sprintf("#%s", c.Name)
		}
		return c.Name
	}
	return c.Number
}

// Color returns the configured color highlight for incoming messages
func (c *Contact) Color() string {
	return c.color
}

// Avatar returns the path to the contact's avatar, if it can find it, otherwise ""
func (c *Contact) Avatar() string {
	folder, err := signal.GetSignalAvatarsFolder()
	if err != nil {
		return ""
	}
	path := filepath.Join(folder, fmt.Sprintf("contact-%s", c.Number))
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

// Configure applies a configuration to the contact (for now, an alias and custom color)
func (c *Contact) Configure(cfg *Config) {
	c.color = cfg.ContactColors[c.Name]
	c.alias = cfg.ContactAliases[c.Name]
}

type ContactList map[PhoneNumber]*Contact
type ConvInfo map[*Contact]*Conversation

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
	Content     string        `json:"content"`
	Timestamp   int64         `json:"timestamp"`
	IsDelivered bool          `json:"is_delivered"`
	IsRead      bool          `json:"is_read"`
	FromSelf    bool          `json:"from_self"`
	Attachments []*Attachment `json:"attachments"`
	From        string        `json:from"`
	FromContact *Contact      `json:from_contact"`
}

func (m *Message) String() string {
	var fromStr, color string
	if !m.FromSelf {
		fromStr = m.FromContact.String()
		color = m.FromContact.Color()
	} else {
		fromStr = " ~ "
	}

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
		data = fmt.Sprintf("[::d]%s[::-]", data)
	} else if m.IsRead == false {
		// bold messages that haven't been read
		data = fmt.Sprintf("[%s::b]%s[-::-]", color, data)
	} else if m.IsRead == true {
		data = fmt.Sprintf("[%s::]%s[-::]", color, data)
	}
	// show attachments
	for _, a := range m.Attachments {
		data = fmt.Sprintf("%s%s\n", data, a)
	}
	return data
}

// AddAttachments currently only is used to track attachments we sent to other people, so that
// they show up in the GUI.
func (m *Message) AddAttachments(paths []string) {
	if m.Attachments == nil {
		m.Attachments = make([]*Attachment, 0)
	}
	for _, path := range paths {
		size := 0
		stats, err := os.Stat(path)
		if err == nil {
			size = int(stats.Size())
		}
		m.Attachments = append(m.Attachments, &Attachment{
			Filename:  path,
			FromSelf:  true,
			Timestamp: m.Timestamp,
			Size:      size,
		})
	}
}

// Attachment is any file sent or received. Received attachments are left in the usual `signal-cli`
// location for now. It seems to automatically delete old attachments, so we may want to come up
// with a way to keep our own copy somewhere in the siggo data folder.
type Attachment struct {
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
	ID          string `json:"id"`
	Size        int    `json:"size"`
	Timestamp   int64  `json:"timestamp"`
	FromSelf    bool   `json:"from_self"`
}

// Path returns the full path to an attachment file
func (a *Attachment) Path() (string, error) {
	if a.ID == "" {
		// TODO: save our own copy of the attachment with our own ID
		// for now, just return the path where we attached it
		return a.Filename, nil
	}
	folder, err := signal.GetSignalFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(folder, "attachments", a.ID), nil
}

// String returns the string representation of the attachment
func (a *Attachment) String() string {
	ts := time.Unix(0, a.Timestamp*1000000).Format("2006-01-02 15:04:05")
	txt := fmt.Sprintf(" 📎| %s | %s | %s | %dB", ts, a.Filename, a.ContentType, a.Size)
	return txt
}

// NewAttachmentFromWire creates a new siggo attachment from a signal.Attachment
func NewAttachmentFromWire(wire *signal.Attachment, timestamp int64, fromSelf bool) *Attachment {
	return &Attachment{
		ContentType: wire.ContentType,
		Filename:    wire.Filename,
		ID:          wire.ID,
		Size:        wire.Size,
		Timestamp:   timestamp,
		FromSelf:    fromSelf,
	}
}

// ConvertAttachments converts signal's wire attachments into our model's attachments
func ConvertAttachments(wire []*signal.Attachment, timestamp int64, fromSelf bool) []*Attachment {
	out := make([]*Attachment, 0, len(wire))
	for _, a := range wire {
		out = append(out, NewAttachmentFromWire(a, timestamp, fromSelf))
	}
	return out
}

// Coversation is a contact or group and its associated messages
type Conversation struct {
	Contact       *Contact // can be a group!
	Messages      map[int64]*Message
	MessageOrder  []int64
	HasNewMessage bool
	StagedMessage string
	// hasNewData tracks whether new data has been added
	// since the last save to disk
	hasNewData        bool
	stagedAttachments []string
}

// String renders the conversation to a single string
func (c *Conversation) String() string {
	out := ""
	for _, k := range c.MessageOrder {
		out += c.Messages[k].String()
	}
	return out
}

// Filter redners the conversation, but filters out any messages that don't have a regex match
// TODO: eliminate code duplication with String()
func (c *Conversation) Filter(pattern string) string {
	if pattern == "" {
		return c.String()
	}
	out := ""
	for _, k := range c.MessageOrder {
		s := c.Messages[k].String()
		if found, err := regexp.MatchString(pattern, s); found == true || err != nil {
			out += s
		}
	}
	return out
}

// AddMessage appends a message to the conversation
func (c *Conversation) AddMessage(message *Message) {
	c.addMessage(message)
}

func (c *Conversation) addMessage(message *Message) {
	_, ok := c.Messages[message.Timestamp]
	c.Messages[message.Timestamp] = message
	if !ok {
		// new messages
		// TODO: this section is to prevent saved pre-groups conversations from breaking when
		// loading.  it ensures that they have a contact. lets remove this after a few releases
		if !message.FromSelf && message.FromContact == nil {
			message.FromContact = c.Contact
		}
		// this we keep
		c.MessageOrder = append(c.MessageOrder, message.Timestamp)
		c.HasNewMessage = true
		c.hasNewData = true
	}
}

// LastMessage returns the most recent message. Can be nil.
func (c *Conversation) LastMessage() *Message {
	nMessage := len(c.MessageOrder)
	if nMessage > 0 {
		lastMsgID := c.MessageOrder[nMessage-1]
		return c.Messages[lastMsgID]
	}
	return nil
}

// StageAttachment attaches a file to be sent in the next message
func (c *Conversation) AddAttachment(path string) error {
	if _, err := os.Stat(path); err != nil {
		// no file there...
		return err
	}
	c.stagedAttachments = append(c.stagedAttachments, path)
	return nil
}

// ClearAttachments removes any staged attachments
func (c *Conversation) ClearAttachments() {
	c.stagedAttachments = []string{}
}

// ClearStagedMessage removes any staged attachments
func (c *Conversation) ClearStagedMessage() {
	c.StagedMessage = ""
}

// ClearStaged clears any staged message or attachment
func (c *Conversation) ClearStaged() {
	c.ClearStagedMessage()
	c.ClearAttachments()
}

// NumAttachments returns the number of staged attachments
func (c *Conversation) NumAttachments() int {
	return len(c.stagedAttachments)
}

// HasStagedMessage returns whether the conversation has a staged message
func (c *Conversation) HasStagedMessage() bool {
	return len(c.StagedMessage) > 0
}

// HasStagedData returns whether the conversation has a staged message or attachment
func (c *Conversation) HasStagedData() bool {
	return c.HasStagedMessage() || c.NumAttachments() != 0
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
// HERE IS WHERE THE COLOR THE LOADED MESSAGES
func (c *Conversation) Load(path string, cfg *Config) error {
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
		if msg.FromContact != nil {
			msg.FromContact.Configure(cfg)
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

		stagedAttachments: make([]string, 0),
	}
}

type SignalAPI interface {
	Send(string, string) (int64, error)
	SendDbus(string, string, ...string) (int64, error)
	SendGroupDbus(string, string, ...string) (int64, error)
	Receive() error
	ReceiveForever()
	Close()
	OnReceived(signal.ReceivedCallback)
	OnReceipt(signal.ReceiptCallback)
	OnSent(signal.SentCallback)
	OnError(signal.ErrorCallback)
}

type Siggo struct {
	config        *Config
	contacts      ContactList
	conversations map[*Contact]*Conversation
	contactOrder  []*Contact
	signal        SignalAPI

	NewInfo    func(*Conversation)
	ErrorEvent func(error)
}

// Send sends a message to a contact.
func (s *Siggo) Send(msg string, contact *Contact) error {
	ts := time.Now().Unix() * 1000
	message := &Message{
		Content:     msg,
		From:        " ~ ",
		Timestamp:   ts,
		IsDelivered: false,
		IsRead:      false,
		FromSelf:    true,
		Attachments: make([]*Attachment, 0),
	}
	conv, ok := s.conversations[contact]
	if !ok {
		log.Infof("new conversation for contact: %v", contact)
		conv = s.newConversation(contact)
	}
	// finally send the message
	var ID int64
	var err error
	if !contact.isGroup {
		log.Debugf("sending message to contact: %v", contact)
		ID, err = s.signal.SendDbus(contact.Number, msg, conv.stagedAttachments...)
	} else {
		log.Debugf("sending message to group: %v", contact)
		ID, err = s.signal.SendGroupDbus(contact.Number, msg, conv.stagedAttachments...)
	}
	if err != nil {
		message.Content = fmt.Sprintf("FAILED TO SEND: %s ERROR: %v", message.Content, err)
		s.NewInfo(conv)
		return err
	}
	// use the official timestamp on success
	message.Timestamp = ID
	conv.CaughtUp()
	message.AddAttachments(conv.stagedAttachments)
	conv.ClearStaged()
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

func (s *Siggo) handleError(err error) {
	s.ErrorEvent(err)
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
		Attachments: ConvertAttachments(sentMsg.Attachments, sentMsg.Timestamp, true),
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
	if receiveMsg.GroupInfo != nil {
		return s.onGroupMessageReceived(msg)
	}
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
		Attachments: ConvertAttachments(receiveMsg.Attachments, receiveMsg.Timestamp, false),
		FromContact: c,
	}
	conv, ok := s.conversations[c]
	if !ok {
		log.Infof("new conversation for contact: %v", c)
		conv = s.newConversation(c)
	}
	conv.AddMessage(message)
	s.NewInfo(conv)
	s.sendNotification(c.String(), message.Content, c.Avatar())
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

func (s *Siggo) onGroupMessageReceived(msg *signal.Message) error {
	// add new message to conversation
	receiveMsg := msg.Envelope.DataMessage
	contactNumber := msg.Envelope.Source
	groupID := msg.Envelope.DataMessage.GroupInfo.GroupID
	groupName := msg.Envelope.DataMessage.GroupInfo.Name

	g, ok := s.contacts[groupID]

	if !ok {
		// group not in contacts
		g = &Contact{
			Number:  groupID,
			Name:    groupName,
			isGroup: true,
		}
		log.Infof("New group: %v", g)
		s.contacts[g.Number] = g
	}

	var fromStr string
	c, ok := s.contacts[contactNumber]
	if !ok {
		// number not currently in contacts
		c = s.newContact(contactNumber)
		log.Infof("New contact: %v", c)
		fromStr = contactNumber
	} else if c.Name == "" {
		fromStr = contactNumber
	} else {
		fromStr = c.Name
	}
	log.Debugf("new group message for group %v from contact %v", g, c)

	message := &Message{
		Content:     receiveMsg.Message,
		From:        fromStr,
		Timestamp:   receiveMsg.Timestamp,
		IsDelivered: true,
		IsRead:      false,
		Attachments: ConvertAttachments(receiveMsg.Attachments, receiveMsg.Timestamp, false),
		FromContact: c,
	}

	conv, ok := s.conversations[g]
	if !ok {
		log.Infof("new conversation for group: %v", g)
		conv = s.newConversation(g)
	}
	conv.AddMessage(message)
	s.NewInfo(conv)
	s.sendNotification(g.String(), message.Content, c.Avatar())
	return nil
}

func (s *Siggo) sendNotification(title, content, iconPath string) {
	if s.config.TerminalBellNotifications {
		fmt.Print("\a")
	}
	if !s.config.DesktopNotifications {
		return
	}
	if !s.config.DesktopNotificationsShowMessage {
		content = ""
	}
	if !s.config.DesktopNotificationsShowAvatar {
		iconPath = ""
	}
	err := beeep.Notify(title, content, iconPath) // title, msg, icon
	if err != nil {
		log.Errorf("failed to send desktop notification: %v", err)
	}
}

// Conversations returns the current converstation book
func (s *Siggo) Conversations() map[*Contact]*Conversation {
	return s.conversations
}

// Contacts returns the current contact list
func (s *Siggo) Contacts() ContactList {
	return s.contacts
}

// Config returns a copy of the current configuration
func (s *Siggo) Config() Config {
	return *s.config
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
	s.signal.Close() // kills the signal-cli daemon
}

// NewSiggo creates a new model
func NewSiggo(sig SignalAPI, config *Config) *Siggo {
	s := &Siggo{
		config: config,
		signal: sig,

		NewInfo:    func(*Conversation) {}, // noop
		ErrorEvent: func(error) {},         // noop
	}
	s.init()
	//sig.OnMessage(s.?)

	sig.OnSent(s.onSent)
	sig.OnReceived(s.onReceived)
	sig.OnReceipt(s.onReceipt)
	sig.OnError(s.handleError)
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
	highestIndex := 0

	// get all contacts from disk
	contacts, err := sig.GetContactList()
	if err != nil {
		log.Warnf("failed to read contacts from disk: %v", err)
		return list
	}
	for _, c := range contacts {
		if c.InboxPosition != nil {
			alias := ""
			if s.config.ContactAliases != nil {
				alias = s.config.ContactAliases[c.Name]
			}
			// check if we have a color for this contact
			color := s.config.ContactColors[c.Name]
			contact := &Contact{
				Number: c.Number,
				Name:   c.Name,
				Index:  *c.InboxPosition,
				alias:  alias,
				color:  color,
			}
			list[c.Number] = contact
			if *c.InboxPosition > highestIndex {
				highestIndex = *c.InboxPosition
			}
		}
	}

	// get all groups from disk
	groups, err := sig.GetGroupList()
	if err != nil {
		log.Warnf("failed to read groups from disk: %v", err)
		return list
	}
	for _, g := range groups {
		if !g.Blocked && !g.Archived {
			alias := ""
			if s.config.ContactAliases != nil {
				alias = s.config.ContactAliases[g.Name]
			}
			highestIndex++
			list[g.GroupID] = &Contact{
				Number:  g.GroupID,
				Name:    g.Name,
				Index:   highestIndex,
				alias:   alias,
				isGroup: true,
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
		// check if we have a conversation file for this contact
		if s.config.SaveMessages {
			convPath := filepath.Join(ConversationFolder(), contact.Number)
			err := conv.Load(convPath, s.config) // if we fail to load, oh well
			if err == nil {
				log.Infof("loaded conversation from: %s", contact.Name)
			}
		}
		conv.CaughtUp()
		conversations[contact] = conv
	}
	return conversations
}
