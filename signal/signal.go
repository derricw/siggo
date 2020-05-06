package signal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SignalDataDir - signal-cli saves user data here
var SignalDataDir string = ".local/share/signal-cli/data"

// SignalContact is the data signal-cli saves for each contact
// in SignalDataDir/<phonenumber>
type SignalContact struct {
	Name                  string `json:"name"`
	Number                string `json:"number"`
	Color                 string `json:"color"`
	MessageExpirationTime int    `json:"messageExpirationTime"`
	ProfileKey            string `json:"profileKey"`
	Blocked               bool   `json:"blocked"`
	InboxPosition         *int   `json:"inboxPosition"`
	Archived              bool   `json:"archived"`
}

// SignalGroup is the data that signal-cli saves for each group
// in SignalDataDir/<phonenumber>
type SignalGroup struct {
	GroupId               string        `json:"groupId"`
	Name                  string        `json:"name"`
	Members               []interface{} `json:"members"`
	Color                 string        `json:"color"`
	Blocked               bool          `json:"blocked"`
	InboxPosition         int           `json:"inboxPosition"`
	Archived              bool          `json:"archived"`
	MessageExpirationTime int           `json:"messageExpirationTime"`
}

// SignalUserData is the data signal saves for a given user
// in SignalDataDir/<phonenumber>
type SignalUserData struct {
	ContactStore struct {
		Contacts []*SignalContact `json:"contacts"`
	} `json:"contactStore"`
	GroupStore struct {
		Groups []*SignalGroup `json:"groups"`
	} `json:"groupStore"`
}

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
		//fmt.Printf("%s\n", wire) // TODO: log to file?
		err = s.ProcessWire(wire)
		if err != nil {
			return err
		}
	}
	return err
}

// ReceiveUntil receives contiuously until it receives a stop signal
func (s *Signal) ReceiveUntil(done chan struct{}) {
	go func() {
		// better to select with timeout?
		for len(done) == 0 {
			err := s.Receive()
			if err != nil {
				log.Printf("receive failed: %v", err)
			}
		}
	}()
}

// Send transmits a message to the specified number
// Destination is a phone number with country code.
// signal-cli likes to have a `+` before the number, so we add one if it isn't there.
func (s *Signal) Send(dest, msg string) error {
	if !strings.HasPrefix(dest, "+") {
		dest = fmt.Sprintf("+%s", dest)
	}
	_, err := s.Exec("-u", s.uname, "send", dest, "-m", msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *Signal) GetUserData() (*SignalUserData, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	homeDir := usr.HomeDir
	dataFile := fmt.Sprintf("%s/%s/%s", homeDir, SignalDataDir, s.uname)
	b, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}
	userData := &SignalUserData{}
	if err = json.Unmarshal(b, userData); err != nil {
		return nil, err
	}
	return userData, nil
}

// GetContactList attempts to read an existing contact list from the signal user directory.
func (s *Signal) GetContactList() ([]*SignalContact, error) {
	userData, err := s.GetUserData()
	if err != nil {
		return nil, err
	}
	return userData.ContactStore.Contacts, nil
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
