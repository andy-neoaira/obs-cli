package obsidian_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestRenderDailyTemplateMatchesContractFixture(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "vault-contract", "v1", "daily-template")
	config, found, err := obsidian.LoadDailyNotesConfig(fixture)
	assert.NoError(t, err)
	assert.True(t, found)
	template, err := os.ReadFile(filepath.Join(fixture, "Templates", "Daily.md"))
	assert.NoError(t, err)
	location, err := time.LoadLocation("Asia/Shanghai")
	assert.NoError(t, err)
	now := time.Date(2026, 7, 24, 9, 30, 0, 0, location)

	var expected struct {
		Assertions []json.RawMessage `json:"assertions"`
	}
	expectedData, err := os.ReadFile(filepath.Join(fixture, "expected.json"))
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(expectedData, &expected))

	layout, err := obsidian.ParseMomentToGoFormat(config.Format)
	assert.NoError(t, err)
	logical := filepath.ToSlash(filepath.Join(config.Folder, now.Format(layout)+".md"))
	assert.Equal(t, "Journal/YYYY/2026-07-24.md", logical)

	content, warnings, err := obsidian.RenderDailyTemplate(string(template), now, config.Format, now.Format(layout))
	assert.NoError(t, err)
	assert.Contains(t, content, "Previous: 2026-07-23")
	assert.Contains(t, content, "Next: 2026-07-25")
	assert.Contains(t, content, "Custom: 2026/07/24")
	assert.Contains(t, content, "Unknown: {{unsupported}}")
	assert.Len(t, warnings, 1)
}

func TestRenderDailyTemplateRejectsUnsupportedToken(t *testing.T) {
	_, _, err := obsidian.RenderDailyTemplate("{{date:YYYY-Qo}}", time.Now(), "YYYY-MM-DD", "daily")
	assert.Error(t, err)
}
