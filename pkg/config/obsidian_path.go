package config

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// WslInteropFile 是 WSL 环境的标志性文件路径，用于检测是否在 WSL 中运行。
var (
	WslInteropFile = "/proc/sys/fs/binfmt_misc/WSLInterop"
	currentGOOS    = runtime.GOOS
)

// ExecCommand 是包级变量，指向 exec.Command(...).Output() 的调用。
// 使用变量方便在测试中替换为 Mock 实现。
var ExecCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// RunningInWSL 检测当前是否在 Windows Subsystem for Linux (WSL) 环境中运行。
func RunningInWSL() bool {
	if currentGOOS != "linux" {
		return false
	}

	// WSL 环境中存在 /proc/sys/fs/binfmt_misc/WSLInterop 文件
	_, err := os.Stat(WslInteropFile)
	return err == nil
}

// ObsidianFile 返回 Obsidian 官方配置文件（obsidian.json）的完整路径。
// 它会尝试多个候选路径，以支持不同安装方式（原生、Flatpak、Snap、WSL 等）。
func ObsidianFile() (obsidianConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}

	defaultPath := filepath.Join(userConfigDir, ObsidianConfigDirectory, ObsidianConfigFile)
	// 默认路径存在则直接返回
	if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
		return defaultPath, nil
	}

	// 非 Linux 系统没有更多候选路径
	if currentGOOS != "linux" {
		return "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}

	// WSL 环境：需要转换 Windows 的 APPDATA 路径
	if RunningInWSL() {
		return resolveWslCandidates(defaultPath)
	}

	// Linux 原生环境：检查 Flatpak 和 Snap 等特殊安装路径
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}

	var candidatePaths []string
	candidatePaths = append(candidatePaths, defaultPath)
	candidatePaths = append(candidatePaths,
		filepath.Join(homeDir, ".var", "app", "md.obsidian.Obsidian", "config", "obsidian", ObsidianConfigFile))
	candidatePaths = append(candidatePaths,
		filepath.Join(homeDir, "snap", "obsidian", "current", ".config", "obsidian", ObsidianConfigFile))

	// Snap 安装时可能有编号子目录（如 ~/snap/obsidian/x1/.config/obsidian/）
	snapNumberedPattern := filepath.Join(homeDir, "snap", "obsidian", "*", ".config", "obsidian", ObsidianConfigFile)
	if matches, globErr := filepath.Glob(snapNumberedPattern); globErr == nil {
		candidatePaths = append(candidatePaths, matches...)
	}

	// 遍历所有候选路径，返回第一个存在的路径
	var firstNonExistErr error
	for _, path := range candidatePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			return path, nil
		} else if !os.IsNotExist(statErr) && firstNonExistErr == nil {
			firstNonExistErr = statErr
		}
	}

	if firstNonExistErr != nil {
		return "", firstNonExistErr
	}

	return defaultPath, nil
}

// resolveWslCandidates 在 WSL 环境中解析 Obsidian 配置文件路径。
// 通过调用 Windows 的 cmd.exe 获取 %APPDATA% 环境变量，再转换为 WSL 的 /mnt/ 路径。
func resolveWslCandidates(defaultPath string) (string, error) {
	// 不能直接使用 os.UserHomeDir，因为 WSL 的 Linux 用户名可能与 Windows 用户名不同
	out, err := ExecCommand("cmd.exe", "/c", "echo %APPDATA%")
	if err != nil {
		log.Print("Failed to extract user APPDATA location. Assuming non-WSL install.")
		return "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}

	// 去除输出中的空白字符和换行
	appDataPath := strings.TrimSpace(string(out))
	if len(appDataPath) > 1 && appDataPath[1] == ':' {
		driveLetter := strings.ToLower(string(appDataPath[0]))
		restPath := appDataPath[3:] // 跳过 "C:\"
		// 将反斜杠替换为正斜杠
		restPath = strings.ReplaceAll(restPath, "\\", "/")
		wslPath := filepath.Join("/mnt", driveLetter, restPath, ObsidianConfigDirectory, ObsidianConfigFile)
		return wslPath, nil
	}

	return "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
}
