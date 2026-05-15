package eino

import "testing"

func TestExtractPreferredImageURLPrefersOgImage(t *testing.T) {
	html := `<html><head><meta property="og:image" content="https://example.com/cover.jpg" /></head></html>`

	got := extractPreferredImageURL(html, "https://example.com/page")

	if got != "https://example.com/cover.jpg" {
		t.Fatalf("expected og:image url, got %s", got)
	}
}

func TestExtractPreferredImageURLResolvesRelativeURL(t *testing.T) {
	html := `<html><body><img src="/assets/product.png" /></body></html>`

	got := extractPreferredImageURL(html, "https://example.com/path/page")

	if got != "https://example.com/assets/product.png" {
		t.Fatalf("expected resolved image url, got %s", got)
	}
}

func TestResolveAssetURLKeepsAbsoluteURL(t *testing.T) {
	got := resolveAssetURL("https://example.com/page", "https://cdn.example.com/a.png")
	if got != "https://cdn.example.com/a.png" {
		t.Fatalf("expected absolute url unchanged, got %s", got)
	}
}
