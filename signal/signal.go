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

// GetSignalAvatarsFolder returns the user's signal-cli avatars folder
func GetSignalAvatarsFolder() (string, error) {
	signalFolder, err := GetSignalFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(signalFolder, "avatars"), nil
}

// SignalContact is the data signal-cli saves for each contact
// in SignalDataDir/<phonenumber>.
// This structure no longer exists as of signal-cli >= 0.8.2
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

// SignalRecipient is the format used in the `recipients-store` file.
type SignalRecipient struct {
	ID                   int         `json:"id"`
	Number               string      `json:"number"`
	UUID                 string      `json:"uuid"`
	Profile              interface{} `json:"profile"`
	ProfileKey           string      `json:"profileKey"`
	ProfileKeyCredential interface{} `json:"profileKeyCredential"`
	Contact              struct {
		Name                  string `json:"name"`
		Color                 string `json:"color"`
		MessageExpirationTime int    `json:"messageExpirationTime"`
		Blocked               bool   `json:"blocked"`
		Archived              bool   `json:"archived"`
	} `json:"contact"`
}

func (r *SignalRecipient) AsContact() *SignalContact {
	return &SignalContact{
		Name:                  r.Contact.Name,
		Number:                r.Number,
		Color:                 r.Contact.Color,
		MessageExpirationTime: r.Contact.MessageExpirationTime,
		ProfileKey:            r.ProfileKey,
		Blocked:               r.Contact.Blocked,
		Archived:              r.Contact.Archived,
	}
}

// SignalRecipientStore
type SignalRecipientStore struct {
	Recipients []*SignalRecipient `json:"recipients"`
}

func (r *SignalRecipientStore) AsContacts() []*SignalContact {
	contacts := make([]*SignalContact, 0)
	for _, c := range r.Recipients {
		contacts = append(contacts, c.AsContact())
	}
	return contacts
}

// SignalGroup is the data that signal-cli saves for each group
// in SignalDataDir/<phonenumber>
type SignalGroup struct {
	GroupID          string `json:"groupId"`
	MasterKey        string `json:"masterKey"`
	Blocked          bool   `json:"blocked"`
	PermissionDenied bool   `json:"permissionDenied"`
}

// SignalGroupInfo is the data that is retrived for each group when we call
// `signal-cli listGroups` As far as I know right now that is the only way to
// get the group names. They don't appear to be saved in signal-cli's data.
type SignalGroupInfo struct {
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	IsMember              bool                `json:"isMember"`
	IsBlocked             bool                `json:"isBlocked"`
	Description           string              `json:"description"`
	Members               []SignalGroupMember `json:"members"`
	Admins                []SignalGroupMember `json:"admins"`
	PendingMembers        []interface{}       `json:"pendingMembers"`
	RequestingMembers     []interface{}       `json:"requestingMembers"`
	GroupInviteLink       interface{}         `json:"groupInviteLink"`
	PermissionAddMember   string              `json:"permissionAddMember"`
	PermissionEditDetails string              `json:"permissionEditDetails"`
	PermissionSendMessage string              `json:"permissionSendMessage"`
}

// SignalGroupMember is a member of a signal group
type SignalGroupMember struct {
	UUID   string `json:"uuid"`
	Number string `json:"number"`
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
	daemon            *exec.Cmd
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
	cmd := exec.Command("signal-cli", "-o", "json", "-u", s.uname, "daemon")

	//  This is the only way to ensure that the signal-cli daemon is killed when we get
	//  SIGKILL, but it isn't available on MacOS, so we leave it commented out for now.
	//  This means that SIGKILL will leave an orphaned signal-cli daemon running that will have to be
	//  killed manually. I have tried unsuccessfully to find a cross-platform solution for this.
	//  Other signals, like SIGTERM and SIGINT should be handled correctly.

	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//Pdeathsig: syscall.SIGKILL,
	//}

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		s.publishError(err)
		return err
	}
	s.daemon = cmd

	scanner := bufio.NewScanner(outReader)
	log.Infof("scanning stdout")
	for scanner.Scan() {
		wire := scanner.Bytes()
		log.Debugf("wire (length %d): %s", len(wire), wire)
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
// Returns the message ID
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

// SendGroupDbus does the same thing as SendDbus but to a group
func (s *Signal) SendGroupDbus(groupID, msg string, attachments ...string) (int64, error) {
	args := []string{"--dbus", "send", "-g", groupID, "-m", msg}
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

// RequestGroupInfo requests info for all groups from the Signal network
func (s *Signal) RequestGroupInfo() ([]SignalGroupInfo, error) {
	cmd := exec.Command("signal-cli", "-o", "json", "-u", s.uname, "listGroups")
	out, err := cmd.Output()
	if err != nil {
		s.publishError(err)
		return nil, err
	}
	groupInfo := []SignalGroupInfo{}
	if err = json.Unmarshal(out, &groupInfo); err != nil {
		return nil, err
	}
	return groupInfo, nil
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
// this is where the contact list is kept for signal-cli < 0.8.2
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

// GetRecipientStore gets the recipient store. This is where the contacts list is kept in
// signal-cli >= 0.8.2
func (s *Signal) GetRecipientStore() (*SignalRecipientStore, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	homeDir := usr.HomeDir
	dataFile := fmt.Sprintf("%s/%s/%s.d/recipients-store", homeDir, SignalDataDir, s.uname)
	b, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}
	store := &SignalRecipientStore{}
	if err = json.Unmarshal(b, store); err != nil {
		return nil, err
	}

	// sometimes there are contacts with no number...
	// idk why but lets filter those out
	filtered := &SignalRecipientStore{}
	for _, recipient := range store.Recipients {
		if recipient.Number != "" {
			filtered.Recipients = append(filtered.Recipients, recipient)
		}
	}
	return filtered, nil
}

// GetContactList attempts to read an existing contact list from the signal user directory.
func (s *Signal) GetContactList() ([]*SignalContact, error) {
	recipients, err := s.GetRecipientStore()
	if err != nil {
		return nil, err
	}
	// recipients stored in `recipients-store` file
	return recipients.AsContacts(), nil
}

// GetGroupList attempts to read an existing contact list from the signal user directory.
func (s *Signal) GetGroupList() ([]*SignalGroup, error) {
	userData, err := s.GetUserData()
	if err != nil {
		return nil, err
	}
	return userData.GroupStore.Groups, nil
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

// Close cleans up any subprocesses
func (s *Signal) Close() {
	if s.daemon != nil {
		log.Debug("killing signal-cli daemon...")
		_ = s.daemon.Process.Signal(os.Interrupt)
	}
}

// NewSignal returns a new signal instance for the specified user.
func NewSignal(uname string) *Signal {
	return &Signal{
		uname: uname,
	}
}
