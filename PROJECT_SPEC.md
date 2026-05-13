# AI 竞品分析系统 — 项目规范文档

> 字节跳动科创竞赛项目 | Arco Design Pro + Go Hertz + Eino

---

## 一、项目概述

### 1.1 项目定位
基于 AI Agent 的多维度竞品分析系统。输入竞品名称，Agent 自动搜索互联网公开信息，通过大模型深度分析，输出结构化竞品分析报告。

### 1.2 技术栈总览

| 模块 | 技术栈 | 说明 |
|------|--------|------|
| **competitor-backend** | Go 1.26+ / Hertz / Eino | 字节自研 Web 框架 + AI Agent 框架，全 Go 栈 |
| **competitor-frontend** | Vue 3 / TypeScript / Vite / Arco Design | 字节官方 UI 设计语言 |
| **AI Agent 引擎** | Eino（Go） | 字节自研大模型应用框架，逐步替换 Python Agent |

### 1.3 技术选型说明

参照字节竞赛评审标准，技术选型优先级：

| 层级 | 首选 | 当前 |
|------|------|------|
| 后端语言 | Go 1.21+ | Go 1.26.3 ✅ |
| Web 框架 | Hertz（CloudWeGo） | Hertz v0.10.4 ✅ |
| AI 框架 | Eino（CloudWeGo） | 迁移中（当前 Python Agent 过渡） |
| 前端框架 | Vue 3+ / Arco Design | Vue 3 + Arco Design ✅ |
| 构建工具 | Vite 5+ | Vite 3.2.11（待升级） |
| 状态管理 | Pinia | Pinia ✅ |
| 代码规范 | ESLint + Prettier | ESLint + Prettier ✅ |
| 容器化 | Docker + Docker Compose | 待建 |
| 数据库 | PostgreSQL / SQLite | 待接入 |

---

## 二、目录结构规范

```
competitor-agent/                              # 项目根目录
├── competitor-backend/                        # Go + Hertz 后端
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                        # 服务启动入口
│   ├── internal/
│   │   ├── biz/                               # 业务逻辑层
│   │   │   ├── handler/                       # HTTP 请求处理
│   │   │   ├── service/                       # 核心业务逻辑
│   │   │   └── model/                         # 业务模型定义
│   │   ├── dal/                               # 数据访问层（待建）
│   │   │   ├── db/                            # 数据库操作
│   │   │   └── cache/                         # 缓存操作
│   │   ├── pkg/                               # 项目内部公共工具（待建）
│   │   │   ├── logger/                        # 日志工具
│   │   │   └── utils/                         # 通用工具函数
│   │   └── config/                            # 配置加载与解析
│   ├── configs/                               # 配置文件模板（待建）
│   ├── scripts/                               # 构建/部署脚本（待建）
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile                             # 容器化配置（待建）
│
├── competitor-frontend/                       # 前端
│   └── arco-design-pro-vite/
│       ├── src/
│       │   ├── api/                           # 接口请求封装
│       │   │   ├── interceptor.ts             # Axios 统一封装
│       │   │   └── analysis.ts                # 分析模块接口
│       │   ├── components/                    # 公共组件
│       │   ├── views/                         # 页面组件
│       │   ├── store/                         # 状态管理（Pinia）
│       │   ├── router/                        # 路由配置
│       │   ├── utils/                         # 工具函数
│       │   ├── assets/                        # 静态资源
│       │   ├── config/                        # 项目配置
│       │   ├── App.vue
│       │   └── main.ts
│       ├── config/                            # Vite 构建配置
│       ├── package.json
│       └── Dockerfile                         # 容器化配置（待建）
│
├── docs/                                      # 项目文档
│   ├── PROJECT_SPEC.md                        # 本文件
│   ├── ARCHITECTURE.md                        # 架构规范
│   ├── UI_DESIGN.md                           # UI 设计
│   └── TESTING.md                             # 测试规范
│
├── docker-compose.yml                         # 容器编排（待建）
├── .gitignore
└── README.md
```

---

## 三、后端规范

### 3.1 接口规范

统一返回格式：

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

状态码规范：

| 场景 | HTTP 状态码 | 业务 code |
|------|------------|-----------|
| 成功 | 200 | 200 |
| 请求参数错误 | 400 | -1 |
| 未授权 | 401 | -1 |
| 资源不存在 | 404 | -1 |
| 服务端错误 | 500 | -1 |

