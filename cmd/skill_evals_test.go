package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/google/jsonschema-go/jsonschema"
)

type skillEvalManifest struct {
	SchemaVersion     int                  `json:"schema_version"`
	MinimumCLIVersion string               `json:"minimum_cli_version"`
	ProtocolVersion   string               `json:"protocol_version"`
	CaseKinds         []string             `json:"case_kinds"`
	Skills            []skillEvalSkill     `json:"skills"`
	CrossCutting      []skillEvalCrossCase `json:"cross_cutting"`
}

type skillEvalSkill struct {
	Name                 string          `json:"name"`
	RequiredCapabilities []string        `json:"required_capabilities"`
	WriteCapabilities    []string        `json:"write_capabilities"`
	Cases                []skillEvalCase `json:"cases"`
}

type skillEvalCase struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"`
	Prompt   string            `json:"prompt"`
	Expected skillEvalExpected `json:"expected"`
}

type skillEvalExpected struct {
	Selected        bool     `json:"selected"`
	Outcome         string   `json:"outcome"`
	Operations      []string `json:"operations"`
	Writes          []string `json:"writes"`
	VaultUnchanged  bool     `json:"vault_unchanged"`
	SilentOverwrite bool     `json:"silent_overwrite"`
	ErrorCode       string   `json:"error_code"`
}

type skillEvalCrossCase struct {
	ID         string   `json:"id"`
	Fixture    string   `json:"fixture"`
	Plan       []string `json:"plan"`
	FinalState struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	} `json:"final_state"`
}

func TestSkillEvalManifestCoverageAndContracts(t *testing.T) {
	manifest := readSkillEvalManifest(t)
	if manifest.SchemaVersion != 1 || manifest.ProtocolVersion != protocol.Version {
		t.Fatalf("unsupported eval manifest contract: %#v", manifest)
	}
	expectedKinds := []string{"conflict", "failure", "non_trigger", "success", "trigger"}
	gotKinds := append([]string(nil), manifest.CaseKinds...)
	sort.Strings(gotKinds)
	if fmt.Sprint(gotKinds) != fmt.Sprint(expectedKinds) {
		t.Fatalf("case kinds = %v, want %v", gotKinds, expectedKinds)
	}

	capabilities := currentCapabilities()
	available := make(map[string]struct{}, len(capabilities.Operations))
	mutating := make(map[string]struct{})
	for _, operation := range capabilities.Operations {
		available[operation.Name] = struct{}{}
		if operation.Mutating {
			mutating[operation.Name] = struct{}{}
		}
	}
	if !skillEvalVersionCompatible(capabilities.CLIVersion, manifest.MinimumCLIVersion) {
		t.Fatalf("CLI %q does not meet Skill minimum %q",
			capabilities.CLIVersion, manifest.MinimumCLIVersion)
	}

	manifestSkills := make([]string, 0, len(manifest.Skills))
	manifestNames := make(map[string]struct{}, len(manifest.Skills))
	caseIDs := make(map[string]struct{})
	for _, skill := range manifest.Skills {
		if _, exists := manifestNames[skill.Name]; exists {
			t.Errorf("duplicate Skill manifest name %q", skill.Name)
		}
		manifestNames[skill.Name] = struct{}{}
		manifestSkills = append(manifestSkills, skill.Name)
		skillText := readSkill(t, filepath.Join("../skills", skill.Name, "SKILL.md"))
		namePattern := regexp.MustCompile(`(?m)^name: ([a-z0-9][a-z0-9-]*)$`)
		nameMatch := namePattern.FindStringSubmatch(skillText)
		if len(nameMatch) != 2 || nameMatch[1] != skill.Name {
			t.Errorf("%s frontmatter name = %q, want directory/manifest name %q",
				skill.Name, nameMatch, skill.Name)
		}
		required := make(map[string]struct{}, len(skill.RequiredCapabilities))
		for _, capability := range skill.RequiredCapabilities {
			if _, ok := available[capability]; !ok {
				t.Errorf("%s requires unavailable capability %q", skill.Name, capability)
			}
			if !strings.Contains(skillText, capability) {
				t.Errorf("%s does not declare manifest capability %q", skill.Name, capability)
			}
			required[capability] = struct{}{}
		}
		for _, capability := range skill.WriteCapabilities {
			if _, ok := required[capability]; !ok {
				t.Errorf("%s write capability %q is not required", skill.Name, capability)
			}
			if _, ok := mutating[capability]; !ok {
				t.Errorf("%s marks read-only capability %q as mutating", skill.Name, capability)
			}
		}
		writeCapabilities := make(map[string]struct{}, len(skill.WriteCapabilities))
		for _, capability := range skill.WriteCapabilities {
			writeCapabilities[capability] = struct{}{}
		}
		for capability := range required {
			if _, isMutating := mutating[capability]; isMutating {
				if _, declared := writeCapabilities[capability]; !declared {
					t.Errorf("%s omits mutating capability %q from write_capabilities",
						skill.Name, capability)
				}
			}
		}

		seenKinds := make(map[string]int)
		for _, scenario := range skill.Cases {
			if scenario.ID == "" || scenario.Prompt == "" {
				t.Errorf("%s contains an unnamed or promptless case: %#v", skill.Name, scenario)
			}
			if _, exists := caseIDs[scenario.ID]; exists {
				t.Errorf("duplicate eval case id %q", scenario.ID)
			}
			caseIDs[scenario.ID] = struct{}{}
			seenKinds[scenario.Kind]++
			validateSkillEvalCase(t, skill.Name, scenario, available, required, writeCapabilities)
		}
		for _, kind := range manifest.CaseKinds {
			if seenKinds[kind] != 1 {
				t.Errorf("%s has %d %s cases, want exactly one", skill.Name, seenKinds[kind], kind)
			}
		}
	}
	sort.Strings(manifestSkills)
	if want := skillDirectories(t); fmt.Sprint(manifestSkills) != fmt.Sprint(want) {
		t.Fatalf("eval Skill coverage = %v, want %v", manifestSkills, want)
	}
	if len(manifest.CrossCutting) < 3 {
		t.Fatalf("cross-cutting evals = %d, want at least 3", len(manifest.CrossCutting))
	}
}

