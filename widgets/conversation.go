package widgets

import (
	"fmt"

	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"

	"github.com/derricw/siggo/model"
)

type ConversationPanel struct {
	*tview.TextView
	hideTitle       bool
	hidePhoneNumber bool
	// only show messages matching a filter
	filter string
}

func (p *ConversationPanel) Update(conv *model.Conversation) {
	p.Clear()
	p.SetText(conv.Filter(p.filter))
	if !p.hideTitle {
		if !p.hidePhoneNumber {
			p.SetTitle(fmt.Sprintf("%s <%s>", conv.Contact.String(), conv.Contact.Number))
		} else {
			p.SetTitle(conv.Contact.String())
		}
	}
	conv.HasNewMessage = false
}

func (p *ConversationPanel) Clear() {
	p.SetText("")
}

func (p *ConversationPanel) Filter(s string) {
	log.Debugf("filtering converstaion: %s", s)
	p.filter = s
}

func NewConversationPanel(siggo *model.Siggo) *ConversationPanel {
	c := &ConversationPanel{
		TextView: tview.NewTextView(),
	}
	c.SetDynamicColors(true)
	c.SetTitle("<name of contact>")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	return c
}
