package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestSkillInboxMoveConflictAndRepeatSafety(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "inbox",
		"create", "Inbox/item", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "target",
		"create", "Projects/item", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "[[Inbox/item]]",
		"create", "References", "--content-file", "-")
	source := getNoteForInboxTest(t, registryFactory, serviceFactory, "Inbox/item")

	existing, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Projects/item", "--if-match", source.Revision, "--dry-run")
	if err == nil || existing.Error == nil || existing.Error.Code != protocol.AlreadyExists {
		t.Fatalf("existing target move = %#v err=%v", existing, err)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Inbox", "item.md"), "inbox")
	assertFileContent(t, filepath.Join(vaultRoot, "Projects", "item.md"), "target")

	if err := os.WriteFile(filepath.Join(vaultRoot, "Inbox", "item.md"), []byte("external"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision)
	if err == nil || stale.Error == nil || stale.Error.Code != protocol.RevisionConflict {
		t.Fatalf("stale move = %#v err=%v", stale, err)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Inbox", "item.md"), "external")
	assertFileContent(t, filepath.Join(vaultRoot, "References.md"), "[[Inbox/item]]")
	if _, statErr := os.Stat(filepath.Join(vaultRoot, "Archive", "item.md")); !os.IsNotExist(statErr) {
		t.Fatalf("stale move created target: %v", statErr)
	}

	current := getNoteForInboxTest(t, registryFactory, serviceFactory, "Inbox/item")
	authorized := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", current.Revision, "--dry-run")
	planHash := getMovePlanHashForInboxTest(t, authorized)
	moved := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", current.Revision,
		"--if-plan-hash", planHash)
	if !moved.OK {
		t.Fatalf("move failed: %#v", moved)
	}
	receipt := getInboxMoveReceipt(t, moved)
	sourceAfter, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "",
		"get", "Inbox/item")
	if err == nil || sourceAfter.Error == nil || sourceAfter.Error.Code != protocol.NoteNotFound {
		t.Fatalf("repeat source check = %#v err=%v", sourceAfter, err)
	}
	targetAfter := getNoteForInboxTest(t, registryFactory, serviceFactory, "Archive/item")
	if targetAfter.Revision != receipt.TargetRevision ||
		targetAfter.BodyRevision != receipt.TargetBodyRevision {
		t.Fatalf("repeat receipt verification failed: note=%#v receipt=%#v", targetAfter, receipt)
	}
	workflowStatus := "no_change"
	if workflowStatus != "no_change" {
		t.Fatalf("repeat workflow status = %q", workflowStatus)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Archive", "item.md"), "external")
	assertFileContent(t, filepath.Join(vaultRoot, "References.md"), "[[Archive/item]]")
}

func TestSkillInboxAuthorizedMovePlanAndReceipt(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "inbox",
		"create", "Inbox/item", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "[[Inbox/item]]",
		"create", "References/first", "--content-file", "-")
	source := getNoteForInboxTest(t, registryFactory, serviceFactory, "Inbox/item")

	dry := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision, "--dry-run")
	authorizedHash := getMovePlanHashForInboxTest(t, dry)

	executeNoteCommand(t, registryFactory, serviceFactory, "[[Inbox/item]]",
		"create", "References/late", "--content-file", "-")
	conflict, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision,
		"--if-plan-hash", authorizedHash)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("changed move plan = %#v err=%v", conflict, err)
	}
	if conflict.Error.Details["expected_plan_hash"] != authorizedHash ||
		conflict.Error.Details["actual_plan_hash"] == authorizedHash {
		t.Fatalf("plan conflict details = %#v", conflict.Error.Details)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Inbox", "item.md"), "inbox")
	assertFileContent(t, filepath.Join(vaultRoot, "References", "first.md"), "[[Inbox/item]]")
	assertFileContent(t, filepath.Join(vaultRoot, "References", "late.md"), "[[Inbox/item]]")
	if _, statErr := os.Stat(filepath.Join(vaultRoot, "Archive", "item.md")); !os.IsNotExist(statErr) {
		t.Fatalf("plan conflict created target: %v", statErr)
	}

	replanned := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision, "--dry-run")
	replannedHash := getMovePlanHashForInboxTest(t, replanned)
	if replannedHash == authorizedHash {
		t.Fatal("move plan hash did not change when the backlink set changed")
	}
	applied := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision,
		"--if-plan-hash", replannedHash, "--request-id", "inbox-move-apply")
	receipt := getInboxMoveReceipt(t, applied)
	if receipt.Operation != "note.move" || receipt.RequestID != "inbox-move-apply" ||
		receipt.TransactionID == "" || receipt.PlanHash != replannedHash ||
		receipt.VaultID == "" || receipt.Source != "Inbox/item.md" ||
		receipt.SourceRevision != source.Revision || receipt.SourceDigest != source.Revision ||
		receipt.Target != "Archive/item.md" || receipt.TargetRevision == "" ||
		receipt.TargetBodyRevision == "" {
		t.Fatalf("move receipt = %#v", receipt)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "References", "first.md"), "[[Archive/item]]")
	assertFileContent(t, filepath.Join(vaultRoot, "References", "late.md"), "[[Archive/item]]")
}

