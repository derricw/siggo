// Package signal provides a minimal wrapper for the signal-cli functionality and data
// types that we need.
package signal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	qr "github.com/mdp/qrterminal/v3"
	log "github.com/sirupsen/logrus"
)

// SignalDir is where signal-cli saves its local data
var SignalDir string = ".local/share/signal-cli"
var SignalDataDir string = fmt.Sprintf("%s/data", SignalDir)
var SignalAttachmentsDir string = fmt.Sprintf("%s/attachments", SignalDir)
var SignalAvatarsDir string = fmt.Sprintf("%s/avatars", SignalDir)

// GetSignalFolder returns the user's signal-cli local storage
func GetSignalFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, SignalDir), nil
}

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
type ErrorCallback func(error)

// Exec invokes signal-cli with the supplied args and returns the bytes that writes to stdout
func Exec(args ...string) ([]byte, error) {
	var out bytes.Buffer
	cmd := exec.Command("signal-cli", args...)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []byte{}, err
	}
	return out.Bytes(), nil
}

// Signal represents a signal-cli session for a given user. It can be run in daemon mode by calling
// the `Daemon` method. It can also be used to send and receive manually using `Receive` and `Send`.
type Signal struct {
	uname             string
	msgCallbacks      []MessageCallback
	sentCallbacks     []SentCallback
	receiptCallbacks  []ReceiptCallback
	receivedCallbacks []ReceivedCallback
	errorCallbacks    []ErrorCallback
}

// OnMessage registers a callback to be executed upon any incoming message of any kind (that we
// know how to process).
func (s *Signal) OnMessage(callback MessageCallback) {
	s.msgCallbacks = append(s.msgCallbacks, callback)
}

// OnSent registers a callback to be executed whenver a sent message appears on the wire (currently,
// this is when a different linked device sends a message).
func (s *Signal) OnSent(callback SentCallback) {
	s.sentCallbacks = append(s.sentCallbacks, callback)
}

// OnReceipt registers a callback to be executed whenever a message receipt (for example a Read
// receipt or a Delivery receipt) is received.
func (s *Signal) OnReceipt(callback ReceiptCallback) {
	s.receiptCallbacks = append(s.receiptCallbacks, callback)
}

// OnReceived registers a callback to be executed whenver an incoming message is received.
func (s *Signal) OnReceived(callback ReceivedCallback) {
	s.receivedCallbacks = append(s.receivedCallbacks, callback)
}

// OnError registers a callback to be executed whenever an error occurs.
func (s *Signal) OnError(callback ErrorCallback) {
	s.errorCallbacks = append(s.errorCallbacks, callback)
}

func (s *Signal) publishError(err error) {
	for _, cb := range s.errorCallbacks {
		cb(err)
	}
}

// Version returns the current version of signal-cli
func (s *Signal) Version() (string, error) {
	b, err := Exec("-v")
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
	b, err := Exec("-u", s.uname, "receive", "--json")
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		wire := scanner.Bytes()
		err = s.ProcessWire(wire)
		if err != nil {
			return err
		}
	}
	return err
}

// ReceiveForever receives contiuously. WARNING: this will continuously start and stop the JVM and
// is not recommended unless you want to simulate the Electon app's recource useage.
func (s *Signal) ReceiveForever() {
	go func() {
		for {
			log.Infof("starting dbus daemon...")
			err := s.Daemon()
			if err != nil {
				log.Error(fmt.Errorf("daemon failed to start... restarting in 5 seconds..."))
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

// Daemon starts the dbus daemon and receives forever.
func (s *Signal) Daemon() error {
	cmd := exec.Command("signal-cli", "-u", s.uname, "daemon", "--json")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		s.publishError(err)
		return err
	}

	scanner := bufio.NewScanner(outReader)
	log.Infof("scanning stdout")
	for scanner.Scan() {
		wire := scanner.Bytes()
		log.Printf("wire (length %d): %s", len(wire), wire)
		err = s.ProcessWire(wire)
		if err != nil {
			return err
		}
	}
	return nil
}

// Send transmits a message to the specified number
// Destination is a phone number with country code.
// signal-cli likes to have a `+` before the number, so we add one if it isn't there.
func (s *Signal) Send(dest, msg string) (int64, error) {
	if !strings.HasPrefix(dest, "+") {
		dest = fmt.Sprintf("+%s", dest)
	}
	cmd := exec.Command("signal-cli", "-u", s.uname, "send", dest, "-m", msg)
	out, err := cmd.Output()
	if err != nil {
		s.publishError(err)
		return 0, err
	}
	ID, err := strconv.Atoi(string(out[:len(out)-1])) //strip newline
	if err != nil {
		return 0, err
	}
	return int64(ID), nil
}

// SendDbus does the same thing as Send but it goes through a running daemon.
func (s *Signal) SendDbus(dest, msg string, attachments ...string) (int64, error) {
	if !strings.HasPrefix(dest, "+") {
		dest = fmt.Sprintf("+%s", dest)
	}
	args := []string{"--dbus", "send", dest, "-m", msg}
	if len(attachments) > 0 {
		// how do I do this in one line?
		args = append(args, "-a")
		args = append(args, attachments...)
	}
	cmd := exec.Command("signal-cli", args...)
	out, err := cmd.Output()
	if err != nil {
		s.publishError(err)
		return 0, err
	}
	ID, err := strconv.Atoi(string(out[:len(out)-1])) //strip newline
	if err != nil {
		return 0, err
	}
	return int64(ID), nil
}

// Link will attempt to link to an existing registered device.
func (s *Signal) Link(deviceName string) error {
	cmd := exec.Command("signal-cli", "link", "-n", deviceName)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	r := bufio.NewReader(out)
	line, _, err := r.ReadLine()
	if err != nil {
		return err
	}
	fmt.Printf("link text: %s\n", line)
	qr.Generate(fmt.Sprintf("%s", line), qr.L, os.Stdout)
	return cmd.Wait()
}

// GetUserData returns the user data for the current user.
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
		return err
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
