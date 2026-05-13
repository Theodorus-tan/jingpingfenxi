import json
import pytest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from tools.search_tool import SearchTool


@pytest.mark.asyncio
async def test_search_tool__basic_query():
    tool = SearchTool()
    result = await tool.execute(query="DJI Action 4 review", max_results=3)

    assert result.success is True
    data = json.loads(result.data)
    assert isinstance(data, list)
    assert len(data) <= 3
    if data and data[0]["title"] != "搜索结果为空" and data[0]["title"] != "搜索暂时不可用":
        assert "title" in data[0]
        assert "snippet" in data[0]


@pytest.mark.asyncio
async def test_search_tool__network_error_graceful():
    tool = SearchTool()

    async def raise_error(*args, **kwargs):
        raise Exception("Connection timeout")

    tool._search_bing = raise_error

    result = await tool.execute(query="xxxnonexistentxxx")
    assert result.success is True
    data = json.loads(result.data)
    assert len(data) == 1
    assert "知识库" in data[0]["snippet"] or "训练数据" in data[0]["snippet"]


@pytest.mark.asyncio
async def test_search_tool__empty_result():
    tool = SearchTool()

    async def return_empty(*args, **kwargs):
        return []

    tool._search_bing = return_empty

    result = await tool.execute(query="zzz_unlikely_query_zzz")
    assert result.success is True
    data = json.loads(result.data)
    assert "知识库" in data[0]["snippet"] or "搜索结果为空" in data[0]["snippet"]