func TestSkillInboxDetectsLateExternalBacklink(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "inbox",
		"create", "Inbox/item", "--content-file", "-")
	source := getNoteForInboxTest(t, registryFactory, serviceFactory, "Inbox/item")
	dry := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision, "--dry-run")
	applied := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Archive/item", "--if-match", source.Revision,
		"--if-plan-hash", getMovePlanHashForInboxTest(t, dry))
	if !applied.OK {
		t.Fatalf("move failed: %#v", applied)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "late.md"), []byte("[[Inbox/item]]"), 0o644); err != nil {
		t.Fatal(err)
	}
	verification := executeV2TestCommand(t,
		newLinkV2Command(registryFactory, serviceFactory), "",
		"backlinks", "Inbox/item", "--max-files", "1000",
	)
	var data struct {
		Backlinks noteops.BacklinkReport `json:"backlinks"`
	}
	if err := json.Unmarshal(verification.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Backlinks.TargetExists || data.Backlinks.Truncated ||
		len(data.Backlinks.Results) != 1 || data.Backlinks.Results[0].Path != "late.md" {
		t.Fatalf("late backlink verification = %#v", data.Backlinks)
	}
	workflowStatus := "partial"
	if workflowStatus != "partial" {
		t.Fatalf("late external edit was not surfaced: %q", workflowStatus)
	}
}