func TestSkillEvalManifestSchema(t *testing.T) {
	schemaBytes, err := os.ReadFile("../skills/evals/scenarios.schema.json")
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
	manifestBytes, err := os.ReadFile("../skills/evals/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if err := resolved.Validate(manifest); err != nil {
		t.Fatal(err)
	}
}

func validateSkillEvalCase(
	t *testing.T,
	skillName string,
	scenario skillEvalCase,
	available map[string]struct{},
	required map[string]struct{},
	writeCapabilities map[string]struct{},
) {
	t.Helper()
	for _, operation := range scenario.Expected.Operations {
		if _, ok := available[operation]; !ok {
			t.Errorf("%s/%s plans unavailable operation %q", skillName, scenario.ID, operation)
		}
		if _, ok := required[operation]; !ok && operation != "capabilities.get" {
			t.Errorf("%s/%s plans undeclared operation %q", skillName, scenario.ID, operation)
		}
	}
	switch scenario.Kind {
	case "trigger":
		if !scenario.Expected.Selected || len(scenario.Expected.Writes) != 0 ||
			!scenario.Expected.VaultUnchanged || scenario.Expected.Outcome != "planned" {
			t.Errorf("%s/%s trigger must select and only plan", skillName, scenario.ID)
		}
		for _, operation := range scenario.Expected.Operations {
			if _, mutating := writeCapabilities[operation]; mutating {
				t.Errorf("%s/%s trigger plan includes mutating operation %q", skillName, scenario.ID, operation)
			}
		}
	case "non_trigger":
		if scenario.Expected.Selected || len(scenario.Expected.Operations) != 0 ||
			len(scenario.Expected.Writes) != 0 || !scenario.Expected.VaultUnchanged ||
			scenario.Expected.Outcome != "not_selected" {
			t.Errorf("%s/%s non-trigger could mutate or execute CLI", skillName, scenario.ID)
		}
	case "success":
		if !scenario.Expected.Selected || scenario.Expected.Outcome != "success" {
			t.Errorf("%s/%s success contract is incomplete", skillName, scenario.ID)
		}
		if len(writeCapabilities) == 0 {
			if len(scenario.Expected.Writes) != 0 || !scenario.Expected.VaultUnchanged {
				t.Errorf("%s/%s read-only success claims a mutation", skillName, scenario.ID)
			}
		} else {
			plannedMutation := false
			for _, operation := range scenario.Expected.Operations {
				_, plannedMutation = writeCapabilities[operation]
				if plannedMutation {
					break
				}
			}
			if !plannedMutation || len(scenario.Expected.Writes) == 0 {
				t.Errorf("%s/%s mutating success lacks operation/write summary", skillName, scenario.ID)
			}
		}
	case "conflict":
		if !scenario.Expected.Selected ||
			(scenario.Expected.Outcome != "conflict" && scenario.Expected.Outcome != "stale") ||
			len(scenario.Expected.Writes) != 0 || !scenario.Expected.VaultUnchanged ||
			scenario.Expected.SilentOverwrite {
			t.Errorf("%s/%s conflict could silently overwrite", skillName, scenario.ID)
		}
	case "failure":
		validOutcome := scenario.Expected.Outcome == "failed" ||
			scenario.Expected.Outcome == "partial" ||
			scenario.Expected.Outcome == "insufficient_evidence" ||
			scenario.Expected.Outcome == "authorization_required" ||
			scenario.Expected.Outcome == "scope_required"
		if !scenario.Expected.Selected || !validOutcome || len(scenario.Expected.Writes) != 0 ||
			!scenario.Expected.VaultUnchanged {
			t.Errorf("%s/%s failure contract is incomplete", skillName, scenario.ID)
		}
		if scenario.Expected.Outcome == "failed" {
			if _, ok := skillEvalErrorCodes[scenario.Expected.ErrorCode]; !ok {
				t.Errorf("%s/%s uses undeclared error code %q", skillName, scenario.ID,
					scenario.Expected.ErrorCode)
			}
		} else if scenario.Expected.ErrorCode != "" {
			t.Errorf("%s/%s result status %q must not invent an error code", skillName,
				scenario.ID, scenario.Expected.Outcome)
		}
	default:
		t.Errorf("%s/%s has unknown kind %q", skillName, scenario.ID, scenario.Kind)
	}
}

