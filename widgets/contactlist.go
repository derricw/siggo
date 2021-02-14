package widgets

import (
	"fmt"

	"github.com/derricw/siggo/model"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

const DraftMarker = "~"

type ContactListPanel struct {
	*tview.TextView
	siggo          *model.Siggo
	parent         *ChatWindow
	sortedContacts []*model.Contact
	currentIndex   int
}

func (cl *ContactListPanel) Next() *model.Contact {
	return cl.GotoIndex(cl.currentIndex + 1)
}

func (cl *ContactListPanel) Previous() *model.Contact {
	return cl.GotoIndex(cl.currentIndex - 1)
}

// GotoIndex goes to a particular contact index and return the Contact. Negative indexing is
// allowed.
func (cl *ContactListPanel) GotoIndex(index int) *model.Contact {
	if index < 0 {
		return cl.GotoIndex(len(cl.sortedContacts) + index)
	}
	if index >= len(cl.sortedContacts) {
		index = 0
	}
	cl.currentIndex = index
	cl.ScrollTo(cl.currentIndex, 0)
	return cl.sortedContacts[index]
}

// GotoContact goes to a particular contact.
// TODO: constant time way to do this?
func (cl *ContactListPanel) GotoContact(contact *model.Contact) {
	for i, c := range cl.sortedContacts {
		if contact == c {
			cl.GotoIndex(i)
		}
	}
}

// Render the contact list
func (cl *ContactListPanel) Render() {
	data := ""
	log.Debug("updating contact panel...")
	// this is dumb, we re-sort every update
	// TODO: don't
	sorted := cl.siggo.Contacts().SortedByIndex()
	convs := cl.siggo.Conversations()
	log.Debugf("sorted contacts: %v", sorted)
	for i, c := range sorted {
		id := c.String()
		line := fmt.Sprintf("%s", id)
		color := c.Color()
		if cl.currentIndex == i {
			line = fmt.Sprintf("[%s::r]%s[-::-]", color, line)
		} else if convs[c].HasNewMessage {
			line = fmt.Sprintf("[%s::b]*%s[-::-]", color, line)
		} else {
			line = fmt.Sprintf("[%s::]%s[-::]", color, line)
		}
		if convs[c].HasStagedData() {
			line += DraftMarker
		}
		data += fmt.Sprintf("%s\n", line)
	}
	cl.sortedContacts = sorted
	cl.SetText(data)
}

// NewContactListPanel creates a new contact list widget
func NewContactListPanel(parent *ChatWindow, siggo *model.Siggo) *ContactListPanel {
	c := &ContactListPanel{
		TextView: tview.NewTextView(),
		siggo:    siggo,
		parent:   parent,
	}
	c.SetDynamicColors(true)
	c.SetTitle("contacts")
	c.SetTitleAlign(0)
	c.SetBorder(true)
	c.SetWrap(false)
	return c
}
