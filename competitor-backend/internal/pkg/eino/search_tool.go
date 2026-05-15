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
type searchIntent string

const (
	searchIntentReview  searchIntent = "review"
	searchIntentMacro   searchIntent = "macro"
	searchIntentFinance searchIntent = "finance"
)

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

	results := bingSearch(rewriteSearchQuery(args.Query, searchIntentReview, time.Now()), args.MaxResults)
	data, _ := json.Marshal(results)
	return string(data), nil
}

type searchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

func rewriteSearchQuery(query string, intent searchIntent, now time.Time) string {
	normalized := strings.TrimSpace(query)
	if normalized == "" {
		return normalized
	}

	tokens := []string{normalized}
	if !strings.Contains(normalized, fmt.Sprintf("%d", now.Year())) {
		tokens = append(tokens, fmt.Sprintf("%d", now.Year()))
	}
	if !strings.Contains(normalized, fmt.Sprintf("%d", now.Year()-1)) {
		tokens = append(tokens, fmt.Sprintf("%d", now.Year()-1))
	}
	if !strings.Contains(normalized, "最新") {
		tokens = append(tokens, "最新")
	}

	switch intent {
	case searchIntentReview:
		tokens = append(tokens, "评测", "用户评价", "体验", "测评")
	case searchIntentMacro:
		tokens = append(tokens, "官网", "创始人", "融资", "商业模式")
	case searchIntentFinance:
		tokens = append(tokens, "财报", "营收", "利润", "融资")
	}

	return strings.Join(deduplicateTokens(tokens), " ")
}

func deduplicateTokens(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		result = append(result, token)
	}
	return result
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
		{Title: "搜索暂时不可用", Snippet: msg + "，请在报告中明确标注缺乏公开数据，禁止编造。", URL: ""},
	}
}

// -----------------------------------------
// 2. 战略广谱抓取 Tool (Macro) + 降级
// -----------------------------------------
type macroSearchTool struct{}

func NewMacroSearchTool() tool.BaseTool {
	return &macroSearchTool{}
}

func (s *macroSearchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "macro_search",
		Desc: "查询企业百科、官网、创始人背景等宏观战略信息。当需要梳理发展历程、商业模式时使用。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Desc:     "搜索关键词，例如 '特斯拉 商业模式' 或 '大疆 创始人 背景'",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (s *macroSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	_ = json.Unmarshal([]byte(argumentsInJSON), &args)

	// 模拟尝试抓取百科/企查查（这里仍用 bing 替代，但在真实场景会接特定 API）
	results := bingSearch(rewriteSearchQuery(args.Query+" 百度百科 企查查", searchIntentMacro, time.Now()), 5)

	// 降级判断：如果结果太少，说明是低调初创企业，退回基础 snippet 并打上标记
	if len(results) < 2 {
		results = bingSearch(rewriteSearchQuery(args.Query, searchIntentMacro, time.Now()), 3)
		for i := range results {
			results[i].Snippet = "[低置信度-非结构化信息] " + results[i].Snippet
		}
	}

	data, _ := json.Marshal(results)
	return string(data), nil
}

// -----------------------------------------
// 3. 财务股市抓取 Tool (Finance) + 熔断
// -----------------------------------------
type financeSearchTool struct{}

func NewFinanceSearchTool() tool.BaseTool {
	return &financeSearchTool{}
}

func (s *financeSearchTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "finance_search",
		Desc: "查询上市公司的财报、营收、毛利率等财务数据。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"company_name": {
				Desc:     "公司名称或股票代码，例如 'AAPL' 或 '比亚迪'",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (s *financeSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		CompanyName string `json:"company_name"`
	}
	_ = json.Unmarshal([]byte(argumentsInJSON), &args)

	// 模拟调用股票 API（这里用 bing 搜索代替）
	results := bingSearch(rewriteSearchQuery(args.CompanyName+" 财报 营收 净利润", searchIntentFinance, time.Now()), 3)

	// 降级与熔断机制 (Circuit Breaker)
	// 如果没有明显包含财务关键字，判定为非上市公司
	hasFinanceData := false
	for _, r := range results {
		if strings.Contains(r.Snippet, "亿") || strings.Contains(r.Snippet, "营收") || strings.Contains(r.Snippet, "财报") {
			hasFinanceData = true
			break
		}
	}

	if !hasFinanceData {
		// 第一步降级：尝试找融资历程
		fundingResults := bingSearch(rewriteSearchQuery(args.CompanyName+" 天眼查 融资 历程", searchIntentFinance, time.Now()), 2)
		hasFunding := false
		for _, r := range fundingResults {
			if strings.Contains(r.Snippet, "轮") || strings.Contains(r.Snippet, "投资") {
				hasFunding = true
				break
			}
		}

		if hasFunding {
			data, _ := json.Marshal(fundingResults)
			return string(data), nil
		}

		// 第二步熔断：彻底没数据，防幻觉，直接返回明确的熔断指令给 LLM
		circuitBreakerMsg := []searchResult{
			{Title: "熔断触发", Snippet: "【系统指令】该企业为非上市公司，且未查到公开融资数据。请在报告中直接输出：『该竞品为非公开市场企业，无法获取有效资本与财务数据。』绝对禁止编造虚假数据。", URL: ""},
		}
		data, _ := json.Marshal(circuitBreakerMsg)
		return string(data), nil
	}

	data, _ := json.Marshal(results)
	return string(data), nil
}
