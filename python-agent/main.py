import asyncio
import os
from dotenv import load_dotenv

from core.llm_client import LLMClient
from core.loop import AgenticLoop
from tools.search_tool import SearchTool

# 尝试加载环境变量 (你需要创建一个 .env 文件)
load_dotenv()

# ==========================================
# 首席分析师 Agent (Lead Agent) System Prompt
# ==========================================
CHIEF_ANALYST_PROMPT = """你是一个顶级的商业竞品分析师（Chief Competitive Analyst）。
你的任务是根据用户提供的竞品名称或线索，进行深度的竞品分析。

你可以使用搜索引擎工具 (web_search) 来查找信息。
分析步骤：
1. 搜索该竞品的官方网站和核心产品信息。
2. 搜索该产品的用户评价和口碑。
3. 搜索该产品的定价和市场定位。
4. 综合以上信息，输出一份结构化的竞品分析报告，包含：
   - 竞品简介
   - 核心功能与卖点
   - 用户真实口碑（优点/缺点）
   - 定价策略
   - SWOT 分析（优势/劣势/机会/威胁）

在最终回答之前，请确保你已经通过工具获取了足够的数据支撑你的观点。
如果搜索不到足够的信息，请基于你的常识进行合理的推断，并在报告中注明。
报告语言必须使用中文，结构清晰，专业客观。
"""

async def main():
    print("="*50)
    print("🤖 字节竞赛 Demo - AI 竞品分析 Agent 启动")
    print("="*50)
    
    # 检查环境变量
    api_key = os.getenv("VOLCENGINE_API_KEY")
    model_ep = os.getenv("DOUBAO_MODEL_EP")
    
    if not api_key or not model_ep:
        print("\n⚠️ 警告：环境变量未配置！")
        print("请在 ai_agent 目录下创建 .env 文件，并配置：")
        print("VOLCENGINE_API_KEY=你的火山引擎API_KEY")
        print("DOUBAO_MODEL_EP=你的模型接入点ID (如 ep-2025xxxx)")
        print("\n由于未配置真实 API，现在将演示代码结构，运行会报错。")
    
    # 初始化组件
    llm = LLMClient()
    search_tool = SearchTool()
    
    # 组装 Agent Loop
    agent = AgenticLoop(
        llm_client=llm,
        tools=[search_tool],
        system_prompt=CHIEF_ANALYST_PROMPT
    )
    
    # 模拟用户输入
    competitor_name = input("\n请输入你要分析的竞品名称 (例如 '特斯拉 Model 3' 或 '大疆 Action 4'): ")
    if not competitor_name:
        competitor_name = "乐歌 FlexiSpot E7 升降桌"
        
    print(f"\n🚀 开始深度分析竞品: {competitor_name} ...\n")
    
    try:
        # 运行 Agent
        final_report = await agent.run(f"请对 '{competitor_name}' 进行全面的竞品分析。")
        
        print("\n" + "="*20 + " 最终竞品分析报告 " + "="*20)
        print(final_report)
        print("="*60)
        
    except Exception as e:
        print(f"\n❌ 运行出错: {e}")
        print("这通常是因为没有配置正确的火山引擎 API Key 和 Model Endpoint。")

if __name__ == "__main__":
    asyncio.run(main())
