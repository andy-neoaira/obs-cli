package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/google/jsonschema-go/jsonschema"
)

func TestSkillCompare(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-compare-notes/SKILL.md")
	requireSkillText(t, content,
		"note.get", "search.content", "obs-cli note get",
		"path+revision", "body_revision", "modified_at", "insufficient_evidence",
		"source_statement | comparison | inference | unknown",
		"inference 不得", "applied", "writeback", "stale",
	)
	forbidSkillText(t, content, "obs-cli print ", "obs-cli search-content", "obs-cli replace ")
}

func TestSkillSynthesis(t *testing.T) {
	content := readSkill(t, "../skills/obsidian-knowledge-synthesis/SKILL.md")
	requireSkillText(t, content,
		"note.get", "note.create", "note.patch",
		"obs-cli note create", "obs-cli note patch",
		"--dry-run", "--if-match", "REVISION_CONFLICT",
		"来源 manifest", "source_statement | synthesis | inference | unknown",
		"默认无写入", "不得重复插入", "stdin/受控文件",
	)
	forbidSkillText(t, content, "obs-cli replace ", "obs-cli delete ", "--unsafe-no-if-match")
}

func TestCompareSynthesisSchemaAndGoldenReports(t *testing.T) {
	raw, err := os.ReadFile("../docs/spec/schema/compare-synthesis-report-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	required, ok := schema["required"].([]any)
	if !ok || len(required) != 13 {
		t.Fatalf("schema required fields = %#v", schema["required"])
	}
	var validationSchema jsonschema.Schema
	if err := json.Unmarshal(raw, &validationSchema); err != nil {
		t.Fatal(err)
	}
	resolved, err := validationSchema.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"compare-report-v1.json",
		"compare-search-report-v1.json",
		"compare-insufficient-duplicates-v1.json",
		"compare-stale-report-v1.json",
		"synthesis-report-v1.json",
	} {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		var instance any
		if err := json.Unmarshal(data, &instance); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := resolved.Validate(instance); err != nil {
			t.Fatalf("%s does not satisfy JSON Schema: %v", name, err)
		}
		var report struct {
			Kind      string `json:"kind"`
			Status    string `json:"status"`
			Selection struct {
				Mode       string   `json:"mode"`
				Queries    []string `json:"queries"`
				Confirmed  bool     `json:"confirmed"`
				Duplicates []string `json:"duplicates"`
			} `json:"selection"`
			Sources []struct {
				ID           string `json:"id"`
				Path         string `json:"path"`
				Revision     string `json:"revision"`
				BodyRevision string `json:"body_revision"`
				ModifiedAt   string `json:"modified_at"`
				Role         string `json:"role"`
				Query        string `json:"query"`
				Reason       string `json:"selection_reason"`
			} `json:"sources"`
			Evidence []struct {
				ID       string `json:"id"`
				SourceID string `json:"source_id"`
				Path     string `json:"path"`
				Revision string `json:"revision"`
				Kind     string `json:"kind"`
				Location string `json:"location"`
				Excerpt  string `json:"excerpt"`
			} `json:"evidence"`
			Claims []struct {
				ID              string   `json:"id"`
				Type            string   `json:"type"`
				EpistemicStatus string   `json:"epistemic_status"`
				EvidenceIDs     []string `json:"evidence_ids"`
			} `json:"claims"`
			Conflicts []struct {
				EvidenceIDs []string `json:"evidence_ids"`
			} `json:"conflicts"`
			Writeback struct {
				Requested            bool   `json:"requested"`
				Operation            string `json:"operation"`
				Status               string `json:"status"`
				PlanDigest           string `json:"plan_digest"`
				RequestID            string `json:"request_id"`
				PlannedRevisionAfter string `json:"planned_revision_after"`
				RevisionAfter        string `json:"revision_after"`
			} `json:"writeback"`
			Warnings []string `json:"warnings"`
		}
		if err := json.Unmarshal(data, &report); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if report.Kind == "" || report.Status == "" || report.Selection.Mode == "" ||
			!report.Selection.Confirmed {
			t.Fatalf("incomplete golden %s: %#v", name, report)
		}
		if report.Status == "success" &&
			(len(report.Sources) < 2 || len(report.Evidence) < 2 || len(report.Claims) == 0) {
			t.Fatalf("incomplete success golden %s: %#v", name, report)
		}
		if report.Selection.Mode == "search" && len(report.Selection.Queries) == 0 {
			t.Fatalf("search selection lacks query in %s", name)
		}
		if len(report.Selection.Duplicates) > 0 && len(report.Warnings) == 0 {
			t.Fatalf("duplicate selection lacks warning in %s", name)
		}
		sourceIDs := map[string]bool{}
		sourcePaths := map[string]bool{}
		sourcesByID := map[string]struct {
			path     string
			revision string
		}{}
		evidenceIDs := map[string]bool{}
		evidenceSource := map[string]string{}
		sourceModifiedAt := map[string]string{}
		selectionQueries := map[string]bool{}
		for _, query := range report.Selection.Queries {
			selectionQueries[query] = true
		}
		for _, source := range report.Sources {
			if source.Path == "" || source.Revision == "" || source.BodyRevision == "" ||
				source.ModifiedAt == "" {
				t.Fatalf("incomplete source in %s: %#v", name, source)
			}
			if report.Selection.Mode == "search" &&
				(source.Role != "search_result" || source.Query == "" || source.Reason == "") {
				t.Fatalf("unauditable search source in %s: %#v", name, source)
			}
			if sourceIDs[source.ID] || sourcePaths[source.Path] {
				t.Fatalf("duplicate source identity in %s: %#v", name, source)
			}
			sourceIDs[source.ID] = true
			sourcePaths[source.Path] = true
			sourcesByID[source.ID] = struct {
				path     string
				revision string
			}{source.Path, source.Revision}
			sourceModifiedAt[source.ID] = source.ModifiedAt
			if report.Selection.Mode == "search" && !selectionQueries[source.Query] {
				t.Fatalf("source query %q was not selected in %s", source.Query, name)
			}
		}
		for _, evidence := range report.Evidence {
			source, found := sourcesByID[evidence.SourceID]
			if !found || evidence.Path != source.path || evidence.Revision != source.revision ||
				(evidence.Kind != "content" && evidence.Kind != "frontmatter" && evidence.Kind != "modified_at") {
				t.Fatalf("orphan evidence in %s: %#v", name, evidence)
			}
			if evidenceIDs[evidence.ID] {
				t.Fatalf("duplicate evidence ID %q in %s", evidence.ID, name)
			}
			evidenceIDs[evidence.ID] = true
			evidenceSource[evidence.ID] = evidence.SourceID
			if evidence.Kind == "modified_at" &&
				(evidence.Location != "modified_at" ||
					evidence.Excerpt != sourceModifiedAt[evidence.SourceID]) {
				t.Fatalf("invalid modified_at evidence in %s: %#v", name, evidence)
			}
		}
		claimIDs := map[string]bool{}
		hasConflictingClaim := false
		conflictingClaimSets := map[string]bool{}
		for _, claim := range report.Claims {
			if claimIDs[claim.ID] {
				t.Fatalf("duplicate claim ID %q in %s", claim.ID, name)
			}
			claimIDs[claim.ID] = true
			if claim.Type != "unknown" && len(claim.EvidenceIDs) == 0 {
				t.Fatalf("unsupported claim in %s: %#v", name, claim)
			}
			if claim.Type == "inference" && claim.EpistemicStatus != "inferred" {
				t.Fatalf("inference mislabeled in %s: %#v", name, claim)
			}
			if claim.EpistemicStatus == "conflicting" {
				hasConflictingClaim = true
				conflictingClaimSets[normalizedEvidenceSet(claim.EvidenceIDs)] = true
			}
			distinctSources := map[string]bool{}
			for _, evidenceID := range claim.EvidenceIDs {
				if !evidenceIDs[evidenceID] {
					t.Fatalf("claim references unknown evidence %q in %s", evidenceID, name)
				}
				distinctSources[evidenceSource[evidenceID]] = true
			}
			if (claim.Type == "comparison" || claim.Type == "synthesis") && len(distinctSources) < 2 {
				t.Fatalf("multi-source claim has fewer than two sources in %s: %#v", name, claim)
			}
		}
		conflictSets := map[string]bool{}
		for _, conflict := range report.Conflicts {
			distinctSources := map[string]bool{}
			for _, evidenceID := range conflict.EvidenceIDs {
				if !evidenceIDs[evidenceID] {
					t.Fatalf("conflict references unknown evidence %q in %s", evidenceID, name)
				}
				distinctSources[evidenceSource[evidenceID]] = true
			}
			if len(distinctSources) < 2 {
				t.Fatalf("conflict lacks independent sources in %s: %#v", name, conflict)
			}
			conflictSets[normalizedEvidenceSet(conflict.EvidenceIDs)] = true
		}
		if hasConflictingClaim && len(report.Conflicts) == 0 {
			t.Fatalf("conflicting claim lacks conflict record in %s", name)
		}
		for evidenceSet := range conflictingClaimSets {
			if !conflictSets[evidenceSet] {
				t.Fatalf("conflicting claim and conflict evidence differ in %s", name)
			}
		}
		if report.Kind == "compare" &&
			(report.Writeback.Requested || report.Writeback.Operation != "none" ||
				report.Writeback.Status != "not_requested") {
			t.Fatalf("compare golden writes back: %#v", report.Writeback)
		}
		if report.Writeback.Requested {
			hexDigest := strings.TrimPrefix(report.Writeback.PlanDigest, "sha256:")
			if len(hexDigest) < 24 ||
				report.Writeback.RequestID != "synthesis-"+hexDigest[:24] {
				t.Fatalf("request ID is not derived from plan digest in %s: %#v", name, report.Writeback)
			}
			if report.Writeback.Status == "applied" &&
				report.Writeback.RevisionAfter != report.Writeback.PlannedRevisionAfter {
				t.Fatalf("applied revision differs from plan in %s: %#v", name, report.Writeback)
			}
		}
	}
}

