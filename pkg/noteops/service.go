package noteops

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/frontmatter"
	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

var (
	ErrNoteNotFound          = errors.New("note not found")
	ErrPatchContextMismatch  = errors.New("patch context does not match")
	ErrPatchContextAmbiguous = errors.New("patch context is not unique")
	ErrSectionNotFound       = errors.New("markdown section not found")
	ErrSectionAmbiguous      = errors.New("markdown section is not unique")
	ErrInvalidFrontmatter    = errors.New("invalid note frontmatter")
	ErrRevisionRequired      = errors.New("revision precondition is required")
	ErrInvalidRevision       = errors.New("invalid revision")
)

type CreateConflict struct {
	Path              string
	ExistingRevision  string
	RequestedRevision string
}

type InvalidFrontmatterError struct {
	Path     string
	Revision string
	Cause    error
}

func (e *InvalidFrontmatterError) Error() string {
	return fmt.Sprintf("invalid frontmatter in %s: %v", e.Path, e.Cause)
}

func (e *InvalidFrontmatterError) Unwrap() error {
	return ErrInvalidFrontmatter
}

func (e *CreateConflict) Error() string {
	return fmt.Sprintf("note already exists: %s", e.Path)
}

func (e *CreateConflict) Unwrap() error {
	return storage.ErrAlreadyExists
}

type Note struct {
	Path         string         `json:"path"`
	Revision     string         `json:"revision"`
	BodyRevision string         `json:"body_revision"`
	ModifiedAt   string         `json:"modified_at"`
	Content      string         `json:"content"`
	Frontmatter  map[string]any `json:"frontmatter"`
}

type Mutation struct {
	Path           string `json:"path"`
	Action         string `json:"action"`
	RevisionBefore string `json:"revision_before,omitempty"`
	RevisionAfter  string `json:"revision_after,omitempty"`
	Changed        bool   `json:"changed"`
	physicalPath   string
	content        []byte
	mode           fs.FileMode
}

type Service struct {
	resolver *pathpolicy.Resolver
	store    *storage.Store
}

func NewService(vaultRoot string, store *storage.Store) (*Service, error) {
	resolver, err := pathpolicy.NewResolver(vaultRoot)
	if err != nil {
		return nil, err
	}
	if store == nil {
		store = storage.DefaultStore()
	}
	return &Service{resolver: resolver, store: store}, nil
}

