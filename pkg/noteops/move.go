package noteops

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

type MoveChange struct {
	Action           string              `json:"action"`
	Path             string              `json:"path"`
	Target           string              `json:"target,omitempty"`
	ExpectedRevision string              `json:"expected_revision,omitempty"`
	RevisionAfter    string              `json:"revision_after,omitempty"`
	LinkEdits        []obsidian.LinkEdit `json:"link_edits"`
}

type MovePlan struct {
	Source     string       `json:"source"`
	Target     string       `json:"target"`
	Changes    []MoveChange `json:"changes"`
	Risks      []string     `json:"risks"`
	mutations  []storage.Mutation
	targetPath string
}

type MoveResult struct {
	TransactionID string       `json:"transaction_id"`
	Source        string       `json:"source"`
	Target        string       `json:"target"`
	Changes       []MoveChange `json:"changes"`
	RevisionAfter string       `json:"revision_after"`
}

type MovePartialFailure struct {
	TransactionID   string
	Completed       []string
	Failed          []string
	RolledBack      []string
	RollbackFailed  []string
	RecoveryActions []string
}

func (e *MovePartialFailure) Error() string {
	return "move transaction requires recovery: " + e.TransactionID
}

func (e *MovePartialFailure) Unwrap() error {
	return storage.ErrPartialFailure
}

func (s *Service) PlanMove(sourceInput, targetInput, expectedRevision string) (MovePlan, error) {
	if expectedRevision == "" {
		return MovePlan{}, ErrRevisionRequired
	}
	if !storage.IsRevision(expectedRevision) {
		return MovePlan{}, ErrInvalidRevision
	}
	source, err := s.resolve(sourceInput, true)
	if err != nil {
		return MovePlan{}, err
	}
	target, err := s.resolve(targetInput, true)
	if err != nil {
		return MovePlan{}, err
	}
	if target.Exists {
		return MovePlan{}, storage.ErrAlreadyExists
	}
	sourceSnapshot, err := storage.ReadSnapshot(source.Path)
	if err != nil {
		return MovePlan{}, mapNotFound(err)
	}
	if sourceSnapshot.Revision != expectedRevision {
		return MovePlan{}, storage.ErrRevisionConflict
	}

	notes, err := s.List()
	if err != nil {
		return MovePlan{}, err
	}
	oldBase := strings.TrimSuffix(filepath.Base(source.Logical), filepath.Ext(source.Logical))
	baseMatches := 0
	for _, note := range notes {
		base := strings.TrimSuffix(filepath.Base(note), filepath.Ext(note))
		if base == oldBase {
			baseMatches++
		}
	}
	includeBase := baseMatches == 1
	rewriter := &obsidian.LinkRewriter{}
	targetContent, sourceEdits, err := rewriter.RewriteStructuredLinks(
		sourceSnapshot.Data, source.Logical, source.Logical, target.Logical, includeBase,
	)
	if err != nil {
		return MovePlan{}, err
	}

	changes := []MoveChange{{
		Action:           "move",
		Path:             source.Logical,
		Target:           target.Logical,
		ExpectedRevision: sourceSnapshot.Revision,
		RevisionAfter:    storage.Revision(targetContent),
		LinkEdits:        sourceEdits,
	}}
	mutations := []storage.Mutation{{
		Path:         target.Path,
		Data:         targetContent,
		Precondition: storage.Preconditions{MustNotExist: true},
		Mode:         sourceSnapshot.Mode,
	}}
	for _, logical := range notes {
		if logical == source.Logical {
			continue
		}
		resolved, err := s.resolve(logical, true)
		if err != nil {
			return MovePlan{}, err
		}
		snapshot, err := storage.ReadSnapshot(resolved.Path)
		if err != nil {
			return MovePlan{}, err
		}
		updated, edits, err := rewriter.RewriteStructuredLinks(
			snapshot.Data, logical, source.Logical, target.Logical, includeBase,
		)
		if err != nil {
			return MovePlan{}, err
		}
		if bytes.Equal(snapshot.Data, updated) {
			continue
		}
		changes = append(changes, MoveChange{
			Action:           "update",
			Path:             logical,
			ExpectedRevision: snapshot.Revision,
			RevisionAfter:    storage.Revision(updated),
			LinkEdits:        edits,
		})
		mutations = append(mutations, storage.Mutation{
			Path:         resolved.Path,
			Data:         updated,
			Precondition: storage.Preconditions{ExpectedRevision: snapshot.Revision},
			Mode:         snapshot.Mode,
		})
	}
	mutations = append(mutations, storage.Mutation{
		Path:         source.Path,
		Delete:       true,
		Precondition: storage.Preconditions{ExpectedRevision: sourceSnapshot.Revision},
	})
	return MovePlan{
		Source:     source.Logical,
		Target:     target.Logical,
		Changes:    changes,
		Risks:      []string{"external applications do not participate in obs-cli cooperative locks"},
		mutations:  mutations,
		targetPath: target.Path,
	}, nil
}

func (s *Service) ApplyMovePlan(plan MovePlan) (MoveResult, error) {
	if len(plan.mutations) == 0 || plan.targetPath == "" {
		return MoveResult{}, errors.New("move plan is incomplete")
	}
	if err := os.MkdirAll(filepath.Dir(plan.targetPath), 0o755); err != nil {
		return MoveResult{}, err
	}
	target, err := s.resolve(plan.Target, true)
	if err != nil {
		return MoveResult{}, err
	}
	if target.Path != plan.targetPath {
		return MoveResult{}, pathpolicy.ErrOutsideVault
	}
	result, err := s.store.ApplyTransaction(plan.mutations)
	if err != nil {
		var partial *storage.PartialFailureError
		if errors.As(err, &partial) {
			return MoveResult{}, &MovePartialFailure{
				TransactionID:   partial.TransactionID,
				Completed:       s.logicalPaths(partial.Completed),
				Failed:          s.logicalPaths(partial.Failed),
				RolledBack:      s.logicalPaths(partial.RolledBack),
				RollbackFailed:  s.logicalPaths(partial.RollbackFailed),
				RecoveryActions: partial.RecoveryActions,
			}
		}
		return MoveResult{}, err
	}
	return MoveResult{
		TransactionID: result.ID,
		Source:        plan.Source,
		Target:        plan.Target,
		Changes:       plan.Changes,
		RevisionAfter: result.Revisions[plan.targetPath],
	}, nil
}

func (s *Service) logicalPaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, filePath := range paths {
		logical, err := filepath.Rel(s.resolver.Root(), filePath)
		if err != nil || strings.HasPrefix(logical, "..") {
			result = append(result, "unresolved")
			continue
		}
		result = append(result, filepath.ToSlash(logical))
	}
	return result
}

func (s *Service) Move(source, target, expectedRevision string) (MoveResult, error) {
	plan, err := s.PlanMove(source, target, expectedRevision)
	if err != nil {
		return MoveResult{}, err
	}
	return s.ApplyMovePlan(plan)
}