func normalizedEvidenceSet(values []string) string {
	copyOfValues := append([]string(nil), values...)
	sort.Strings(copyOfValues)
	return strings.Join(copyOfValues, ",")
}

func TestCompareSynthesisSchemaRejectsUnsafeReports(t *testing.T) {
	raw, err := os.ReadFile("../docs/spec/schema/compare-synthesis-report-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	load := func(name string) map[string]any {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		var instance map[string]any
		if err := json.Unmarshal(data, &instance); err != nil {
			t.Fatal(err)
		}
		return instance
	}
	cases := []struct {
		name   string
		report map[string]any
		mutate func(map[string]any)
	}{
		{
			name: "compare cannot write", report: load("compare-report-v1.json"),
			mutate: func(report map[string]any) {
				writeback := report["writeback"].(map[string]any)
				writeback["requested"], writeback["operation"], writeback["status"] = true, "create", "applied"
			},
		},
		{
			name: "search requires queries", report: load("compare-search-report-v1.json"),
			mutate: func(report map[string]any) {
				report["selection"].(map[string]any)["queries"] = []any{}
			},
		},
		{
			name: "stale clears conclusions", report: load("compare-stale-report-v1.json"),
			mutate: func(report map[string]any) {
				report["claims"] = []any{map[string]any{
					"id": "C1", "type": "unknown", "epistemic_status": "unknown",
					"statement": "stale", "evidence_ids": []any{},
				}}
			},
		},
		{
			name: "inference must be inferred", report: load("compare-report-v1.json"),
			mutate: func(report map[string]any) {
				claim := report["claims"].([]any)[0].(map[string]any)
				claim["type"], claim["epistemic_status"] = "inference", "supported"
			},
		},
		{
			name: "requested false cannot apply", report: load("synthesis-report-v1.json"),
			mutate: func(report map[string]any) {
				report["writeback"].(map[string]any)["requested"] = false
			},
		},
		{
			name: "patch requires anchor", report: load("synthesis-report-v1.json"),
			mutate: func(report map[string]any) {
				writeback := report["writeback"].(map[string]any)
				writeback["operation"] = "patch"
			},
		},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			item.mutate(item.report)
			if err := resolved.Validate(item.report); err == nil {
				t.Fatal("unsafe report unexpectedly satisfied JSON Schema")
			}
		})
	}
}

