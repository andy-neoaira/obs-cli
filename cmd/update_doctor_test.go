package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

type fakeUpdater struct {
	check updateCheckData
	plan  updatePlan
	apply updateApplyData
	err   error
}

func (f fakeUpdater) Check(context.Context, string) (updateCheckData, error) {
	return f.check, f.err
}

func (f fakeUpdater) Plan(context.Context, string, string) (updatePlan, error) {
	return f.plan, f.err
}

func (f fakeUpdater) Apply(context.Context, updatePlan) (updateApplyData, error) {
	return f.apply, f.err
}

func TestUpdateCommandsAreExplicitAndDryRunDoesNotApply(t *testing.T) {
	service := fakeUpdater{
		check: updateCheckData{
			CurrentVersion: "v1.0.0-rc.1", LatestVersion: "v1.0.0", UpdateAvailable: true,
		},
		plan: updatePlan{
			CurrentVersion: "v1.0.0-rc.1", TargetVersion: "v1.0.0",
			Asset: "obs-cli_darwin_all.tar.gz", Destination: "/tmp/obs-cli", Changed: true,
		},
		apply: updateApplyData{Applied: true},
	}
	check := executeJSONCommand(t, newUpdateCommand(service), "check", "--request-id", "update-check")
	if !check.OK || check.Operation != "update.check" {
		t.Fatalf("update check = %#v", check)
	}
	dryRun := executeJSONCommand(
		t,
		newUpdateCommand(service),
		"apply",
		"--dry-run",
		"--request-id",
		"update-plan",
	)
	if !dryRun.OK || dryRun.Operation != "update.apply" {
		t.Fatalf("update dry-run = %#v", dryRun)
	}
	var data updateApplyData
	if err := json.Unmarshal(dryRun.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.DryRun || data.Applied || !data.Changed || data.Plan.TargetVersion != "v1.0.0" {
		t.Fatalf("update dry-run applied or lost plan: %#v", data)
	}
}

func TestDoctorAuditsManagedSkillsAndOnlyChecksNetworkWhenRequested(t *testing.T) {
	configRoot := t.TempDir()
	codexRoot := t.TempDir()
	t.Setenv("OBS_CLI_CONFIG_HOME", configRoot)
	t.Setenv("CODEX_HOME", codexRoot)
	skillsRoot := filepath.Join(codexRoot, "skills")
	for _, name := range officialSkillNames {
		directory := filepath.Join(skillsRoot, name)
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("---\nname: " + name + "\n---\n")
		if err := os.WriteFile(filepath.Join(directory, "SKILL.md"), content, 0o644); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		metadata := managedSkillMetadata{
			ManagedBy: "obs-cli", Version: "dev", Source: defaultUpdateRepository,
			Skill: name, SkillDigest: "sha256:" + hex.EncodeToString(digest[:]),
		}
		raw, err := json.Marshal(metadata)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(
			filepath.Join(directory, ".obs-cli-managed.json"),
			raw,
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}
	service := fakeUpdater{check: updateCheckData{
		CurrentVersion: "dev", LatestVersion: "v1.0.0", UpdateAvailable: true,
	}}
	offline := executeJSONCommand(t, newDoctorCommand(service), "--request-id", "doctor-offline")
	var offlineData doctorData
	if err := json.Unmarshal(offline.Data, &offlineData); err != nil {
		t.Fatal(err)
	}
	if offlineData.NetworkChecked || offlineData.Update != nil {
		t.Fatalf("offline doctor performed update check: %#v", offlineData)
	}
	online := executeJSONCommand(
		t,
		newDoctorCommand(service),
		"--online",
		"--request-id",
		"doctor-online",
	)
	var onlineData doctorData
	if err := json.Unmarshal(online.Data, &onlineData); err != nil {
		t.Fatal(err)
	}
	if !onlineData.NetworkChecked || onlineData.Update == nil ||
		onlineData.Update.LatestVersion != "v1.0.0" {
		t.Fatalf("online doctor lost update result: %#v", onlineData)
	}
}

func TestDoctorSupportsDocumentedAgentSkillHosts(t *testing.T) {
	override := filepath.Join(t.TempDir(), "skills")
	cases := map[string]string{
		"codex":       "codex",
		"claude":      "claude-code",
		"claude-code": "claude-code",
		"opencode":    "opencode",
		"cursor":      "cursor",
		"kimi":        "kimi-code",
		"kimicode":    "kimi-code",
		"kimi-code":   "kimi-code",
	}
	for input, expected := range cases {
		agent, path, err := resolveAgentSkillsPath(input, override)
		if err != nil || agent != expected || path != override {
			t.Errorf(
				"resolveAgentSkillsPath(%q) = %q, %q, %v; want %q, %q",
				input,
				agent,
				path,
				err,
				expected,
				override,
			)
		}
	}
	if _, _, err := resolveAgentSkillsPath("unknown", "relative/path"); err == nil {
		t.Fatal("relative custom Skill path was accepted")
	}
}

func TestGitHubUpdaterPlanAndApplyVerifiedArchive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self replacement is intentionally unsupported on Windows")
	}
	assetName, err := updateAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skip(err)
	}
	if !strings.HasSuffix(assetName, ".tar.gz") {
		t.Skip("test fixture is a tar.gz archive")
	}
	candidateScript := []byte(`#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "obs-cli version v1.0.0"
  exit 0
fi
if [ "$1" = "capabilities" ]; then
  echo '{"protocol_version":"obs-cli/v1","ok":true,"operation":"capabilities.get","request_id":"test","data":{},"warnings":[]}'
  exit 0
fi
exit 1
`)
	archive := makeUpdateTarball(t, candidateScript)
	digest := sha256.Sum256(archive)
	checksums := []byte(hex.EncodeToString(digest[:]) + "  " + assetName + "\n")

	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body []byte
		status := http.StatusOK
		switch request.URL.Path {
		case "/repos/owner/repo/releases/latest":
			body, _ = json.Marshal(releaseMetadata{
				TagName: "v1.0.0",
				Assets: []releaseAsset{
					{Name: assetName, URL: "https://assets.test/asset"},
					{Name: "checksums.txt", URL: "https://assets.test/checksums"},
				},
			})
		case "/asset":
			body = archive
		case "/checksums":
			body = checksums
		default:
			status = http.StatusNotFound
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})

	directory := t.TempDir()
	current := filepath.Join(directory, "obs-cli")
	if err := os.WriteFile(current, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := &githubUpdater{
		client: &http.Client{Transport: transport}, apiBase: "https://api.test",
		repository:     "owner/repo",
		executablePath: func() (string, error) { return current, nil },
	}
	plan, err := service.Plan(context.Background(), "v1.0.0-rc.1", "")
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Apply(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied {
		t.Fatalf("update was not applied: %#v", result)
	}
	updated, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(updated, candidateScript) {
		t.Fatal("current executable does not contain verified candidate")
	}
	backup, err := os.ReadFile(current + ".previous")
	if err != nil || string(backup) != "old binary" {
		t.Fatalf("backup = %q err=%v", backup, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestReleaseVersionOrdering(t *testing.T) {
	cases := []struct {
		left, right string
		want        int
	}{
		{"v1.0.0-rc.1", "v1.0.0", -1},
		{"v1.0.0-rc.2", "v1.0.0-rc.1", 1},
		{"v1.0.0", "v1.0.0", 0},
		{"v2.0.0", "v1.9.9", 1},
	}
	for _, test := range cases {
		got, err := compareReleaseVersions(test.left, test.right)
		if err != nil || got != test.want {
			t.Errorf("compare(%s, %s) = %d, %v; want %d", test.left, test.right, got, err, test.want)
		}
	}
}

func TestGitHubUpdaterCheckPlanAndHTTPFailures(t *testing.T) {
	asset, err := updateAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skip(err)
	}
	releaseBody, _ := json.Marshal(releaseMetadata{TagName: "v1.2.0", Assets: []releaseAsset{
		{Name: asset, URL: "https://assets.test/archive"},
		{Name: "checksums.txt", URL: "https://assets.test/checksums"},
	}})
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		status := http.StatusOK
		body := releaseBody
		switch {
		case strings.Contains(request.URL.Path, "/server-error/"):
			status = http.StatusInternalServerError
		case strings.Contains(request.URL.Path, "/bad-json/"):
			body = []byte("{")
		case strings.Contains(request.URL.Path, "/empty-release/"):
			body = []byte(`{"assets":[]}`)
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header), Request: request}, nil
	})
	current := filepath.Join(t.TempDir(), "obs-cli")
	if err := os.WriteFile(current, []byte("current"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := &githubUpdater{client: &http.Client{Transport: transport}, apiBase: "https://api.test", repository: "owner/repo", executablePath: func() (string, error) { return current, nil }}
	checked, err := service.Check(context.Background(), "v1.0.0")
	if err != nil || !checked.UpdateAvailable || checked.LatestVersion != "v1.2.0" || checked.Channel != "stable" {
		t.Fatalf("Check = %#v, %v", checked, err)
	}
	if checked, err = service.Check(context.Background(), "dev"); err != nil || !checked.UpdateAvailable {
		t.Fatalf("development Check = %#v, %v", checked, err)
	}
	plan, err := service.Plan(context.Background(), "v1.0.0", "")
	if err != nil || !plan.Changed || plan.Asset != asset {
		t.Fatalf("Plan = %#v, %v", plan, err)
	}
	if _, err := service.Plan(context.Background(), "v2.0.0", ""); err == nil {
		t.Fatal("downgrade plan should fail")
	}
	if result, err := service.Apply(context.Background(), updatePlan{Changed: false}); err != nil || result.Applied || result.Changed {
		t.Fatalf("no-change Apply = %#v, %v", result, err)
	}

	badRepository := *service
	badRepository.repository = "bad repository"
	if _, err := badRepository.Check(context.Background(), "v1.0.0"); err == nil {
		t.Fatal("invalid repository should fail")
	}
	insecure := *service
	insecure.apiBase = "http://api.test"
	if _, err := insecure.Check(context.Background(), "v1.0.0"); err == nil {
		t.Fatal("insecure API should fail")
	}
	for path := range map[string]struct{}{"/server-error": {}, "/bad-json": {}, "/empty-release": {}} {
		broken := *service
		broken.apiBase = "https://api.test" + path
		broken.repository = "owner/repo"
		if _, err := broken.Check(context.Background(), "v1.0.0"); err == nil {
			t.Fatalf("API failure %s should fail", path)
		}
	}
	missingAssets := *service
	missingAssets.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := []byte(`{"tag_name":"v1.2.0","assets":[]}`)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header), Request: request}, nil
	})}
	if _, err := missingAssets.Plan(context.Background(), "v1.0.0", ""); err == nil {
		t.Fatal("release without assets should fail")
	}
	executableFailure := *service
	executableFailure.executablePath = func() (string, error) { return "", errors.New("executable failed") }
	if _, err := executableFailure.Plan(context.Background(), "v1.0.0", ""); err == nil {
		t.Fatal("executable lookup failure should fail")
	}
}