func TestSkillEvalTemporaryVaultSafety(t *testing.T) {
	manifest := readSkillEvalManifest(t)
	crossCases := make(map[string]skillEvalCrossCase, len(manifest.CrossCutting))
	for _, scenario := range manifest.CrossCutting {
		crossCases[scenario.ID] = scenario
	}
	dangerous := crossCases["dangerous-unicode-multiline"]
	if dangerous.Fixture == "" || dangerous.FinalState.Content != "exact_fixture_bytes" {
		t.Fatalf("dangerous content eval plan is incomplete: %#v", dangerous)
	}
	rawPayload, err := os.ReadFile(filepath.Join("../skills/evals", dangerous.Fixture))
	if err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(t.TempDir(), "shell-side-effect")
	payload := []byte(strings.ReplaceAll(string(rawPayload), "__SENTINEL__", sentinel))
	if !strings.Contains(string(payload), "$(touch ") || !strings.Contains(string(payload), "🧪") ||
		strings.Count(string(payload), "\n") < 8 {
		t.Fatal("dangerous Unicode fixture lost required coverage")
	}

	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	beforeDryRun := directoryDigest(t, vaultRoot)
	dryRun := executeNoteCommand(t, registryFactory, serviceFactory, string(payload),
		"create", dangerous.FinalState.Path, "--content-file", "-", "--dry-run")
	if !dryRun.OK || directoryDigest(t, vaultRoot) != beforeDryRun {
		t.Fatal("dry-run changed the temporary Vault")
	}
	applied := executeNoteCommand(t, registryFactory, serviceFactory, string(payload),
		"create", dangerous.FinalState.Path, "--content-file", "-")
	if !applied.OK {
		t.Fatalf("dangerous content apply = %#v", applied)
	}
	assertFileContent(t, filepath.Join(vaultRoot, dangerous.FinalState.Path), string(payload))
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatalf("dangerous Markdown executed shell side effect at %s", sentinel)
	}
	beforeRead := directoryDigest(t, vaultRoot)
	if got := getSkillEvalNote(t, registryFactory, serviceFactory, dangerous.FinalState.Path); got.Content != string(payload) {
		t.Fatalf("read-only eval changed content: %#v", got)
	}
	if afterRead := directoryDigest(t, vaultRoot); afterRead != beforeRead {
		t.Fatal("read-only note.get changed the temporary Vault")
	}

	baseCase := crossCases["stale-revision-preserves-vault"]
	base, err := os.ReadFile(filepath.Join("../skills/evals", baseCase.Fixture))
	if err != nil {
		t.Fatal(err)
	}
	executeNoteCommand(t, registryFactory, serviceFactory, string(base),
		"create", baseCase.FinalState.Path, "--content-file", "-")
	current := getSkillEvalNote(t, registryFactory, serviceFactory, baseCase.FinalState.Path)
	external := string(base) + "\nexternal change\n"
	if err := os.WriteFile(filepath.Join(vaultRoot, baseCase.FinalState.Path), []byte(external), 0o644); err != nil {
		t.Fatal(err)
	}
	beforeConflict := directoryDigest(t, vaultRoot)
	conflict, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "agent append\n",
		"append", baseCase.FinalState.Path, "--content-file", "-", "--if-match", current.Revision)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("stale eval = %#v err=%v", conflict, err)
	}
	if after := directoryDigest(t, vaultRoot); after != beforeConflict {
		t.Fatal("revision conflict changed the temporary Vault")
	}
	assertFileContent(t, filepath.Join(vaultRoot, baseCase.FinalState.Path), external)

	executeNoteCommand(t, registryFactory, serviceFactory, "same\nsame\n",
		"create", "Eval/Ambiguous.md", "--content-file", "-")
	ambiguous := getSkillEvalNote(t, registryFactory, serviceFactory, "Eval/Ambiguous.md")
	matchFile := filepath.Join(t.TempDir(), "match.md")
	if err := os.WriteFile(matchFile, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	beforeAmbiguous := directoryDigest(t, vaultRoot)
	failed, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "changed",
		"patch", "Eval/Ambiguous.md", "--match-file", matchFile, "--content-file", "-",
		"--if-match", ambiguous.Revision)
	if err == nil || failed.Error == nil || failed.Error.Code != protocol.AmbiguousNote {
		t.Fatalf("ambiguous eval = %#v err=%v", failed, err)
	}
	if afterAmbiguous := directoryDigest(t, vaultRoot); afterAmbiguous != beforeAmbiguous {
		t.Fatal("ambiguous patch failure changed the temporary Vault")
	}
}

