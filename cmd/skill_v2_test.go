package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestSkillV2Setup(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-vault-setup/SKILL.md")

	requireSkillText(t, content,
		"vault.discover",
		"vault.list",
		"vault.add",
		"vault.set-default",
		"obs-cli vault discover",
		"obs-cli vault list",
		"obs-cli vault add",
		"obs-cli vault set-default",
		"--dry-run",
		"禁止直接写 `.obsidian/`",
		"no_change",
		"升级到提供 obs-cli/v2",
		"--request-id \"<stable-id>\"",
		"同路径已注册但名称不同",
		"结构化 argv",
	)
	forbidSkillText(t, content,
		"obs-cli list-vaults",
		"obs-cli add-vault",
		"obs-cli set-default-vault",
	)
}

func TestSkillV2Capture(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-capture/SKILL.md")

	requireSkillText(t, content,
		"note.get",
		"note.create",
		"note.append",
		"obs-cli note get",
		"obs-cli note create",
		"obs-cli note append",
		"--content-file -",
		"--dry-run",
		"--if-match",
		"REVISION_CONFLICT",
		"new_revision:",
		"create 内容与预期完全一致可返回 `no_change`",
		"append 不自动重放",
		"升级到支持对应 obs-cli/v2",
		"--request-id \"<stable-id>\"",
		"多段内容合并为一次 append",
		"写前正文保持为写后正文的原样前缀",
		"结构化 argv",
	)
	forbidSkillText(t, content,
		"obs-cli create ",
		"--overwrite",
	)
}

func TestSkillV2Daily(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-daily-log/SKILL.md")
	requireSkillText(t, content,
		"daily.get", "daily.create", "daily.append",
		"obs-cli daily get", "obs-cli daily create", "obs-cli daily append",
		"--dry-run", "--if-match", "REVISION_CONFLICT",
		"不硬编码 `Dailies/`", "禁止重复创建章节", "payload_added_once:",
	)
	forbidSkillText(t, content, "printf ", "--overwrite")
}

func TestSkillV2Project(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-project-note/SKILL.md")
	requireSkillText(t, content,
		"note.get", "note.create", "note.append", "metadata.get", "metadata.set",
		"obs-cli metadata get", "obs-cli metadata set",
		"--dry-run", "--if-match", "REVISION_CONFLICT",
		"保留未知字段", "unknown_metadata_preserved:",
	)
	forbidSkillText(t, content,
		"obs-cli create ", "obs-cli frontmatter ",
	)
}

func readSkill(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read skill %q: %v", path, err)
	}
	return string(content)
}

func requireSkillText(t *testing.T, content string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(content, value) {
			t.Errorf("skill is missing %q", value)
		}
	}
}

func forbidSkillText(t *testing.T, content string, values ...string) {
	t.Helper()
	for _, value := range values {
		if strings.Contains(content, value) {
			t.Errorf("skill contains forbidden V1 or unsafe form %q", value)
		}
	}
}
