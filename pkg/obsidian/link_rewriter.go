package obsidian

import (
	"bytes"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// LinkRewriter 专门负责 Obsidian 链接重写。
//
// 它和 Note 的职责不同：
//   - Note 负责文件级操作：读、写、移动、删除、搜索。
//   - LinkRewriter 负责内容级操作：生成链接替换规则、判断同名歧义、跳过代码块、批量改写。
//
// 把这部分单独拆出来后，后续如果要支持更完整的 Markdown parser、
// 相对路径重算、block link 或更复杂的 Obsidian 链接规则，不需要继续膨胀 note.go。
type LinkRewriter struct{}

// LinkRewriteManager 定义链接重写能力。
// Move 业务只需要这个接口，不应该依赖完整的 NoteManager。
type LinkRewriteManager interface {
	UpdateLinks(string, string, string) error
}

// UpdateLinks 遍历 vault 中所有 Markdown 笔记，将指向旧笔记的链接更新为新笔记的链接。
// 这是 move 命令的重要后续操作，保证笔记间的引用关系不会断裂。
func (r *LinkRewriter) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	excluded := ExcludedPaths(vaultPath)
	includeBaseLinks := r.shouldUpdateBasenameLinks(vaultPath, oldNoteName)
	replacements := r.GenerateReplacements(oldNoteName, newNoteName, LinkRewriteOptions{
		IncludeBaseLinks: includeBaseLinks,
	})

	err := filepath.Walk(vaultPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New(VaultAccessError)
		}

		if ShouldSkipDirectoryOrFile(info) {
			return nil
		}
		relPath, relErr := filepath.Rel(vaultPath, filePath)
		if relErr != nil {
			return errors.New(VaultAccessError)
		}
		if IsExcluded(relPath, excluded) {
			return nil
		}

		originalContent, err := os.ReadFile(filePath)
		if err != nil {
			return errors.New(VaultReadError)
		}

		// 执行替换时跳过 fenced code block，避免修改代码示例中的链接文本。
		updatedContent := r.ReplaceContentSkippingFencedCode(originalContent, replacements)

		// 如果没有实际变化，跳过写入以提高性能并保留文件修改时间。
		if bytes.Equal(originalContent, updatedContent) {
			return nil
		}

		if err := os.WriteFile(filePath, updatedContent, info.Mode()); err != nil {
			return errors.New(VaultWriteError)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// LinkRewriteOptions 控制链接替换规则的生成策略。
type LinkRewriteOptions struct {
	// IncludeBaseLinks 控制是否替换 [[basename]] 这类不带目录的链接。
	// 当 vault 中存在多个同名笔记时，basename 链接无法唯一确定目标，应关闭此选项。
	IncludeBaseLinks bool
}

// GenerateReplacements 创建移动笔记时需要替换的链接映射表。
// 它会处理各种 Obsidian 链接格式：简单 wikilink、路径 wikilink、Markdown 链接。
// 所有路径都会被归一化为正斜杠，以保证跨平台一致。
func (r *LinkRewriter) GenerateReplacements(oldNotePath, newNotePath string, options LinkRewriteOptions) map[string]string {
	replacements := make(map[string]string)

	// 将路径归一化为正斜杠，确保匹配 Obsidian 链接格式。
	oldNormalized := normalizePathSeparators(oldNotePath)
	newNormalized := normalizePathSeparators(newNotePath)

	// 取 basename（不含扩展名）和完整路径（不含扩展名）。
	oldBase := RemoveMdSuffix(path.Base(oldNormalized))
	newBase := RemoveMdSuffix(path.Base(newNormalized))
	oldPathNoExt := RemoveMdSuffix(oldNormalized)
	newPathNoExt := RemoveMdSuffix(newNormalized)

	// 1. 简单 wikilink（仅 basename）——仅在调用方确认无歧义时替换。
	if options.IncludeBaseLinks {
		replacements["[["+oldBase+"]]"] = "[[" + newBase + "]]"
		replacements["[["+oldBase+"|"] = "[[" + newBase + "|"
		replacements["[["+oldBase+"#"] = "[[" + newBase + "#"
	}

	// 2. 基于路径的 wikilink（路径与 basename 不同时）。
	if oldPathNoExt != oldBase {
		replacements["[["+oldPathNoExt+"]]"] = "[[" + newPathNoExt + "]]"
		replacements["[["+oldPathNoExt+"|"] = "[[" + newPathNoExt + "|"
		replacements["[["+oldPathNoExt+"#"] = "[[" + newPathNoExt + "#"
	}

	// 3. Markdown 链接（多种格式）。
	oldMd := AddMdSuffix(oldNormalized)
	newMd := AddMdSuffix(newNormalized)
	replacements["]("+oldMd+")"] = "](" + newMd + ")"
	replacements["]("+oldPathNoExt+")"] = "](" + newPathNoExt + ")"
	replacements["](./"+oldMd+")"] = "](./" + newMd + ")"
	replacements["](./"+oldPathNoExt+")"] = "](./" + newPathNoExt + ")"

	return replacements
}

// ReplaceContent 批量替换 content 中的字符串，使用 replacements map 中的键值对。
func (r *LinkRewriter) ReplaceContent(content []byte, replacements map[string]string) []byte {
	for oldText, newText := range replacements {
		content = bytes.ReplaceAll(content, []byte(oldText), []byte(newText))
	}
	return content
}

// ReplaceContentSkippingFencedCode 批量替换 Markdown 内容，但跳过 ``` 或 ~~~ 包裹的代码块。
// 这不能替代完整 Markdown parser，但能避免 move 命令修改代码示例中的链接文本，
// 对当前 CLI 的依赖体积和行为稳定性是更务实的折中。
func (r *LinkRewriter) ReplaceContentSkippingFencedCode(content []byte, replacements map[string]string) []byte {
	lines := bytes.SplitAfter(content, []byte("\n"))
	inFence := false
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~")) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		lines[i] = r.ReplaceContent(line, replacements)
	}
	return bytes.Join(lines, nil)
}

// shouldUpdateBasenameLinks 判断 move 时是否可以安全更新 [[basename]] 链接。
//
// 当 oldNoteName 本身就是纯文件名时，用户表达的就是 basename 级移动/重命名，可以更新。
// 当 oldNoteName 带目录时，如果 vault 内有多个同名 Markdown 文件，[[basename]] 无法唯一指向旧文件，
// 此时只更新 [[folder/name]] 和 Markdown 路径链接，避免误改其他同名笔记的引用。
func (r *LinkRewriter) shouldUpdateBasenameLinks(vaultPath, oldNoteName string) bool {
	oldNormalized := normalizePathSeparators(oldNoteName)
	oldPathNoExt := RemoveMdSuffix(oldNormalized)
	oldBase := RemoveMdSuffix(path.Base(oldNormalized))
	if oldPathNoExt == oldBase {
		return true
	}

	matches, err := r.countMarkdownFilesByBase(vaultPath, oldBase)
	if err != nil {
		return false
	}
	return matches <= 1
}

// countMarkdownFilesByBase 统计 vault 中同 basename 的 Markdown 文件数量。
// 统计遵守隐藏目录和 Obsidian ignore 配置，和用户可见的笔记集合保持一致。
func (r *LinkRewriter) countMarkdownFilesByBase(vaultPath, baseNoExt string) (int, error) {
	excluded := ExcludedPaths(vaultPath)
	target := AddMdSuffix(baseNoExt)
	count := 0
	err := filepath.WalkDir(vaultPath, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if isHiddenDir(d) {
			return filepath.SkipDir
		}
		relPath, err := filepath.Rel(vaultPath, filePath)
		if err != nil {
			return err
		}
		if relPath != "." && IsExcluded(relPath, excluded) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() && filepath.Base(filePath) == target {
			count++
		}
		return nil
	})
	return count, err
}

// normalizePathSeparators 将反斜杠转换为正斜杠，保证跨平台一致性。
// Obsidian 在所有操作系统中都使用正斜杠作为链接分隔符。
func normalizePathSeparators(notePath string) string {
	return strings.ReplaceAll(notePath, "\\", "/")
}
