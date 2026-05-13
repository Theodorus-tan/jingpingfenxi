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

const ChiefAnalystPrompt = `你是一个顶级的商业竞品分析师（Chief Competitive Analyst）。
你的任务是根据用户提供的竞品名称或线索，进行深度的竞品分析。

重要规则（必须严格遵守）：
1. 你的训练数据可能过时，你必须使用 web_search 工具获取最新信息。
2. **最多搜索2次**（1次搜索核心信息 + 最多1次补充搜索），之后必须立即输出最终报告。绝对不要搜索第3次。
3. 报告中每个结论都必须标注数据来源：
   - 来自搜索结果的数据标注为：🔍 [来源：搜索标题]
   - 基于行业常识的推演标注为：📊 [基于行业知识推演]
4. 搜索时请使用中英文关键词结合。

分析步骤：
1. 搜索该竞品的核心功能、官方网站和最新动态。
2. 如果第一次搜索结果充分，直接输出报告；如果信息不足，再搜索一次用户评价和定价信息。
3. 综合以上信息，输出一份专业竞品分析报告，包含以下模块：

---

# 竞品分析报告：[竞品名称]

**生成时间：** [当前时间]
**分析依据：** 🔍 互联网公开信息 + 📊 行业知识推演

---

## 1. 执行摘要
用 3-5 句话概括核心发现，让读者 30 秒内了解全貌。

## 2. 竞品简介与公司背景
- 产品定位、公司背景、发布时间

## 3. 核心功能与卖点
- 关键规格、差异化功能、技术创新点

## 4. 用户口碑分析
- 优点（附来源）
- 缺点（附来源）

## 5. 定价与市场定位
- 价格区间、目标人群、定价策略分析

## 6. 优势与劣势

| 优势 | 劣势 |
|------|------|
| ... | ... |

## 7. 机会与威胁

## 8. 综合评价与建议
- 评分（满分 10 分）
- 适合谁 / 不适合谁
- 综合建议

---

记住：搜索最多2次，之后必须立即输出最终报告，不要犹豫！
报告语言必须使用中文，结构清晰，专业客观。
每个数据点尽量标注来源链接，让读者可追溯。`

type AgentConfig struct {
	DeepSeekAPIKey string
	DeepSeekModel  string
}

type EinoAgent struct {
	rAgent *react.Agent
}

func NewEinoAgent(config *AgentConfig) (*EinoAgent, error) {
	chatModel := NewDeepSeekChatModel(config.DeepSeekAPIKey, config.DeepSeekModel)

	searchTool := NewSearchTool()

	rAgent, err := react.NewAgent(context.Background(), &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{searchTool},
		},
		MaxStep: 30,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ReAct Agent 失败: %w", err)
	}

	return &EinoAgent{rAgent: rAgent}, nil
}

type AgentEvent struct {
	Type         string   `json:"type"`
	Message      string   `json:"message,omitempty"`
	Query        string   `json:"query,omitempty"`
	Report       string   `json:"report,omitempty"`
	Titles       []string `json:"titles,omitempty"`
	ResultsCount int      `json:"results_count,omitempty"`
}

type eventCallbackKey struct{}

func (a *EinoAgent) Run(ctx context.Context, competitorName string, onEvent func(AgentEvent)) (string, error) {
	onEvent(AgentEvent{Type: "thinking", Message: "开始分析任务..."})

	ctx = context.WithValue(ctx, eventCallbackKey{}, onEvent)

	toolCb := &ub.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			if cb, ok := ctx.Value(eventCallbackKey{}).(func(AgentEvent)); ok {
				var args map[string]any
				if err := json.Unmarshal([]byte(input.ArgumentsInJSON), &args); err == nil {
					if q, ok := args["query"].(string); ok {
						cb(AgentEvent{Type: "searching", Query: q})
					}
				}
			}
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			if cb, ok := ctx.Value(eventCallbackKey{}).(func(AgentEvent)); ok {
				var results []searchResult
				if err := json.Unmarshal([]byte(output.Response), &results); err == nil {
					titles := make([]string, 0, len(results))
					for _, r := range results {
						titles = append(titles, r.Title)
					}
					cb(AgentEvent{
						Type:         "search_result",
						ResultsCount: len(results),
						Titles:       titles,
					})
				}
			}
			return ctx
		},
	}

	modelCb := &ub.ModelCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			if cb, ok := ctx.Value(eventCallbackKey{}).(func(AgentEvent)); ok {
				cb(AgentEvent{Type: "thinking", Message: "正在调用大模型推理..."})
			}
			return ctx
		},
	}

	handler := react.BuildAgentCallback(modelCb, toolCb)

	sr, err := a.rAgent.Stream(ctx, []*schema.Message{
		{Role: schema.System, Content: ChiefAnalystPrompt},
		{Role: schema.User, Content: fmt.Sprintf("请对 '%s' 进行全面的竞品分析。", competitorName)},
	}, agent.WithComposeOptions(compose.WithCallbacks(handler)))
	if err != nil {
		return "", fmt.Errorf("Agent Stream 失败: %w", err)
	}
	defer sr.Close()

	var fullReport string
	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("接收流失败: %w", err)
		}
		fullReport += msg.Content
	}

	onEvent(AgentEvent{Type: "writing", Message: "报告生成完成"})

	return fullReport, nil
}

func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		DeepSeekAPIKey: os.Getenv("VOLCENGINE_API_KEY"),
		DeepSeekModel:  os.Getenv("DOUBAO_MODEL_EP"),
	}
}
