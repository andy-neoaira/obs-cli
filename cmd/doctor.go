package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

var officialSkillNames = []string{
	"obsidian-vault-setup",
	"obsidian-capture",
	"obsidian-compare-notes",
	"obsidian-daily-log",
	"obsidian-inbox-triage",
	"obsidian-knowledge-search",
	"obsidian-knowledge-synthesis",
	"obsidian-project-note",
	"obsidian-project-status",
	"obsidian-safe-note-update",
	"obsidian-vault-audit",
}

type doctorCheck struct {
	ID      string         `json:"id"`
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type doctorData struct {
	Status         string           `json:"status"`
	CLIVersion     string           `json:"cli_version"`
	Protocol       string           `json:"protocol"`
	Platform       string           `json:"platform"`
	ConfigPath     string           `json:"config_path"`
	SkillsPath     string           `json:"skills_path"`
	Agent          string           `json:"agent"`
	Checks         []doctorCheck    `json:"checks"`
	Update         *updateCheckData `json:"update,omitempty"`
	NetworkChecked bool             `json:"network_checked"`
}

type managedSkillMetadata struct {
	ManagedBy   string `json:"managed_by"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	Skill       string `json:"skill"`
	SkillDigest string `json:"skill_digest"`
}

func newDoctorCommand(updateService updater) *cobra.Command {
	var common commonFlags
	var online bool
	var agent string
	var skillsPath string
	command := &cobra.Command{
		Use:   "doctor",
		Short: "Audit the local obs-cli, Vault registry, and managed Skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			requestID, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				requestID, _ = protocol.ResolveRequestID("")
				return renderEnvelope(cmd, "doctor.audit", requestID, func() (any, error) {
					return nil, err
				})
			}
			if common.Output != "json" {
				err := protocol.NewError(
					protocol.InvalidArgument,
					"doctor only supports JSON output",
					map[string]any{"field": "output"},
				)
				return renderEnvelope(cmd, "doctor.audit", requestID, func() (any, error) {
					return nil, err
				})
			}
			return renderEnvelope(cmd, "doctor.audit", requestID, func() (any, error) {
				return runDoctor(cmd.Context(), updateService, online, agent, skillsPath)
			})
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{Output: true, RequestID: true}, false)
	command.Flags().BoolVar(&online, "online", false, "also check the latest GitHub Release")
	command.Flags().StringVar(
		&agent,
		"agent",
		"codex",
		"skill host: codex, claude-code, opencode, cursor, or kimi-code",
	)
	command.Flags().StringVar(
		&skillsPath,
		"skills-path",
		"",
		"override the audited Agent skill directory",
	)
	return command
}

func runDoctor(
	ctx context.Context,
	updateService updater,
	online bool,
	requestedAgent string,
	requestedSkillsPath string,
) (doctorData, error) {
	configDir, configPath, configPathErr := config.ConfigPath()
	agent, skillsPath, agentErr := resolveAgentSkillsPath(requestedAgent, requestedSkillsPath)
	result := doctorData{
		Status:         "healthy",
		CLIVersion:     resolveVersion(),
		Protocol:       protocol.Version,
		Platform:       runtime.GOOS + "/" + runtime.GOARCH,
		ConfigPath:     configPath,
		SkillsPath:     skillsPath,
		Agent:          agent,
		Checks:         []doctorCheck{},
		NetworkChecked: online,
	}
	add := func(check doctorCheck) {
		result.Checks = append(result.Checks, check)
		if check.Status == "error" {
			result.Status = "error"
		} else if check.Status == "warning" && result.Status == "healthy" {
			result.Status = "warning"
		}
	}
	versionStatus := "ok"
	versionMessage := "CLI has a release version"
	if result.CLIVersion == "dev" {
		versionStatus = "warning"
		versionMessage = "development build cannot be matched to a release"
	}
	add(doctorCheck{ID: "cli.version", Status: versionStatus, Message: versionMessage})
	if agentErr != nil {
		add(doctorCheck{ID: "skills.agent", Status: "error", Message: agentErr.Error()})
	}

	executable, executableErr := os.Executable()
	if executableErr != nil {
		add(doctorCheck{ID: "cli.executable", Status: "error", Message: executableErr.Error()})
	} else {
		resolved, err := filepath.EvalSymlinks(executable)
		if err != nil {
			add(doctorCheck{ID: "cli.executable", Status: "error", Message: err.Error()})
		} else {
			add(doctorCheck{
				ID: "cli.executable", Status: "ok", Message: "executable path is resolvable",
				Details: map[string]any{"path": resolved},
			})
		}
	}

	if configPathErr != nil {
		add(doctorCheck{ID: "config.path", Status: "error", Message: configPathErr.Error()})
	} else {
		store := config.NewStore(configPath)
		cfg, err := store.Load()
		if err != nil {
			add(doctorCheck{ID: "config.registry", Status: "error", Message: err.Error()})
		} else {
			status := "ok"
			message := "Vault registry is valid"
			if len(cfg.Vaults) == 0 {
				status = "warning"
				message = "no Vault is registered"
			}
			add(doctorCheck{
				ID: "config.registry", Status: status, Message: message,
				Details: map[string]any{
					"directory": configDir, "vault_count": len(cfg.Vaults),
					"default_vault_id": cfg.DefaultVaultID,
				},
			})
			for _, vault := range config.SortedVaults(cfg) {
				info, statErr := os.Stat(vault.Path)
				check := doctorCheck{
					ID: "vault." + vault.ID, Status: "ok", Message: "registered Vault is accessible",
					Details: map[string]any{"name": vault.Name, "path": vault.Path},
				}
				if statErr != nil || !info.IsDir() {
					check.Status = "error"
					check.Message = "registered Vault is not an accessible directory"
				}
				add(check)
			}
		}
	}

	if agentErr == nil {
		auditSkills(skillsPath, result.CLIVersion, add)
	}
	if online {
		update, err := updateService.Check(ctx, result.CLIVersion)
		if err != nil {
			add(doctorCheck{ID: "update.release", Status: "warning", Message: err.Error()})
		} else {
			result.Update = &update
			status := "ok"
			message := "CLI is up to date"
			if update.UpdateAvailable {
				status = "warning"
				message = "a newer CLI release is available"
			}
			add(doctorCheck{
				ID: "update.release", Status: status, Message: message,
				Details: map[string]any{
					"current_version": update.CurrentVersion,
					"latest_version":  update.LatestVersion,
				},
			})
		}
	}
	return result, nil
}

func auditSkills(root, cliVersion string, add func(doctorCheck)) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		add(doctorCheck{
			ID: "skills.directory", Status: "warning",
			Message: "Agent Skill directory is not available",
			Details: map[string]any{"path": root},
		})
		return
	}
	add(doctorCheck{
		ID: "skills.directory", Status: "ok", Message: "Agent Skill directory is available",
		Details: map[string]any{"path": root},
	})
	for _, name := range officialSkillNames {
		skillFile := filepath.Join(root, name, "SKILL.md")
		content, err := os.ReadFile(skillFile)
		if err != nil {
			add(doctorCheck{
				ID: "skill." + name, Status: "warning", Message: "official Skill is not installed",
			})
			continue
		}
		metadataPath := filepath.Join(root, name, ".obs-cli-managed.json")
		metadataRaw, err := os.ReadFile(metadataPath)
		if err != nil {
			add(doctorCheck{
				ID: "skill." + name, Status: "warning",
				Message: "Skill is installed but has no managed upgrade metadata",
			})
			continue
		}
		var metadata managedSkillMetadata
		if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
			add(doctorCheck{
				ID: "skill." + name, Status: "error", Message: "Skill upgrade metadata is invalid",
			})
			continue
		}
		digest := sha256.Sum256(content)
		actualDigest := "sha256:" + hex.EncodeToString(digest[:])
		check := doctorCheck{
			ID: "skill." + name, Status: "ok", Message: "managed Skill is intact",
			Details: map[string]any{"version": metadata.Version},
		}
		switch {
		case metadata.ManagedBy != "obs-cli" || metadata.Skill != name:
			check.Status = "error"
			check.Message = "Skill upgrade metadata identity is invalid"
		case metadata.SkillDigest != actualDigest:
			check.Status = "warning"
			check.Message = "managed Skill has local modifications"
		case cliVersion != "dev" && metadata.Version != cliVersion:
			check.Status = "warning"
			check.Message = "Skill version does not match the CLI version"
		}
		add(check)
	}
}

func resolveAgentSkillsPath(requested, override string) (string, string, error) {
	agent := requested
	switch agent {
	case "claude":
		agent = "claude-code"
	case "kimi", "kimicode":
		agent = "kimi-code"
	}
	if override != "" {
		if !filepath.IsAbs(override) {
			return agent, "", fmt.Errorf("skills path must be absolute: %s", override)
		}
		return agent, filepath.Clean(override), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return agent, "", err
	}
	switch agent {
	case "codex":
		root := os.Getenv("CODEX_HOME")
		if root == "" {
			root = filepath.Join(home, ".codex")
		}
		return agent, filepath.Join(root, "skills"), nil
	case "claude-code":
		return agent, filepath.Join(home, ".claude", "skills"), nil
	case "opencode":
		root := os.Getenv("XDG_CONFIG_HOME")
		if root == "" {
			root = filepath.Join(home, ".config")
		}
		return agent, filepath.Join(root, "opencode", "skills"), nil
	case "cursor":
		return agent, filepath.Join(home, ".cursor", "skills"), nil
	case "kimi-code":
		root := os.Getenv("KIMI_CODE_HOME")
		if root == "" {
			root = filepath.Join(home, ".kimi-code")
		}
		return agent, filepath.Join(root, "skills"), nil
	default:
		return agent, "", fmt.Errorf("unsupported Agent %q", requested)
	}
}
