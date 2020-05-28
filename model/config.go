package model

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var configFilename string = "config.yml"
var configFolder string = ".siggo"

func DefaultConfigFolder() string {
	d, _ := os.UserHomeDir()
	return filepath.Join(d, configFolder)
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigFolder(), configFilename)
}

func ConversationFolder() string {
	return filepath.Join(DefaultConfigFolder(), "conversations")
}

var DefaultConfig Config = Config{
	UserName: "self",
}

type Config struct {
	UserNumber string `yaml:"user_number"`
	UserName   string `yaml:"user_name"`
	// SaveMessages enables message saving. You will still load any (previously) saved messages
	// at startup.
	SaveMessages bool `yaml:"save_messages"`
	// doesn't do anything yet
	MaxConversationLength int               `yaml:"max_coversation_length"`
	HidePanelTitles       bool              `yaml:"hide_panel_titles"`
	ContactColors         map[string]string `yaml:"contact_colors"`
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
	return c.SaveAs(DefaultConfigPath())
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
	cfg := Config{}
	err = yaml.Unmarshal(b, &cfg)
	return &cfg, err
}

// NewConfigFile makes a new config file at `path` and returns the default config.
func NewConfigFile(path string) (*Config, error) {
	cfg := DefaultConfig
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
	return &cfg, nil
}

// GetConfig returns the current configuration from the
// default config location, creates a new one if it isn't there
func GetConfig() (*Config, error) {
	path := DefaultConfigPath()
	if _, err := os.Stat(path); err != nil {
		// config doesn't exist so lets save default
		return NewConfigFile(path)
	}
	return LoadConfig(path)
}
