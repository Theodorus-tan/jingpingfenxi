package eino

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
)

func loadEnv() {
	f, err := os.Open("../python-agent/.env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func init() {
	loadEnv()
}

func buildSystemPrompt(scenario string, dimension string) string {
	return buildSystemPromptAt(scenario, dimension, time.Now())
}

func buildSystemPromptAt(scenario string, dimension string, now time.Time) string {
	var rolePrompt, taskPrompt string

	if dimension == "review" {
		rolePrompt = "你是一位资深的电商产品经理和消费者心理学专家，精通通过海量用户评论（Review）洞察产品优劣势与市场切入点。"
		if scenario == "Product_Improvement" {
			taskPrompt = "1. 差评归因：提取痛点。\n2. 爽点提炼：分析超出预期的功能。\n3. 对比竞品痛点，为我方产品提出至少3条差异化竞争与迭代优化的落地建议。"
		} else {
			taskPrompt = "1. 差评归因：提取痛点。\n2. 爽点提炼：分析超出预期的功能。\n3. 基于评论反馈的满足度，评估该品类是否存在未被满足的刚需，并给出新产品冷启动切入点建议。"
		}
	} else if dimension == "macro" {
		rolePrompt = "你是一位顶级商业战略分析师和行业研究员，擅长从碎片化的互联网信息中拼凑出企业的全貌，并评估商业可行性。"
		if scenario == "Product_Improvement" {
			taskPrompt = "1. 发展里程碑梳理。\n2. 核心团队与壁垒剖析。\n3. 梳理竞品当前的商业模式和主要获客渠道，寻找我方可以抢占的渠道空白。"
		} else {
			taskPrompt = "1. 发展里程碑梳理。\n2. 核心团队与壁垒剖析。\n3. 评估该领域的行业壁垒（技术/资金/渠道），并判断对于新入局者的难度评级。"
		}
	} else if dimension == "finance" {
		rolePrompt = "你是一位持有 CFA 证书的资深证券分析师，擅长透过公开数据与新闻看透企业真实的经营健康度与行业盈利空间。"
		if scenario == "Product_Improvement" {
			taskPrompt = "1. 盈利质量与资金链评估。\n2. 分析竞品的研发投入和库存周转率，判断其未来的发力方向和当前的销售疲软点。"
		} else {
			taskPrompt = "1. 盈利质量与资金链评估。\n2. 测算该品类头部企业的平均毛利率，评估新玩家入局的资金门槛和预期投资回报周期。"
		}
	}

	return fmt.Sprintf(`【角色设定】
%s

【时间上下文】
今天是 %s，当前处于 %d 年第 %d 季度。凡是“最新”“近期”“今年”的判断，都必须优先依据检索结果，不得退回训练记忆。

【执行任务】
%s

【输出目标】
你不是在写最终面向用户的长报告，而是在为 Master Agent 产出可汇总的中间结论。输出必须精炼、可执行、少空话。

【输出格式】
严格遵循 Markdown，且必须只包含以下四段：

## 发现
- 仅写 3 条最重要发现，每条都尽量短。

## 风险
- 仅写 2 到 3 条风险，直接说明会影响什么。

## 机会
- 仅写 2 到 3 条机会，直接说明我方可怎么打。

## 证据
- 列出 3 到 5 条证据，每条必须包含来源标题；若拿到了时间信息，要写明时间。

【重要规则】
1. 必须使用 web_search 工具获取最新信息。最多搜索 2 次。
2. 报告中每个结论都必须标注数据来源：🔍 [来源：搜索标题]
3. 数据降级与防幻觉：如果没有搜到相关数据，请明确标注“缺乏公开数据”，绝对禁止编造虚假数据。
4. 禁止写成冗长背景介绍，禁止大而全，禁止竞品自夸视角。
`, rolePrompt, now.Format("2006-01-02"), now.Year(), (int(now.Month())-1)/3+1, taskPrompt)
}

type AgentConfig struct {
	DeepSeekAPIKey string
	DeepSeekModel  string
	Scenario       string // 动态场景注入
}

type EinoAgent struct {
	chatModel model.ToolCallingChatModel
	config    *AgentConfig
}

func NewEinoAgent(config *AgentConfig) (*EinoAgent, error) {
	chatModel := NewDeepSeekChatModel(config.DeepSeekAPIKey, config.DeepSeekModel)

	return &EinoAgent{
		chatModel: chatModel,
		config:    config,
	}, nil
}

type AgentEvent struct {
	Type         string         `json:"type"`
	Dimension    string         `json:"dimension,omitempty"` // review, macro, finance, master
	Message      string         `json:"message,omitempty"`
	Query        string         `json:"query,omitempty"`
	Report       string         `json:"report,omitempty"`
	Titles       []string       `json:"titles,omitempty"`
	Evidences    []searchResult `json:"evidences,omitempty"`
	ResultsCount int            `json:"results_count,omitempty"`
}

type FinalReportPayload struct {
	Version      int                  `json:"version"`
	Competitor   string               `json:"competitor"`
	Scenario     string               `json:"scenario"`
	GeneratedAt  string               `json:"generated_at"`
	Summary      string               `json:"summary"`
	Review       string               `json:"review,omitempty"`
	Macro        string               `json:"macro,omitempty"`
	Finance      string               `json:"finance,omitempty"`
	EvidencePool []SharedEvidenceItem `json:"evidence_pool,omitempty"`
}

type SharedEvidenceItem struct {
	Dimension string `json:"dimension"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet,omitempty"`
	URL       string `json:"url,omitempty"`
}

type eventCallbackKey struct{}
type dimensionKey struct{}

func (a *EinoAgent) runDimension(ctx context.Context, competitorName string, dimension string, onEvent func(AgentEvent)) (string, []searchResult, error) {
	ctx = context.WithValue(ctx, dimensionKey{}, dimension)
	ctx = context.WithValue(ctx, eventCallbackKey{}, onEvent)
	collectedEvidence := make([]searchResult, 0, 6)

	toolCb := &ub.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			if cb, ok := ctx.Value(eventCallbackKey{}).(func(AgentEvent)); ok {
				dim := ctx.Value(dimensionKey{}).(string)
				var args map[string]any
				if err := json.Unmarshal([]byte(input.ArgumentsInJSON), &args); err == nil {
					if q, ok := args["query"].(string); ok {
						cb(AgentEvent{Type: "searching", Dimension: dim, Query: q})
					}
				}
			}
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			if cb, ok := ctx.Value(eventCallbackKey{}).(func(AgentEvent)); ok {
				dim := ctx.Value(dimensionKey{}).(string)
				var results []searchResult
				if err := json.Unmarshal([]byte(output.Response), &results); err == nil {
					collectedEvidence = appendUniqueSearchResults(collectedEvidence, results...)
					titles := make([]string, 0, len(results))
					for _, r := range results {
						titles = append(titles, r.Title)
					}
					cb(AgentEvent{
						Type:         "search_result",
						Dimension:    dim,
						ResultsCount: len(results),
						Titles:       titles,
						Evidences:    results,
					})
				}
			}
			return ctx
		},
	}

	modelCb := &ub.ModelCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			if cb, ok := ctx.Value(eventCallbackKey{}).(func(AgentEvent)); ok {
				dim := ctx.Value(dimensionKey{}).(string)
				cb(AgentEvent{Type: "thinking", Dimension: dim, Message: "正在调用大模型推理..."})
			}
			return ctx
		},
	}

	handler := react.BuildAgentCallback(modelCb, toolCb)

	var currentTool tool.BaseTool
	if dimension == "macro" {
		currentTool = NewMacroSearchTool()
	} else if dimension == "finance" {
		currentTool = NewFinanceSearchTool()
	} else {
		currentTool = NewSearchTool() // Review 默认用常规搜索
	}

	rAgent, err := react.NewAgent(context.Background(), &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{currentTool},
		},
		MaxStep: 30,
	})
	if err != nil {
		return "", nil, fmt.Errorf("创建 ReAct Agent 失败 (%s): %w", dimension, err)
	}

	prompt := buildSystemPrompt(a.config.Scenario, dimension)

	sr, err := rAgent.Stream(ctx, []*schema.Message{
		{Role: schema.System, Content: prompt},
		{Role: schema.User, Content: fmt.Sprintf("请对 '%s' 进行全面的分析。", competitorName)},
	}, agent.WithComposeOptions(compose.WithCallbacks(handler)))
	if err != nil {
		return "", nil, fmt.Errorf("Agent Stream 失败 (%s): %w", dimension, err)
	}
	defer sr.Close()

	var fullReport string
	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", nil, fmt.Errorf("接收流失败 (%s): %w", dimension, err)
		}
		fullReport += msg.Content
	}

	onEvent(AgentEvent{Type: "writing", Dimension: dimension, Message: "报告生成完成"})

	return fullReport, collectedEvidence, nil
}

