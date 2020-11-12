package signal

import ()

type Message struct {
	Envelope *Envelope `json:"envelope"`
}

type Envelope struct {
	Source         string          `json:"source"`
	Timestamp      int64           `json:"timestamp"`
	IsReceipt      bool            `json:"isReceipt"`
	SyncMessage    *SyncMessage    `json:"syncMessage"`
	CallMessage    *CallMessage    `json:"callMessage"`
	ReceiptMessage *ReceiptMessage `json:"receiptMessage"`
	DataMessage    *DataMessage    `json:"dataMessage"`
	SourceDevice   int             `json:"sourceDevice"`
}

type SyncMessage struct {
	SentMessage  *SentMessage `json:"sentMessage"`
	Type         interface{}  `json:"type"`
	ReadMessages interface{}  `json:"readMessages"`
}

type SentMessage struct {
	Timestamp        int64         `json:"timestamp"`
	Message          string        `json:"message"`
	ExpiresInSeconds int64         `json:"expiresInSeconds"`
	Attachments      []*Attachment `json:"attachments"`
	GroupInfo        GroupInfo     `json:"groupInfo"`
	Destination      string        `json:"destination"`
}

type DataMessage struct {
	Timestamp        int64         `json:"timestamp"`
	Message          string        `json:"message"`
	ExpiresInSeconds int64         `json:"expiresInSeconds"`
	Attachments      []*Attachment `json:"attachments"`
	GroupInfo        *GroupInfo    `json:"groupInfo"`
}

type CallMessage interface{}

type ReceiptMessage struct {
	When       int64   `json:"when"`
	IsDelivery bool    `json:"isDelivery"`
	IsRead     bool    `json:"isRead"`
	Timestamps []int64 `json:"timestamps"`
}

type GroupInfo struct {
	GroupID string   `json:"groupId"`
	Members []string `json:"members"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
}

type Attachment struct {
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
	ID          string `json:"id"`
	Size        int    `json:"size"`
}
