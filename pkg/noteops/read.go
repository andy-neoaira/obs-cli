package noteops

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

type SearchMatch struct {
	Path      string `json:"path"`
	Revision  string `json:"revision"`
	Size      int64  `json:"size"`
	Line      int    `json:"line"`
	Snippet   string `json:"snippet"`
	MatchType string `json:"match_type"`
}

type SearchPage struct {
	Query        string        `json:"query"`
	Scope        string        `json:"scope"`
	Page         int           `json:"page"`
	PageSize     int           `json:"page_size"`
	TotalResults int           `json:"total_results"`
	HasMore      bool          `json:"has_more"`
	ScannedFiles int           `json:"scanned_files"`
	Truncated    bool          `json:"truncated"`
	Results      []SearchMatch `json:"results"`
}

type Backlink struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
	Line     int    `json:"line"`
	Kind     string `json:"kind"`
	Link     string `json:"link"`
}

type BacklinkReport struct {
	Target       string     `json:"target"`
	TargetExists bool       `json:"target_exists"`
	Scope        string     `json:"scope"`
	ScannedFiles int        `json:"scanned_files"`
	Truncated    bool       `json:"truncated"`
	Results      []Backlink `json:"results"`
}

var (
	readWikilinkPattern = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	readMarkdownPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)`)
)

func (s *Service) Search(query, scope string, pageNumber, pageSize, maxFiles int) (SearchPage, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchPage{}, fmt.Errorf("search query must not be empty")
	}
	prefix, err := s.scopePrefix(scope)
	if err != nil {
		return SearchPage{}, err
	}
	notes, err := s.List()
	if err != nil {
		return SearchPage{}, err
	}
	matches := make([]SearchMatch, 0)
	scanned := 0
	truncated := false
	lowerQuery := strings.ToLower(query)
	for _, logical := range notes {
		if prefix != "" && logical != strings.TrimSuffix(prefix, "/") && !strings.HasPrefix(logical, prefix) {
			continue
		}
		if scanned >= maxFiles {
			truncated = true
			break
		}
		scanned++
		snapshot, err := s.readRaw(logical)
		if err != nil {
			return SearchPage{}, err
		}
		if strings.Contains(strings.ToLower(path.Base(logical)), lowerQuery) {
			matches = append(matches, SearchMatch{
				Path: logical, Revision: snapshot.Revision, Size: snapshot.Size, MatchType: "filename",
				Snippet: path.Base(logical),
			})
		}
		if bytes.IndexByte(snapshot.Data, 0) >= 0 {
			continue
		}
		for index, line := range strings.Split(string(snapshot.Data), "\n") {
			if strings.Contains(strings.ToLower(line), lowerQuery) {
				matches = append(matches, SearchMatch{
					Path: logical, Revision: snapshot.Revision, Size: snapshot.Size, Line: index + 1,
					Snippet: boundedSnippet(line, 240), MatchType: "content",
				})
			}
		}
	}
	start := (pageNumber - 1) * pageSize
	end := start + pageSize
	if start > len(matches) {
		start = len(matches)
	}
	if end > len(matches) {
		end = len(matches)
	}
	results := append([]SearchMatch{}, matches[start:end]...)
	return SearchPage{
		Query: query, Scope: strings.TrimSuffix(prefix, "/"), Page: pageNumber, PageSize: pageSize,
		TotalResults: len(matches), HasMore: end < len(matches), ScannedFiles: scanned,
		Truncated: truncated, Results: results,
	}, nil
}

func (s *Service) Backlinks(target, scope string, maxFiles int) (BacklinkReport, error) {
	resolved, err := s.resolve(target, true)
	if err != nil {
		return BacklinkReport{}, err
	}
	targetID := strings.TrimSuffix(resolved.Logical, filepath.Ext(resolved.Logical))
	notes, err := s.List()
	if err != nil {
		return BacklinkReport{}, err
	}
	prefix, err := s.scopePrefix(scope)
	if err != nil {
		return BacklinkReport{}, err
	}
	report := BacklinkReport{
		Target: resolved.Logical, TargetExists: resolved.Exists,
		Scope: strings.TrimSuffix(prefix, "/"), Results: []Backlink{},
	}
	for _, logical := range notes {
		if prefix != "" && logical != strings.TrimSuffix(prefix, "/") && !strings.HasPrefix(logical, prefix) {
			continue
		}
		if report.ScannedFiles >= maxFiles {
			report.Truncated = true
			break
		}
		report.ScannedFiles++
		snapshot, err := s.readRaw(logical)
		if err != nil {
			return BacklinkReport{}, err
		}
		if bytes.IndexByte(snapshot.Data, 0) >= 0 {
			continue
		}
		for index, line := range strings.Split(string(snapshot.Data), "\n") {
			for _, match := range readWikilinkPattern.FindAllStringSubmatch(line, -1) {
				destination := strings.SplitN(strings.SplitN(match[1], "#", 2)[0], "^", 2)[0]
				destination = strings.TrimSuffix(strings.ReplaceAll(destination, "\\", "/"), ".md")
				if wikilinkTargets(destination, targetID) {
					report.Results = append(report.Results, Backlink{
						Path: logical, Revision: snapshot.Revision, Line: index + 1,
						Kind: "wikilink", Link: match[0],
					})
				}
			}
			for _, match := range readMarkdownPattern.FindAllStringSubmatch(line, -1) {
				if markdownLinkTargets(logical, match[1], targetID) {
					report.Results = append(report.Results, Backlink{
						Path: logical, Revision: snapshot.Revision, Line: index + 1,
						Kind: "markdown", Link: match[0],
					})
				}
			}
		}
	}
	return report, nil
}

func (s *Service) scopePrefix(scope string) (string, error) {
	if strings.TrimSpace(scope) == "" {
		return "", nil
	}
	resolved, err := s.resolver.Resolve(scope, pathpolicy.ResolveOptions{})
	if err != nil {
		return "", err
	}
	if !resolved.Exists {
		return "", ErrNoteNotFound
	}
	info, err := os.Stat(resolved.Path)
	if err != nil || !info.IsDir() {
		return "", ErrNoteNotFound
	}
	return strings.TrimSuffix(resolved.Logical, "/") + "/", nil
}

func (s *Service) readRaw(logical string) (storage.Snapshot, error) {
	resolved, err := s.resolve(logical, false)
	if err != nil {
		return storage.Snapshot{}, err
	}
	snapshot, err := storage.ReadSnapshot(resolved.Path)
	if err != nil {
		return storage.Snapshot{}, mapNotFound(err)
	}
	return snapshot, nil
}

func wikilinkTargets(destination, targetID string) bool {
	if strings.Contains(destination, "/") {
		return destination == targetID
	}
	return destination == path.Base(targetID)
}

func markdownLinkTargets(source, destination, targetID string) bool {
	if strings.Contains(destination, "://") || strings.HasPrefix(destination, "/") {
		return false
	}
	pathPart := strings.SplitN(destination, "#", 2)[0]
	decoded, err := url.PathUnescape(pathPart)
	if err != nil {
		return false
	}
	sourceDir := path.Dir(strings.ReplaceAll(source, "\\", "/"))
	resolved := path.Clean(path.Join(sourceDir, strings.ReplaceAll(decoded, "\\", "/")))
	return strings.TrimSuffix(resolved, ".md") == targetID
}

func boundedSnippet(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "…"
}