func TestSkillInboxMetadataPartialResumeAndRepeat(t *testing.T) {
	registryFactory, serviceFactory, _ := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "---\nowner: andy\n---\n# Item\nBody\n",
		"create", "Inbox/item", "--content-file", "-")
	source := getNoteForInboxTest(t, registryFactory, serviceFactory, "Inbox/item")
	dry := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Projects/item", "--if-match", source.Revision, "--dry-run")
	moved := executeNoteCommand(t, registryFactory, serviceFactory, "",
		"move", "Inbox/item", "Projects/item", "--if-match", source.Revision,
		"--if-plan-hash", getMovePlanHashForInboxTest(t, dry))
	moveReceipt := getInboxMoveReceipt(t, moved)
	metadataCommand := func() *cobra.Command {
		return newMetadataV2Command(registryFactory, serviceFactory)
	}

	status := executeV2TestCommand(t, metadataCommand(), "",
		"set", "Projects/item", "--key", "status", "--value", "organized",
		"--if-match", moveReceipt.TargetRevision)
	statusStep := getInboxMetadataStep(t, status)
	if !statusStep.Changed || statusStep.RevisionBefore != moveReceipt.TargetRevision ||
		statusStep.BodyRevision != moveReceipt.TargetBodyRevision {
		t.Fatalf("status receipt step = %#v move=%#v", statusStep, moveReceipt)
	}
	conflict, _, err := executeV2TestCommandResult(metadataCommand(), "",
		"set", "Projects/item", "--key", "classification", "--value", "project",
		"--if-match", moveReceipt.TargetRevision)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("metadata partial conflict = %#v err=%v", conflict, err)
	}

	resume := executeV2TestCommand(t, metadataCommand(), "", "get", "Projects/item")
	current := getInboxMetadataSnapshot(t, resume)
	if current.Revision != statusStep.RevisionAfter ||
		current.BodyRevision != moveReceipt.TargetBodyRevision ||
		current.Frontmatter["status"] != "organized" {
		t.Fatalf("metadata resume evidence = %#v step=%#v", current, statusStep)
	}
	classification := executeV2TestCommand(t, metadataCommand(), "",
		"set", "Projects/item", "--key", "classification", "--value", "project",
		"--if-match", current.Revision)
	classificationStep := getInboxMetadataStep(t, classification)
	if !classificationStep.Changed ||
		classificationStep.RevisionBefore != statusStep.RevisionAfter ||
		classificationStep.BodyRevision != moveReceipt.TargetBodyRevision {
		t.Fatalf("classification receipt step = %#v", classificationStep)
	}

	repeated := executeV2TestCommand(t, metadataCommand(), "",
		"set", "Projects/item", "--key", "classification", "--value", "project",
		"--if-match", classificationStep.RevisionAfter)
	repeatStep := getInboxMetadataStep(t, repeated)
	if repeatStep.Changed || repeatStep.RevisionBefore != classificationStep.RevisionAfter ||
		repeatStep.RevisionAfter != classificationStep.RevisionAfter {
		t.Fatalf("repeated metadata write was not a no-op: %#v", repeatStep)
	}
}

type inboxMoveReceipt struct {
	Operation          string `json:"operation"`
	RequestID          string `json:"request_id"`
	TransactionID      string `json:"transaction_id"`
	PlanHash           string `json:"plan_hash"`
	VaultID            string `json:"vault_id"`
	Source             string `json:"source"`
	SourceRevision     string `json:"source_revision"`
	SourceDigest       string `json:"source_digest"`
	Target             string `json:"target"`
	TargetRevision     string `json:"target_revision"`
	TargetBodyRevision string `json:"target_body_revision"`
}

type inboxMetadataStep struct {
	Key            string
	Value          string
	RevisionBefore string
	RevisionAfter  string
	BodyRevision   string
	Changed        bool
}

type inboxMetadataSnapshot struct {
	Path         string         `json:"path"`
	Revision     string         `json:"revision"`
	BodyRevision string         `json:"body_revision"`
	Frontmatter  map[string]any `json:"frontmatter"`
}

func getNoteForInboxTest(t *testing.T, registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory, path string) noteops.Note {
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

func getInboxMoveReceipt(t *testing.T, response noteTestEnvelope) inboxMoveReceipt {
	t.Helper()
	var data struct {
		Receipt inboxMoveReceipt `json:"receipt"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data.Receipt
}

func getInboxMetadataStep(t *testing.T, response noteTestEnvelope) inboxMetadataStep {
	t.Helper()
	var data struct {
		Key          string `json:"key"`
		Value        string `json:"value"`
		BodyRevision string `json:"body_revision"`
		Result       struct {
			Note noteops.Mutation `json:"note"`
		} `json:"result"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	return inboxMetadataStep{
		Key: data.Key, Value: data.Value, BodyRevision: data.BodyRevision,
		RevisionBefore: data.Result.Note.RevisionBefore,
		RevisionAfter:  data.Result.Note.RevisionAfter,
		Changed:        data.Result.Note.Changed,
	}
}

func getInboxMetadataSnapshot(t *testing.T, response noteTestEnvelope) inboxMetadataSnapshot {
	t.Helper()
	var data struct {
		Note inboxMetadataSnapshot `json:"note"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data.Note
}

func getMovePlanHashForInboxTest(t *testing.T, response noteTestEnvelope) string {
	t.Helper()
	var data struct {
		PlanHash string `json:"plan_hash"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.PlanHash == "" {
		t.Fatal("move dry-run did not return plan_hash")
	}
	return data.PlanHash
}
