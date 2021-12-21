package widgets

import (
	"fmt"
	"strings"

	"github.com/derricw/siggo/model"
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

type SendPanel struct {
	*tview.InputField
	parent *ChatWindow
	siggo  *model.Siggo
}

func (s *SendPanel) Send() {
	if !s.isDataStaged() {
		return
	}
	msg := s.GetText()
	contact := s.parent.currentContact
	s.parent.ShowTempSentMsg(msg)
	go s.siggo.Send(msg, contact)
	log.Infof("sent message: %s to contact: %s", msg, contact)
	s.SetText("")
	s.SetLabel("")
}

func (s *SendPanel) Clear() {
	s.SetText("")
	conv, err := s.parent.currentConversation()
	if err != nil {
		return
	}
	conv.ClearAttachments()
	s.Update()
}

// returns true if there is either a message or attachment to send
func (s *SendPanel) isDataStaged() bool {
	if len(s.GetText()) > 0 {
		return true
	}
	if conv, err := s.parent.currentConversation(); err != nil {
		return false
	} else if conv.NumAttachments() > 0 {
		return true
	}
	return false
}

func (s *SendPanel) Defocus() {
	s.parent.NormalMode()
}

func (s *SendPanel) Update() {
	conv, err := s.parent.currentConversation()
	if err != nil {
		return
	}
	nAttachments := conv.NumAttachments()
	if nAttachments > 0 {
		s.SetLabel(fmt.Sprintf("📎(%d) ", nAttachments))
	} else {
		s.SetLabel("")
	}
	if conv.StagedMessage != "" {
		s.SetText(conv.StagedMessage)
	}
}

// emojify is a custom input change handler that provides emoji support
func (s *SendPanel) emojify(input string) {
	if strings.HasSuffix(input, ":") {
		emojified := emoji.Sprint(input)
		if emojified != input {
			s.SetText(emojified)
		}
	}
}

// NewSendPanel creates a new SendPanel that is primarily a tview.InputField
func NewSendPanel(parent *ChatWindow, siggo *model.Siggo) *SendPanel {
	s := &SendPanel{
		InputField: tview.NewInputField(),
		siggo:      siggo,
		parent:     parent,
	}
	s.SetTitle(" send: ")
	s.SetTitleAlign(0)
	s.SetBorder(true)
	//s.SetFieldBackgroundColor(tcell.ColorDefault)
	s.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	s.SetChangedFunc(s.emojify)
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			s.Defocus()
			return nil
		case tcell.KeyEnter:
			s.Send()
			return nil
		case tcell.KeyCtrlQ:
			s.parent.Quit()
		case tcell.KeyCtrlL:
			s.Clear()
			return nil
		}
		return event
	})
	return s
}
