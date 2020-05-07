package signal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

var fakeSendReceipt *Message = &Message{
	Envelope: &Envelope{
		Source:       "",
		SourceDevice: 5,
		Timestamp:    0,
		IsReceipt:    false,
		DataMessage:  nil,
		SyncMessage: &SyncMessage{
			SentMessage: &SentMessage{
				Timestamp:   0,
				Message:     "",
				Destination: "",
			},
		},
	},
}

// MockSignal implements siggo's SignalAPI interface
type MockSignal struct {
	*Signal
	exampleData []byte
	userNumber  string
}

func (ms *MockSignal) Send(dest, msg string) error {
	// send a fake message, just puts in on the "wire"
	timestamp := time.Now().Unix()
	fakeWire := fakeSendReceipt
	fakeWire.Envelope.Timestamp = timestamp
	fakeWire.Envelope.SyncMessage.SentMessage.Timestamp = timestamp
	fakeWire.Envelope.SyncMessage.SentMessage.Message = msg
	fakeWire.Envelope.SyncMessage.SentMessage.Destination = dest

	log.Printf("%v", fakeWire)
	b, err := json.Marshal(fakeWire)
	if err != nil {
		return fmt.Errorf("failed to marshal send receipt: %v", err)
	}
	ms.exampleData = append(ms.exampleData, b...)
	return nil
}

func (ms *MockSignal) SendDbus(dest, msg string) error {
	return ms.Send(dest, msg)
}

func (ms *MockSignal) Receive() error {
	r := bytes.NewReader(ms.exampleData)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		wire := scanner.Bytes()
		err := ms.ProcessWire(wire)
		if err != nil {
			return err
		}
	}
	ms.exampleData = []byte{}
	return nil
}

func (ms *MockSignal) ReceiveForever() {
	go func() {
		ms.Receive()
		time.Sleep(time.Second * 1)
	}()
}

func NewMockSignal(userNumber string, exampleData []byte) *MockSignal {
	return &MockSignal{
		Signal:      NewSignal(userNumber),
		exampleData: exampleData,
		userNumber:  userNumber,
	}
}
