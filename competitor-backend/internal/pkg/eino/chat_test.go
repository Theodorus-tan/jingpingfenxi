package eino

import (
	"strings"
	"testing"
	"time"
)

func TestBuildChatSystemPromptIncludesConciseRulesAndTime(t *testing.T) {
	prompt := buildChatSystemPrompt("报告内容", time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))

	requiredSnippets := []string{
		"今天是 2026-06-01",
		"默认使用简明模式",
		"默认控制在 150 到 300 字",
		"报告未提供",
		"报告内容",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("prompt missing snippet %q", snippet)
		}
	}
}