func TestSkillEvalCompatibilityFailureIsExplicit(t *testing.T) {
	manifest := readSkillEvalManifest(t)
	if skillEvalVersionCompatible("v0.9.9", manifest.MinimumCLIVersion) {
		t.Fatalf("obsolete CLI unexpectedly meets %s", manifest.MinimumCLIVersion)
	}
	if !skillEvalVersionCompatible("v1.0.0-rc.1", manifest.MinimumCLIVersion) ||
		!skillEvalVersionCompatible("v1.0.0", manifest.MinimumCLIVersion) ||
		!skillEvalVersionCompatible("dev", manifest.MinimumCLIVersion) {
		t.Fatal("compatible release or local development build was rejected")
	}
	if requested := os.Getenv("SKILL_EVAL_CLI_VERSION"); requested != "" &&
		!skillEvalVersionCompatible(requested, manifest.MinimumCLIVersion) {
		t.Fatalf("release CLI %q does not meet Skill minimum %q",
			requested, manifest.MinimumCLIVersion)
	}
	response, _, err := executeCapabilities(t, "--require", "eval.unsupported")
	if err == nil || response.Error == nil || response.Error.Code != protocol.CapabilityUnsupported {
		t.Fatalf("missing capability did not fail explicitly: %#v err=%v", response, err)
	}
}

