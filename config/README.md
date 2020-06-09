# Configuring siggo

Siggo keeps a configuration `yaml` file at `~/.siggo/config.yml`. You can and should edit it manually, but there are some helper commands for certain things.

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