func (s *Service) List() ([]string, error) {
	notes := make([]string, 0)
	err := filepath.WalkDir(s.resolver.Root(), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == s.resolver.Root() {
			return nil
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return nil
		}
		logical, err := filepath.Rel(s.resolver.Root(), path)
		if err != nil {
			return err
		}
		logical = filepath.ToSlash(logical)
		result, err := s.resolver.Resolve(logical, pathpolicy.ResolveOptions{})
		if err != nil {
			return err
		}
		if result.Exists {
			notes = append(notes, result.Logical)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(notes)
	return notes, nil
}

func (s *Service) Get(input string) (Note, error) {
	resolved, err := s.resolve(input, false)
	if err != nil {
		return Note{}, err
	}
	snapshot, err := storage.ReadSnapshot(resolved.Path)
	if err != nil {
		return Note{}, mapNotFound(err)
	}
	fm := map[string]any{}
	body := string(snapshot.Data)
	if frontmatter.HasFrontmatter(string(snapshot.Data)) {
		parsed, parsedBody, err := frontmatter.Parse(string(snapshot.Data))
		if err != nil {
			return Note{}, &InvalidFrontmatterError{
				Path: resolved.Logical, Revision: snapshot.Revision, Cause: err,
			}
		}
		body = parsedBody
		if parsed != nil {
			fm = parsed
		}
	}
	return Note{
		Path:         resolved.Logical,
		Revision:     snapshot.Revision,
		BodyRevision: storage.Revision([]byte(body)),
		ModifiedAt:   snapshot.ModifiedAt.Format(time.RFC3339Nano),
		Content:      string(snapshot.Data),
		Frontmatter:  fm,
	}, nil
}

func (s *Service) PlanCreate(input string, content []byte) (Mutation, error) {
	resolved, err := s.resolve(input, true)
	if err != nil {
		return Mutation{}, err
	}
	if resolved.Exists {
		snapshot, err := storage.ReadSnapshot(resolved.Path)
		if err != nil {
			return Mutation{}, err
		}
		return Mutation{}, &CreateConflict{
			Path:              resolved.Logical,
			ExistingRevision:  snapshot.Revision,
			RequestedRevision: storage.Revision(content),
		}
	}
	return Mutation{
		Path:          resolved.Logical,
		Action:        "create",
		RevisionAfter: storage.Revision(content),
		Changed:       true,
		physicalPath:  resolved.Path,
		content:       append([]byte(nil), content...),
		mode:          0o644,
	}, nil
}

func (s *Service) Create(input string, content []byte) (Mutation, error) {
	plan, err := s.PlanCreate(input, content)
	if err != nil {
		return Mutation{}, err
	}
	if err := os.MkdirAll(filepath.Dir(plan.physicalPath), 0o755); err != nil {
		return Mutation{}, err
	}
	resolved, err := s.resolve(plan.Path, true)
	if err != nil {
		return Mutation{}, err
	}
	if resolved.Path != plan.physicalPath {
		return Mutation{}, fmt.Errorf("%w: target identity changed while creating parent directories", pathpolicy.ErrOutsideVault)
	}
	result, err := s.store.WriteAtomic(
		plan.physicalPath,
		plan.content,
		storage.Preconditions{MustNotExist: true},
		storage.WriteOptions{Mode: plan.mode},
	)
	if err != nil {
		return Mutation{}, err
	}
	plan.RevisionAfter = result.RevisionAfter
	plan.Changed = result.Changed
	return plan, nil
}

func (s *Service) PlanAppend(input string, addition []byte, section, expectedRevision string) (Mutation, error) {
	return s.planUpdate(input, expectedRevision, func(current []byte) ([]byte, error) {
		if section == "" {
			return appendWithBoundary(current, addition), nil
		}
		return appendToSection(current, addition, section)
	})
}

func (s *Service) Append(input string, addition []byte, section, expectedRevision string) (Mutation, error) {
	plan, err := s.PlanAppend(input, addition, section, expectedRevision)
	if err != nil {
		return Mutation{}, err
	}
	return s.applyUpdate(plan)
}

func (s *Service) PlanPatch(input string, context, replacement []byte, expectedRevision string) (Mutation, error) {
	if expectedRevision == "" {
		return Mutation{}, ErrRevisionRequired
	}
	if len(context) == 0 {
		return Mutation{}, ErrPatchContextMismatch
	}
	return s.planUpdate(input, expectedRevision, func(current []byte) ([]byte, error) {
		switch bytes.Count(current, context) {
		case 0:
			return nil, ErrPatchContextMismatch
		case 1:
			return bytes.Replace(current, context, replacement, 1), nil
		default:
			return nil, ErrPatchContextAmbiguous
		}
	})
}

func (s *Service) Patch(input string, context, replacement []byte, expectedRevision string) (Mutation, error) {
	plan, err := s.PlanPatch(input, context, replacement, expectedRevision)
	if err != nil {
		return Mutation{}, err
	}
	return s.applyUpdate(plan)
}

func (s *Service) PlanReplace(input string, content []byte, expectedRevision string) (Mutation, error) {
	if expectedRevision == "" {
		return Mutation{}, ErrRevisionRequired
	}
	return s.planUpdate(input, expectedRevision, func([]byte) ([]byte, error) {
		return append([]byte(nil), content...), nil
	})
}

func (s *Service) Replace(input string, content []byte, expectedRevision string) (Mutation, error) {
	plan, err := s.PlanReplace(input, content, expectedRevision)
	if err != nil {
		return Mutation{}, err
	}
	return s.applyUpdate(plan)
}

func (s *Service) PlanDelete(input, expectedRevision string) (Mutation, error) {
	if expectedRevision == "" {
		return Mutation{}, ErrRevisionRequired
	}
	if !storage.IsRevision(expectedRevision) {
		return Mutation{}, ErrInvalidRevision
	}
	resolved, err := s.resolve(input, true)
	if err != nil {
		return Mutation{}, err
	}
	snapshot, err := storage.ReadSnapshot(resolved.Path)
	if err != nil {
		return Mutation{}, mapNotFound(err)
	}
	if snapshot.Revision != expectedRevision {
		return Mutation{}, storage.ErrRevisionConflict
	}
	return Mutation{
		Path:           resolved.Logical,
		Action:         "delete",
		RevisionBefore: snapshot.Revision,
		Changed:        true,
		physicalPath:   resolved.Path,
		mode:           snapshot.Mode,
	}, nil
}

func (s *Service) Delete(input, expectedRevision string) (Mutation, error) {
	plan, err := s.PlanDelete(input, expectedRevision)
	if err != nil {
		return Mutation{}, err
	}
	if _, err := s.store.DeleteAtomic(plan.physicalPath, plan.RevisionBefore); err != nil {
		return Mutation{}, err
	}
	return plan, nil
}

func (s *Service) planUpdate(
	input, expectedRevision string,
	transform func([]byte) ([]byte, error),
) (Mutation, error) {
	if expectedRevision != "" && !storage.IsRevision(expectedRevision) {
		return Mutation{}, ErrInvalidRevision
	}
	resolved, err := s.resolve(input, true)
	if err != nil {
		return Mutation{}, err
	}
	snapshot, err := storage.ReadSnapshot(resolved.Path)
	if err != nil {
		return Mutation{}, mapNotFound(err)
	}
	if expectedRevision != "" && snapshot.Revision != expectedRevision {
		return Mutation{}, storage.ErrRevisionConflict
	}
	next, err := transform(snapshot.Data)
	if err != nil {
		return Mutation{}, err
	}
	return Mutation{
		Path:           resolved.Logical,
		Action:         "update",
		RevisionBefore: snapshot.Revision,
		RevisionAfter:  storage.Revision(next),
		Changed:        !bytes.Equal(snapshot.Data, next),
		physicalPath:   resolved.Path,
		content:        next,
		mode:           snapshot.Mode,
	}, nil
}

func (s *Service) applyUpdate(plan Mutation) (Mutation, error) {
	result, err := s.store.WriteAtomic(
		plan.physicalPath,
		plan.content,
		storage.Preconditions{ExpectedRevision: plan.RevisionBefore},
		storage.WriteOptions{Mode: plan.mode},
	)
	if err != nil {
		return Mutation{}, err
	}
	plan.RevisionBefore = result.RevisionBefore
	plan.RevisionAfter = result.RevisionAfter
	plan.Changed = result.Changed
	return plan, nil
}

func (s *Service) resolve(input string, mutation bool) (pathpolicy.Result, error) {
	logical := input
	if !strings.EqualFold(filepath.Ext(logical), ".md") {
		logical += ".md"
	}
	result, err := s.resolver.Resolve(logical, pathpolicy.ResolveOptions{})
	if err != nil {
		return pathpolicy.Result{}, err
	}
	if mutation && result.ThroughSymlink {
		return pathpolicy.Result{}, fmt.Errorf("%w: mutation through symbolic link", pathpolicy.ErrOutsideVault)
	}
	if !mutation && !result.Exists {
		return pathpolicy.Result{}, ErrNoteNotFound
	}
	return result, nil
}

func mapNotFound(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return ErrNoteNotFound
	}
	return err
}

func appendWithBoundary(current, addition []byte) []byte {
	result := append([]byte(nil), current...)
	if len(result) != 0 && len(addition) != 0 && result[len(result)-1] != '\n' {
		result = append(result, '\n')
	}
	return append(result, addition...)
}

func appendToSection(current, addition []byte, section string) ([]byte, error) {
	lines := splitLines(current)
	type heading struct {
		index int
		level int
	}
	headings := make([]heading, 0)
	inFence := false
	var fence byte
	var fenceLength int
	for index, line := range lines {
		if marker, length, ok := parseFence(line); ok {
			if !inFence {
				inFence, fence, fenceLength = true, marker, length
			} else if marker == fence && length >= fenceLength {
				inFence = false
			}
			continue
		}
		if inFence {
			continue
		}
		level, _, ok := parseHeading(line)
		if ok {
			headings = append(headings, heading{index: index, level: level})
		}
	}
	matches := make([]heading, 0, 1)
	for _, item := range headings {
		_, title, _ := parseHeading(lines[item.index])
		if title == section {
			matches = append(matches, item)
		}
	}
	if len(matches) == 0 {
		return nil, ErrSectionNotFound
	}
	if len(matches) != 1 {
		return nil, ErrSectionAmbiguous
	}
	match := matches[0]
	insertAt := len(lines)
	for _, item := range headings {
		if item.index > match.index && item.level <= match.level {
			insertAt = item.index
			break
		}
	}
	before := bytes.Join(lines[:insertAt], []byte("\n"))
	after := bytes.Join(lines[insertAt:], []byte("\n"))
	result := appendWithBoundary(before, addition)
	if len(after) != 0 {
		result = appendWithBoundary(result, after)
	}
	return result, nil
}

func parseFence(line []byte) (byte, int, bool) {
	trimmed := strings.TrimSpace(string(line))
	if len(trimmed) < 3 || (trimmed[0] != '`' && trimmed[0] != '~') {
		return 0, 0, false
	}
	length := 0
	for length < len(trimmed) && trimmed[length] == trimmed[0] {
		length++
	}
	return trimmed[0], length, length >= 3
}

func splitLines(content []byte) [][]byte {
	return bytes.Split(content, []byte("\n"))
}

func parseHeading(line []byte) (int, string, bool) {
	trimmed := strings.TrimSpace(string(line))
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level == len(trimmed) || trimmed[level] != ' ' {
		return 0, "", false
	}
	title := strings.TrimSpace(strings.TrimRight(trimmed[level+1:], "#"))
	return level, title, title != ""
}
