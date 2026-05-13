# AI 竞品分析系统 — 架构与接口规范

> 本文档对标字节系科创竞赛技术栈标准，描述系统架构、数据链路与接口规范。

---

## 一、系统架构全景

```
┌─────────────────────────────────────────────────────────────┐
│                       用户（浏览器）                          │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                  Arco Design 前端（Vue 3）                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │  工作台   │  │ 竞品管理  │  │ 新建分析  │  │  分析报告   │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────┬─────┘  │
│       └──────────────┼─────────────┼───────────────┘        │
│                      │             │                         │
│               Axios HTTP      EventSource/SSE                │
└──────────────────────┼─────────────┼─────────────────────────┘
                       │             │
                       ▼             ▼
┌─────────────────────────────────────────────────────────────┐
│                  Go + Hertz 后端（:8888）                     │
│                                                              │
│  POST /api/analysis/task    GET /api/competitors             │
│  POST /analyze/stream (SSE 穿透 → Agent)                    │
│                                                              │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    HTTP POST /analyze
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              Python AI Agent（:8000，过渡方案）               │
│                                                              │
│  ┌──────────────────────────────┐                            │
│  │  AgenticLoop（LLM 推理循环）  │                            │
│  │  ├─ web_search (ddgs)       │                            │
│  │  ├─ DeepSeek Chat API       │                            │
│  │  └─ on_event SSE 回调       │                            │
│  └──────────────────────────────┘                            │
│                                                              │
│  ⬇ 迁移目标：Eino（CloudWeGo Go AI 框架）                    │
│  ┌──────────────────────────────┐                            │
│  │  Eino ReAct Agent            │                            │
│  │  ├─ ChatModel (DeepSeek)    │                            │
│  │  ├─ Tools (web_search)      │                            │
│  │  └─ Graph 编排              │                            │
│  └──────────────────────────────┘                            │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、API 接口规范

### 2.1 统一响应格式

对标字节规范，所有接口统一返回格式：

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

状态码：

| HTTP 状态码 | 业务场景 |
|------------|---------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务端错误 |

### 2.2 接口列表

#### 健康检查

```
GET /api/ping
```

```json
// Response
{ "code": 200, "msg": "success", "data": { "message": "pong" } }
```

#### 竞品列表

```
GET /api/competitors
```

```json
// Response
{
  "code": 200,
  "msg": "success",
  "data": [
    { "id": 1, "name": "特斯拉 Model 3", "industry": "新能源汽车" }
  ]
}
```

#### 提交分析任务

```
POST /api/analysis/task
Content-Type: application/json

{
  "competitor_name": "GoPro Hero 13"
}
```

```json
// Response
{
  "code": 200,
  "msg": "success",
  "data": {
    "report": "## 竞品分析报告...（完整 Markdown）"
  }
}
```

#### SSE 流式分析

```
POST /analyze/stream
Content-Type: application/json

{
  "competitor_name": "GoPro Hero 13"
}
```

响应为 `text/event-stream`，事件格式：

```
data: {"type":"thinking","message":"开始分析任务..."}

data: {"type":"searching","query":"GoPro Hero 13 详细规格"}

data: {"type":"search_result","query":"...","results_count":5,"titles":["GoPro 官网","Wikipedia",...]}

data: {"type":"writing","message":"报告生成完成"}

data: {"type":"done","report":"## 完整报告..."}
```

### 2.3 接口规范要求

| 规则 | 约定 |
|------|------|
| 基础路径 | `/api` 前缀 |
| 请求体 | JSON（`Content-Type: application/json`） |
| 时间格式 | ISO 8601 |
| 分页 | `{ page, page_size, total, items }` |
| 字段命名 | snake_case |

---

## 三、数据链路

### 3.1 同步分析链路

```
前端                    Go 后端                 Python Agent             数据源
 │                       │                       │                       │
 │ POST /api/analysis    │                       │                       │
 │ /task                 │                       │                       │
 │──────────────────────>│                       │                       │
 │                       │ POST /analyze         │                       │
 │                       │──────────────────────>│                       │
 │                       │                       │ ddgs.text()           │
 │                       │                       │──────────────────────>│ DuckDuckGo
 │                       │                       │<──────────────────────│
 │                       │                       │                       │
 │                       │                       │ DeepSeek Chat API     │
 │                       │                       │──────────────────────>│ DeepSeek
 │                       │                       │<──────────────────────│
 │                       │      报告 (report)     │                       │
 │                       │<──────────────────────│                       │
 │   报告 (JSON)         │                       │                       │
 │<──────────────────────│                       │                       │
