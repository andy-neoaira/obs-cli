package obsidian

import (
	"path"
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/config"
)

var (
	ObsidianConfigFile = config.ObsidianFile
	RunningInWSL       = config.RunningInWSL
)

func adjustForWslMount(directory string) string {
	if len(directory) < 2 || directory[1] != ':' {
		return directory
	}
	drive := directory[0]
	if !((drive >= 'A' && drive <= 'Z') || (drive >= 'a' && drive <= 'z')) {
		return directory
	}
	mounted := "/mnt/" + strings.ToLower(string(drive)) + directory[2:]
	return strings.ReplaceAll(mounted, "\\", "/")
}

func vaultBaseName(vaultPath string) string {
	return path.Base(strings.ReplaceAll(vaultPath, "\\", "/"))
}
