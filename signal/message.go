package signal

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
	Attachments      []interface{} `json:"attachments"`
	GroupInfo        interface{}   `json:"groupInfo"`
	Destination      string        `json:"destination"`
}

type DataMessage struct {
	Timestamp        int64         `json:"timestamp"`
	Message          string        `json:"message"`
	ExpiresInSeconds int64         `json:"expiresInSeconds"`
	Attachments      []interface{} `json:"attachments"`
	GroupInfo        interface{}   `json:"groupInfo"`
}

type CallMessage interface{}

type ReceiptMessage struct {
	When       int64   `json:"when"`
	IsDelivery bool    `json:"isDelivery"`
	IsRead     bool    `json:"isRead"`
	Timestamps []int64 `json:"timestamps"`
}