```

### 3.2 SSE 流式链路

```
前端                    Python Agent
 │                       │
 │ POST /analyze/stream  │
 │──────────────────────>│
 │                       │
 │ data: thinking        │  ← Agent 开始分析
 │<──────────────────────│
 │                       │
 │ data: searching       │  ← Agent 正在搜索
 │<──────────────────────│
 │                       │
 │ data: search_result   │  ← 搜索到结果
 │<──────────────────────│
 │                       │
 │ data: done            │  ← 报告生成完成
 │<──────────────────────│
```

### 3.3 通信协议

| 方向 | 协议 | 说明 |
|------|------|------|
| 前端 → Go 后端 | HTTP/REST | Axios 封装，同步请求 |
| 前端 → Python Agent | HTTP/SSE | Vite proxy 转发，流式推送 |
| Go 后端 → Python Agent | HTTP/REST | resty 客户端，同步调用 |
| Agent → 搜索引擎 | HTTP | ddgs 库异步搜索 |
| Agent → DeepSeek API | HTTP | OpenAI 兼容 API |

---

## 四、数据持久化方案（当前）

| 数据 | 存储方式 | 说明 |
|------|---------|------|
| 分析报告 | localStorage | `latest_analysis_report` Key |
| 分析历史 | localStorage | `analysis_history` JSON 数组 |
| 竞品列表 | 前端内存 | Mock 数据，刷新丢失 |

### 4.1 后续演进（规划中）

接入数据库后的目标架构：

```
┌──────────────────┐
│  PostgreSQL 16    │
│  ├─ competitors   │  ← 竞品表
│  ├─ analysis_     │
│  │  tasks         │  ← 分析任务表
│  └─ analysis_     │
│     reports       │  ← 分析报告表
└──────────────────┘
         │
         ▼
┌──────────────────┐
│  Redis 7          │
│  ├─ 报告缓存      │
│  ├─ 搜索缓存      │
│  └─ 请求限流      │
└──────────────────┘
```

---

## 五、AI Agent 架构

### 5.1 当前架构（Python）

```
AgenticLoop
  │
  ├── LLMClient (DeepSeek API)
  │     └── AsyncOpenAI (OpenAI 兼容接口)
  │
  ├── SearchTool (ddgs)
  │     └── DDGS.text() → DuckDuckGo 搜索
  │
  ├── System Prompt (首席分析师)
  │     ├── 角色定义
  │     ├── 搜索规则（强制搜索，来源标注）
  │     └── 报告模板（8 大模块）
  │
  ├── on_event 回调 → SSE 推送
  │     ├── thinking
  │     ├── searching
  │     ├── search_result
  │     └── done
  │
  └── max_iterations = 6
```

### 5.2 迁移目标（Eino）

```
Eino ReAct Agent
  │
  ├── ChatModel (OpenAI 兼容)
  │     └── base_url: https://api.deepseek.com
  │
  ├── Tool 工具节点
  │     └── SearchTool (ddgs)
  │
  ├── Graph 编排
  │     ├── ChatTemplate → 系统提示词
  │     ├── ToolsNode → 工具调用
  │     └── 条件判断 → 继续搜索 / 输出报告
  │
  └── StreamCallback → SSE 推送
```

---

## 六、本地开发环境

### 6.1 启动顺序

```bash
# 1. Python Agent（过渡）
cd python-agent
source venv/bin/activate
uvicorn api:app --host 0.0.0.0 --port 8000

# 2. Go 后端
cd competitor-backend
go run main.go

# 3. 前端
cd competitor-frontend/arco-design-pro-vite
npx vite --config ./config/vite.config.dev.ts --host
```

### 6.2 端口分配

| 服务 | 端口 |
|------|------|
| Python Agent | 8000 |
| Go Hertz 后端 | 8888 |
| Vite 前端 | 5173 |

### 6.3 代理配置

Vite 开发服务器代理：

```
/api/*         → :8888 (Go 后端)
/analyze/*     → :8000 (Python Agent SSE)
```

---

## 七、容器化部署（待建）

```yaml
# docker-compose.yml（目标配置）
version: "3.9"
services:
  agent:
    build: ./competitor-backend  # Eino Agent 内置
    ports: ["8000:8000"]
  backend:
    build: ./competitor-backend
    ports: ["8888:8888"]
  frontend:
    build: ./competitor-frontend/arco-design-pro-vite
    ports: ["80:80"]
```

---

## 八、未来演进路线

| 阶段 | 内容 |
|------|------|
| **Phase 1（当前）** | Python Agent 过渡 + Go Hertz 后端 + Vue 前端 |
| **Phase 2** | Eino 替换 Python Agent，全 Go 栈闭环 |
| **Phase 3** | 接入数据库 PostgreSQL，实现持久化 |
| **Phase 4** | Docker 容器化，一键部署 |
| **Phase 5** | 字节官方数据源替换 ddgs 搜索 |
