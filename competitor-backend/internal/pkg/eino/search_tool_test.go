package eino

import (
	"strings"
	"testing"
	"time"
)

func TestRewriteSearchQueryAddsTimeWindowAndReviewSources(t *testing.T) {
	now := time.Date(2026, time.May, 14, 0, 0, 0, 0, time.UTC)

	query := rewriteSearchQuery("Mac mini M4", searchIntentReview, now)

	requiredTokens := []string{"Mac mini M4", "2026", "2025", "最新", "评测", "用户评价", "体验", "测评"}
	for _, token := range requiredTokens {
		if !strings.Contains(query, token) {
			t.Fatalf("rewritten query missing token %q: %s", token, query)
		}
	}
}

func TestRewriteSearchQueryDeduplicatesTokens(t *testing.T) {
	now := time.Date(2026, time.May, 14, 0, 0, 0, 0, time.UTC)

	query := rewriteSearchQuery("2026 最新 GoPro", searchIntentFinance, now)

	if strings.Count(query, "2026") != 1 {
		t.Fatalf("expected deduplicated year token, got %s", query)
	}
	if strings.Count(query, "最新") != 1 {
		t.Fatalf("expected deduplicated latest token, got %s", query)
	}
}

func TestFallbackResultsWarnsAgainstHallucination(t *testing.T) {
	results := fallbackResults("未找到公开信息")
	if len(results) != 1 {
		t.Fatalf("expected single fallback result")
	}
	if !strings.Contains(results[0].Snippet, "禁止编造") {
		t.Fatalf("fallback snippet should warn against hallucination: %s", results[0].Snippet)
	}
}
