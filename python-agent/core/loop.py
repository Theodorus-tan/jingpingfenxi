import json
import logging
from typing import List, Dict, Any, Callable, Optional
from core.llm_client import LLMClient
from tools.base import BaseTool

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("AgentLoop")

class AgenticLoop:
    """
    极简版 Agent 核心循环
    支持多轮推理和工具调用 (Function Calling)
    支持通过 on_event 回调流式输出思考过程
    """
    def __init__(self, llm_client: LLMClient, tools: List[BaseTool], system_prompt: str):
        self.llm = llm_client
        self.tools = {tool.name: tool for tool in tools}
        self.system_prompt = system_prompt
        self.max_iterations = 6
        
    def _get_openai_tools_schema(self) -> List[Dict[str, Any]]:
        """将 Python Tools 转换为 OpenAI Function Calling Schema"""
        schemas = []
        for tool in self.tools.values():
            schemas.append({
                "type": "function",
                "function": {
                    "name": tool.name,
                    "description": tool.description,
                    "parameters": tool.input_schema
                }
            })
        return schemas if schemas else None

    async def run(self, user_input: str, on_event: Optional[Callable] = None) -> str:
        """运行 Agent 循环
        Args:
            user_input: 用户输入
            on_event: 可选回调，接收 (event_type: str, data: dict)
                      事件类型: 'thinking', 'searching', 'search_result', 'writing', 'done', 'error'
        """
        messages = [
            {"role": "system", "content": self.system_prompt},
            {"role": "user", "content": user_input}
        ]
        
        tools_schema = self._get_openai_tools_schema()
        
        if on_event:
            await on_event("thinking", {"message": "开始分析任务..."})
        
        for iteration in range(self.max_iterations):
            logger.info(f"--- Iteration {iteration + 1} ---")
            
            if on_event:
                await on_event("thinking", {"message": f"第 {iteration + 1} 轮分析..."})
            
            # 1. 调用 LLM
            response_message = await self.llm.chat(messages=messages, tools=tools_schema)
            
            # OpenAI 的返回可能没有 tool_calls 字段或为空
            has_tool_calls = hasattr(response_message, "tool_calls") and response_message.tool_calls
            
            # 深拷贝一份，否则可能引发 Pydantic 模型解析错误
            msg_dict = {"role": "assistant"}
            if response_message.content is not None:
                msg_dict["content"] = response_message.content
            
            if has_tool_calls:
                msg_dict["tool_calls"] = [
                    {
                        "id": t.id,
                        "type": t.type,
                        "function": {"name": t.function.name, "arguments": t.function.arguments}
                    } for t in response_message.tool_calls
                ]
            messages.append(msg_dict)
            
            # 2. 检查是否有工具调用
            if not has_tool_calls:
                # 如果 AI 没有调用工具，强制注入一条搜索指令提醒它
                if iteration == 0:
                    logger.warning("AI 未调用搜索工具，强制注入搜索指令...")
                    if on_event:
                        await on_event("thinking", {"message": "正在准备搜索信息..."})
                    messages.append({
                        "role": "user",
                        "content": "【系统提示】你刚才没有使用搜索工具。请立即使用 web_search 工具搜索相关信息，不要直接给出结论。先搜索，再分析。"
                    })
                    continue
                if on_event:
                    await on_event("writing", {"message": "报告生成完成"})
                logger.info("Agent 完成任务，输出最终结果。")
                return response_message.content
                
            # 3. 执行工具调用
            for tool_call in response_message.tool_calls:
                func_name = tool_call.function.name
                func_args = json.loads(tool_call.function.arguments)
                
                logger.info(f"🛠️ 调用工具: {func_name}, 参数: {func_args}")
                
                query = func_args.get("query", "")
                if on_event:
                    await on_event("searching", {"query": query})
                
                if func_name not in self.tools:
                    tool_result = f"Error: Tool {func_name} not found."
                else:
                    tool = self.tools[func_name]
                    result = await tool.execute(**func_args)
                    if result.success:
                        tool_result = result.data
                    else:
                        tool_result = f"Error: {result.error}"
                        
                logger.info(f"✅ 工具返回 (前100字符): {str(tool_result)[:100]}...")
                
                titles = []
                try:
                    parsed = json.loads(tool_result)
                    if isinstance(parsed, list):
                        titles = [item.get("title", "") for item in parsed if item.get("title")]
                except json.JSONDecodeError:
                    pass
                
                if on_event:
                    await on_event("search_result", {
                        "query": query,
                        "results_count": len(titles),
                        "titles": titles[:3]
                    })
                
                # 将工具结果追加到历史记录中
                messages.append({
                    "role": "tool",
                    "tool_call_id": tool_call.id,
                    "name": func_name,
                    "content": str(tool_result)
                })
                
        logger.warning("达到最大迭代次数，强制结束。")
        if on_event:
            await on_event("error", {"message": "分析超时，未能在限定步数内完成"})
        return "达到最大思考次数，强制停止。"
