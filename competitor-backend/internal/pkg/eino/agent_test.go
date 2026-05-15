package eino

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildSystemPromptAtIncludesTimeAndIntermediateSections(t *testing.T) {
	now := time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)

	prompt := buildSystemPromptAt("Product_Improvement", "review", now)

	requiredSnippets := []string{
		"今天是 2026-01-15",
		"当前处于 2026 年第 1 季度",
		"你不是在写最终面向用户的长报告",
		"## 发现",
		"## 风险",
		"## 机会",
		"## 证据",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("prompt missing snippet %q", snippet)
		}
	}
}

func TestBuildFinalReportPayloadIncludesSummaryAndDimensions(t *testing.T) {
	generatedAt := time.Date(2026, time.March, 2, 8, 30, 0, 0, time.UTC)
	reports := map[string]string{
		"review":  "review report",
		"macro":   "macro report",
		"finance": "finance report",
	}
	evidencePool := []SharedEvidenceItem{
		{Dimension: "review", Title: "来源 1", Snippet: "用户抱怨", URL: "https://example.com/1"},
	}

	raw, err := buildFinalReportPayload("Mac mini", "Product_Improvement", "# 一句话结论\n- 可打", reports, evidencePool, generatedAt)
	if err != nil {
		t.Fatalf("buildFinalReportPayload returned error: %v", err)
	}

	var payload FinalReportPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("payload should be valid json: %v", err)
	}

	if payload.Version != 3 {
		t.Fatalf("expected version 3, got %d", payload.Version)
	}
	if payload.Competitor != "Mac mini" {
		t.Fatalf("unexpected competitor: %s", payload.Competitor)
	}
	if payload.Summary == "" || payload.Review == "" || payload.Macro == "" || payload.Finance == "" {
		t.Fatalf("payload should contain summary and all dimension reports: %+v", payload)
	}
	if payload.GeneratedAt != generatedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected generatedAt: %s", payload.GeneratedAt)
	}
	if len(payload.EvidencePool) != 1 || payload.EvidencePool[0].Title != "来源 1" {
		t.Fatalf("payload should include evidence pool: %+v", payload.EvidencePool)
	}
}

func TestBuildFallbackExecutiveSummaryPrioritizesReview(t *testing.T) {
	summary := buildFallbackExecutiveSummary("Mac mini", map[string]string{
		"review": "## 发现\n- 内存价格高\n",
	})

	if !strings.Contains(summary, "# 一句话结论") {
		t.Fatalf("fallback summary should contain headline")
	}
	if !strings.Contains(summary, "Review 中间结论如下") {
		t.Fatalf("fallback summary should reference review findings")
	}
	if !strings.Contains(summary, "内存价格高") {
		t.Fatalf("fallback summary should embed review content")
	}
}

func TestBuildSharedEvidencePoolKeepsDimensionLabels(t *testing.T) {
	pool := buildSharedEvidencePool(map[string][]searchResult{
		"review": {
			{Title: "评测 A", Snippet: "摘要 A", URL: "https://example.com/a"},
		},
		"macro": {
			{Title: "官网 B", Snippet: "摘要 B", URL: "https://example.com/b"},
		},
	})

	if len(pool) != 2 {
		t.Fatalf("expected 2 evidence items, got %d", len(pool))
	}
	if pool[0].Dimension != "review" || pool[1].Dimension != "macro" {
		t.Fatalf("unexpected dimensions: %+v", pool)
	}
}

func TestAppendUniqueSearchResultsDeduplicatesByURL(t *testing.T) {
	items := appendUniqueSearchResults(
		[]searchResult{{Title: "A", URL: "https://example.com/a"}},
		searchResult{Title: "A2", URL: "https://example.com/a"},
		searchResult{Title: "B", URL: "https://example.com/b"},
	)

	if len(items) != 2 {
		t.Fatalf("expected deduplicated results length 2, got %d", len(items))
	}
}
