package widgets

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
)

// CommandInput is an input field that appears at the bottom of the window and allows for various
// commands
type CommandInput struct {
	*tview.InputField
	parent *ChatWindow
}

// NewAttachInput is a command input that selects an attachment and attaches it to the current
// conversation to be sent in the next message.
func NewAttachInput(parent *ChatWindow) *CommandInput {
	ci := &CommandInput{
		InputField: tview.NewInputField(),
		parent:     parent,
	}
	ci.SetLabel("📎: ")
	ci.SetText("~/")
	ci.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	ci.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Setup keys
		log.Debugf("Key Event <ATTACH>: %v mods: %v rune: %v", event.Key(), event.Modifiers(), event.Rune())
		switch event.Key() {
		case tcell.KeyESC:
			ci.parent.HideCommandInput()
			return nil
		case tcell.KeyTAB:
			ci.SetText(CompletePath(ci.GetText()))
			return nil
		case tcell.KeyEnter:
			path := ci.GetText()
			ci.parent.HideCommandInput()
			if path == "" {
				return nil
			}
			conv, err := ci.parent.currentConversation()
			if err != nil {
				ci.parent.SetErrorStatus(fmt.Errorf("couldn't find conversation: %v", err))
				return nil
			}
			err = conv.AddAttachment(path)
			if err != nil {
				ci.parent.SetErrorStatus(fmt.Errorf("failed to attach: %s - %v", path, err))
				return nil
			}
			ci.parent.InsertMode()
			ci.parent.sendPanel.Update()
			return nil
		}
		return event
	})
	return ci
}

// AttachFromClipboard attaches a file directly from clipboard. Text is just pasted into the
// message.
func AttachFromClipboard(parent *ChatWindow) error {
	content, err := clipboard.ReadAll()
	if err != nil {
		return err
	}

	contentBytes := []byte(content)
	mimetype := http.DetectContentType(contentBytes)

	if strings.HasPrefix(mimetype, "text/") {
		// If clipboard contains text, paste it instead of attaching
		parent.sendPanel.SetText(parent.sendPanel.GetText() + content)
		parent.InsertMode()
		return nil
	}

	ext, err := mime.ExtensionsByType(mimetype)
	if err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile("", "*"+ext[0])
	if err != nil {
		return err
	}
	if _, err = tmpFile.Write(contentBytes); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	conv, err := parent.currentConversation()
	if err != nil {
		return err
	}
	if err := conv.AddAttachment(tmpFile.Name()); err != nil {
		return err
	}
	parent.sendPanel.Update()
	parent.InsertMode()

	return nil
}