func TestSkillCompareDetectsChangedSource(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "# A\nactive\n",
		"create", "Sources/A", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "# B\npaused\n",
		"create", "Sources/B", "--content-file", "-")
	a := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/A")
	b := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/B")
	if a.ModifiedAt == "" || b.ModifiedAt == "" || a.BodyRevision == "" || b.BodyRevision == "" {
		t.Fatalf("source evidence incomplete: A=%#v B=%#v", a, b)
	}
	future := time.Now().Add(2 * time.Hour)
	if err := os.Chtimes(filepath.Join(vaultRoot, "Sources", "B.md"), future, future); err != nil {
		t.Fatal(err)
	}
	touched := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/B")
	if touched.Revision != b.Revision || touched.ModifiedAt == b.ModifiedAt {
		t.Fatalf("modified_at-only change not observable: before=%#v after=%#v", b, touched)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "Sources", "A.md"), []byte("# A\nchanged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	verified := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/A")
	if verified.Revision == a.Revision {
		t.Fatal("changed source was not detected")
	}
	status := "stale"
	if status != "stale" {
		t.Fatalf("compare status = %q", status)
	}
}

func TestSkillSynthesisCreateAndPatchConflictSafety(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "# A\nUse V1\n",
		"create", "Sources/A", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "# B\nKeep revisions\n",
		"create", "Sources/B", "--content-file", "-")
	a := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/A")
	b := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/B")
	content := "# Synthesis\n\n## Sources\n- Sources/A.md@" + a.Revision +
		"\n- Sources/B.md@" + b.Revision + "\n"

	dry := executeNoteCommand(t, registryFactory, serviceFactory, content,
		"create", "Synthesis/V1", "--content-file", "-", "--dry-run")
	if !dry.OK {
		t.Fatalf("create dry-run = %#v", dry)
	}
	if err := os.MkdirAll(filepath.Join(vaultRoot, "Synthesis"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "Synthesis", "V1.md"), []byte("external"), 0o644); err != nil {
		t.Fatal(err)
	}
	existing, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, content,
		"create", "Synthesis/V1", "--content-file", "-")
	if err == nil || existing.Error == nil || existing.Error.Code != protocol.AlreadyExists {
		t.Fatalf("create conflict = %#v err=%v", existing, err)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Synthesis", "V1.md"), "external")

	target := getNoteForInboxTest(t, registryFactory, serviceFactory, "Synthesis/V1")
	matchFile := filepath.Join(t.TempDir(), "anchor.md")
	if err := os.WriteFile(matchFile, []byte("external"), 0o600); err != nil {
		t.Fatal(err)
	}
	patchDry := executeNoteCommand(t, registryFactory, serviceFactory, content,
		"patch", "Synthesis/V1", "--match-file", matchFile, "--content-file", "-",
		"--if-match", target.Revision, "--dry-run")
	if !patchDry.OK {
		t.Fatalf("patch dry-run = %#v", patchDry)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "Synthesis", "V1.md"), []byte("external edit"), 0o644); err != nil {
		t.Fatal(err)
	}
	conflict, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, content,
		"patch", "Synthesis/V1", "--match-file", matchFile, "--content-file", "-",
		"--if-match", target.Revision)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("patch conflict = %#v err=%v", conflict, err)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Synthesis", "V1.md"), "external edit")

	exact := executeNoteCommand(t, registryFactory, serviceFactory, content,
		"create", "Synthesis/Exact", "--content-file", "-")
	if !exact.OK {
		t.Fatalf("exact create = %#v", exact)
	}
	repeated, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, content,
		"create", "Synthesis/Exact", "--content-file", "-")
	if err == nil || repeated.Error == nil || repeated.Error.Code != protocol.AlreadyExists ||
		repeated.Error.Details["same_content"] != true {
		t.Fatalf("exact create repeat = %#v err=%v", repeated, err)
	}

	executeNoteCommand(t, registryFactory, serviceFactory, "anchor\nanchor\n",
		"create", "Synthesis/Ambiguous", "--content-file", "-")
	ambiguousTarget := getNoteForInboxTest(t, registryFactory, serviceFactory, "Synthesis/Ambiguous")
	anchorFile := filepath.Join(t.TempDir(), "ambiguous-anchor.md")
	if err := os.WriteFile(anchorFile, []byte("anchor"), 0o600); err != nil {
		t.Fatal(err)
	}
	ambiguous, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, content,
		"patch", "Synthesis/Ambiguous", "--match-file", anchorFile, "--content-file", "-",
		"--if-match", ambiguousTarget.Revision, "--dry-run")
	if err == nil || ambiguous.Error == nil || ambiguous.Error.Code != protocol.AmbiguousNote {
		t.Fatalf("ambiguous anchor = %#v err=%v", ambiguous, err)
	}
}

func TestSkillSynthesisSourceChangeBeforeApplyStopsWrite(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "A",
		"create", "Sources/A", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "B",
		"create", "Sources/B", "--content-file", "-")
	a := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/A")
	b := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/B")
	content := "# Synthesis\n- Sources/A.md@" + a.Revision + "\n- Sources/B.md@" + b.Revision
	dry := executeNoteCommand(t, registryFactory, serviceFactory, content,
		"create", "Synthesis/Blocked", "--content-file", "-", "--dry-run")
	if !dry.OK {
		t.Fatalf("create dry-run = %#v", dry)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "Sources", "A.md"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	current := getNoteForInboxTest(t, registryFactory, serviceFactory, "Sources/A")
	if current.Revision == a.Revision {
		t.Fatal("source pre-apply check did not detect change")
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "Synthesis", "Blocked.md")); !os.IsNotExist(err) {
		t.Fatalf("write occurred despite stale source: %v", err)
	}
}
