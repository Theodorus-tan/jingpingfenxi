package eino

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func buildChatSystemPrompt(report string, now time.Time) string {
	return fmt.Sprintf(`你是一个专业的商业分析 AI 助手，负责基于【竞品分析报告】做简明、业务可执行的回答。

今天是 %s。涉及“最新”“近期”“今年”的表述，只能引用报告中已有内容，不能把训练记忆当成实时信息。

【回答规则】
1. 默认使用简明模式：先结论，再 3 条要点，最后给动作建议。
2. 默认控制在 150 到 300 字；除非用户明确要求展开，否则不要长篇大论。
3. 不要写空泛背景，不要写“作为 AI”，不要复述整份报告。
4. 如果问题超出报告范围，要明确说“报告未提供”，再补一条一般性建议。

【竞品分析报告】
%s`, now.Format("2006-01-02"), report)
}

// ChatWithReport 对话接口
func ChatWithReport(ctx context.Context, report, message string, onChunk func(string)) error {
	apiKey := os.Getenv("VOLCENGINE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
	}

	modelName := os.Getenv("DOUBAO_MODEL_EP")
	if modelName == "" {
		modelName = os.Getenv("DEEPSEEK_MODEL")
	}
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	baseURL := os.Getenv("VOLCENGINE_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("DEEPSEEK_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	baseURL = strings.TrimRight(baseURL, "/") + "/v1"

	if apiKey == "" {
		return fmt.Errorf("未配置对话模型 API Key")
	}

	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	})
	if err != nil {
		return err
	}

	sysPrompt := buildChatSystemPrompt(report, time.Now())

	messages := []*schema.Message{
		schema.SystemMessage(sysPrompt),
		schema.UserMessage(message),
	}

	stream, err := model.Stream(ctx, messages)
	if err != nil {
		return err
	}

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" || err.Error() == "stop" || err.Error() == "closed" {
				break
			}
			return err
		}
		if chunk.Content != "" {
			onChunk(chunk.Content)
		}
	}

	return nil
}
