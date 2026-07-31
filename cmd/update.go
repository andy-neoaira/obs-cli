package cmd

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

const defaultUpdateRepository = "andy-neoaira/obs-cli"

var updateRepositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type releaseMetadata struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type updateCheckData struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	Channel         string `json:"channel"`
	ReleaseURL      string `json:"release_url"`
}

type updatePlan struct {
	CurrentVersion string `json:"current_version"`
	TargetVersion  string `json:"target_version"`
	Asset          string `json:"asset"`
	AssetURL       string `json:"asset_url"`
	ChecksumURL    string `json:"checksum_url"`
	Destination    string `json:"destination"`
	Backup         string `json:"backup"`
	Changed        bool   `json:"changed"`
}

type updateApplyData struct {
	Plan    updatePlan `json:"plan"`
	DryRun  bool       `json:"dry_run"`
	Applied bool       `json:"applied"`
	Changed bool       `json:"changed"`
	Backup  string     `json:"backup"`
}

type updater interface {
	Check(context.Context, string) (updateCheckData, error)
	Plan(context.Context, string, string) (updatePlan, error)
	Apply(context.Context, updatePlan) (updateApplyData, error)
}

type githubUpdater struct {
	client         *http.Client
	apiBase        string
	repository     string
	executablePath func() (string, error)
}

func defaultUpdater() *githubUpdater {
	apiBase := strings.TrimRight(os.Getenv("OBS_CLI_UPDATE_API_URL"), "/")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	repository := os.Getenv("OBS_CLI_UPDATE_REPOSITORY")
	if repository == "" {
		repository = defaultUpdateRepository
	}
	return &githubUpdater{
		client:         &http.Client{Timeout: 15 * time.Second},
		apiBase:        apiBase,
		repository:     repository,
		executablePath: os.Executable,
	}
}

func (u *githubUpdater) Check(ctx context.Context, current string) (updateCheckData, error) {
	release, err := u.release(ctx, "")
	if err != nil {
		return updateCheckData{}, err
	}
	comparison, err := compareReleaseVersions(current, release.TagName)
	if err != nil && current != "dev" {
		return updateCheckData{}, err
	}
	return updateCheckData{
		CurrentVersion:  current,
		LatestVersion:   release.TagName,
		UpdateAvailable: current == "dev" || comparison < 0,
		Channel:         releaseChannel(release.TagName),
		ReleaseURL:      "https://github.com/" + u.repository + "/releases/tag/" + release.TagName,
	}, nil
}

func (u *githubUpdater) Plan(ctx context.Context, current, requested string) (updatePlan, error) {
	release, err := u.release(ctx, requested)
	if err != nil {
		return updatePlan{}, err
	}
	if _, err := parseReleaseVersion(release.TagName); err != nil {
		return updatePlan{}, err
	}
	changed := true
	if current != "dev" {
		comparison, err := compareReleaseVersions(current, release.TagName)
		if err != nil {
			return updatePlan{}, err
		}
		if comparison > 0 {
			return updatePlan{}, fmt.Errorf(
				"target version %s is older than current version %s",
				release.TagName,
				current,
			)
		}
		changed = comparison < 0
	}
	assetName, err := updateAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return updatePlan{}, err
	}
	assetURL := findReleaseAsset(release.Assets, assetName)
	checksumURL := findReleaseAsset(release.Assets, "checksums.txt")
	if assetURL == "" || checksumURL == "" {
		return updatePlan{}, fmt.Errorf(
			"release %s does not contain %s and checksums.txt",
			release.TagName,
			assetName,
		)
	}
	executablePath, err := u.executablePath()
	if err != nil {
		return updatePlan{}, fmt.Errorf("resolve current executable: %w", err)
	}
	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		return updatePlan{}, fmt.Errorf("resolve executable symlinks: %w", err)
	}
	return updatePlan{
		CurrentVersion: current,
		TargetVersion:  release.TagName,
		Asset:          assetName,
		AssetURL:       assetURL,
		ChecksumURL:    checksumURL,
		Destination:    executablePath,
		Backup:         executablePath + ".previous",
		Changed:        changed,
	}, nil
}

