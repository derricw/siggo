package model

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg := &DefaultConfig
	dir := os.TempDir()
	path := filepath.Join(dir, "tempconfig.yaml")
	err := cfg.SaveAs(path)
	if err != nil {
		t.Fatalf("failed to save test config @ %s - %s\n", path, err)
	}
	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("failed to load test config @ %s - %s\n", path, err)
	}

	assert.Equal(t, loaded, cfg)
}
