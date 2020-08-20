package model

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	configFilename   string = "config.yml"
	configFolderName string = "siggo"
	dataFolderName   string = "siggo"
)

// FindConfigFolder returns $XDG_CONFIG_HOME/siggo/ if it exists, otherwise returns $HOME/.config/siggo/
func FindConfigFolder() string {
	XDGConfig := os.Getenv("XDG_CONFIG_HOME")
	if XDGConfig != "" {
		return filepath.Join(XDGConfig, configFolderName)
	}
	d, _ := os.UserHomeDir()
	return filepath.Join(d, ".config", configFolderName)
}

// ConfigPath returns the config file path
func ConfigPath() string {
	return filepath.Join(FindConfigFolder(), configFilename)
}

// FindDataFolder returns $XDG_DATA_HOME if it exists, otherwise returns $HOME/.local/share/siggo/
func FindDataFolder() string {
	XDGData := os.Getenv("XDG_DATA_HOME")
	if XDGData != "" {
		return filepath.Join(XDGData, dataFolderName)
	}
	d, _ := os.UserHomeDir()
	return filepath.Join(d, ".local", "share", dataFolderName)
}

// ConversationFolder returns the folder where conversations are saved
func ConversationFolder() string {
	return filepath.Join(FindDataFolder(), "conversations")
}

// LogPath returns the log file path
func LogPath() string {
	return filepath.Join(FindDataFolder(), "siggo.log")
}

func DefaultConfig() *Config {
	return &Config{
		UserName:       "self",
		ContactColors:  make(map[string]string),
		ContactAliases: make(map[string]string),
	}
}

// Config includes both siggo and UI config
type Config struct {
	UserNumber string `yaml:"user_number"`
	UserName   string `yaml:"user_name"`
	// SaveMessages enables message saving. You will still load any (previously) saved messages
	// at startup.
	SaveMessages bool `yaml:"save_messages"`
	// Attempt to send desktop notifications
	DesktopNotifications            bool `yaml:"desktop_notifications"`
	DesktopNotificationsShowMessage bool `yaml:"desktop_notifications_show_message"`
	DesktopNotificationsShowAvatar  bool `yaml:"desktop_notifications_show_avatar"`
	// Terminal bell
	TerminalBellNotifications bool `yaml:"terminal_bell_notifications"`
	// doesn't do anything yet
	MaxConversationLength int               `yaml:"max_coversation_length"`
	HidePanelTitles       bool              `yaml:"hide_panel_titles"`
	HidePhoneNumbers      bool              `yaml:"hide_phone_numbers"`
	ContactColors         map[string]string `yaml:"contact_colors"`
	ContactAliases        map[string]string `yaml:"contact_aliases"`

	// No rotation provided, use at your own risk!
	LogFilePath string `yaml:"log_file"`
}

// SaveAs writes the config to `path`
func (c *Config) SaveAs(path string) error {
	d, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(d)
	return err
}

// Save saves the config to the default location
func (c *Config) Save() error {
	return c.SaveAs(ConfigPath())
}

// Print pretty-prints the configuration
func (c *Config) Print() {
	b, _ := yaml.Marshal(c)
	fmt.Printf("%s", string(b))
}

// LoadConfig loads the config located @ `path`
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := DefaultConfig()
	err = yaml.Unmarshal(b, &cfg)
	return cfg, err
}

// NewConfigFile makes a new config file at `path` and returns the default config.
func NewConfigFile(path string) (*Config, error) {
	cfg := DefaultConfig()
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = cfg.SaveAs(path)
	if err != nil {
		log.Printf("failed to save config @ %s", path)
	}
	log.Printf("default config saved @ %s", path)
	return cfg, nil
}

// GetConfig returns the current configuration from the
// default config location, creates a new one if it isn't there
func GetConfig() (*Config, error) {
	path := ConfigPath()
	if _, err := os.Stat(path); err != nil {
		// config doesn't exist so lets save default
		return NewConfigFile(path)
	}
	return LoadConfig(path)
}