func (u *githubUpdater) Apply(ctx context.Context, plan updatePlan) (updateApplyData, error) {
	if !plan.Changed {
		return updateApplyData{
			Plan: plan, DryRun: false, Applied: false, Changed: false,
		}, nil
	}
	if runtime.GOOS == "windows" {
		return updateApplyData{}, errors.New(
			"self-update apply is not supported on Windows; use scripts/install.sh --force",
		)
	}
	workDir, err := os.MkdirTemp("", "obs-cli-update-*")
	if err != nil {
		return updateApplyData{}, fmt.Errorf("create update workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	archivePath := filepath.Join(workDir, plan.Asset)
	checksumPath := filepath.Join(workDir, "checksums.txt")
	if err := u.download(ctx, plan.AssetURL, archivePath); err != nil {
		return updateApplyData{}, err
	}
	if err := u.download(ctx, plan.ChecksumURL, checksumPath); err != nil {
		return updateApplyData{}, err
	}
	if err := verifyReleaseChecksum(archivePath, checksumPath, plan.Asset); err != nil {
		return updateApplyData{}, err
	}
	candidate, err := extractUpdateBinary(archivePath, workDir)
	if err != nil {
		return updateApplyData{}, err
	}
	if err := os.Chmod(candidate, 0o755); err != nil {
		return updateApplyData{}, fmt.Errorf("make update candidate executable: %w", err)
	}
	if err := validateUpdateBinary(ctx, candidate, plan.TargetVersion); err != nil {
		return updateApplyData{}, err
	}

	info, err := os.Stat(plan.Destination)
	if err != nil {
		return updateApplyData{}, fmt.Errorf("stat current executable: %w", err)
	}
	stage, err := os.CreateTemp(filepath.Dir(plan.Destination), ".obs-cli-update-*")
	if err != nil {
		return updateApplyData{}, fmt.Errorf("stage update beside executable: %w", err)
	}
	stagePath := stage.Name()
	stage.Close()
	defer os.Remove(stagePath)
	if err := copyFile(candidate, stagePath, info.Mode().Perm()|0o111); err != nil {
		return updateApplyData{}, err
	}

	if _, err := os.Stat(plan.Backup); err == nil {
		return updateApplyData{}, fmt.Errorf(
			"backup already exists at %s; preserve or remove it before upgrading",
			plan.Backup,
		)
	} else if !errors.Is(err, os.ErrNotExist) {
		return updateApplyData{}, fmt.Errorf("inspect update backup: %w", err)
	}
	if err := os.Rename(plan.Destination, plan.Backup); err != nil {
		return updateApplyData{}, fmt.Errorf("backup current executable: %w", err)
	}
	if err := os.Rename(stagePath, plan.Destination); err != nil {
		rollbackErr := os.Rename(plan.Backup, plan.Destination)
		return updateApplyData{}, fmt.Errorf(
			"activate update: %w (rollback error: %v)",
			err,
			rollbackErr,
		)
	}
	return updateApplyData{
		Plan: plan, DryRun: false, Applied: true, Changed: true, Backup: plan.Backup,
	}, nil
}

func (u *githubUpdater) release(ctx context.Context, version string) (releaseMetadata, error) {
	if !updateRepositoryPattern.MatchString(u.repository) {
		return releaseMetadata{}, fmt.Errorf("invalid update repository %q", u.repository)
	}
	endpoint := u.apiBase + "/repos/" + u.repository + "/releases/latest"
	if version != "" {
		if _, err := parseReleaseVersion(version); err != nil {
			return releaseMetadata{}, err
		}
		endpoint = u.apiBase + "/repos/" + u.repository + "/releases/tags/" + version
	}
	parsedEndpoint, err := parseUpdateURL(endpoint)
	if err != nil {
		return releaseMetadata{}, err
	}
	if parsedEndpoint.Scheme != "https" {
		return releaseMetadata{}, fmt.Errorf("update API URL must use HTTPS: %s", endpoint)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return releaseMetadata{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "obs-cli/"+resolveVersion())
	response, err := u.client.Do(request)
	if err != nil {
		return releaseMetadata{}, fmt.Errorf("query GitHub release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return releaseMetadata{}, fmt.Errorf("query GitHub release: HTTP %d", response.StatusCode)
	}
	var release releaseMetadata
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&release); err != nil {
		return releaseMetadata{}, fmt.Errorf("decode GitHub release: %w", err)
	}
	if release.TagName == "" {
		return releaseMetadata{}, errors.New("GitHub release has no tag name")
	}
	return release, nil
}

func (u *githubUpdater) download(ctx context.Context, url, destination string) error {
	parsedURL, err := parseUpdateURL(url)
	if err != nil {
		return err
	}
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("update asset URL must use HTTPS: %s", url)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "obs-cli/"+resolveVersion())
	response, err := u.client.Do(request)
	if err != nil {
		return fmt.Errorf("download update asset: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download update asset: HTTP %d", response.StatusCode)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, io.LimitReader(response.Body, 256<<20))
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("write update asset: %w", copyErr)
	}
	return closeErr
}

func newUpdateCommand(service updater) *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "Explicitly check for and apply verified obs-cli releases",
	}
	command.AddCommand(newUpdateCheckCommand(service), newUpdateApplyCommand(service))
	return command
}

