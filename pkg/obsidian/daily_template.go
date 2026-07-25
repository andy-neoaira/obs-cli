package obsidian

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var templateVariablePattern = regexp.MustCompile(`\{\{([^{}]+)\}\}`)

type momentToken struct {
	moment string
	goFmt  string
}

var sharedMomentTokens = []momentToken{
	{"YYYY", "2006"},
	{"MMMM", "January"},
	{"dddd", "Monday"},
	{"MMM", "Jan"},
	{"ddd", "Mon"},
	{"YY", "06"},
	{"MM", "01"},
	{"DD", "02"},
	{"HH", "15"},
	{"hh", "03"},
	{"mm", "04"},
	{"ss", "05"},
	{"A", "PM"},
	{"a", "pm"},
}

// ParseMomentToGoFormat converts the shared Obsidian/Moment subset and rejects
// unsupported alphabetic tokens instead of emitting a misleading filename.
func ParseMomentToGoFormat(format string) (string, error) {
	if format == "" {
		return "", fmt.Errorf("daily date format must not be empty")
	}

	var result strings.Builder
	for index := 0; index < len(format); {
		if format[index] == '[' {
			end := strings.IndexByte(format[index+1:], ']')
			if end < 0 {
				return "", fmt.Errorf("daily date format contains an unclosed literal")
			}
			end += index + 1
			result.WriteString(format[index+1 : end])
			index = end + 1
			continue
		}

		char := format[index]
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			matched := false
			for _, token := range sharedMomentTokens {
				if strings.HasPrefix(format[index:], token.moment) {
					result.WriteString(token.goFmt)
					index += len(token.moment)
					matched = true
					break
				}
			}
			if !matched {
				end := index + 1
				for end < len(format) {
					next := format[end]
					if !((next >= 'A' && next <= 'Z') || (next >= 'a' && next <= 'z')) {
						break
					}
					end++
				}
				return "", fmt.Errorf("unsupported Moment date token %q", format[index:end])
			}
			continue
		}

		result.WriteByte(char)
		index++
	}
	return result.String(), nil
}

// RenderDailyTemplate renders the shared vault-contract/v1 variables.
// Unknown variables remain byte-for-byte intact and are returned as warnings.
func RenderDailyTemplate(content string, now time.Time, dateFormat, title string) (string, []string, error) {
	layout, err := ParseMomentToGoFormat(dateFormat)
	if err != nil {
		return "", nil, err
	}
	values := map[string]string{
		"date":      now.Format(layout),
		"time":      now.Format("15:04"),
		"title":     title,
		"filename":  title,
		"yesterday": now.AddDate(0, 0, -1).Format(layout),
		"tomorrow":  now.AddDate(0, 0, 1).Format(layout),
	}

	warnings := make([]string, 0)
	warned := make(map[string]bool)
	var renderErr error
	rendered := templateVariablePattern.ReplaceAllStringFunc(content, func(variable string) string {
		expression := variable[2 : len(variable)-2]
		if value, ok := values[strings.ToLower(expression)]; ok {
			return value
		}
		if len(expression) > 5 && strings.EqualFold(expression[:5], "date:") {
			customLayout, parseErr := ParseMomentToGoFormat(expression[5:])
			if parseErr != nil {
				renderErr = parseErr
				return variable
			}
			return now.Format(customLayout)
		}
		if !warned[variable] {
			warned[variable] = true
			warnings = append(warnings, "preserved unknown template variable "+variable)
		}
		return variable
	})
	if renderErr != nil {
		return "", warnings, renderErr
	}
	return rendered, warnings, nil
}
