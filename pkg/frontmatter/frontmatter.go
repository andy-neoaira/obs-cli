package frontmatter

import (
	"errors"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

const (
	Delimiter               = "---" // YAML frontmatter 的起止分隔符
	InvalidFrontmatterError = "frontmatter contains invalid YAML"
)

// Parse 从笔记内容中提取并解析 YAML frontmatter。
// 返回 frontmatter 的 map 表示、正文内容、以及可能的错误。
func Parse(content string) (map[string]interface{}, string, error) {
	var fm map[string]interface{}
	rest, err := frontmatter.Parse(strings.NewReader(content), &fm)
	if err != nil {
		return nil, "", errors.New(InvalidFrontmatterError)
	}
	return fm, string(rest), nil
}

// HasFrontmatter 检查内容是否以 frontmatter 分隔符开头。
func HasFrontmatter(content string) bool {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return false
	}
	return strings.TrimSpace(lines[0]) == Delimiter
}

// SetKey 更新或添加 frontmatter 中的指定 key，返回更新后的完整内容。
// 如果笔记原本没有 frontmatter，会自动创建一个新的 frontmatter 块。
func SetKey(content, key, value string) (string, error) {
	parsedValue := parseValue(value)

	if !HasFrontmatter(content) {
		// 原笔记没有 frontmatter：新建一个 frontmatter 块并拼接到原内容前面
		fm := map[string]interface{}{key: parsedValue}
		fmStr, err := yaml.Marshal(fm)
		if err != nil {
			return "", err
		}
		return Delimiter + "\n" + string(fmStr) + Delimiter + "\n" + content, nil
	}

	// 解析已有 frontmatter
	fm, body, err := Parse(content)
	if err != nil {
		return "", err
	}

	if fm == nil {
		fm = make(map[string]interface{})
	}

	// 更新 key
	fm[key] = parsedValue

	// 重新组装内容
	fmStr, err := yaml.Marshal(fm)
	if err != nil {
		return "", err
	}

	return Delimiter + "\n" + string(fmStr) + Delimiter + "\n" + body, nil
}

// parseValue 尝试将字符串值解析为合适的 Go 类型。
// 支持布尔值、数组（方括号包裹的逗号分隔值）、字符串。
func parseValue(value string) interface{} {
	// 尝试布尔值
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// 尝试数组（方括号包裹的逗号分隔值）
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inner := value[1 : len(value)-1]
		if inner == "" {
			return []string{}
		}
		parts := strings.Split(inner, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			result = append(result, strings.TrimSpace(p))
		}
		return result
	}

	// 默认作为字符串返回
	return value
}
