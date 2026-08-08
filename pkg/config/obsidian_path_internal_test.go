package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWSLDetectionAndCandidateResolution(t *testing.T) {
	originalGOOS, originalInterop, originalExec := currentGOOS, WslInteropFile, ExecCommand
	t.Cleanup(func() {
		currentGOOS, WslInteropFile, ExecCommand = originalGOOS, originalInterop, originalExec
	})
	currentGOOS = "linux"
	WslInteropFile = filepath.Join(t.TempDir(), "WSLInterop")
	if RunningInWSL() {
		t.Fatal("missing interop marker reported WSL")
	}
	if err := os.WriteFile(WslInteropFile, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if !RunningInWSL() {
		t.Fatal("interop marker did not report WSL")
	}

	ExecCommand = func(string, ...string) ([]byte, error) {
		return []byte("D:\\Users\\andy\\AppData\\Roaming\r\n"), nil
	}
	got, err := resolveWslCandidates("unused")
	if err != nil || got != "/mnt/d/Users/andy/AppData/Roaming/obsidian/obsidian.json" {
		t.Fatalf("resolveWslCandidates = %q, %v", got, err)
	}
	ExecCommand = func(string, ...string) ([]byte, error) { return []byte("invalid"), nil }
	if _, err := resolveWslCandidates("unused"); err == nil {
		t.Fatal("invalid APPDATA should fail")
	}
	ExecCommand = func(string, ...string) ([]byte, error) { return nil, errors.New("cmd failed") }
	if _, err := resolveWslCandidates("unused"); err == nil {
		t.Fatal("cmd failure should fail")
	}
}

func TestObsidianFileLinuxFallbackWithoutConfig(t *testing.T) {
	originalGOOS, originalDir, originalInterop := currentGOOS, UserConfigDirectory, WslInteropFile
	t.Cleanup(func() {
		currentGOOS, UserConfigDirectory, WslInteropFile = originalGOOS, originalDir, originalInterop
	})
	root := t.TempDir()
	currentGOOS = "linux"
	WslInteropFile = filepath.Join(root, "missing-wsl")
	UserConfigDirectory = func() (string, error) { return filepath.Join(root, ".config"), nil }
	t.Setenv("HOME", root)
	want := filepath.Join(root, ".config", "obsidian", "obsidian.json")
	got, err := ObsidianFile()
	if err != nil || got != want {
		t.Fatalf("ObsidianFile = %q, %v; want %q", got, err, want)
	}
}
