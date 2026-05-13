import json
import asyncio
import os
import httpx
from .base import BaseTool, ToolResult


class SearchTool(BaseTool):
    name = "web_search"
    description = "在互联网上查找最新信息。当需要了解竞品最新动态、新闻、价格或用户评价时使用。"
    input_schema = {
        "type": "object",
        "properties": {
            "query": {
                "type": "string",
                "description": "搜索关键词，例如 'MacBook M4 review' 或 '大疆 Action 5 Pro 评测'"
            },
            "max_results": {
                "type": "integer",
                "description": "返回的最大结果数，默认 5"
            }
        },
        "required": ["query"]
    }

    _bing_api_key = os.getenv("BING_API_KEY", "")
    _bing_endpoint = "https://api.bing.microsoft.com/v7.0/search"

    async def execute(self, query: str, max_results: int = 5, **kwargs) -> ToolResult:
        # 优先使用 Bing Search API
        if self._bing_api_key:
            return await self._search_bing(query, max_results)
        # 没有 Bing Key 时使用 ddgs
        return await self._search_ddgs(query, max_results)

    async def _search_bing(self, query: str, max_results: int) -> ToolResult:
        """Bing Search API v7"""
        try:
            async with httpx.AsyncClient(timeout=10) as client:
                resp = await client.get(
                    self._bing_endpoint,
                    params={"q": query, "count": max_results, "mkt": "zh-CN"},
                    headers={"Ocp-Apim-Subscription-Key": self._bing_api_key},
                )
                resp.raise_for_status()
                data = resp.json()

            web_pages = data.get("webPages", {}).get("value", [])
            if not web_pages:
                return ToolResult(
                    success=True,
                    data=json.dumps([{
                        "title": "搜索结果为空",
                        "snippet": "未找到搜索结果，我会基于已有知识继续分析。",
                        "url": ""
                    }], ensure_ascii=False),
                )

            formatted = [
                {
                    "title": item.get("name", ""),
                    "snippet": item.get("snippet", ""),
                    "url": item.get("url", ""),
                }
                for item in web_pages[:max_results]
            ]
            return ToolResult(success=True, data=json.dumps(formatted, ensure_ascii=False))

        except Exception as e:
            return ToolResult(
                success=True,
                data=json.dumps([{
                    "title": "搜索暂时不可用",
                    "snippet": f"搜索服务暂时不可用（{str(e)[:50]}），我会基于训练数据和常识继续分析。",
                    "url": ""
                }], ensure_ascii=False),
            )

    async def _search_ddgs(self, query: str, max_results: int) -> ToolResult:
        """降级方案：DuckDuckGo 搜索（无需 API Key）"""
        try:
            from ddgs import DDGS
            results = await asyncio.to_thread(
                DDGS().text, query, max_results=max_results
            )
            if not results:
                return ToolResult(
                    success=True,
                    data=json.dumps([{
                        "title": "搜索结果为空",
                        "snippet": "未找到搜索结果，我会基于已有知识继续分析。",
                        "url": ""
                    }], ensure_ascii=False),
                )
            formatted = [
                {
                    "title": r["title"],
                    "snippet": r["body"],
                    "url": r["href"],
                }
                for r in results
            ]
            return ToolResult(success=True, data=json.dumps(formatted, ensure_ascii=False))
        except Exception as e:
            return ToolResult(
                success=True,
                data=json.dumps([{
                    "title": "搜索暂时不可用",
                    "snippet": "搜索服务暂时不可用，我会基于训练数据和常识继续分析。",
                    "url": ""
                }], ensure_ascii=False),
            )