func newUpdateCheckCommand(service updater) *cobra.Command {
	var common commonFlags
	command := &cobra.Command{
		Use:   "check",
		Short: "Check GitHub Releases for a newer version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			requestID, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				requestID, _ = protocol.ResolveRequestID("")
				return renderEnvelope(cmd, "update.check", requestID, func() (any, error) {
					return nil, err
				})
			}
			if common.Output != "json" {
				err := protocol.NewError(
					protocol.InvalidArgument,
					"update check only supports JSON output",
					map[string]any{"field": "output"},
				)
				return renderEnvelope(cmd, "update.check", requestID, func() (any, error) {
					return nil, err
				})
			}
			return renderEnvelope(cmd, "update.check", requestID, func() (any, error) {
				return service.Check(cmd.Context(), resolveVersion())
			})
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{Output: true, RequestID: true}, false)
	return command
}

func newUpdateApplyCommand(service updater) *cobra.Command {
	var common commonFlags
	var version string
	command := &cobra.Command{
		Use:   "apply",
		Short: "Download, verify, and explicitly replace the current binary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			requestID, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				requestID, _ = protocol.ResolveRequestID("")
				return renderEnvelope(cmd, "update.apply", requestID, func() (any, error) {
					return nil, err
				})
			}
			if common.Output != "json" {
				err := protocol.NewError(
					protocol.InvalidArgument,
					"update apply only supports JSON output",
					map[string]any{"field": "output"},
				)
				return renderEnvelope(cmd, "update.apply", requestID, func() (any, error) {
					return nil, err
				})
			}
			return renderEnvelope(cmd, "update.apply", requestID, func() (any, error) {
				plan, err := service.Plan(cmd.Context(), resolveVersion(), version)
				if err != nil {
					return nil, err
				}
				if common.DryRun {
					return updateApplyData{
						Plan: plan, DryRun: true, Applied: false, Changed: plan.Changed,
					}, nil
				}
				return service.Apply(cmd.Context(), plan)
			})
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{
		Output: true, RequestID: true, DryRun: true,
	}, false)
	command.Flags().StringVar(&version, "version", "", "release tag; default is latest")
	return command
}

func parseUpdateURL(value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("parse update URL: %w", err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("update URL has no host: %s", value)
	}
	return parsed, nil
}

