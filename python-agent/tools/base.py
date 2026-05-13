"""
工具基类定义
"""
from abc import ABC, abstractmethod
from typing import Any, Dict
from pydantic import BaseModel

class ToolResult(BaseModel):
    """工具执行结果"""
    success: bool
    data: Any = None
    error: str = None

class BaseTool(ABC):
    """工具基类"""
    name: str = "base_tool"
    description: str = "Base tool description"
    input_schema: Dict[str, Any] = {}

    @abstractmethod
    async def execute(self, **kwargs) -> ToolResult:
        """执行工具"""
        pass
