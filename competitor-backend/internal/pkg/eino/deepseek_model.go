package eino

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type deepSeekChatModel struct {
	apiKey   string
	model    string
	baseURL  string
	client   *http.Client
	tools    []*schema.ToolInfo
}

func NewDeepSeekChatModel(apiKey, modelName string) model.ToolCallingChatModel {
	return &deepSeekChatModel{
		apiKey:  apiKey,
		model:   modelName,
		baseURL: "https://api.deepseek.com/chat/completions",
		client:  &http.Client{},
	}
}

type chatRequest struct {
	Model    string          `json:"model"`
	Messages []chatMessage   `json:"messages"`
	Tools    json.RawMessage `json:"tools,omitempty"`
}

type chatMessage struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	ToolCalls  []toolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	Name       string      `json:"name,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message chatMessage `json:"message"`
}

func (d *deepSeekChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	req := d.buildRequest(messages)
	return d.call(ctx, req)
}

func (d *deepSeekChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	req := d.buildRequest(messages)
	msg, err := d.call(ctx, req)
	if err != nil {
		return nil, err
	}
	sr, sw := schema.Pipe[*schema.Message](1)
	sw.Send(msg, nil)
	sw.Close()
	return sr, nil
}

func (d *deepSeekChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	newModel := &deepSeekChatModel{
		apiKey:  d.apiKey,
		model:   d.model,
		baseURL: d.baseURL,
		client:  d.client,
		tools:   tools,
	}
	return newModel, nil
}

func (d *deepSeekChatModel) BindTools(tools []*schema.ToolInfo) error {
	d.tools = tools
	return nil
}

func (d *deepSeekChatModel) buildToolsJSON() json.RawMessage {
	if len(d.tools) == 0 {
		return nil
	}
	var openAITools []map[string]any
	for _, t := range d.tools {
		params := map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
		}

		if t.ParamsOneOf != nil {
			js, err := t.ParamsOneOf.ToJSONSchema()
			if err == nil && js != nil {
				b, _ := json.Marshal(js)
				var raw map[string]any
				json.Unmarshal(b, &raw)

				if p, ok := raw["properties"]; ok {
					props := p.(map[string]any)
					for k, v := range props {
						prop := v.(map[string]any)
						cleanProp := map[string]any{
							"type": prop["type"],
						}
						if desc, ok := prop["description"]; ok {
							cleanProp["description"] = desc
						}
						params["properties"].(map[string]any)[k] = cleanProp
					}
				}
				if r, ok := raw["required"]; ok {
					params["required"] = r
				}
			}
		}

		openAITools = append(openAITools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Desc,
				"parameters":  params,
			},
		})
	}
	b, _ := json.Marshal(openAITools)
	return b
}

func (d *deepSeekChatModel) buildRequest(messages []*schema.Message) chatRequest {
	req := chatRequest{
		Model:    d.model,
		Messages: make([]chatMessage, 0, len(messages)),
		Tools:    d.buildToolsJSON(),
	}
	for _, m := range messages {
		cm := chatMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				cm.ToolCalls = append(cm.ToolCalls, toolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: toolFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
		}
		if m.ToolCallID != "" {
			cm.ToolCallID = m.ToolCallID
			cm.Name = m.Name
		}
		req.Messages = append(req.Messages, cm)
	}
	return req
}

func (d *deepSeekChatModel) call(ctx context.Context, req chatRequest) (*schema.Message, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", d.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (status %d, body: %s)", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("parse response (body: %s): %w", string(respBody), err)
	}

	if len(chatResp.Choices) == 0 {
		return &schema.Message{Role: schema.Assistant, Content: ""}, nil
	}

	msg := &schema.Message{
		Role:    schema.Assistant,
		Content: chatResp.Choices[0].Message.Content,
	}

	for _, tc := range chatResp.Choices[0].Message.ToolCalls {
		msg.ToolCalls = append(msg.ToolCalls, schema.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: schema.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	return msg, nil
}

