import os
import json
from typing import List, Dict, Any, Optional
from openai import AsyncOpenAI
from dotenv import load_dotenv

class LLMClient:
    """封装火山引擎（豆包）大模型 API"""
    
    def __init__(self, api_key: str = None, base_url: str = None, default_model: str = None):
        self.api_key = api_key or os.getenv("VOLCENGINE_API_KEY")
        self.base_url = base_url or os.getenv("VOLCENGINE_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3")
        self.default_model = default_model or os.getenv("DOUBAO_MODEL_EP", "ep-xxxxx") # 需要替换为实际的接入点
        
        if not self.api_key:
            print("Warning: VOLCENGINE_API_KEY is not set. LLM calls will fail.")
            
        self.client = AsyncOpenAI(
            api_key=self.api_key,
            base_url=self.base_url,
        )

    async def chat(
        self, 
        messages: List[Dict[str, str]], 
        model: str = None,
        tools: List[Dict[str, Any]] = None,
        temperature: float = 0.7
    ) -> Any:
        """
        发起对话请求，支持 Function Calling (Tools)
        """
        target_model = model or self.default_model
        
        kwargs = {
            "model": target_model,
            "messages": messages,
            "temperature": temperature,
        }
        
        if tools:
            kwargs["tools"] = tools
            kwargs["tool_choice"] = "auto"
            
        try:
            response = await self.client.chat.completions.create(**kwargs)
            return response.choices[0].message
        except Exception as e:
            print(f"LLM Call Error: {str(e)}")
            raise e
