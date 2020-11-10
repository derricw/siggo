package widgets

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
)

// LinksInput is a widget that allows us to select from links in the conversation
type LinksInput struct {
	*tview.List
	parent *ChatWindow
	links  []string
}

func (li *LinksInput) Close() {
	li.parent.Grid.RemoveItem(li)
	li.parent.ShowConversation()
	li.parent.FocusMe()
}

// init populates the list with links
func (li *LinksInput) init() {
	li.Clear()
	l := li.parent.getLinks() // should be sorted by date
	li.links = l
	for _, item := range l {
		li.AddItem(fmt.Sprintf(" %s", item), "", 0, nil)
	}
}

func (li *LinksInput) Previous() {
	current := li.GetCurrentItem()
	li.SetCurrentItem(current - 1)
}

func (li *LinksInput) Next() {
	current := li.GetCurrentItem()
	li.SetCurrentItem(current + 1)
}

// OpenSelected opens whichever link is selected
func (li *LinksInput) OpenSelected() {
	nlinks := len(li.links)
	selected := li.GetCurrentItem()
	if nlinks == 0 || selected >= nlinks {
		return
	}
	li.OpenLink(li.links[selected])
}

// OpenLink opens any URL
func (li *LinksInput) OpenLink(link string) {
	err := open.Run(link)
	if err != nil {
		li.parent.SetErrorStatus(fmt.Errorf("<OPEN FAILED: %v>", err))
	} else {
		li.parent.SetStatus(fmt.Sprintf("📂%s", link))
	}
}

// YankSelected copies copies the currently selected link to the clipboard
func (li *LinksInput) YankSelected() {
	li.Close()
	nlinks := len(li.links)
	selected := li.GetCurrentItem()
	if nlinks == 0 || selected >= nlinks {
		return
	}
	link := li.links[selected]
	if err := clipboard.WriteAll(link); err != nil {
		li.parent.SetErrorStatus(err)
		return
	}
	li.parent.SetStatus(fmt.Sprintf("📋%s", link))
}

// OpenLast opens the most recent link
func (li *LinksInput) OpenLast() {
	li.Close()
	links := li.links
	if len(links) > 0 {
		last := links[len(links)-1]
		li.OpenLink(last)
	} else {
		li.parent.SetStatus(fmt.Sprintf("📂<NO MATCHES>"))
	}
}

func NewLinksInput(parent *ChatWindow) *LinksInput {
	li := &LinksInput{
		List:   tview.NewList(),
		parent: parent,
	}
	inputHandler := li.List.InputHandler()
	li.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Debugf("Key Event <LINK>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyESC:
			li.Close()
			li.parent.NormalMode()
			return nil
		//case tcell.KeyUp:
		//li.Next()
		//return nil
		//case tcell.KeyDown:
		//li.Previous()
		//return nil
		case tcell.KeyPgUp:
			inputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyPgDn:
			inputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyEnd:
			inputHandler(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyHome:
			inputHandler(event, func(p tview.Primitive) {})
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case 108: // l
				li.OpenLast()
				return nil
			case 106: // j
				li.Next()
				return nil
			case 107: // k
				li.Previous()
				return nil
			case 121: // y
				li.YankSelected()
				return nil
			}
		case tcell.KeyEnter:
			li.OpenSelected()
			return nil
		}

		return event
	})

	//li.SetDynamicColors(true)
	li.SetHighlightFullLine(true)
	li.ShowSecondaryText(false)
	li.SetBorder(true)
	li.init()
	li.SetCurrentItem(-1)

	return li
}
