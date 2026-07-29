package obsidian

import (
	"bytes"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// LinkRewriter rewrites structured Obsidian links for the current transactional
// note move service. File discovery, preconditions, and writes stay in noteops.
type LinkRewriter struct{}

type LinkEdit struct {
	Kind   string `json:"kind"`
	Before string `json:"before"`
	After  string `json:"after"`
}

var (
	wikilinkPattern     = regexp.MustCompile(`\[\[([^\]|#]+)((?:#[^\]|]+)?(?:\|[^\]]+)?)\]\]`)
	markdownLinkPattern = regexp.MustCompile(`(\[[^\]]*\]\()([^\s)]+)(\))`)
)

// RewriteStructuredLinks only rewrites parsed Wikilink and Markdown link
// destinations. It preserves aliases/fragments and skips frontmatter, fenced
// code, inline code, and HTML comments.
func (r *LinkRewriter) RewriteStructuredLinks(
	content []byte,
	containingNote, oldNote, newNote string,
	includeBaseLinks bool,
) ([]byte, []LinkEdit, error) {
	lines := bytes.SplitAfter(content, []byte("\n"))
	inFence := false
	inComment := false
	inFrontmatter := len(lines) != 0 && string(bytes.TrimSpace(lines[0])) == "---"
	frontmatterClosed := !inFrontmatter
	edits := make([]LinkEdit, 0)
	for index, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if inFrontmatter && index == 0 {
			continue
		}
		if inFrontmatter && !frontmatterClosed {
			if string(trimmed) == "---" {
				frontmatterClosed = true
				inFrontmatter = false
			}
			continue
		}
		if bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~")) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		rewritten, lineEdits, commentState, err := rewriteLinkLine(
			line, containingNote, oldNote, newNote, includeBaseLinks, inComment,
		)
		if err != nil {
			return nil, nil, err
		}
		inComment = commentState
		lines[index] = rewritten
		edits = append(edits, lineEdits...)
	}
	return bytes.Join(lines, nil), edits, nil
}

func rewriteLinkLine(
	line []byte,
	containingNote, oldNote, newNote string,
	includeBaseLinks, inComment bool,
) ([]byte, []LinkEdit, bool, error) {
	var output bytes.Buffer
	edits := make([]LinkEdit, 0)
	for offset := 0; offset < len(line); {
		if inComment {
			end := bytes.Index(line[offset:], []byte("-->"))
			if end < 0 {
				output.Write(line[offset:])
				return output.Bytes(), edits, true, nil
			}
			end += offset + 3
			output.Write(line[offset:end])
			offset = end
			inComment = false
			continue
		}
		if bytes.HasPrefix(line[offset:], []byte("<!--")) {
			inComment = true
			continue
		}
		if line[offset] == '`' {
			run := 1
			for offset+run < len(line) && line[offset+run] == '`' {
				run++
			}
			delimiter := bytes.Repeat([]byte{'`'}, run)
			end := bytes.Index(line[offset+run:], delimiter)
			if end < 0 {
				output.Write(line[offset:])
				return output.Bytes(), edits, inComment, nil
			}
			end += offset + run*2
			output.Write(line[offset:end])
			offset = end
			continue
		}
		next := len(line)
		for _, marker := range [][]byte{[]byte("<!--"), []byte("`")} {
			if found := bytes.Index(line[offset:], marker); found >= 0 && offset+found < next {
				next = offset + found
			}
		}
		segment, segmentEdits, err := rewriteLinkSyntax(
			string(line[offset:next]), containingNote, oldNote, newNote, includeBaseLinks,
		)
		if err != nil {
			return nil, nil, inComment, err
		}
		output.WriteString(segment)
		edits = append(edits, segmentEdits...)
		offset = next
	}
	return output.Bytes(), edits, inComment, nil
}

func rewriteLinkSyntax(
	text, containingNote, oldNote, newNote string,
	includeBaseLinks bool,
) (string, []LinkEdit, error) {
	edits := make([]LinkEdit, 0)
	text = wikilinkPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := wikilinkPattern.FindStringSubmatch(match)
		target := removeMarkdownSuffix(normalizePathSeparators(parts[1]))
		oldID := removeMarkdownSuffix(normalizePathSeparators(oldNote))
		replacement := ""
		if includeBaseLinks && !strings.Contains(target, "/") && target == path.Base(oldID) {
			replacement = path.Base(removeMarkdownSuffix(normalizePathSeparators(newNote)))
		} else if target == oldID {
			replacement = removeMarkdownSuffix(normalizePathSeparators(newNote))
		}
		if replacement == "" {
			return match
		}
		updated := "[[" + replacement + parts[2] + "]]"
		edits = append(edits, LinkEdit{Kind: "wikilink", Before: match, After: updated})
		return updated
	})
	var rewriteErr error
	text = markdownLinkPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := markdownLinkPattern.FindStringSubmatch(match)
		destination := parts[2]
		updatedDestination, changed, err := rewriteMarkdownDestination(
			destination, containingNote, oldNote, newNote,
		)
		if err != nil {
			rewriteErr = err
			return match
		}
		if !changed {
			return match
		}
		updated := parts[1] + updatedDestination + parts[3]
		edits = append(edits, LinkEdit{Kind: "markdown", Before: match, After: updated})
		return updated
	})
	return text, edits, rewriteErr
}

func rewriteMarkdownDestination(destination, containingNote, oldNote, newNote string) (string, bool, error) {
	if strings.HasPrefix(destination, "/") || strings.Contains(strings.Split(destination, "/")[0], ":") {
		return destination, false, nil
	}
	pathPart, fragment, _ := strings.Cut(destination, "#")
	decoded, err := url.PathUnescape(pathPart)
	if err != nil {
		return "", false, err
	}
	hadExtension := strings.EqualFold(path.Ext(decoded), ".md")
	containingDir := path.Dir(normalizePathSeparators(containingNote))
	resolved := path.Clean(path.Join(containingDir, decoded))
	oldWithExtension := addMarkdownSuffix(normalizePathSeparators(oldNote))
	if resolved != oldWithExtension && addMarkdownSuffix(resolved) != oldWithExtension {
		return destination, false, nil
	}
	relative, err := filepath.Rel(
		filepath.FromSlash(containingDir),
		filepath.FromSlash(addMarkdownSuffix(normalizePathSeparators(newNote))),
	)
	if err != nil {
		return "", false, err
	}
	relative = filepath.ToSlash(relative)
	if !hadExtension {
		relative = removeMarkdownSuffix(relative)
	}
	if strings.HasPrefix(pathPart, "./") && !strings.HasPrefix(relative, ".") {
		relative = "./" + relative
	}
	segments := strings.Split(relative, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	updated := strings.Join(segments, "/")
	if fragment != "" {
		updated += "#" + fragment
	}
	return updated, true, nil
}

func addMarkdownSuffix(value string) string {
	if strings.HasSuffix(value, ".md") {
		return value
	}
	return value + ".md"
}

func removeMarkdownSuffix(value string) string {
	return strings.TrimSuffix(value, ".md")
}

// normalizePathSeparators 将反斜杠转换为正斜杠，保证跨平台一致性。
// Obsidian 在所有操作系统中都使用正斜杠作为链接分隔符。
func normalizePathSeparators(notePath string) string {
	return strings.ReplaceAll(notePath, "\\", "/")
}
