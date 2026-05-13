# AI 竞品分析系统

> 字节跳动科创竞赛项目 | 双引擎 AI Agent 自动生成结构化竞品分析报告

输入竞品名称，AI Agent 自动搜索互联网公开信息，通过大模型深度分析，输出专业竞品分析报告。支持 **Python Agent（流式）** 和 **Eino Agent（快速）** 两种模式。

---

## 技术栈

| 模块 | 技术栈 | 说明 |
|------|--------|------|
| **前端** | Vue 3 + TypeScript + Vite + Arco Design | 字节官方 UI 设计语言 |
| **后端** | Go 1.26 + Hertz（CloudWeGo） | 字节自研 Web 框架 |
| **AI Agent 引擎** | Python（FastAPI）/ Eino（CloudWeGo Go 框架） | 双引擎可选 |
| **大模型** | DeepSeek Chat API | OpenAI 兼容接口 |
| **搜索引擎** | 原生 Bing HTML 解析（Go）/ Bing API / ddgs（Python） | 无需第三方 Key |

---

## 快速开始

### 前置条件

- Go 1.26+
- Python 3.13+
- Node.js 18+
- pnpm 或 npm

### 1. 配置 API Key

```bash
cp python-agent/.env.example python-agent/.env
# 填写你的 DeepSeek API Key
```

```env
VOLCENGINE_API_KEY=sk-你的API_KEY
DOUBAO_MODEL_EP=deepseek-chat
VOLCENGINE_BASE_URL=https://api.deepseek.com
```

### 2. 启动服务

```bash
# 终端 1：Python Agent（如果使用 Eino 模式则无需启动）
cd python-agent
source venv/bin/activate
uvicorn api:app --host 0.0.0.0 --port 8000

# 终端 2：Go 后端
cd competitor-backend
go run main.go

# 终端 3：前端
cd competitor-frontend/arco-design-pro-vite
npm run dev
```

访问 [http://localhost:5173](http://localhost:5173)

### 端口一览

| 服务 | 端口 |
|------|------|
| 前端 | 5173 |
| Go 后端（Eino Agent） | 8888 |
| Python Agent | 8000 |

---

## 功能特性

- **双引擎可选** — 同一页面切换 Python Agent（流式）和 Eino Agent（快速）
- **实时透明** — Agent 的搜索、思考过程通过 SSE 实时推送，清晰可见
- **结构化报告** — 自动生成包含执行摘要、核心功能、用户口碑、SWOT 分析等模块的专业报告
- **动态菜单** — 侧边栏菜单由后端驱动，前端无硬编码
- **竞品管理** — 手动添加竞品 + 自动从分析历史沉淀竞品库
- **数据持久化** — 分析历史、竞品列表、报告均持久化到 localStorage，刷新不丢失

---

## 侧边栏菜单

| 菜单 | 路径 | 说明 |
|------|------|------|
| 📊 工作台 | `/dashboard/workbench` | 快捷入口 + 概览统计 + 最近分析 |
| 📋 竞品管理 | `/competitors/list` | 由后端 `/api/menus` 驱动 |
| 📈 分析概览 | `/analysis/new` | 由后端 `/api/menus` 驱动 |

---

## 项目结构

```
competitor-agent/
├── competitor-backend/         # Go + Hertz 后端
│   ├── main.go                 # 服务入口 + 路由
│   └── internal/pkg/eino/      # Eino Agent 实现
│       ├── agent.go            # ReAct Agent 定义
│       ├── deepseek_model.go   # DeepSeek 模型适配器
│       └── search_tool.go      # Bing 搜索工具
├── competitor-frontend/        # Vue 3 前端
│   └── arco-design-pro-vite/
│       ├── src/views/          # 页面组件
│       ├── src/api/            # 接口封装
│       └── src/router/         # 路由配置
├── python-agent/               # Python Agent（过渡方案）
│   ├── api.py                  # FastAPI 服务
│   ├── core/loop.py            # Agent 推理循环
│   └── tools/search_tool.py    # 搜索工具
└── docs/                       # 项目文档
```

---

## 分析流程

```
用户输入竞品名
    │
    ▼
前端选择分析引擎（Python / Eino）
    │
    ▼
AI Agent 执行分析：
    ├─ 搜索互联网公开信息（Bing / ddgs）
    ├─ DeepSeek LLM 多轮推理
    └─ 生成结构化分析报告
    │
    ▼
前端实时展示搜索过程和思考日志（SSE）
    │
    ▼
完成 → Markdown 渲染报告
```

---

## 测试

```bash
# Python Agent
cd python-agent && pytest tests/ -v

# Go 后端
cd competitor-backend && go test ./... -v

# 前端
cd competitor-frontend/arco-design-pro-vite && npx vitest
```

---

## License

MIT
