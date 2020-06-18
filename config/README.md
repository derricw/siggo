# Configuring siggo

Siggo keeps a configuration `yaml` file at `~/.config/siggo/config.yml`. You can and should edit it manually, but there are some helper commands for certain things.

Most optional behavior is opt-in (disabled by default). For example, siggo does not save messages unless you turn on `save_messages`.

### Phone number config

Be sure to prefix your phone number with `+` and your country code, just like `signal-cli` requires.

### Print Current Configuration to stdout

```
siggo cfg
```

### Print Default Configuration to stdout

```
siggo cfg default
```

### Configure Contact Colors

Supports W3C and hex.

Note: Matches on contact name, so if you have two people with the same name they will get the same color.

```
siggo cfg color "Leloo Dallas" DarkViolet
siggo cfg color "Ruby Rhod" "#00FF00"
```

### Configure Contact Aliases

```
siggo cfg alias "John Smith" "Ruby Rhod"
```

