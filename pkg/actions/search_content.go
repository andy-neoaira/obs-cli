package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

// 支持的输出格式常量
const (
	searchContentFormatText = "text"
	searchContentFormatJSON = "json"
)

// SearchContentOptions 定义了 search-content 命令的高级选项。
type SearchContentOptions struct {
	UseEditor           bool      // 是否使用编辑器打开
	EditorFlagExplicit  bool      // --editor 是否是用户显式传入的
	NoInteractive       bool      // 是否禁用交互式选择
	Format              string    // 输出格式：text 或 json
	InteractiveTerminal bool      // 当前是否在交互式终端中
	Output              io.Writer // 输出目标，默 os.Stdout
	Page                int       // 分页页码
	PageSize            int       // 每页结果数
}

// searchContentJSONMatch 是 JSON 输出时单条结果的数据结构。
type searchContentJSONMatch struct {
	File      string `json:"file"`       // 笔记相对路径
	Line      int    `json:"line"`       // 匹配行号
	Content   string `json:"content"`    // 匹配行的内容摘要
	MatchType string `json:"match_type"` // 匹配类型：filename 或 content
}

// searchContentPaginatedJSON 是带分页信息的 JSON 输出结构。
type searchContentPaginatedJSON struct {
	Page            int                      `json:"page"`
	PageSize        int                      `json:"page_size"`
	TotalPages      int                      `json:"total_pages"`
	TotalResults    int                      `json:"total_results"`
	ReturnedResults int                      `json:"returned_results"`
	HasMore         bool                     `json:"has_more"`
	Results         []searchContentJSONMatch `json:"results"`
}

// 分页默认值与上限
const (
	defaultPageSize = 25
	maxPageSize     = 100
)

// SearchNotesContentWithOptions 是 "search-content" 命令的完整业务核心。
// 支持交互式模糊选择、纯文本输出、JSON 输出、分页等多种模式。
func SearchNotesContentWithOptions(vault obsidian.VaultManager, note obsidian.NoteManager, uri obsidian.UriManager, fuzzyFinder obsidian.FuzzyFinderManager, searchTerm string, options SearchContentOptions) error {
	// 规范化并校验输出格式
	format, err := normalizeSearchContentFormat(options.Format)
	if err != nil {
		return err
	}

	nonInteractiveMode := shouldUseNonInteractiveMode(options, format)
	useEditor := options.UseEditor

	// 非交互模式下不允许使用 --editor（因为无法选择具体打开哪一个）
	if nonInteractiveMode && options.EditorFlagExplicit && options.UseEditor {
		return errors.New("--editor cannot be used with non-interactive search-content output")
	}

	// 如果 editor 来自配置默认值而非显式 flag，在非交互模式下优先输出到 stdout（便于脚本使用）
	if nonInteractiveMode {
		useEditor = false
	}

	output := options.Output
	if output == nil {
		output = os.Stdout
	}

	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	// 执行全文搜索，返回所有匹配的笔记片段
	matches, err := note.SearchNotesWithSnippets(vaultPath, searchTerm)
	if err != nil {
		return err
	}

	// 非交互模式：直接打印结果
	if nonInteractiveMode {
		return printMatches(matches, searchTerm, format, output, options)
	}

	// 交互模式：无结果、单结果、多结果分别处理
	if len(matches) == 0 {
		_, _ = fmt.Fprintf(output, "No notes found containing '%s'\n", searchTerm)
		return nil
	}

	// 只有一条结果时直接打开，无需让用户选择
	if len(matches) == 1 {
		_, _ = fmt.Fprintf(output, "Opening note: %s\n", matches[0].FilePath)
		if useEditor {
			filePath := filepath.Join(vaultPath, matches[0].FilePath)
			return obsidian.OpenInEditor(filePath)
		}
		obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
			"file":  matches[0].FilePath,
			"vault": vaultName,
		})
		return uri.Execute(obsidianUri)
	}

	// 多条结果：启动模糊搜索界面，让用户选择
	displayItems := formatMatchesForDisplay(matches)

	index, err := fuzzyFinder.Find(displayItems, func(i int) string {
		return displayItems[i]
	})
	if err != nil {
		return err
	}

	selectedMatch := matches[index]
	if useEditor {
		filePath := filepath.Join(vaultPath, selectedMatch.FilePath)
		_, _ = fmt.Fprintf(output, "Opening note: %s\n", selectedMatch.FilePath)
		return obsidian.OpenInEditor(filePath)
	}
	obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  selectedMatch.FilePath,
		"vault": vaultName,
	})
	return uri.Execute(obsidianUri)
}

// shouldUseNonInteractiveMode 判断是否应进入非交互模式。
func shouldUseNonInteractiveMode(options SearchContentOptions, format string) bool {
	if options.NoInteractive {
		return true
	}
	if format == searchContentFormatJSON {
		return true
	}
	if isPaginationRequested(options) {
		return true
	}
	return !options.InteractiveTerminal
}

// normalizeSearchContentFormat 规范化并校验输出格式字符串。
func normalizeSearchContentFormat(format string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(format))
	if trimmed == "" {
		return searchContentFormatText, nil
	}

	switch trimmed {
	case searchContentFormatText, searchContentFormatJSON:
		return trimmed, nil
	default:
		return "", fmt.Errorf("invalid format '%s': expected one of text, json", format)
	}
}

