package eino

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ImageCandidate struct {
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
	PageURL  string `json:"page_url"`
	ImageURL string `json:"image_url"`
}

func FetchCompetitorImages(ctx context.Context, competitorName string) ([]ImageCandidate, error) {
	query := rewriteSearchQuery(
		fmt.Sprintf("%s 官网 产品图 主图 开箱", competitorName),
		searchIntentReview,
		time.Now(),
	)
	results := bingSearch(query, 5)
	client := &http.Client{Timeout: 8 * time.Second}
	candidates := make([]ImageCandidate, 0, 4)
	seenImages := map[string]struct{}{}

	for _, result := range results {
		if result.URL == "" {
			continue
		}
		imageURL, err := fetchPrimaryImageURL(ctx, client, result.URL)
		if err != nil || imageURL == "" {
			continue
		}
		if _, exists := seenImages[imageURL]; exists {
			continue
		}
		seenImages[imageURL] = struct{}{}
		candidates = append(candidates, ImageCandidate{
			Title:    result.Title,
			Snippet:  result.Snippet,
			PageURL:  result.URL,
			ImageURL: imageURL,
		})
		if len(candidates) >= 3 {
			break
		}
	}

	return candidates, nil
}

func fetchPrimaryImageURL(ctx context.Context, client *http.Client, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	return extractPreferredImageURL(string(body), pageURL), nil
}

func extractPreferredImageURL(html string, pageURL string) string {
	for _, property := range []string{"og:image", "twitter:image"} {
		if value := extractMetaContent(html, property); value != "" {
			return resolveAssetURL(pageURL, value)
		}
	}

	for _, marker := range []string{`<img src="`, `<img data-src="`, `<img class="`} {
		index := strings.Index(html, marker)
		if index == -1 {
			continue
		}
		segment := html[index:]
		for _, attr := range []string{`src="`, `data-src="`} {
			attrIndex := strings.Index(segment, attr)
			if attrIndex == -1 {
				continue
			}
			start := attrIndex + len(attr)
			end := strings.Index(segment[start:], `"`)
			if end == -1 {
				continue
			}
			value := segment[start : start+end]
			if strings.TrimSpace(value) == "" {
				continue
			}
			return resolveAssetURL(pageURL, value)
		}
	}

	return ""
}

func extractMetaContent(html string, property string) string {
	patterns := []string{
		fmt.Sprintf(`property="%s" content="`, property),
		fmt.Sprintf(`name="%s" content="`, property),
	}
	for _, pattern := range patterns {
		index := strings.Index(html, pattern)
		if index == -1 {
			continue
		}
		start := index + len(pattern)
		end := strings.Index(html[start:], `"`)
		if end == -1 {
			continue
		}
		candidate := html[start : start+end]
		if strings.Contains(pattern, `content="`) && candidate == property {
			continue
		}
		if candidate != "" {
			return candidate
		}
	}

	altPattern := `content="`
	index := strings.Index(html, fmt.Sprintf(`property="%s"`, property))
	if index == -1 {
		index = strings.Index(html, fmt.Sprintf(`name="%s"`, property))
	}
	if index != -1 {
		segment := html[index:]
		contentIndex := strings.Index(segment, altPattern)
		if contentIndex != -1 {
			start := contentIndex + len(altPattern)
			end := strings.Index(segment[start:], `"`)
			if end != -1 {
				return segment[start : start+end]
			}
		}
	}

	return ""
}

func resolveAssetURL(pageURL string, assetURL string) string {
	assetURL = strings.TrimSpace(assetURL)
	if assetURL == "" {
		return ""
	}
	if strings.HasPrefix(assetURL, "http://") || strings.HasPrefix(assetURL, "https://") {
		return assetURL
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return assetURL
	}
	ref, err := url.Parse(assetURL)
	if err != nil {
		return assetURL
	}
	return base.ResolveReference(ref).String()
}
