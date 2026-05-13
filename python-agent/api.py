import json
import asyncio
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
import uvicorn
from core.llm_client import LLMClient
from core.loop import AgenticLoop
from tools.search_tool import SearchTool
import os
from dotenv import load_dotenv

load_dotenv()

app = FastAPI(title="Competitor Analysis Agent API")

search_tool = SearchTool()

class SearchRequest(BaseModel):
    query: str
    max_results: int = 5

@app.post("/web_search")
async def web_search(req: SearchRequest):
    result = await search_tool.execute(query=req.query, max_results=req.max_results)
    return {"code": 200, "msg": "success", "data": json.loads(result.data)}

CHIEF_ANALYST_PROMPT = """你是一个顶级的商业竞品分析师（Chief Competitive Analyst）。
你的任务是根据用户提供的竞品名称或线索，进行深度的竞品分析。

重要规则：
1. 你的训练数据可能过时，你必须使用 web_search 工具获取最新信息。
2. 先搜索，再分析。进行 1-2 次搜索获取足够信息后，立即输出报告。不要反复搜索。
3. 报告中每个结论都必须标注数据来源：
   - 来自搜索结果的数据标注为：🔍 [来源：搜索标题]
   - 基于行业常识的推演标注为：📊 [基于行业知识推演]
4. 搜索时请使用中英文关键词结合。

分析步骤：
1. 搜索该竞品的核心功能、官方网站和最新动态。
2. 搜索该产品的用户评价和定价信息。
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

获得搜索结果后，立即停止搜索并输出最终报告。
报告语言必须使用中文，结构清晰，专业客观。
每个数据点尽量标注来源链接，让读者可追溯。
"""

class AnalyzeRequest(BaseModel):
    competitor_name: str

class AnalyzeResponse(BaseModel):
    report: str

@app.post("/analyze", response_model=AnalyzeResponse)
async def analyze_competitor(req: AnalyzeRequest):
    llm = LLMClient()
    search_tool = SearchTool()
    
    agent = AgenticLoop(
        llm_client=llm,
        tools=[search_tool],
        system_prompt=CHIEF_ANALYST_PROMPT
    )
    
    final_report = await agent.run(f"请对 '{req.competitor_name}' 进行全面的竞品分析。")
    return AnalyzeResponse(report=final_report)

@app.post("/analyze/stream")
async def analyze_competitor_stream(req: AnalyzeRequest):
    llm = LLMClient()
    search_tool = SearchTool()

    agent = AgenticLoop(
        llm_client=llm,
        tools=[search_tool],
        system_prompt=CHIEF_ANALYST_PROMPT
    )

    async def event_generator():
        queue = asyncio.Queue()

        async def on_event(event_type: str, data: dict):
            await queue.put({"type": event_type, **data})

        async def run_agent():
            try:
                report = await agent.run(
                    f"请对 '{req.competitor_name}' 进行全面的竞品分析。",
                    on_event=on_event
                )
                await queue.put({"type": "done", "report": report})
            except Exception as e:
                await queue.put({"type": "error", "message": str(e)})

        task = asyncio.create_task(run_agent())

        while True:
            try:
                msg = await asyncio.wait_for(queue.get(), timeout=0.5)
                yield f"data: {json.dumps(msg, ensure_ascii=False)}\n\n"
                if msg.get("type") in ("done", "error"):
                    break
            except asyncio.TimeoutError:
                if task.done() and queue.empty():
                    break
                continue

        if not task.done():
            task.cancel()

    return StreamingResponse(
        event_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        }
    )

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
