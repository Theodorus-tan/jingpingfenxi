# AI 竞品分析系统 — 测试规范

> 每搞一个新模块，先写测试。没测试的代码不算完成。

---

## 一、测试金字塔策略

本项目采用经典三层测试金字塔：

```
          ╱── UI E2E ──╲           ← 少量，覆盖核心用户旅程（待建）
         ╱──── API ────╲          ← 中层，覆盖接口逻辑
        ╱─── 单元测试 ───╲         ← 大量，覆盖每个函数/组件
       ╱─────────────────╲
      ╱  Python  │  Go  │  Vue   ╲
     ╱─────────────────────────────╲
```

| 层级 | 覆盖目标 | 数量占比 | 运行速度 |
|------|---------|---------|---------|
| 单元测试 | 函数、方法、组件 | 70% | 毫秒级 |
| API 测试 | Handler、Service、路由 | 20% | 秒级 |
| UI E2E | 核心用户旅程 | 10% | 分钟级（待建） |

---

## 二、各模块测试框架选型

| 模块 | 框架 | 断言库 | Mock | 覆盖率工具 |
|------|------|--------|------|-----------|
| **Python Agent** | `pytest` 8.x | `assert` | `pytest-asyncio` + `unittest.mock` | `pytest-cov` |
| **Go 后端** | `go test` + `testify` | `testify/assert` | `testify/mock` | `go test -cover` |
| **前端** | `vitest` + `vue-test-utils` | `vitest` 内置 | `vi.mock()` | `vitest --coverage` |

---

## 三、Python-Agent 测试规范

### 3.1 环境准备

```bash
cd python-agent
pip install pytest pytest-asyncio pytest-cov pytest-mock
```

### 3.2 目录结构

```
python-agent/
├── core/
│   ├── llm_client.py
│   └── loop.py
├── tools/
│   ├── base.py
│   └── search_tool.py
├── tests/
│   ├── test_search_tool.py          # 搜索工具测试
│   ├── test_agentic_loop.py         # Agent 循环测试
│   └── mocks/
│       └── llm_response.py          # Mock 数据
```

### 3.3 命名规范

```python
# 文件命名：test_{module_name}.py
# 函数命名：test_{function_name}__{scenario}

def test_search_tool__basic_query(): ...
def test_search_tool__empty_result(): ...
def test_agentic_loop__no_tool_call(): ...
def test_agentic_loop__max_iterations(): ...
```

### 3.4 搜索工具测试

```python
# tests/test_search_tool.py
@pytest.mark.asyncio
async def test_search_tool__basic_query():
    """搜索工具：正常关键词应该返回结果列表"""
    tool = SearchTool()
    result = await tool.execute(query="DJI Action 4 review", max_results=3)
    assert result.success is True
    data = json.loads(result.data)
    assert isinstance(data, list)
    assert len(data) <= 3


@pytest.mark.asyncio
async def test_search_tool__network_error_graceful():
    """搜索工具：网络异常优雅降级，返回 success=True + 提示文本"""
    tool = SearchTool()
    # Mock ddgs 抛异常
    tool.execute = AsyncMock(side_effect=Exception("Connection timeout"))

    result = await tool.execute(query="xxxnonexistentxxx")
    assert result.success is True
```

### 3.5 Agent Loop 测试

```python
# tests/test_agentic_loop.py
@pytest.mark.asyncio
async def test_loop__no_tool_call_returns_directly():
    """Agent 循环：LLM 没有调用工具时，直接返回内容"""
    mock_llm = MockLLM()
    mock_llm.set_responses([make_final_response("分析结果")])

    loop = AgenticLoop(llm_client=mock_llm, tools=[MockTool()], system_prompt="你是一个分析师")
    result = await loop.run("分析一下")
    assert "分析结果" in result
    assert mock_llm.call_count == 1


@pytest.mark.asyncio
async def test_loop__tool_call_then_final():
    """Agent 循环：LLM 先调用工具，再用结果生成最终回答"""
    mock_llm = MockLLM()
    mock_llm.set_responses([
        make_tool_call_response("mock_tool", '{"input": "test"}'),
        make_final_response("最终报告"),
    ])

    loop = AgenticLoop(llm_client=mock_llm, tools=[MockTool()], system_prompt="你是一个分析师")
    result = await loop.run("分析竞品")
    assert "最终报告" in result
    assert mock_llm.call_count == 2
```

### 3.6 运行测试

```bash
# 全部测试
pytest python-agent/tests/ -v

# 带覆盖率
pytest python-agent/tests/ --cov=core --cov=tools --cov-report=term

# 指定文件
pytest python-agent/tests/test_search_tool.py -v
```

### 3.7 覆盖率要求

| 模块 | 当前覆盖率 | 目标 |
|------|-----------|------|
| `tools/` | 73% | 80% |
| `core/loop.py` | 73% | 85% |

---

## 四、Go 后端测试规范

### 4.1 环境准备

