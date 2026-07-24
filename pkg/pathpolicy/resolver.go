// Package pathpolicy resolves Vault logical paths without allowing filesystem
// traversal or symbolic-link escape.
package pathpolicy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	CodeOutsideVault  = "PATH_OUTSIDE_VAULT"
	RuleLogicalPath   = "VC-1.2"
	RuleCanonicalPath = "VC-1.3"
	RuleSymlink       = "VC-1.4"
	RuleHiddenPath    = "VC-2.3"
)

var ErrOutsideVault = errors.New(CodeOutsideVault)

type ResolveOptions struct {
	AllowRoot   bool
	AllowHidden bool
}

type Result struct {
	// Logical is the normalized protocol path and always uses '/'.
	Logical string
	// Path is the canonical filesystem target. Existing symbolic links are
	// resolved; for a new target, its nearest existing parent is resolved.
	Path string
	// Exists reports whether the final target existed during resolution.
	Exists bool
	// ThroughSymlink reports that at least one target path component resolved
	// to a different physical path. Reads may allow this when it stays inside
	// the Vault; mutations must require an unambiguous canonical target.
	ThroughSymlink bool
}

type Error struct {
	Code   string
	Rule   string
	Input  string
	Reason string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s): %s: %q", e.Code, e.Rule, e.Reason, e.Input)
}

func (e *Error) Unwrap() error {
	return ErrOutsideVault
}

type Resolver struct {
	root string
}

func NewResolver(vaultRoot string) (*Resolver, error) {
	absolute, err := filepath.Abs(vaultRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve vault root: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, fmt.Errorf("resolve vault root symlinks: %w", err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return nil, fmt.Errorf("stat vault root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault root is not a directory")
	}
	return &Resolver{root: filepath.Clean(canonical)}, nil
}

func (r *Resolver) Root() string {
	return r.root
}

func (r *Resolver) Resolve(input string, options ResolveOptions) (Result, error) {
	logical, err := validateLogicalPath(input, options)
	if err != nil {
		return Result{}, err
	}
	if logical == "" {
		return Result{Path: r.root, Exists: true}, nil
	}

	candidate := filepath.Join(r.root, filepath.FromSlash(logical))
	resolved, exists, throughSymlink, err := r.resolveFilesystem(candidate, input)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Logical:        logical,
		Path:           resolved,
		Exists:         exists,
		ThroughSymlink: throughSymlink,
	}, nil
}

func validateLogicalPath(input string, options ResolveOptions) (string, error) {
	reject := func(rule, reason string) (string, error) {
		return "", &Error{Code: CodeOutsideVault, Rule: rule, Input: input, Reason: reason}
	}

	if strings.ContainsRune(input, '\x00') {
		return reject(RuleLogicalPath, "NUL is forbidden")
	}
	if input == "" {
		if options.AllowRoot {
			return "", nil
		}
		return reject(RuleLogicalPath, "empty logical path is forbidden")
	}
	if strings.HasPrefix(input, "~") {
		return reject(RuleLogicalPath, "home-relative path is forbidden")
	}
	if filepath.IsAbs(input) || strings.HasPrefix(input, "/") || strings.HasPrefix(input, `\`) {
		return reject(RuleLogicalPath, "absolute path is forbidden")
	}
	if hasWindowsVolume(input) {
		return reject(RuleLogicalPath, "Windows volume or UNC path is forbidden")
	}

	logical := strings.ReplaceAll(input, `\`, "/")
	segments := strings.Split(logical, "/")
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return reject(RuleLogicalPath, "empty, dot, and parent segments are forbidden")
		}
		if strings.Contains(segment, ":") {
			return reject(RuleLogicalPath, "alternate data stream syntax is forbidden")
		}
		if !options.AllowHidden && strings.HasPrefix(segment, ".") {
			return reject(RuleHiddenPath, "hidden path requires an explicit audited capability")
		}
	}
	return strings.Join(segments, "/"), nil
}

func (r *Resolver) resolveFilesystem(candidate, logicalInput string) (string, bool, bool, error) {
	if _, err := os.Lstat(candidate); err == nil {
		canonical, evalErr := filepath.EvalSymlinks(candidate)
		if evalErr != nil {
			return "", false, false, fmt.Errorf("resolve target symlinks: %w", evalErr)
		}
		if !withinRoot(r.root, canonical) {
			return "", false, false, outsideError(logicalInput, RuleSymlink, "symbolic link resolves outside Vault")
		}
		canonical = filepath.Clean(canonical)
		return canonical, true, filepath.Clean(candidate) != canonical, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, false, fmt.Errorf("inspect target path: %w", err)
	}

	ancestor := candidate
	missing := make([]string, 0, 2)
	for {
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", false, false, outsideError(logicalInput, RuleCanonicalPath, "no existing parent inside Vault")
		}
		missing = append(missing, filepath.Base(ancestor))
		ancestor = parent
		if _, err := os.Lstat(ancestor); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", false, false, fmt.Errorf("inspect target parent: %w", err)
		}
	}

	canonicalParent, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", false, false, fmt.Errorf("resolve target parent symlinks: %w", err)
	}
	if !withinRoot(r.root, canonicalParent) {
		return "", false, false, outsideError(logicalInput, RuleSymlink, "nearest existing parent resolves outside Vault")
	}
	throughSymlink := filepath.Clean(ancestor) != filepath.Clean(canonicalParent)
	for index := len(missing) - 1; index >= 0; index-- {
		canonicalParent = filepath.Join(canonicalParent, missing[index])
	}
	return filepath.Clean(canonicalParent), false, throughSymlink, nil
}

func withinRoot(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func hasWindowsVolume(path string) bool {
	if strings.HasPrefix(path, `\\`) || strings.HasPrefix(path, "//") {
		return true
	}
	return len(path) >= 2 &&
		((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) &&
		path[1] == ':'
}

func outsideError(input, rule, reason string) error {
	return &Error{Code: CodeOutsideVault, Rule: rule, Input: input, Reason: reason}
}
