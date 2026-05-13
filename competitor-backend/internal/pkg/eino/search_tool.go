package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type searchTool struct{}

func NewSearchTool() tool.BaseTool {
	return &searchTool{}
}

func (s *searchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: "在互联网上查找最新信息。当需要了解竞品最新动态、新闻、价格或用户评价时使用。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Desc:     "搜索关键词，例如 'MacBook M4 review' 或 '大疆 Action 5 Pro 评测'",
				Type:     schema.String,
				Required: true,
			},
			"max_results": {
				Desc: "返回的最大结果数，默认 5",
				Type: schema.Integer,
			},
		}),
	}, nil
}

func (s *searchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}
	if args.MaxResults <= 0 {
		args.MaxResults = 5
	}

	results := bingSearch(args.Query, args.MaxResults)
	data, _ := json.Marshal(results)
	return string(data), nil
}

type searchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

func bingSearch(query string, maxResults int) []searchResult {
	client := &http.Client{Timeout: 10 * time.Second}
	searchURL := fmt.Sprintf("https://cn.bing.com/search?q=%s&count=%d", url.QueryEscape(query), maxResults)

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return fallbackResults("搜索暂时不可用")
	}
	defer resp.Body.Close()

	html, _ := io.ReadAll(resp.Body)
	return parseBingResults(string(html), maxResults)
}

func parseBingResults(html string, maxResults int) []searchResult {
	var results []searchResult
	parts := strings.Split(html, `<li class="b_algo`)
	for _, part := range parts[1:] {
		if len(results) >= maxResults {
			break
		}

		var title, snippet, link string

		// 找真实链接（跳过 stylesheet 链接）
		for _, prefix := range []string{`"` + "https://", `"http://`} {
			searchFrom := 0
			for {
				hrefStart := strings.Index(part[searchFrom:], `href=`+prefix)
				if hrefStart == -1 {
					break
				}
				hrefStart += 6 + searchFrom
				hrefEnd := strings.Index(part[hrefStart:], `"`)
				if hrefEnd == -1 {
					break
				}
				candidate := part[hrefStart : hrefStart+hrefEnd]
				if !strings.Contains(candidate, "bing.com") && !strings.HasSuffix(candidate, ".css") {
					link = candidate
					break
				}
				searchFrom = hrefStart + hrefEnd
			}
			if link != "" {
				break
			}
		}

		// 提取标题文本：找 <a ...>标题</a> 中非空文本
		aTags := strings.Split(part, `<a`)
		for _, aTag := range aTags {
			gtIdx := strings.Index(aTag, `>`)
			if gtIdx == -1 {
				continue
			}
			text := aTag[gtIdx+1:]
			if closeA := strings.Index(text, `</a>`); closeA != -1 {
				text = text[:closeA]
			}
			text = stripHTMLTags(text)
			text = strings.TrimSpace(text)
			if len(text) > 3 && !strings.Contains(text, " ") {
				continue // skip icon/short labels
			}
			if len(text) > 5 && title == "" {
				title = text
			}
		}

		// 提取摘要：找 <p> 标签
		for _, pTag := range []string{`<p>`, `<p `} {
			pStart := strings.Index(part, pTag)
			if pStart == -1 {
				continue
			}
			pContent := part[pStart:]
			gtIdx := strings.Index(pContent, `>`)
			if gtIdx == -1 {
				continue
			}
			pText := pContent[gtIdx+1:]
			if pEnd := strings.Index(pText, `</p>`); pEnd != -1 {
				snippet = stripHTMLTags(pText[:pEnd])
				snippet = strings.TrimSpace(snippet)
				if len(snippet) > 10 {
					break
				}
			}
		}

		if title != "" && link != "" {
			results = append(results, searchResult{
				Title:   title,
				Snippet: snippet,
				URL:     link,
			})
		}
	}

	if len(results) == 0 {
		return fallbackResults("未找到相关搜索结果")
	}
	return results
}

func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

func fallbackResults(msg string) []searchResult {
	return []searchResult{
		{Title: "搜索暂时不可用", Snippet: msg + "，将基于训练数据和行业知识进行分析。", URL: ""},
	}
}