var releaseVersionPattern = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)(?:-rc\.([0-9]+))?$`)

type releaseVersion struct {
	major, minor, patch int
	rc                  int
	prerelease          bool
}

func parseReleaseVersion(value string) (releaseVersion, error) {
	match := releaseVersionPattern.FindStringSubmatch(value)
	if match == nil {
		return releaseVersion{}, fmt.Errorf("invalid release version %q", value)
	}
	var result releaseVersion
	_, err := fmt.Sscanf(match[1]+" "+match[2]+" "+match[3], "%d %d %d", &result.major, &result.minor, &result.patch)
	if err != nil {
		return releaseVersion{}, err
	}
	if match[4] != "" {
		result.prerelease = true
		_, err = fmt.Sscanf(match[4], "%d", &result.rc)
	}
	return result, err
}

func compareReleaseVersions(left, right string) (int, error) {
	a, err := parseReleaseVersion(left)
	if err != nil {
		return 0, err
	}
	b, err := parseReleaseVersion(right)
	if err != nil {
		return 0, err
	}
	valuesA := []int{a.major, a.minor, a.patch}
	valuesB := []int{b.major, b.minor, b.patch}
	for index := range valuesA {
		if valuesA[index] < valuesB[index] {
			return -1, nil
		}
		if valuesA[index] > valuesB[index] {
			return 1, nil
		}
	}
	if a.prerelease != b.prerelease {
		if a.prerelease {
			return -1, nil
		}
		return 1, nil
	}
	if a.rc < b.rc {
		return -1, nil
	}
	if a.rc > b.rc {
		return 1, nil
	}
	return 0, nil
}

func releaseChannel(version string) string {
	if strings.Contains(version, "-rc.") {
		return "prerelease"
	}
	return "stable"
}

func updateAssetName(goos, goarch string) (string, error) {
	switch goos {
	case "darwin":
		return "obs-cli_darwin_all.tar.gz", nil
	case "linux":
		if goarch == "amd64" || goarch == "arm64" {
			return "obs-cli_linux_" + goarch + ".tar.gz", nil
		}
	case "windows":
		if goarch == "amd64" || goarch == "arm64" {
			return "obs-cli_windows_" + goarch + ".zip", nil
		}
	}
	return "", fmt.Errorf("unsupported update platform: %s/%s", goos, goarch)
}

func findReleaseAsset(assets []releaseAsset, name string) string {
	for _, asset := range assets {
		if asset.Name == name {
			return asset.URL
		}
	}
	return ""
}

func verifyReleaseChecksum(archivePath, checksumPath, asset string) error {
	file, err := os.Open(checksumPath)
	if err != nil {
		return err
	}
	defer file.Close()
	expected := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == asset {
			expected = fields[0]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if expected == "" {
		return fmt.Errorf("checksums.txt does not contain %s", asset)
	}
	archive, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, archive); err != nil {
		return err
	}
	actual := hex.EncodeToString(digest.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s", asset)
	}
	return nil
}

func extractUpdateBinary(archivePath, destination string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		reader, err := zip.OpenReader(archivePath)
		if err != nil {
			return "", err
		}
		defer reader.Close()
		var candidate string
		for _, file := range reader.File {
			if filepath.Base(file.Name) != "obs-cli.exe" || file.FileInfo().IsDir() {
				continue
			}
			if candidate != "" {
				return "", errors.New("update archive contains multiple obs-cli binaries")
			}
			input, err := file.Open()
			if err != nil {
				return "", err
			}
			candidate = filepath.Join(destination, "candidate-obs-cli.exe")
			output, err := os.OpenFile(candidate, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
			if err == nil {
				_, err = io.Copy(output, io.LimitReader(input, 128<<20))
				closeErr := output.Close()
				if err == nil {
					err = closeErr
				}
			}
			input.Close()
			if err != nil {
				return "", err
			}
		}
		if candidate == "" {
			return "", errors.New("update archive does not contain obs-cli.exe")
		}
		return candidate, nil
	}
	archive, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer archive.Close()
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	candidate := ""
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if filepath.Base(header.Name) != "obs-cli" || header.Typeflag != tar.TypeReg {
			continue
		}
		if candidate != "" {
			return "", errors.New("update archive contains multiple obs-cli binaries")
		}
		candidate = filepath.Join(destination, "candidate-obs-cli")
		output, err := os.OpenFile(candidate, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(output, io.LimitReader(tarReader, 128<<20))
		closeErr := output.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
	}
	if candidate == "" {
		return "", errors.New("update archive does not contain obs-cli")
	}
	return candidate, nil
}

func validateUpdateBinary(ctx context.Context, candidate, version string) error {
	versionCommand := exec.CommandContext(ctx, candidate, "--version")
	versionOutput, err := versionCommand.Output()
	if err != nil {
		return fmt.Errorf("validate update version: %w", err)
	}
	if strings.TrimSpace(string(versionOutput)) != "obs-cli version "+version {
		return fmt.Errorf("update binary version does not match %s", version)
	}
	capabilitiesCommand := exec.CommandContext(ctx, candidate, "capabilities", "--output", "json")
	capabilitiesOutput, err := capabilitiesCommand.Output()
	if err != nil {
		return fmt.Errorf("validate update capabilities: %w", err)
	}
	var envelope protocol.Envelope
	if err := json.Unmarshal(capabilitiesOutput, &envelope); err != nil {
		return fmt.Errorf("validate update capabilities JSON: %w", err)
	}
	if !envelope.OK || envelope.ProtocolVersion != protocol.Version {
		return errors.New("update binary does not provide a valid obs-cli/v1 capability envelope")
	}
	return nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Chmod(destination, mode)
}
