# siggo
[![GoDoc](https://godoc.org/github.com/derricw/siggo?status.svg)](https://godoc.org/github.com/derricw/siggo)
[![Go Report](https://goreportcard.com/badge/github.com/derricw/siggo)](https://goreportcard.com/badge/github.com/derricw/siggo)
![Build](https://github.com/derricw/siggo/workflows/Test/badge.svg)

A terminal ui for signal-cli, written in Go.

![Alt text](media/screenshot.jpg?raw=true "Screenshot")

### Unable to keep pace with `signal-cli`

I have not been keeping up with `signal-cli` development. They change the location and format of their local data constantly. If you want your install to work, use signal cli version 0.9.2 until I find time to update.

### Features

* vim-style ux
* useful for quick messages or use $EDITOR to compose fancy ones
* emoji support, just use colons, like `:cat:` or [the kitty emoji picker](https://sw.kovidgoyal.net/kitty/kittens/unicode-input.html)
* configurable contact [colors](config/README.md#configure-contact-colors)
* can use [fzf](https://github.com/junegunn/fzf) to fuzzy-find files to attach
* support for groups! (but not creating new groups)
* quickly filter messages by providing a regex pattern

### Dependencies

* [signal-cli](https://github.com/AsamK/signal-cli). (==0.9.2)

siggo uses the dbus daemon feature of signal-cli, so `libunixsocket-java` (Debian), `libmatthew-java` (Fedora) or `libmatthew-unix-java` (AUR) is required. There seems to be a `brew` [forumla](https://formulae.brew.sh/formula/dbus) for dbus on MacOS.

Install signal-cli and put it somewhere safe in your path. You will need to follow its instructions to either [link](https://github.com/AsamK/signal-cli/wiki/Linking-other-devices-(Provisioning)) or [register](https://github.com/AsamK/signal-cli#usage) your device. Alternatively, the `siggo link <phonenumber> <devicename>` subcommand has been added to make linking more user-friendly. Be sure to prefix with `+` and country code (for example `+12345678901`).

When setup is finished, you should be able to run without error:

```
signal-cli -u +<yourphonenumber> -o json receive
```
You are now ready to use `siggo`.

#### Optional Dependencies

* [fzf](https://github.com/junegunn/fzf)

Some nice features are available if you have `fzf` in your PATH.

### Security

siggo shells out to `signal-cli`, so if that worries you, don't use it, for now. I have lofty goals of eventually replacing this with [libsignal](https://github.com/signalapp/libsignal-protocol-c).

### Build

siggo should build on Linux or MacOS, but has primarily been tested on Linux.

```
make build
```

### Run

```
bin/siggo
```

### Updating

If you are updating from a previous version, I recommend deleting your conversation files first. See below.

### Keybinds

* `j` - Scroll Down
* `k` - Scroll Up
* `J` - Next Contact
* `K` - Previous Contact
* `n` - Move to next conversation with unread messages
* `t` - Use fzf to goto contact with fuzzy matching
* `a` - Attach file (sent with next message)
* `A` - Use fzf to attach a file
* `/` - Filter conversation by providing a pattern
* `i` - Insert Mode
  * `CTRL+L` - Clear input field (also clears staged attachments)
* `I` - Compose (opens $EDITOR and lets you make a fancy message)
* `y` - Yank Mode
  * `yy` - Yank Last Message (from current conversation)
  * `yl` - Yank Last URL
* `o` - Open Mode
  * `Enter` - Open selected attachment
  * `oo` - Open Last Attachment
* `l` - Link Mode
  * `Enter` - Open selected link in browser
  * `ll` - Open Last URL
  * `y` - Yank selected link to clipboard
* `p` or `CTRL+V` - Paste text/attach file in clipboard
* `ESC` - Normal Mode
* `CTRL+Q` - Quit (`CTRL+C` _should_ also work)

### Configuration

See the configuration README [here](config/README.md).

### Message History

Message saving is an opt-in feature.

If you enable it, conversations are stored in plain text in `~/.local/share/siggo/conversations`.

Delete them like this:

```
rm ~/.local/share/siggo/conversations/*
```

### Troubleshooting

I've started a wiki [here](https://github.com/derricw/siggo/wiki/Troubleshooting).

### Development

Honestly the code is a hot mess right now, and I don't recommend trying to contribute yet. But I will absolutely take a PR if you want to throw one at me.

If you save the output of signal-cli like so:

```
signal-cli -u +<yourphonenumber> receive --json > example_messages.json
```
You can then run siggo using it as mock input. This is useful for development and testing.
```
bin/siggo -m example_messages.json
```
This way you can test without sending yourself messages.

### Similar Projects / Inspiration

* [signal-curses](https://github.com/jwoglom/signal-curses)
* [scli](https://github.com/isamert/scli)

### Roadmap

Here is a list of things that are currently broken.
* Send read receipts for incoming messages (`signal-cli` limitation, but might be fixed soon)

Here is a list of features I'd like to add soonish.
* Better Attachments Support
  * signal-cli seems to delete old attachments after a while. maybe I should move them somewhere where they wont get deleted?
* default color list for contacts instead of white
* better mode indication
* gui configuration
  * colors and border styles
* let user re-sort contact list (for example alphabetically)
* use dbus to send instead of signal-cli, to avoid having to spin up the JVM
* there is still some data that I'm dropping on the floor (I believe it to be the "typing indicator" messages)
* weechat/BitlBee plugin that uses the siggo model without the UI
* wouldn't tests be neat?
