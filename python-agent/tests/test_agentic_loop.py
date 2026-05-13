import json
import pytest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from core.loop import AgenticLoop
from tools.base import BaseTool, ToolResult
from tests.mocks.llm_response import make_tool_call_response, make_final_response


class MockTool(BaseTool):
    name = "mock_tool"
    description = "test tool"
    input_schema = {
        "type": "object",
        "properties": {"input": {"type": "string"}},
        "required": ["input"],
    }

    async def execute(self, **kwargs):
        return ToolResult(success=True, data=json.dumps({"result": "ok"}))


class MockLLM:
    def __init__(self):
        self.responses = []

    def set_responses(self, responses):
        self.responses = responses
        self.call_count = 0

    async def chat(self, **kwargs):
        resp = self.responses[self.call_count]
        self.call_count += 1
        return resp


@pytest.mark.asyncio
async def test_loop__no_tool_call_returns_directly():
    mock_llm = MockLLM()
    mock_llm.set_responses([make_final_response("分析结果")])

    loop = AgenticLoop(llm_client=mock_llm, tools=[MockTool()], system_prompt="你是一个分析师")

    result = await loop.run("分析一下")
    assert "分析结果" in result
    assert mock_llm.call_count == 1


@pytest.mark.asyncio
async def test_loop__tool_call_then_final():
    mock_llm = MockLLM()
    mock_llm.set_responses(
        [
            make_tool_call_response("mock_tool", json.dumps({"input": "test"})),
            make_final_response("最终报告"),
        ]
    )

    loop = AgenticLoop(llm_client=mock_llm, tools=[MockTool()], system_prompt="你是一个分析师")

    result = await loop.run("分析竞品")
    assert "最终报告" in result
    assert mock_llm.call_count == 2


@pytest.mark.asyncio
async def test_loop__max_iterations_limit():
    mock_llm = MockLLM()
    mock_llm.set_responses(
        [make_tool_call_response("mock_tool", json.dumps({"input": "loop"}))] * 5
    )

    loop = AgenticLoop(llm_client=mock_llm, tools=[MockTool()], system_prompt="你是一个分析师")
    loop.max_iterations = 3

    result = await loop.run("分析")
    assert "最大思考次数" in result
    assert mock_llm.call_count == 3
