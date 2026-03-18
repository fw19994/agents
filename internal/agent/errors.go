package agent

import (
	"errors"
	"strings"
)

// FormatErrorChain 将错误及其 Unwrap 链串联，便于展示模型 API、HTTP 等底层返回信息。
func FormatErrorChain(err error) string {
	if err == nil {
		return ""
	}
	var parts []string
	seen := make(map[string]bool)
	for e := err; e != nil; e = errors.Unwrap(e) {
		s := strings.TrimSpace(e.Error())
		if s != "" && !seen[s] {
			seen[s] = true
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return err.Error()
	}
	return strings.Join(parts, " | ")
}