func TestSkillEvalCompatibilityMatrixAndReleaseReport(t *testing.T) {
	manifest := readSkillEvalManifest(t)
	matrixBytes, err := os.ReadFile("../skills/evals/compatibility-matrix.md")
	if err != nil {
		t.Fatal(err)
	}
	matrix := string(matrixBytes)
	for _, skill := range manifest.Skills {
		var row string
		for _, line := range strings.Split(matrix, "\n") {
			if strings.HasPrefix(line, "| `"+skill.Name+"` |") {
				row = line
				break
			}
		}
		if row == "" {
			t.Errorf("compatibility matrix is missing %s", skill.Name)
		}
		for _, capability := range skill.RequiredCapabilities {
			if !strings.Contains(row, "`"+capability+"`") {
				t.Errorf("compatibility matrix row %s is missing capability %s", skill.Name, capability)
			}
		}
	}
	report, err := os.ReadFile("../skills/evals/release-report.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		"55 / 55", "确定性评测", "模型主观质量评测", "run-skill-evals.sh",
		manifest.MinimumCLIVersion, manifest.ProtocolVersion,
	} {
		if !strings.Contains(string(report), marker) {
			t.Errorf("release report is missing %q", marker)
		}
	}
}

func readSkillEvalManifest(t *testing.T) skillEvalManifest {
	t.Helper()
	raw, err := os.ReadFile("../skills/evals/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest skillEvalManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func skillDirectories(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("../skills")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "_template" && entry.Name() != "evals" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}

func getSkillEvalNote(
	t *testing.T,
	registryFactory vaultRegistryFactory,
	serviceFactory noteServiceFactory,
	path string,
) noteops.Note {
	t.Helper()
	response := executeNoteCommand(t, registryFactory, serviceFactory, "", "get", path)
	var data struct {
		Note noteops.Note `json:"note"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data.Note
}

var skillEvalSemverPattern = regexp.MustCompile(
	`^v([0-9]+)\.([0-9]+)\.([0-9]+)(?:-rc\.([0-9]+))?$`,
)

var skillEvalErrorCodes = map[string]struct{}{
	string(protocol.InvalidArgument):       {},
	string(protocol.InvalidFrontmatter):    {},
	string(protocol.VaultNotFound):         {},
	string(protocol.NoteNotFound):          {},
	string(protocol.AlreadyExists):         {},
	string(protocol.AmbiguousNote):         {},
	string(protocol.RevisionConflict):      {},
	string(protocol.PathOutsideVault):      {},
	string(protocol.PartialFailure):        {},
	string(protocol.CapabilityUnsupported): {},
	string(protocol.InternalError):         {},
}

func skillEvalVersionCompatible(current, minimum string) bool {
	if current == "dev" {
		return true
	}
	currentParts := skillEvalSemverPattern.FindStringSubmatch(current)
	minimumParts := skillEvalSemverPattern.FindStringSubmatch(minimum)
	if currentParts == nil || minimumParts == nil {
		return false
	}
	for index := 1; index <= 3; index++ {
		currentValue, _ := strconv.Atoi(currentParts[index])
		minimumValue, _ := strconv.Atoi(minimumParts[index])
		if currentValue != minimumValue {
			return currentValue > minimumValue
		}
	}
	currentRC, minimumRC := currentParts[4], minimumParts[4]
	if currentRC == "" {
		return true
	}
	if minimumRC == "" {
		return false
	}
	currentValue, _ := strconv.Atoi(currentRC)
	minimumValue, _ := strconv.Atoi(minimumRC)
	return currentValue >= minimumValue
}
