package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

const Version = "obs-cli/v2"

type Code string

const (
	InvalidArgument       Code = "INVALID_ARGUMENT"
	InvalidFrontmatter    Code = "INVALID_FRONTMATTER"
	VaultNotFound         Code = "VAULT_NOT_FOUND"
	NoteNotFound          Code = "NOTE_NOT_FOUND"
	AlreadyExists         Code = "ALREADY_EXISTS"
	AmbiguousNote         Code = "AMBIGUOUS_NOTE"
	RevisionConflict      Code = "REVISION_CONFLICT"
	PathOutsideVault      Code = "PATH_OUTSIDE_VAULT"
	PartialFailure        Code = "PARTIAL_FAILURE"
	CapabilityUnsupported Code = "CAPABILITY_UNSUPPORTED"
	InternalError         Code = "INTERNAL_ERROR"
)

type DomainError struct {
	Code      Code           `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details"`
	cause     error
}

func (e *DomainError) Error() string {
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.cause
}

func NewError(code Code, message string, details map[string]any) *DomainError {
	if details == nil {
		details = map[string]any{}
	}
	return &DomainError{
		Code:      code,
		Message:   message,
		Retryable: code == RevisionConflict,
		Details:   details,
	}
}

func Wrap(code Code, message string, cause error, details map[string]any) *DomainError {
	result := NewError(code, message, details)
	result.cause = cause
	return result
}

type Warning struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Envelope struct {
	ProtocolVersion string       `json:"protocol_version"`
	OK              bool         `json:"ok"`
	Operation       string       `json:"operation"`
	RequestID       string       `json:"request_id"`
	Data            any          `json:"data,omitempty"`
	Error           *DomainError `json:"error,omitempty"`
	Warnings        []Warning    `json:"warnings"`
}

func Success(operation, requestID string, data any, warnings []Warning) Envelope {
	if warnings == nil {
		warnings = []Warning{}
	}
	return Envelope{
		ProtocolVersion: Version,
		OK:              true,
		Operation:       operation,
		RequestID:       requestID,
		Data:            data,
		Warnings:        warnings,
	}
}

func Failure(operation, requestID string, err error, warnings []Warning) Envelope {
	if warnings == nil {
		warnings = []Warning{}
	}
	return Envelope{
		ProtocolVersion: Version,
		OK:              false,
		Operation:       operation,
		RequestID:       requestID,
		Error:           MapError(err),
		Warnings:        warnings,
	}
}

func Render(writer io.Writer, envelope Envelope) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(envelope)
}

func MapError(err error) *DomainError {
	var domain *DomainError
	if errors.As(err, &domain) {
		return domain
	}
	var createConflict *noteops.CreateConflict
	if errors.As(err, &createConflict) {
		return Wrap(AlreadyExists, "The note already exists.", err, map[string]any{
			"path":               createConflict.Path,
			"existing_revision":  createConflict.ExistingRevision,
			"requested_revision": createConflict.RequestedRevision,
			"same_content":       createConflict.ExistingRevision == createConflict.RequestedRevision,
		})
	}
	var invalidFrontmatter *noteops.InvalidFrontmatterError
	if errors.As(err, &invalidFrontmatter) {
		return Wrap(InvalidFrontmatter, "The note contains invalid frontmatter.", err, map[string]any{
			"path":     invalidFrontmatter.Path,
			"revision": invalidFrontmatter.Revision,
		})
	}
	var partial *storage.PartialFailureError
	if errors.As(err, &partial) {
		return Wrap(PartialFailure, "The operation requires manual recovery.", err, map[string]any{
			"transaction_id":   partial.TransactionID,
			"completed":        partial.Completed,
			"failed":           partial.Failed,
			"rolled_back":      partial.RolledBack,
			"rollback_failed":  partial.RollbackFailed,
			"recovery_actions": partial.RecoveryActions,
		})
	}
	var movePartial *noteops.MovePartialFailure
	if errors.As(err, &movePartial) {
		return Wrap(PartialFailure, "The move requires manual recovery.", err, map[string]any{
			"transaction_id":   movePartial.TransactionID,
			"completed":        movePartial.Completed,
			"failed":           movePartial.Failed,
			"rolled_back":      movePartial.RolledBack,
			"rollback_failed":  movePartial.RollbackFailed,
			"recovery_actions": movePartial.RecoveryActions,
		})
	}
	switch {
	case errors.Is(err, storage.ErrRevisionConflict):
		return Wrap(RevisionConflict, "The target changed after it was read.", err, nil)
	case errors.Is(err, noteops.ErrPatchContextMismatch), errors.Is(err, noteops.ErrSectionNotFound):
		return Wrap(RevisionConflict, "The requested content context no longer matches.", err, nil)
	case errors.Is(err, noteops.ErrPatchContextAmbiguous), errors.Is(err, noteops.ErrSectionAmbiguous):
		return Wrap(AmbiguousNote, "The requested content context is not unique.", err, nil)
	case errors.Is(err, storage.ErrAlreadyExists), errors.Is(err, obsidian.ErrVaultAlreadyExists), errors.Is(err, obsidian.ErrVaultNameConflict):
		return Wrap(AlreadyExists, "The target already exists.", err, nil)
	case errors.Is(err, storage.ErrPartialFailure):
		return Wrap(PartialFailure, "The operation could not be completely applied or rolled back.", err, nil)
	case errors.Is(err, pathpolicy.ErrOutsideVault):
		return Wrap(PathOutsideVault, "The path is outside the Vault policy boundary.", err, nil)
	case errors.Is(err, obsidian.ErrVaultNotFound):
		return Wrap(VaultNotFound, "The Vault was not found.", err, nil)
	case errors.Is(err, noteops.ErrNoteNotFound):
		return Wrap(NoteNotFound, "The note was not found.", err, nil)
	case errors.Is(err, noteops.ErrInvalidFrontmatter):
		return Wrap(InvalidFrontmatter, "The note contains invalid frontmatter.", err, nil)
	case errors.Is(err, noteops.ErrRevisionRequired), errors.Is(err, noteops.ErrInvalidRevision), errors.Is(err, storage.ErrPrecondition):
		return Wrap(InvalidArgument, "A required write precondition is missing.", err, nil)
	default:
		return Wrap(InternalError, "The operation failed due to an internal error.", err, nil)
	}
}

func ExitCodeFor(err error) int {
	switch MapError(err).Code {
	case InvalidArgument, InvalidFrontmatter:
		return 2
	case VaultNotFound, NoteNotFound:
		return 3
	case AlreadyExists, AmbiguousNote, RevisionConflict:
		return 4
	case PathOutsideVault:
		return 5
	case PartialFailure:
		return 7
	case CapabilityUnsupported:
		return 8
	default:
		return 10
	}
}

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

func ResolveRequestID(input string) (string, error) {
	if input != "" {
		if !requestIDPattern.MatchString(input) {
			return "", NewError(InvalidArgument, "request ID contains invalid characters or length", map[string]any{
				"field": "request_id",
			})
		}
		return input, nil
	}
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}
	return "req-" + hex.EncodeToString(value), nil
}
