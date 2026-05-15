package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-resty/resty/v2"
	"github.com/hertz-contrib/sse"

	einoagent "competitor-backend/internal/pkg/eino"
)

func main() {
	h := server.Default(server.WithHostPorts("127.0.0.1:8888"))

	// CORS setup
	h.Use(func(c context.Context, ctx *app.RequestContext) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if string(ctx.Request.Method()) == "OPTIONS" {
			ctx.AbortWithStatus(consts.StatusOK)
			return
		}
		ctx.Next(c)
	})

	api := h.Group("/api")
	{
		api.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
		})

		// 获取动态菜单（侧边栏由后端驱动）
		api.GET("/menus", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{
				"code": 20000,
				"data": []utils.H{
					{
						"path": "/competitors",
						"name": "competitors",
						"meta": utils.H{
							"locale":       "menu.competitors",
							"requiresAuth": true,
							"icon":         "icon-list",
							"order":        1,
						},
						"children": []utils.H{
							{
								"path": "list",
								"name": "competitorsList",
								"meta": utils.H{
									"locale":       "menu.competitors",
									"requiresAuth": true,
								},
							},
						},
					},
					{
						"path": "/analysis",
						"name": "analysis",
						"meta": utils.H{
							"locale":       "menu.analysis.overview",
							"requiresAuth": true,
							"icon":         "icon-bar-chart",
							"order":        2,
						},
						"children": []utils.H{
							{
								"path": "new",
								"name": "analysisNew",
								"meta": utils.H{
									"locale":       "menu.analysis.new",
									"requiresAuth": true,
								},
							},
						},
					},
				},
			})
		})

		// 竞品相关路由 (暂时Mock，因为没有DB)
		api.GET("/competitors", func(c context.Context, ctx *app.RequestContext) {
			// Mock data
			ctx.JSON(consts.StatusOK, utils.H{
				"code": 20000,
				"data": []utils.H{
					{"id": 1, "name": "特斯拉 Model 3", "industry": "新能源汽车"},
				},
			})
		})

		// 触发分析任务
		api.POST("/analysis/task", func(c context.Context, ctx *app.RequestContext) {
			// 接收前端请求
			var req struct {
				CompetitorName string `json:"competitor_name"`
			}
			if err := ctx.BindAndValidate(&req); err != nil {
				ctx.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
				return
			}

			// 调用 Python Agent API
			client := resty.New()
			client.SetTimeout(3 * time.Minute) // Agent 分析需要时间

			var agentResp struct {
				Report string `json:"report"`
			}

			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(map[string]string{"competitor_name": req.CompetitorName}).
				SetResult(&agentResp).
				Post("http://localhost:8000/analyze")

			if err != nil {
				ctx.JSON(consts.StatusInternalServerError, utils.H{"error": "Failed to call Python Agent: " + err.Error()})
				return
			}

			if resp.IsError() {
				ctx.JSON(consts.StatusInternalServerError, utils.H{"error": "Agent API returned error: " + resp.String()})
				return
			}

			// 返回分析结果给前端
			ctx.JSON(consts.StatusOK, utils.H{
				"code":    20000,
				"message": "分析完成",
				"data": utils.H{
					"report": agentResp.Report,
				},
			})
		})
	}

	// Eino Agent
	api.POST("/analysis/eino", func(c context.Context, ctx *app.RequestContext) {
		var req struct {
			CompetitorName string `json:"competitor_name"`
			Scenario       string `json:"scenario"`
			Project        string `json:"project"`
		}
		if err := ctx.BindAndValidate(&req); err != nil {
			ctx.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
			return
		}

		ctx.SetStatusCode(consts.StatusOK)
		s := sse.NewStream(ctx)
		var streamMu sync.Mutex

		sendEvent := func(event einoagent.AgentEvent) {
			data, _ := json.Marshal(event)
			streamMu.Lock()
			defer streamMu.Unlock()
			_ = s.Publish(&sse.Event{
				Data: data,
			})
		}

		// 将场景透传给 Agent
		config := einoagent.DefaultConfig()
		config.Scenario = req.Scenario // 需要在 Config 中新增这个字段

		agent, err := einoagent.NewEinoAgent(config)
		if err != nil {
			sendEvent(einoagent.AgentEvent{Type: "error", Message: "初始化失败: " + err.Error()})
			return
		}

		report, err := agent.Run(c, req.CompetitorName, sendEvent)
		if err != nil {
			sendEvent(einoagent.AgentEvent{Type: "error", Message: err.Error()})
			return
		}

		sendEvent(einoagent.AgentEvent{Type: "done", Report: report})
	})

	api.POST("/analysis/chat", func(c context.Context, ctx *app.RequestContext) {
		var req struct {
			Report  string `json:"report"`
			Message string `json:"message"`
		}
		if err := ctx.BindAndValidate(&req); err != nil {
			ctx.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
			return
		}

		ctx.SetStatusCode(consts.StatusOK)
		s := sse.NewStream(ctx)
		var streamMu sync.Mutex

		// 调用简单的对话逻辑
		err := einoagent.ChatWithReport(c, req.Report, req.Message, func(chunk string) {
			data, _ := json.Marshal(utils.H{"chunk": chunk})
			streamMu.Lock()
			defer streamMu.Unlock()
			_ = s.Publish(&sse.Event{Data: data})
		})

		if err != nil {
			data, _ := json.Marshal(utils.H{"error": err.Error()})
			streamMu.Lock()
			_ = s.Publish(&sse.Event{Data: data})
			streamMu.Unlock()
		} else {
			data, _ := json.Marshal(utils.H{"done": true})
			streamMu.Lock()
			_ = s.Publish(&sse.Event{Data: data})
			streamMu.Unlock()
		}
	})

	api.GET("/analysis/images", func(c context.Context, ctx *app.RequestContext) {
		competitorName := string(ctx.Query("competitor_name"))
		if competitorName == "" {
			ctx.JSON(consts.StatusBadRequest, utils.H{"error": "competitor_name is required"})
			return
		}

		images, err := einoagent.FetchCompetitorImages(c, competitorName)
		if err != nil {
			ctx.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
			return
		}

		ctx.JSON(consts.StatusOK, utils.H{
			"code": 20000,
			"data": images,
		})
	})

	log.Println("Hertz Server started on :8888")
	h.Spin()
}
