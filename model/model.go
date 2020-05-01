package model

import (
	"github.com/derricw/siggo/signal"
)

type Config struct {
	user string
}

type ContactList struct{}

type Contact struct {
	Number string
	Name   string
}

type Message struct {
	Content     string
	From        string
	Timestamp   int64
	IsDelivered bool
	IsRead      bool
}

type Conversation struct {
	Contact       *Contact
	Messages      []*Message
	HasNewMessage bool
}

type Siggo struct {
	config        *Config
	contacts      *ContactList
	conversations []*Conversation
	signal        *signal.Signal
}

func (s *Siggo) Send(msg string, contact *Contact) {
}