func buildMasterSummaryPrompt(scenario string, competitorName string, reports map[string]string, evidencePool []SharedEvidenceItem, now time.Time) string {
	scenarioText := "已有产品求改进"
	if scenario == "Market_Entry" {
		scenarioText = "无产品求入局"
	}

	return fmt.Sprintf(`【角色设定】
你是 Master Agent，负责把 Review、宏观、财务三位分析师的中间结论汇总成一份业务可直接使用的精炼综述。

【时间上下文】
今天是 %s，当前处于 %d 年第 %d 季度。涉及“最新”“近期”“今年”的结论，只能依据下方分析师中已经引用的证据，不得自行脑补。

【任务背景】
竞品：%s
场景：%s

【输出目标】
只输出一份综述报告，不要重复三份分报告，不要写学术化长文。Review 视角权重最高，宏观与财务只做补充判断。

【输出格式】
严格使用 Markdown，并且只输出以下结构：

# 一句话结论
- 1 句话说透值不值得打、该怎么打。

## 用户最痛 3 点
- 每条都写成接近用户抱怨的话。

## 竞品最强 3 点
- 只写真正构成威胁的优势。

## 我们可攻击 3 点
- 每条都直接说可攻击点和攻击方式。

## 这周可执行动作
- 只写 3 到 5 条能马上落地的动作。

## 补充背景
- 仅保留必要的宏观或财务背景，不超过 3 条。

【重要规则】
1. 结论必须短、狠、可执行。
2. 不能堆大段背景，不能写成咨询报告。
3. 如果某维度缺乏公开数据，要明确说“缺乏公开数据”。
4. 禁止输出“作为 AI”之类废话。

【分析师中间结论】
%s

【共享证据池】
%s
`, now.Format("2006-01-02"), now.Year(), (int(now.Month())-1)/3+1, competitorName, scenarioText, formatReportsForMaster(reports), formatEvidencePoolForMaster(evidencePool))
}