### 3.2 API 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/ping | 健康检查 |
| POST | /api/analysis/task | 提交分析任务 |
| POST | /analyze/stream | SSE 流式分析 |
| GET | /api/competitors | 竞品列表 |

### 3.3 编码规范

- 包名：全小写单数（`handler`、`service`、`model`）
- 导出标识符：大驼峰（`UserService`）
- 私有标识符：小驼峰（`userID`）
- 接口名：以 `er` 结尾（`Storer`、`Reader`）
- 使用 `gofmt` 自动格式化
- 错误必须处理，禁止忽略
- 日志结构化，区分级别

---

## 四、前端规范

### 4.1 技术选型

| 类别 | 选型 |
|------|------|
| 框架 | Vue 3 + Composition API + `<script setup>` |
| 构建 | Vite 3.2.x |
| 语言 | TypeScript（严格模式） |
| UI | Arco Design Vue 2.x |
| 状态管理 | Pinia |
| 路由 | Vue Router 4 |
| HTTP | Axios（封装在 `interceptor.ts`） |
| 样式 | Less + Arco Token 系统 |

### 4.2 编码规范

| 规则 | 要求 |
|------|------|
| 文件命名 | kebab-case（`task-list.vue`） |
| 组件命名 | PascalCase（`CompetitorCard.vue`） |
| 变量/函数 | camelCase |
| 缩进 | 2 个空格 |
| 字符串 | 单引号，结尾不加分号 |
| 类型 | 必须声明，禁止 `any` |

### 4.3 页面路由

| 路由 | 页面 | 说明 |
|------|------|------|
| `/` | 起始页 | 进入竞品分析入口 |
| `/dashboard/workbench` | 工作台 | 概览 + 最近分析 |
| `/competitors/list` | 竞品管理 | 竞品列表 CRUD |
| `/analysis/new` | 新建分析 | 输入竞品名，启动 Agent 分析 |
| `/analysis-report/view` | 分析报告 | AI 生成的完整报告展示 |

---

## 五、模块间通信

### 5.1 数据流

```
前端 (Vue) ──HTTP──→ Go 后端 (Hertz) ──HTTP──→ AI Agent (Eino/Python)
                                                      │
                                                      ├── ddgs 搜索工具
                                                      └── DeepSeek API
```

### 5.2 通信协议

| 方向 | 协议 | 说明 |
|------|------|------|
| 前端 → 后端 | HTTP/REST | 同步请求 |
| 前端 → Agent（SSE） | HTTP/SSE | 流式获取 Agent 思考过程 |
| 后端 → Agent | HTTP/REST | 触发分析任务 |
| Agent → 外部 | HTTP | ddgs 搜索、DeepSeek API |

### 5.3 分析流程时序

```
用户输入竞品名 →
前端表单 →
POST /api/analysis/task → Go 后端 →
POST Python Agent /analyze →
  ├─ ddgs 搜索竞品信息
  ├─ DeepSeek LLM 分析推理
  └─ 返回结构化报告 →
Go 后端返回报告 →
前端渲染 Markdown 报告
```

---

## 六、Git 规范

### 6.1 提交规范

```
<类型>: <描述>

feat: 新增竞品分析任务创建接口
fix: 修复搜索工具超时问题
docs: 更新架构文档
style: 调整页面样式
refactor: 重构 Agent 循环逻辑
test: 添加搜索工具单元测试
chore: 更新依赖
```

### 6.2 分支策略

```
main       ← 稳定版本
  └── dev  ← 开发主分支
```

### 6.3 .gitignore

```gitignore
# Python
python-agent/venv/
python-agent/.env
python-agent/__pycache__/
*.pyc

# Node
node_modules/
dist/
.env.development.local
.env.production.local

# Go
competitor-backend/main_app
*.exe

# IDE
.idea/
.vscode/
*.swp
.DS_Store
```

---

## 七、容器化（待建）

- Dockerfile（Go 后端多阶段构建）
- Dockerfile（前端 Nginx 静态部署）
- docker-compose.yml（一键启动）

---

## 八、设计原则

1. **字节原生偏好**：优先使用 Hertz、Eino、Arco Design 等字节自研技术
2. **Fail Gracefully**：外部服务不可用时优雅降级
3. **先写测试**：每搞一个新模块，先写单元测试
4. **决策透明**：Agent 思考过程实时可见，结论标注数据来源
