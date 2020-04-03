package signal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type MessageCallback func(*Message) error
type SentCallback func(*Message) error
type ReceiptCallback func(*Message) error
type ReceivedCallback func(*Message) error

type Signal struct {
	uname             string
	msgCallbacks      []MessageCallback
	sentCallbacks     []SentCallback
	receiptCallbacks  []ReceiptCallback
	receivedCallbacks []ReceivedCallback
}

func (s *Signal) OnMessage(callback MessageCallback) {
	s.msgCallbacks = append(s.msgCallbacks, callback)
}

func (s *Signal) OnSent(callback SentCallback) {
	s.sentCallbacks = append(s.sentCallbacks, callback)
}

func (s *Signal) OnReceipt(callback ReceiptCallback) {
	s.receiptCallbacks = append(s.receiptCallbacks, callback)
}

func (s *Signal) OnReceived(callback ReceivedCallback) {
	s.receivedCallbacks = append(s.receivedCallbacks, callback)
}

func (s *Signal) Exec(args ...string) ([]byte, error) {
	var out bytes.Buffer
	cmd := exec.Command("signal-cli", args...)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []byte{}, err
	}
	return out.Bytes(), nil
}

// Link a device
// not implemented but wouldn't it be cool
func (s *Signal) Link() {}

// Version returns the current version of signal-cli
func (s *Signal) Version() (string, error) {
	b, err := s.Exec("-v")
	if err != nil {
		return "", err
	}
	versionStr := fmt.Sprintf("%s", b)
	versionNum := strings.Split(versionStr, " ")
	if len(versionNum) == 0 {
		return "", err
	} else if len(versionNum) == 1 {
		return versionNum[0], nil
	}
	return versionNum[1], nil
}

// Receive receives and processes all outstanding messages
func (s *Signal) Receive() error {
	b, err := s.Exec("-u", s.uname, "receive", "--json")
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		wire := scanner.Bytes()
		fmt.Printf("%s\n", wire)
		err = s.ProcessWire(wire)
		if err != nil {
			return err
		}
	}
	return err
}

// ProcessWire processes a single wire message, executing any callbacks we
// have registered.
func (s *Signal) ProcessWire(wire []byte) error {
	var msg Message
	err := json.Unmarshal(wire, &msg)
	if err != nil {
		log.Printf("failed to unmarshal message: %s - %s", wire, err)
	}
	for _, cb := range s.msgCallbacks {
		err = cb(&msg)
		if err != nil {
			return err
		}
	}
	if msg.Envelope.DataMessage != nil {
		for _, cb := range s.receivedCallbacks {
			err = cb(&msg)
			if err != nil {
				return err
			}
		}
	}
	if msg.Envelope.SyncMessage != nil {
		if msg.Envelope.SyncMessage.SentMessage != nil {
			for _, cb := range s.sentCallbacks {
				err = cb(&msg)
				if err != nil {
					return err
				}
			}
		}
	}
	if msg.Envelope.ReceiptMessage != nil {
		for _, cb := range s.receiptCallbacks {
			err = cb(&msg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NewSignal returns a new signal instance for the specified user.
func NewSignal(uname string) *Signal {
	return &Signal{
		uname: uname,
	}
}