func formatReportsForMaster(reports map[string]string) string {
	ordered := []struct {
		key   string
		label string
	}{
		{key: "review", label: "Review"},
		{key: "macro", label: "宏观"},
		{key: "finance", label: "财务"},
	}

	var builder strings.Builder
	for _, item := range ordered {
		builder.WriteString("### ")
		builder.WriteString(item.label)
		builder.WriteString("\n")
		content := strings.TrimSpace(reports[item.key])
		if content == "" {
			builder.WriteString("缺乏公开数据\n\n")
			continue
		}
		builder.WriteString(content)
		builder.WriteString("\n\n")
	}

	return strings.TrimSpace(builder.String())
}

func formatEvidencePoolForMaster(evidencePool []SharedEvidenceItem) string {
	if len(evidencePool) == 0 {
		return "缺乏公开证据"
	}

	var builder strings.Builder
	for _, item := range evidencePool {
		builder.WriteString("- [")
		builder.WriteString(item.Dimension)
		builder.WriteString("] ")
		builder.WriteString(item.Title)
		if item.Snippet != "" {
			builder.WriteString(" | 摘录：")
			builder.WriteString(item.Snippet)
		}
		if item.URL != "" {
			builder.WriteString(" | 链接：")
			builder.WriteString(item.URL)
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}

func (a *EinoAgent) buildExecutiveSummary(ctx context.Context, competitorName string, reports map[string]string, evidencePool []SharedEvidenceItem) (string, error) {
	resp, err := a.chatModel.Generate(ctx, []*schema.Message{
		{
			Role:    schema.System,
			Content: buildMasterSummaryPrompt(a.config.Scenario, competitorName, reports, evidencePool, time.Now()),
		},
		{
			Role:    schema.User,
			Content: "请严格按指定结构输出一份精炼综述报告。",
		},
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

func buildFallbackExecutiveSummary(competitorName string, reports map[string]string) string {
	review := strings.TrimSpace(reports["review"])
	if review == "" {
		review = "缺乏公开数据"
	}

	var builder strings.Builder
	builder.WriteString("# 一句话结论\n")
	builder.WriteString("- 当前自动汇总降级为基于 Review 主视角的简版综述，建议优先核查用户痛点与可攻击点。\n\n")
	builder.WriteString("## 用户最痛 3 点\n")
	builder.WriteString("- 请先查看下方 Review 中间结论提炼用户抱怨。\n")
	builder.WriteString("- 若需更具体结论，请补充更明确的搜索关键词。\n")
	builder.WriteString("- 宏观与财务仅作为补充判断。\n\n")
	builder.WriteString("## 竞品最强 3 点\n")
	builder.WriteString("- 该部分需结合 Review 与宏观结论交叉判断。\n")
	builder.WriteString("- 若证据不足，应明确标注缺乏公开数据。\n")
	builder.WriteString("- 不以训练记忆替代检索结果。\n\n")
	builder.WriteString("## 我们可攻击 3 点\n")
	builder.WriteString("- 优先攻击用户集中抱怨且竞品短期难修复的问题。\n")
	builder.WriteString("- 把攻击点写成页面卖点、客服话术和投放文案。\n")
	builder.WriteString("- 用补充证据继续收紧结论。\n\n")
	builder.WriteString("## 这周可执行动作\n")
	builder.WriteString("- 复核 Review 检索结果并补齐高质量来源。\n")
	builder.WriteString("- 将用户抱怨转成产品需求与营销文案。\n")
	builder.WriteString("- 基于可攻击点输出下一轮对标方案。\n\n")
	builder.WriteString("## 补充背景\n")
	builder.WriteString("- 竞品：")
	builder.WriteString(competitorName)
	builder.WriteString("\n")
	builder.WriteString("- Review 中间结论如下：\n\n")
	builder.WriteString(review)

	return builder.String()
}

func appendUniqueSearchResults(existing []searchResult, incoming ...searchResult) []searchResult {
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		key := strings.TrimSpace(item.URL)
		if key == "" {
			key = strings.TrimSpace(item.Title)
		}
		if key != "" {
			seen[key] = struct{}{}
		}
	}

	for _, item := range incoming {
		key := strings.TrimSpace(item.URL)
		if key == "" {
			key = strings.TrimSpace(item.Title)
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		existing = append(existing, item)
	}

	return existing
}

func buildSharedEvidencePool(dimensionResults map[string][]searchResult) []SharedEvidenceItem {
	order := []string{"review", "macro", "finance"}
	items := make([]SharedEvidenceItem, 0, 12)

	for _, dimension := range order {
		results := dimensionResults[dimension]
		limit := len(results)
		if limit > 4 {
			limit = 4
		}
		for _, result := range results[:limit] {
			items = append(items, SharedEvidenceItem{
				Dimension: dimension,
				Title:     strings.TrimSpace(result.Title),
				Snippet:   strings.TrimSpace(result.Snippet),
				URL:       strings.TrimSpace(result.URL),
			})
		}
	}

	return items
}

func buildFinalReportPayload(competitorName string, scenario string, summary string, reports map[string]string, evidencePool []SharedEvidenceItem, generatedAt time.Time) (string, error) {
	payload := FinalReportPayload{
		Version:      3,
		Competitor:   competitorName,
		Scenario:     scenario,
		GeneratedAt:  generatedAt.Format(time.RFC3339),
		Summary:      strings.TrimSpace(summary),
		Review:       strings.TrimSpace(reports["review"]),
		Macro:        strings.TrimSpace(reports["macro"]),
		Finance:      strings.TrimSpace(reports["finance"]),
		EvidencePool: evidencePool,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (a *EinoAgent) Run(ctx context.Context, competitorName string, onEvent func(AgentEvent)) (string, error) {
	onEvent(AgentEvent{Type: "thinking", Dimension: "master", Message: "Master Agent 开始拆解任务并分发给 3D 领域专家..."})

	type dimResult struct {
		dimension string
		report    string
		evidences []searchResult
		err       error
	}

	results := make(chan dimResult, 3)
	dimensions := []string{"review", "macro", "finance"}

	for _, dim := range dimensions {
		go func(d string) {
			onEvent(AgentEvent{Type: "thinking", Dimension: d, Message: fmt.Sprintf("[%s] 专家 Agent 启动...", d)})
			report, evidences, err := a.runDimension(ctx, competitorName, d, onEvent)
			results <- dimResult{dimension: d, report: report, evidences: evidences, err: err}
		}(dim)
	}

	reports := make(map[string]string)
	evidenceByDimension := make(map[string][]searchResult)
	for i := 0; i < 3; i++ {
		res := <-results
		if res.err != nil {
			onEvent(AgentEvent{Type: "error", Dimension: res.dimension, Message: res.err.Error()})
			reports[res.dimension] = fmt.Sprintf("分析失败: %v", res.err)
		} else {
			reports[res.dimension] = res.report
			evidenceByDimension[res.dimension] = res.evidences
		}
	}

	onEvent(AgentEvent{Type: "thinking", Dimension: "master", Message: "所有专家 Agent 执行完毕，Master Agent 正在汇总 3D 报告..."})
	evidencePool := buildSharedEvidencePool(evidenceByDimension)

	summary, err := a.buildExecutiveSummary(ctx, competitorName, reports, evidencePool)
	if err != nil {
		onEvent(AgentEvent{Type: "error", Dimension: "master", Message: fmt.Sprintf("Master 汇总失败，已降级为本地综述：%v", err)})
		summary = buildFallbackExecutiveSummary(competitorName, reports)
	}

	onEvent(AgentEvent{Type: "writing", Dimension: "master", Message: "综述报告已汇总完成"})

	return buildFinalReportPayload(competitorName, a.config.Scenario, summary, reports, evidencePool, time.Now())
}

func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		DeepSeekAPIKey: os.Getenv("VOLCENGINE_API_KEY"),
		DeepSeekModel:  os.Getenv("DOUBAO_MODEL_EP"),
		Scenario:       "Product_Improvement", // 默认场景
	}
}
