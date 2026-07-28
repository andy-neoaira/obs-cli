package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
	"github.com/google/jsonschema-go/jsonschema"
)

func TestSkillProjectStatus(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-project-status/SKILL.md")
	requireSkillText(t, content,
		"note.get", "daily.get", "search.content", "note.append",
		"period_id", "当前 period heading 已存在且完整 block bytes/digest 等于计划时返回 `no_change`",
		"--dry-run", "--if-match", "REVISION_CONFLICT",
		"progress", "risks", "decisions", "next_steps",
	)
	forbidSkillText(t, content, "obs-cli replace ")
}

func TestSkillSafeUpdate(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-safe-note-update/SKILL.md")
	requireSkillText(t, content,
		"note.get", "note.patch", "obs-cli note patch",
		"--match-file", "--content-file", "--dry-run", "--if-match",
		"REVISION_CONFLICT", "three_way", "abandon",
		"planned revision_after", "expected bytes", "frontmatter", "UNKNOWN_OUTCOME",
		"不能等于整篇 base", "unrelated_bytes_preserved",
	)
	forbidSkillText(t, content,
		"obs-cli replace ", "obs-cli delete ",
	)
}

func TestProjectStatusSchemaGolden(t *testing.T) {
	schemaBytes, err := os.ReadFile("../docs/spec/schema/project-status-report-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatal(err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"project-status-report-v1.json",
		"project-status-applied-v1.json",
		"project-status-no-change-v1.json",
		"project-status-conflict-v1.json",
	} {
		t.Run(name, func(t *testing.T) {
			report := readProjectStatusGolden(t, name)
			if err := resolved.Validate(report); err != nil {
				t.Fatal(err)
			}
			validateProjectStatusSemantics(t, report)
		})
	}

	invalidPeriod := readProjectStatusGolden(t, "project-status-report-v1.json")
	invalidPeriod["period_id"] = "2026-W00"
	if err := resolved.Validate(invalidPeriod); err == nil {
		t.Fatal("project status schema accepted invalid ISO week W00")
	}

	invalidWriteback := readProjectStatusGolden(t, "project-status-report-v1.json")
	invalidWriteback["writeback"].(map[string]any)["status"] = "applied"
	if err := resolved.Validate(invalidWriteback); err == nil {
		t.Fatal("project status schema accepted applied writeback without request")
	}
}

