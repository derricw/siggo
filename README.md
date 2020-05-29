siggo
---

A terminal gui for signal-cli, written in Go.

### Dependencies

* [signal-cli](https://github.com/AsamK/signal-cli).

siggo uses the dbus daemon feature of signal-cli, so `libunixsocket-java` (Debian) or `libmatthew-unix-java` (AUR) is required.

Install signal-cli and put it somewhere safe in your path. You will need to follow its instructions to either [link](https://github.com/AsamK/signal-cli/wiki/Linking-other-devices-(Provisioning)) or [register](https://github.com/AsamK/signal-cli#usage) your device. The `siggo link <phonenumber>` subcommand has been added to make linking more user-friendly, but has not been tested sufficiently. Be sure to prefix with `+` and country code (for example `+12345678901`).

When setup is finished, you should be able to run without error:

```
signal-cli -u +<yourphonenumber> receive --json
```
You are now ready to use `siggo`.

### Build

```
make build
```

### Run

```
bin/siggo
```

### Keybinds

* `j` - Scroll Down
* `k` - Scroll Up
* `J` - Next Contact
* `K` - Previous Contact
* `i` - Insert Mode
  * `CTRL+L` - Clear input field
* `I` - Compose (opens $EDITOR and lets you make a fancy message)
* `y` - Yank Mode
  * `yy` - Yank Last Message (from current conversation)
  * `yl` - Yank Last URL
* `o` - Open Mode (doesn't work yet)
  * `oo` - Open Last Attachment
  * `ol` - Open Last Link
* `ESC` - Normal Mode
* `ALT+Q` - Quit

### Configure Contact Colors

Supports W3C and hex.

Note: Matches on contact name, so if you have two people with the same name they will get the same color.

```
siggo cfg color "Leloo Dallas" DarkViolet
siggo cfg color "Ruby Rhod" "#00FF00"
```

### Development

If you save the output of signal-cli like so:

```
signal-cli -u +<yourphonenumber> receive --json > example_messages.json
```
You can then run siggo using it as mock input. This is useful for development and testing.
```
bin/siggo -m example_messages.json
```
This way you can test without sending yourself messages.

### Roadmap

Here is a list of things that are currently broken.
* Send read receipts for incoming messages (`signal-cli` limitation, but might be fixed soon)

Here is a list of features I'd like to add.
* Attachments support
* gui configuration
* let user re-sort contact list (for example alphabetically)
* command to go to next contact with message waiting
* command to go to contact with fuzzy matching
* groups support
* use dbus to send instead of signal-cli, to avoid having to spin up the JVM (might also fix the read receipt issue)
* there is still some data that i'm dropping on the floor
