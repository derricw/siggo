package widgets

import (
	"fmt"

	"github.com/derricw/siggo/model"
	"github.com/rivo/tview"
)

type ConversationPanel struct {
	*tview.TextView
	hideTitle       bool
	hidePhoneNumber bool
}

func (p *ConversationPanel) Update(conv *model.Conversation) {
	p.Clear()
	p.SetText(conv.String())
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