```bash
cd competitor-backend
go get github.com/stretchr/testify
```

### 4.2 目录结构

```
competitor-backend/
├── internal/
│   ├── handler/
│   │   ├── competitor_handler.go      # 待重构为独立文件
│   │   └── competitor_handler_test.go # 存在但不可运行
│   ├── service/
│   │   ├── competitor_service.go
│   │   └── competitor_service_test.go
│   └── model/
│       └── competitor.go
├── main.go                            # 当前逻辑在此，待重构到 cmd/api/
└── go.mod
```

当前状态说明：`main.go` 中所有逻辑写在 `main()` 匿名函数中，尚未拆分为独立 handler 文件。现有的 handler/service 层代码（`internal/`）未被 `main.go` 引用。计划重构目录结构对齐字节标准后，补充完整测试。

### 4.3 运行测试

```bash
# 全部测试
cd competitor-backend && go test ./... -v

# 带覆盖率
go test ./... -cover -coverprofile=coverage.out

# 指定包
go test ./internal/service/... -v
```

### 4.4 覆盖率要求

| 层 | 当前覆盖率 | 目标 |
|----|-----------|------|
| `service/` | 88.2% | 85% ✅ |
| `handler/` | 未接入 | 75% |

---

## 五、前端测试规范

### 5.1 环境准备

```bash
cd competitor-frontend/arco-design-pro-vite
pnpm add -D vitest @vue/test-utils jsdom @vitest/coverage-v8
```

### 5.2 目录结构

```
arco-design-pro-vite/
├── src/
│   ├── utils/
│   │   ├── format.ts
│   │   └── __tests__/
│   │       └── format.test.ts
│   └── test/
│       └── setup.ts
```

### 5.3 工具函数测试

```typescript
// src/utils/__tests__/format.test.ts
describe('formatPrice', () => {
  it('整数价格返回带货币符号格式', () => {
    expect(formatPrice(2599)).toBe('¥2,599')
  })
  it('0 元返回 "免费"', () => {
    expect(formatPrice(0)).toBe('免费')
  })
})

describe('truncateText', () => {
  it('短于限制不截断', () => {
    expect(truncateText('Hello', 10)).toBe('Hello')
  })
  it('长于限制添加省略号', () => {
    expect(truncateText('这是一段非常长的文本', 5)).toBe('这是一段...')
  })
})

describe('calcScoreColor', () => {
  it('>= 8 分返回绿色', () => { expect(calcScoreColor(8.5)).toBe('green') })
  it('>= 5 分返回黄色', () => { expect(calcScoreColor(7)).toBe('yellow') })
  it('< 5 分返回红色', () => { expect(calcScoreColor(3)).toBe('red') })
})
```

### 5.4 运行测试

```bash
# 全部测试
npx vitest

# 带覆盖率
npx vitest --coverage

# 指定文件
npx vitest src/utils/__tests__/format.test.ts
```

### 5.5 覆盖率要求

| 类型 | 当前 | 目标 |
|------|------|------|
| Utils | ✅ 12 测试全部通过 | 90% |

---

## 六、Mock 规范

### 6.1 Mock 层级

```
层级 1: LLM 客户端 Mock
        → MockLLM 类（预设响应）

层级 2: 搜索引擎 Mock
        → 设置 execute 返回预设 JSON 或抛异常

层级 3: 外部 API Mock
        → Python: unittest.mock / AsyncMock
        → Go: testify/mock
        → Vue: vi.mock()
```

### 6.2 MockLLM 实现

```python
class MockLLM:
    def __init__(self):
        self.responses = []
        self.call_count = 0

    def set_responses(self, responses):
        self.responses = responses

    async def chat(self, **kwargs):
        resp = self.responses[self.call_count]
        self.call_count += 1
        return resp
```

---

## 七、命令速查表

```bash
# ── Python Agent ──
pytest python-agent/tests/ -v                                      # 全部
pytest python-agent/tests/ --cov=core --cov=tools --cov-report=term # 覆盖率
pytest python-agent/tests/ -k "search or loop" -v                  # 匹配测试名

# ── Go 后端 ──
go test ./... -v                                                    # 全部
go test ./internal/service/... -v                                   # 指定包
go test ./... -cover                                                # 覆盖率

# ── 前端 ──
npx vitest                                                          # 全部
npx vitest --coverage                                               # 覆盖率
npx vitest src/utils/__tests__/format.test.ts                       # 指定文件
```

## 八、当前覆盖状态总览

| 模块 | 测试数 | 覆盖状态 |
|------|--------|---------|
| Python Agent (tools) | 3 | ✅ 通过 |
| Python Agent (loop) | 3 | ✅ 通过 |
| Go 后端 (service) | 5 | ✅ 通过（88.2%） |
| Go 后端 (handler) | — | ❌ main.go 未拆分为独立文件 |
| 前端 (utils) | 12 | ✅ 通过 |
| 组件/API/E2E | — | ⏳ 待建 |