// paginationResult 保存分页后的结果和元数据。
type paginationResult struct {
	items      []obsidian.NoteMatch // 当前页的数据
	page       int                  // 当前页码
	pageSize   int                  // 每页大小
	totalPages int                  // 总页数
	hasMore    bool                 // 是否还有更多页
}

// paginateMatches 对搜索结果进行分页。
// 注意：当请求页码超过总页数时，保留用户请求的页码并返回空结果，
// 而不是静默夹到最后一页。这样脚本调用方可以准确判断分页已经结束。
func paginateMatches(matches []obsidian.NoteMatch, options SearchContentOptions) paginationResult {
	page := options.Page
	pageSize := options.PageSize

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	total := len(matches)
	totalPages := (total + pageSize - 1) / pageSize // 向上取整计算总页数
	if totalPages == 0 {
		totalPages = 1
	}

	start := (page - 1) * pageSize
	if start >= total {
		return paginationResult{page: page, pageSize: pageSize, totalPages: totalPages}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return paginationResult{
		items:      matches[start:end],
		page:       page,
		pageSize:   pageSize,
		totalPages: totalPages,
		hasMore:    end < total,
	}
}

// isPaginationRequested 判断用户是否显式请求了分页。
func isPaginationRequested(options SearchContentOptions) bool {
	return options.Page > 0 || options.PageSize > 0
}

// toJSONMatches 将内部 NoteMatch 切片转换为 JSON 输出用的结构体切片。
func toJSONMatches(matches []obsidian.NoteMatch) []searchContentJSONMatch {
	result := make([]searchContentJSONMatch, 0, len(matches))
	for _, match := range matches {
		result = append(result, searchContentJSONMatch{
			File:      match.FilePath,
			Line:      match.LineNumber,
			Content:   match.MatchLine,
			MatchType: getMatchType(match),
		})
	}
	return result
}

// printMatches 将搜索结果按指定格式输出到 output。
func printMatches(matches []obsidian.NoteMatch, searchTerm string, format string, output io.Writer, options SearchContentOptions) error {
	paginate := isPaginationRequested(options)

	switch format {
	case searchContentFormatText:
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "No notes found containing '%s'\n", searchTerm)
			return nil
		}

		if paginate {
			pg := paginateMatches(matches, options)
			for _, match := range pg.items {
				_, _ = fmt.Fprintln(output, formatMatchForList(match))
			}
			_, _ = fmt.Fprintf(output, "-- Page %d/%d (%d of %d results) --\n", pg.page, pg.totalPages, len(pg.items), len(matches))
			return nil
		}

		for _, match := range matches {
			_, _ = fmt.Fprintln(output, formatMatchForList(match))
		}
		return nil
	case searchContentFormatJSON:
		if paginate {
			pg := paginateMatches(matches, options)
			result := toJSONMatches(pg.items)
			encoder := json.NewEncoder(output)
			encoder.SetEscapeHTML(false)
			return encoder.Encode(searchContentPaginatedJSON{
				Page:            pg.page,
				PageSize:        pg.pageSize,
				TotalPages:      pg.totalPages,
				TotalResults:    len(matches),
				ReturnedResults: len(result),
				HasMore:         pg.hasMore,
				Results:         result,
			})
		}

		encoder := json.NewEncoder(output)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(toJSONMatches(matches))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// formatMatchForList 将单条匹配结果格式化为适合终端打印的字符串。
func formatMatchForList(match obsidian.NoteMatch) string {
	if match.LineNumber > 0 {
		return fmt.Sprintf("%s:%d: %s", match.FilePath, match.LineNumber, match.MatchLine)
	}
	return fmt.Sprintf("%s: %s", match.FilePath, match.MatchLine)
}

// getMatchType 判断匹配类型是文件名匹配还是内容匹配。
func getMatchType(match obsidian.NoteMatch) string {
	if match.LineNumber == 0 {
		return "filename"
	}
	return "content"
}

// formatMatchesForDisplay 将所有匹配结果格式化为模糊搜索界面显示的字符串列表。
func formatMatchesForDisplay(matches []obsidian.NoteMatch) []string {
	maxPathLength := calculateMaxPathLength(matches)

	var displayItems []string
	for _, match := range matches {
		displayStr := formatSingleMatch(match, maxPathLength)
		displayItems = append(displayItems, displayStr)
	}

	return displayItems
}

// calculateMaxPathLength 计算所有匹配结果中最长的路径长度，用于对齐输出。
func calculateMaxPathLength(matches []obsidian.NoteMatch) int {
	maxLength := 0
	for _, match := range matches {
		pathWithLine := formatPathWithLine(match)
		if len(pathWithLine) > maxLength {
			maxLength = len(pathWithLine)
		}
	}
	return maxLength
}

// formatPathWithLine 将文件路径与行号组合成显示字符串。
func formatPathWithLine(match obsidian.NoteMatch) string {
	if match.LineNumber > 0 {
		return fmt.Sprintf("%s:%d", match.FilePath, match.LineNumber)
	}
	return match.FilePath
}

// formatSingleMatch 将单条匹配结果格式化为对齐的显示字符串。
func formatSingleMatch(match obsidian.NoteMatch, maxPathLength int) string {
	pathWithLine := formatPathWithLine(match)
	if match.LineNumber == 0 {
		// 文件名匹配：显示路径并标注
		return fmt.Sprintf("%-*s | %s", maxPathLength, pathWithLine, match.MatchLine)
	}
	// 内容匹配：显示路径:行号 | 内容摘要
	return fmt.Sprintf("%-*s | %s", maxPathLength, pathWithLine, match.MatchLine)
}
