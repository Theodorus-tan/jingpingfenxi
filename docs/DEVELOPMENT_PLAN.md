# 竞品分析系统 — 开发落地计划清单 (Development Roadmap)

基于 [BUSINESS_ARCHITECTURE.md](./BUSINESS_ARCHITECTURE.md) 中确立的“层级架构（企业-选品-分析）”、“3D分析模型”与“场景驱动机制”，制定以下落地开发计划。

---

## 阶段一：基础骨架与数据流重构 (Phase 1: Skeleton & Data Flow)
**目标**：打通包含“业务场景 (Scenario)”与“分析维度 (3D)”的端到端数据流。

- [ ] **1.1 后端数据模型升级 (Go/Hertz)**
  - 在 `internal/biz/model` 中定义完整的层级结构：`Enterprise` -> `Project` -> `AnalysisTask`。
  - 为 `AnalysisTask` 结构体新增核心字段：
    - `Scenario` (枚举：`Product_Improvement` / `Market_Entry`)
    - `Status_Review` (维度1状态)
    - `Status_Strategy` (维度2状态)
    - `Status_Finance` (维度3状态)
- [ ] **1.2 前端“新建分析”交互改造 (Vue3)**
  - 修改 `src/views/analysis/new`，将原本单一的输入框改为“场景化表单”。
  - 增加“分析诉求”单选框（已有产品求改进 / 无产品求入局）。
  - 增加“所属选品工程”下拉选择。
- [ ] **1.3 前端“分析报告看板”改造 (Vue3)**
  - 将单篇 Markdown 报告页，改造为包含三个 Tab 的大屏看板：`产品 Review 诊断` | `宏观商业战略` | `财务与生命力评估`。
  - 增加一个置顶的“基于当前场景的 AI 核心执行建议”高亮模块。

---

## 阶段二：Agent 编排与提示词注入 (Phase 2: Agent Orchestration)
**目标**：将当前的单体 Agent 升级为具备路由分发能力的“多路并发专家系统”。

- [ ] **2.1 主控 Agent 开发 (Master Router)**
  - 接收用户任务及场景变量，将其拆解并并行派发给三个维度的子 Agent。
- [ ] **2.2 维度子 Agent 提示词工程落地**
  - 根据 `BUSINESS_ARCHITECTURE.md`，在代码中实现三个 Prompt 模板，并打通 `{{Scenario}}` 变量的动态注入：
    - [ ] `Review 分析师 Agent`
    - [ ] `泛战略分析师 Agent`
    - [ ] `股市金融分析师 Agent`
- [ ] **2.3 Eino Graph 编排构建**
  - 使用 Eino 框架的 Graph 机制，构建并行的执行流，确保三个 Agent 能够同时工作并支持独立的 SSE 进度推送。

---

## 阶段三：垂直领域数据源接入与降级机制 (Phase 3: Domain Data Sources & Fallback)
**目标**：为各个领域的分析师 Agent 配备专属的数据抓取工具 (Tools)，并实现健壮的数据降级与熔断机制。

- [ ] **3.1 Review 数据源与降级**
  - **首选**：接入开源的电商评论抓取 API 或爬虫（如亚马逊、京东商品页）。
  - **降级 (Fallback)**：若无电商数据，切换至通用搜索工具，以“产品名+坑/怎么样”为关键词抓取社交媒体（知乎/Reddit）舆情。
- [ ] **3.2 战略广谱数据源与降级**
  - **首选**：企查查/天眼查 API，或针对百科、官网的深度结构化提取。
  - **降级 (Fallback)**：若为低调初创企业，退回基础搜索引擎的 Snippet 摘要总结，并标记“低置信度”。
- [ ] **3.3 财务股市数据源与熔断**
  - **首选**：接入免费的金融/股票 API（如 AlphaVantage、Tushare）获取上市公司三表数据。
  - **降级 (Fallback)**：若未上市，尝试搜索其天眼查融资历程。
  - **熔断 (Circuit Breaker)**：若完全无资本数据，Agent 触发熔断返回空状态，前端展示缺省状态（防幻觉）。

---

## 阶段四：系统加固与商业化预备 (Phase 4: Hardening & Persistence)
**目标**：完成持久化存储，保障系统在真实企业环境下的高可用性。

- [ ] **4.1 数据库接入**
  - 引入 GORM，连接 PostgreSQL 或 MySQL。
  - 将 `localStorage` 中的历史记录迁移至真实的数据库表中，实现“企业级数据隔离”。
- [ ] **4.2 缓存与并发控制**
  - 引入 Redis，控制同企业下的高并发调用频次（限流 Rate Limiting）。
  - 缓存热门竞品的通用分析结果，降低大模型 Token 成本。
- [ ] **4.3 用户系统与权限验证**
  - 引入 JWT Token，实现基于 Enterprise ID 的租户隔离。