siggo
---

A terminal gui for signal-cli, written in Go.

### Dependencies

* [signal-cli](https://github.com/AsamK/signal-cli).

siggo uses the dbus daemon feature of signal-cli, so `libunixsocket-java` (Debian) or `libmatthew-unix-java` (AUR) is required.

Install signal-cli and put it somewhere safe in your path. You will need to follow its instructions to either [link](https://github.com/AsamK/signal-cli/wiki/Linking-other-devices-(Provisioning)) or register your device. The `siggo link <phonenumber>` subcommand has been added to make linking more user-friendly, but has not been tested sufficiently. Be sure to prefix with `+` and country code (for example `+12345678901`).

When setup is finished, you should be able to run without error:

```
signal-cli -u +<yourphonenumber> receive --json
```

### Build

```
make build
```

### Run

```
bin/siggo ui
```

### Keybinds

* `i` - Insert Mode
* `I` - Compose
* `y` - Yank Mode
* `ESC` - Normal Mode
* `CTRL+L` - (insert mode) Clear input field
* `ALT+Q` - Quit

### Development

If you save the output of signal-cli like so:

```
signal-cli -u +<yourphonenumber> receive --json > example_messages.json
```
You can then run siggo using it as mock input. This is useful for development and testing.
```
bin/siggo ui -m example_messages.json
```
This way you can test without sending yourself messages.

### Roadmap

Here is a list of things that are currently broken.
* Read receipts for outgoing messages (`signal-cli` limitation)

Here is a list of features I'd like to add.
* get rid of the `ui` command and make that the default behavior of the binary
* Attachments support
* gui configuration
* let user re-sort contact list (for example alphabetically)
* command to go to next contact with message waiting
* command to go to contact with fuzzy matching
* command to yank last posted web link to make sharing between converstaions easier
* optional colors for each contact and their messages, like the electron app has
* groups support
* use dbus to send instead of signal-cli, to avoid having to spin up the JVM (might also fix the read receipt issue)
* there is still some data that i'm dropping on the floor
