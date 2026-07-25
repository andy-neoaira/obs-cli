package actions_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andy-neoaira/obs-cli/mocks"
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestDailyNote(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	t.Run("Creates daily note in vault root with defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(tmpDir, today+".md"))
	})

	t.Run("Creates daily note in configured folder", func(t *testing.T) {
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		if err := os.MkdirAll(obsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(`{
			"folder": "Daily"
		}`), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(tmpDir, "Daily", today+".md"))
	})

	t.Run("Uses template when configured", func(t *testing.T) {
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		if err := os.MkdirAll(obsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(`{
			"template": "Templates/Daily"
		}`), 0644); err != nil {
			t.Fatal(err)
		}

		// Create template file
		if err := os.MkdirAll(filepath.Join(tmpDir, "Templates"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "Templates", "Daily.md"), []byte("# Daily Note\n- [ ] Task"), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)

		content, _ := os.ReadFile(filepath.Join(tmpDir, today+".md"))
		assert.Equal(t, "# Daily Note\n- [ ] Task", string(content))
	})

	t.Run("Renders shared template variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Templates"), 0755))
		assert.NoError(t, os.MkdirAll(obsDir, 0755))
		assert.NoError(t, os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(`{
			"folder": "Journal/YYYY",
			"format": "YYYY-MM-DD",
			"template": "Templates/Daily"
		}`), 0644))
		assert.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, "Templates", "Daily.md"),
			[]byte("{{date}} {{yesterday}} {{tomorrow}} {{date:YYYY/MM/DD}} {{unsupported}}"),
			0644,
		))
		location, err := time.LoadLocation("Asia/Shanghai")
		assert.NoError(t, err)
		now := time.Date(2026, 7, 24, 9, 30, 0, 0, location)

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		err = actions.DailyNote(&vault, &mocks.MockUriManager{}, actions.DailyParams{Now: now})
		assert.NoError(t, err)
		data, err := os.ReadFile(filepath.Join(tmpDir, "Journal", "YYYY", "2026-07-24.md"))
		assert.NoError(t, err)
		assert.Equal(t, "2026-07-24 2026-07-23 2026-07-25 2026/07/24 {{unsupported}}", string(data))
	})

	t.Run("Fails when configured template is missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".obsidian"), 0755))
		assert.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, ".obsidian", "daily-notes.json"),
			[]byte(`{"template":"Templates/Missing"}`),
			0644,
		))
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		err := actions.DailyNote(&vault, &mocks.MockUriManager{}, actions.DailyParams{})
		assert.Error(t, err)
	})

	t.Run("Fails when configured template basename is ambiguous", func(t *testing.T) {
		tmpDir := t.TempDir()
		assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".obsidian"), 0755))
		assert.NoError(t, os.WriteFile(
			filepath.Join(tmpDir, ".obsidian", "daily-notes.json"),
			[]byte(`{"template":"Daily"}`),
			0644,
		))
		for _, folder := range []string{"Templates/A", "Templates/B"} {
			assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, folder), 0755))
			assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, folder, "Daily.md"), []byte(folder), 0644))
		}
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		err := actions.DailyNote(&vault, &mocks.MockUriManager{}, actions.DailyParams{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ambiguous note id")
	})

	t.Run("Rejects template path outside vault", func(t *testing.T) {
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		if err := os.MkdirAll(obsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(`{
			"template": "../outside"
		}`), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.ErrorIs(t, err, obsidian.ErrPathTraversal)
	})

	t.Run("Does not overwrite existing daily note", func(t *testing.T) {
		tmpDir := t.TempDir()
		notePath := filepath.Join(tmpDir, today+".md")
		if err := os.WriteFile(notePath, []byte("existing content"), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)

		content, _ := os.ReadFile(notePath)
		assert.Equal(t, "existing content", string(content))
	})

	t.Run("Opens in Obsidian by default", func(t *testing.T) {
		tmpDir := t.TempDir()
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)
		assert.Equal(t, "myVault", uri.LastParams["vault"])
		assert.Equal(t, today, uri.LastParams["file"])
	})

	t.Run("Opens in editor when UseEditor is true", func(t *testing.T) {
		tmpDir := t.TempDir()
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		originalEditor := os.Getenv("EDITOR")
		defer os.Setenv("EDITOR", originalEditor) //nolint:errcheck
		if err := os.Setenv("EDITOR", "true"); err != nil {
			t.Fatal(err)
		}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{UseEditor: true})
		assert.NoError(t, err)
	})

	t.Run("vault.DefaultName returns an error", func(t *testing.T) {
		vaultDefaultNameErr := errors.New("Failed to get vault name")
		vaultOp := &mocks.MockVaultOperator{
			DefaultNameErr: vaultDefaultNameErr,
		}
		err := actions.DailyNote(vaultOp, &mocks.MockUriManager{}, actions.DailyParams{})
		assert.Error(t, err, vaultDefaultNameErr)
	})

	t.Run("vault.Path returns an error", func(t *testing.T) {
		vault := mocks.MockVaultOperator{
			Name:      "myVault",
			PathError: errors.New("path error"),
		}
		err := actions.DailyNote(&vault, &mocks.MockUriManager{}, actions.DailyParams{})
		assert.Equal(t, vault.PathError, err)
	})

	t.Run("uri.Execute returns an error", func(t *testing.T) {
		tmpDir := t.TempDir()
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{
			ExecuteErr: errors.New("Failed to execute URI"),
		}
		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.Equal(t, uri.ExecuteErr, err)
	})

	t.Run("Creates daily note with content", func(t *testing.T) {
		tmpDir := t.TempDir()
		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{
			Content: "hello world",
		})
		assert.NoError(t, err)

		data, _ := os.ReadFile(filepath.Join(tmpDir, today+".md"))
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("Creates daily note with template and content", func(t *testing.T) {
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		if err := os.MkdirAll(obsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(`{
			"template": "Templates/Daily"
		}`), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(tmpDir, "Templates"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "Templates", "Daily.md"), []byte("# Daily Note\n"), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{
			Content: "- task 1",
		})
		assert.NoError(t, err)

		data, _ := os.ReadFile(filepath.Join(tmpDir, today+".md"))
		assert.Equal(t, "# Daily Note\n- task 1", string(data))
	})

	t.Run("Appends content to existing daily note", func(t *testing.T) {
		tmpDir := t.TempDir()
		notePath := filepath.Join(tmpDir, today+".md")
		if err := os.WriteFile(notePath, []byte("existing content"), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{
			Content: "\nnew line",
		})
		assert.NoError(t, err)

		data, _ := os.ReadFile(notePath)
		assert.Equal(t, "existing content\nnew line", string(data))
	})

	t.Run("Does not modify existing daily note without content", func(t *testing.T) {
		tmpDir := t.TempDir()
		notePath := filepath.Join(tmpDir, today+".md")
		if err := os.WriteFile(notePath, []byte("existing content"), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)

		data, _ := os.ReadFile(notePath)
		assert.Equal(t, "existing content", string(data))
	})

	t.Run("Creates daily note with custom format", func(t *testing.T) {
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		if err := os.MkdirAll(obsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(`{
			"format": "DD-MM-YYYY"
		}`), 0644); err != nil {
			t.Fatal(err)
		}

		vault := mocks.MockVaultOperator{Name: "myVault", PathValue: tmpDir}
		uri := mocks.MockUriManager{}

		err := actions.DailyNote(&vault, &uri, actions.DailyParams{})
		assert.NoError(t, err)

		expectedName := time.Now().Format(obsidian.MomentToGoFormat("DD-MM-YYYY"))
		assert.FileExists(t, filepath.Join(tmpDir, expectedName+".md"))
	})
}