func TestUpdateArchiveURLChecksumAndPlatformBranches(t *testing.T) {
	for _, item := range []struct{ goos, arch string }{
		{"darwin", "amd64"}, {"linux", "amd64"}, {"linux", "arm64"}, {"windows", "amd64"}, {"windows", "arm64"},
	} {
		if _, err := updateAssetName(item.goos, item.arch); err != nil {
			t.Fatalf("updateAssetName(%s/%s): %v", item.goos, item.arch, err)
		}
	}
	if _, err := updateAssetName("plan9", "amd64"); err == nil {
		t.Fatal("unsupported platform should fail")
	}
	if releaseChannel("v1.0.0-rc.2") != "prerelease" || releaseChannel("v1.0.0") != "stable" {
		t.Fatal("release channel classification failed")
	}
	if _, err := parseUpdateURL("/relative"); err == nil {
		t.Fatal("relative update URL should fail")
	}

	root := t.TempDir()
	archivePath := filepath.Join(root, "asset.zip")
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	entry, err := writer.Create("nested/obs-cli.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("binary")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archivePath, archive.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	candidate, err := extractUpdateBinary(archivePath, root)
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(candidate); err != nil || string(data) != "binary" {
		t.Fatalf("zip candidate = %q, %v", data, err)
	}

	checksumPath := filepath.Join(root, "checksums.txt")
	digest := sha256.Sum256(archive.Bytes())
	if err := os.WriteFile(checksumPath, []byte(hex.EncodeToString(digest[:])+"  asset.zip\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleaseChecksum(archivePath, checksumPath, "asset.zip"); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleaseChecksum(archivePath, checksumPath, "missing.zip"); err == nil {
		t.Fatal("missing checksum entry should fail")
	}
	if err := os.WriteFile(checksumPath, []byte(strings.Repeat("0", 64)+"  asset.zip\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyReleaseChecksum(archivePath, checksumPath, "asset.zip"); err == nil {
		t.Fatal("checksum mismatch should fail")
	}

	service := &githubUpdater{client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("missing")), Header: make(http.Header), Request: request}, nil
	})}}
	if err := service.download(context.Background(), "http://assets.test/file", filepath.Join(root, "insecure")); err == nil {
		t.Fatal("insecure asset URL should fail")
	}
	if err := service.download(context.Background(), "https://assets.test/file", filepath.Join(root, "missing")); err == nil {
		t.Fatal("HTTP download failure should fail")
	}
}

type testEnvelope struct {
	OK        bool                  `json:"ok"`
	Operation string                `json:"operation"`
	Data      json.RawMessage       `json:"data"`
	Error     *protocol.DomainError `json:"error"`
}

func executeJSONCommand(t *testing.T, cobraCommand *cobra.Command, args ...string) testEnvelope {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cobraCommand.SetOut(&stdout)
	cobraCommand.SetErr(&stderr)
	cobraCommand.SilenceErrors = true
	cobraCommand.SilenceUsage = true
	cobraCommand.SetArgs(args)
	_ = cobraCommand.Execute()
	var response testEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	return response
}

func makeUpdateTarball(t *testing.T, content []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{
		Name: "obs-cli", Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
