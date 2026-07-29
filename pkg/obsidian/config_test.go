package obsidian_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

func TestLoadDailyNotesConfig(t *testing.T) {
	root := t.TempDir()

	config, found, err := obsidian.LoadDailyNotesConfig(root)
	if err != nil || found || config != (obsidian.DailyNotesConfig{}) {
		t.Fatalf("missing config = %#v found=%v err=%v", config, found, err)
	}

	obsidianDir := filepath.Join(root, ".obsidian")
	if err := os.MkdirAll(obsidianDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(obsidianDir, "daily-notes.json")
	if err := os.WriteFile(configPath, []byte(`{
		"folder": "Daily",
		"format": "YYYY-MM-DD",
		"template": "Templates/Daily"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	config, found, err = obsidian.LoadDailyNotesConfig(root)
	if err != nil || !found {
		t.Fatalf("valid config found=%v err=%v", found, err)
	}
	if config.Folder != "Daily" || config.Format != "YYYY-MM-DD" ||
		config.Template != "Templates/Daily" {
		t.Fatalf("valid config = %#v", config)
	}

	if err := os.WriteFile(configPath, []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, found, err = obsidian.LoadDailyNotesConfig(root)
	var configErr *obsidian.ConfigFileError
	if !found || !errors.As(err, &configErr) ||
		configErr.File != ".obsidian/daily-notes.json" || configErr.Kind != "parse" {
		t.Fatalf("malformed config found=%v error=%#v", found, err)
	}
}