func readProjectStatusGolden(t *testing.T, name string) map[string]any {
	t.Helper()
	goldenBytes, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(goldenBytes, &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func validateProjectStatusSemantics(t *testing.T, report map[string]any) {
	t.Helper()
	sources := make(map[string]map[string]any)
	for _, raw := range report["sources"].([]any) {
		source := raw.(map[string]any)
		id := source["id"].(string)
		if _, exists := sources[id]; exists {
			t.Fatalf("duplicate project status source id %q", id)
		}
		sources[id] = source
	}
	evidenceByID := make(map[string]map[string]any)
	for _, raw := range report["evidence"].([]any) {
		evidence := raw.(map[string]any)
		id := evidence["id"].(string)
		if _, exists := evidenceByID[id]; exists {
			t.Fatalf("duplicate project status evidence id %q", id)
		}
		source, ok := sources[evidence["source_id"].(string)]
		if !ok || evidence["path"] != source["path"] || evidence["revision"] != source["revision"] {
			t.Fatalf("project status evidence is not traceable: source=%#v evidence=%#v", source, evidence)
		}
		evidenceByID[id] = evidence
	}
	categories := map[string]string{
		"progress": "progress", "risks": "risk",
		"decisions": "decision", "next_steps": "next_step",
	}
	itemIDs := make(map[string]struct{})
	for field, category := range categories {
		for _, raw := range report[field].([]any) {
			item := raw.(map[string]any)
			id := item["id"].(string)
			if _, exists := itemIDs[id]; exists {
				t.Fatalf("duplicate project status item id %q", id)
			}
			itemIDs[id] = struct{}{}
			if item["category"] != category {
				t.Fatalf("%s item has category %q", field, item["category"])
			}
			for _, rawEvidenceID := range item["evidence_ids"].([]any) {
				evidenceID := rawEvidenceID.(string)
				if _, ok := evidenceByID[evidenceID]; !ok {
					t.Fatalf("item %q references missing evidence %q", id, evidenceID)
				}
			}
		}
	}
	if rawBaseline := report["baseline"]; rawBaseline != nil {
		baseline := rawBaseline.(map[string]any)
		if baseline["period_id"].(string) >= report["period_id"].(string) {
			t.Fatalf("baseline period %q is not earlier than %q", baseline["period_id"], report["period_id"])
		}
		evidence, ok := evidenceByID[baseline["evidence_id"].(string)]
		source := sources[evidence["source_id"].(string)]
		if !ok || source["role"] != "project" ||
			evidence["path"] != baseline["path"] || evidence["revision"] != baseline["revision"] {
			t.Fatalf("baseline is not traceable to project evidence: baseline=%#v evidence=%#v", baseline, evidence)
		}
	}
	writeback := report["writeback"].(map[string]any)
	if writeback["status"] == "applied" || writeback["status"] == "no_change" {
		plan := report["writeback_plan"].(map[string]any)
		verified := report["verified"].(map[string]any)
		if writeback["revision_after"] != plan["planned_revision_after"] ||
			verified["period_heading_count"] != float64(1) ||
			verified["payload_count"] != float64(1) ||
			verified["sources_current"] != true {
			t.Fatalf("applied writeback is not fully verified: writeback=%#v plan=%#v verified=%#v",
				writeback, plan, verified)
		}
	}
}

func TestSkillProjectStatusDoesNotDuplicatePeriod(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	initial := "# Project\n\n## Status\n\n### 2026-W30\nold\n\n## Other\nkeep\n"
	executeNoteCommand(t, registryFactory, serviceFactory, initial,
		"create", "Projects/Demo", "--content-file", "-")
	note := getNoteForInboxTest(t, registryFactory, serviceFactory, "Projects/Demo")
	if strings.Count(note.Content, "### 2026-W30") != 1 {
		t.Fatalf("period baseline = %q", note.Content)
	}
	existingBlock := projectPeriodBlock(note.Content, "2026-W30")
	expectedExisting := "### 2026-W30\nold\n\n"
	status := "conflict"
	if existingBlock == expectedExisting {
		status = "no_change"
	}
	if status != "no_change" {
		t.Fatalf("duplicate period status = %q", status)
	}
	if projectPeriodBlock(note.Content, "2026-W30") == "### 2026-W30\nnew\n\n" {
		status = "no_change"
	} else {
		status = "conflict"
	}
	if status != "conflict" {
		t.Fatalf("different same-period payload status = %q", status)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Projects", "Demo.md"), initial)

	payload := "### 2026-W31\n- Progress\n"
	dry := executeNoteCommand(t, registryFactory, serviceFactory, payload,
		"append", "Projects/Demo", "--section", "Status", "--content-file", "-",
		"--if-match", note.Revision, "--dry-run")
	if !dry.OK {
		t.Fatalf("status dry-run = %#v", dry)
	}
	var dryData struct {
		Vault struct {
			ID string `json:"id"`
		} `json:"vault"`
		Plan protocol.Plan `json:"plan"`
	}
	if err := json.Unmarshal(dry.Data, &dryData); err != nil {
		t.Fatal(err)
	}
	planned, _ := dryData.Plan.Changes[0].Details["revision_after"].(string)
	applied := executeNoteCommand(t, registryFactory, serviceFactory, payload,
		"append", "Projects/Demo", "--section", "Status", "--content-file", "-",
		"--if-match", note.Revision)
	var data struct {
		Note noteops.Mutation `json:"note"`
	}
	if err := json.Unmarshal(applied.Data, &data); err != nil {
		t.Fatal(err)
	}
	after := getNoteForInboxTest(t, registryFactory, serviceFactory, "Projects/Demo")
	if planned == "" || planned != data.Note.RevisionAfter || data.Note.RevisionAfter != after.Revision ||
		strings.Count(after.Content, "### 2026-W31") != 1 ||
		strings.Count(after.Content, "- Progress") != 1 ||
		!strings.Contains(after.Content, "## Other\nkeep\n") {
		t.Fatalf("status verification = %#v content=%q", data.Note, after.Content)
	}
	beforeRepeat := after.Content
	if projectPeriodBlock(after.Content, "2026-W31") == payload {
		status = "no_change"
	} else {
		status = "conflict"
	}
	if status != "no_change" ||
		getNoteForInboxTest(t, registryFactory, serviceFactory, "Projects/Demo").Content != beforeRepeat {
		t.Fatal("repeated project status changed the note")
	}
}

func projectPeriodBlock(content, period string) string {
	statusHeading := "\n## Status\n"
	if strings.Count(content, statusHeading) != 1 {
		return ""
	}
	statusStart := strings.Index(content, statusHeading) + len(statusHeading)
	statusEnd := len(content)
	if index := strings.Index(content[statusStart:], "\n## "); index >= 0 {
		statusEnd = statusStart + index + 1
	}
	status := content[statusStart:statusEnd]
	heading := "### " + period
	if strings.Count(status, heading) != 1 {
		return ""
	}
	start := strings.Index(status, heading)
	rest := status[start+len(heading):]
	end := len(rest)
	if index := strings.Index(rest, "\n### "); index >= 0 {
		end = index + 1
	}
	return status[start : start+len(heading)+end]
}

func TestSkillSafeUpdatePlanApplyVerifyAndConflict(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "prefix\nold value\nsuffix\n",
		"create", "Notes/Demo", "--content-file", "-")
	base := getNoteForInboxTest(t, registryFactory, serviceFactory, "Notes/Demo")
	matchFile := filepath.Join(t.TempDir(), "match.md")
	if err := os.WriteFile(matchFile, []byte("old value"), 0o600); err != nil {
		t.Fatal(err)
	}
	replacementFile := filepath.Join(t.TempDir(), "replacement.md")
	if err := os.WriteFile(replacementFile, []byte("new value"), 0o600); err != nil {
		t.Fatal(err)
	}
	dry := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"patch", "Notes/Demo", "--match-file", matchFile, "--content-file", replacementFile,
		"--if-match", base.Revision, "--dry-run")
	var dryData struct {
		Vault struct {
			ID string `json:"id"`
		} `json:"vault"`
		Plan protocol.Plan `json:"plan"`
	}
	if err := json.Unmarshal(dry.Data, &dryData); err != nil {
		t.Fatal(err)
	}
	planned, _ := dryData.Plan.Changes[0].Details["revision_after"].(string)
	matchBytes, _ := os.ReadFile(matchFile)
	replacementBytes, _ := os.ReadFile(replacementFile)
	planBytes, err := json.Marshal(struct {
		VaultID           string `json:"vault_id"`
		Target            string `json:"target"`
		BaseRevision      string `json:"base_revision"`
		MatchDigest       string `json:"match_digest"`
		ReplacementDigest string `json:"replacement_digest"`
		ExpectedDigest    string `json:"expected_digest"`
		PlannedRevision   string `json:"planned_revision_after"`
	}{
		VaultID: dryData.Vault.ID, Target: "Notes/Demo.md", BaseRevision: base.Revision,
		MatchDigest: storage.Revision(matchBytes), ReplacementDigest: storage.Revision(replacementBytes),
		ExpectedDigest: planned, PlannedRevision: planned,
	})
	if err != nil {
		t.Fatal(err)
	}
	planDigest := strings.TrimPrefix(storage.Revision(planBytes), "sha256:")
	applyRequestID := "safe-update-" + planDigest[:24]
	applied := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"patch", "Notes/Demo", "--match-file", matchFile, "--content-file", replacementFile,
		"--if-match", base.Revision, "--request-id", applyRequestID)
	var applyData struct {
		Note noteops.Mutation `json:"note"`
	}
	if err := json.Unmarshal(applied.Data, &applyData); err != nil {
		t.Fatal(err)
	}
	verified := getNoteForInboxTest(t, registryFactory, serviceFactory, "Notes/Demo")
	if planned == "" || planned != applyData.Note.RevisionAfter || planned != verified.Revision ||
		verified.Content != "prefix\nnew value\nsuffix\n" {
		t.Fatalf("safe update mismatch planned=%q apply=%#v verify=%#v", planned, applyData.Note, verified)
	}

	stale := verified
	if err := os.WriteFile(filepath.Join(vaultRoot, "Notes", "Demo.md"),
		[]byte("external prefix\nold value\nsuffix\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(replacementFile, []byte("another"), 0o600); err != nil {
		t.Fatal(err)
	}
	conflict, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "",
		"patch", "Notes/Demo", "--match-file", matchFile, "--content-file", replacementFile,
		"--if-match", stale.Revision)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("safe update conflict = %#v err=%v", conflict, err)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Notes", "Demo.md"), "external prefix\nold value\nsuffix\n")
}

func TestSkillSafeUpdateRejectsChangedControlledInput(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "prefix\nold\nsuffix\n",
		"create", "Notes/Guard", "--content-file", "-")
	base := getNoteForInboxTest(t, registryFactory, serviceFactory, "Notes/Guard")
	matchFile := filepath.Join(t.TempDir(), "match.md")
	replacementFile := filepath.Join(t.TempDir(), "replacement.md")
	if err := os.WriteFile(matchFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(replacementFile, []byte("approved"), 0o600); err != nil {
		t.Fatal(err)
	}
	approvedDigest := storage.Revision([]byte("approved"))
	dry := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"patch", "Notes/Guard", "--match-file", matchFile, "--content-file", replacementFile,
		"--if-match", base.Revision, "--dry-run")
	if !dry.OK {
		t.Fatalf("guard dry-run = %#v", dry)
	}
	if err := os.WriteFile(replacementFile, []byte("swapped"), 0o600); err != nil {
		t.Fatal(err)
	}
	currentBytes, _ := os.ReadFile(replacementFile)
	if storage.Revision(currentBytes) == approvedDigest {
		t.Fatal("controlled input replacement was not detected")
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Notes", "Guard.md"), "prefix\nold\nsuffix\n")
}
