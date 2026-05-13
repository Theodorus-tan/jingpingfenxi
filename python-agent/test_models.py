import asyncio, os
from dotenv import load_dotenv
from openai import AsyncOpenAI

load_dotenv()

client = AsyncOpenAI(
    api_key=os.getenv("VOLCENGINE_API_KEY"),
    base_url=os.getenv("VOLCENGINE_BASE_URL"),
)

models = [
    "doubao-seed-2-0-lite-260428",
    "doubao-seed-2-0-260428",
    "doubao-1-6-lite",
    "doubao-1-6-pro",
]

async def test(name):
    try:
        r = await client.chat.completions.create(
            model=name,
            messages=[{"role": "user", "content": "hi"}],
            max_tokens=20,
        )
        return f"OK  {name}: {r.choices[0].message.content}"
    except Exception as e:
        code = getattr(e, "status_code", "")
        msg = str(e)[:150]
        return f"FAIL {name}: {code} {msg}"

async def main():
    for m in models:
        print(await test(m))

asyncio.run(main())
