import os

from openai import OpenAI


client = OpenAI(
    base_url=os.getenv("AGENT_GATEWAY_BASE_URL", "http://127.0.0.1:8765/v1"),
    api_key=os.getenv("AGENT_GATEWAY_API_KEY", "local-secret"),
)

response = client.chat.completions.create(
    model=os.getenv("AGENT_GATEWAY_MODEL", "claude-sonnet"),
    messages=[
        {"role": "user", "content": "Write one short Korean greeting."},
    ],
)

print(response.choices[0].message.content)
