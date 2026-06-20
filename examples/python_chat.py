import json
import os
import urllib.request


base_url = os.getenv("AGENT_GATEWAY_BASE_URL", "http://127.0.0.1:8765/v1")
api_key = os.getenv("AGENT_GATEWAY_API_KEY", "local-secret")
model = os.getenv("AGENT_GATEWAY_MODEL", "claude-sonnet")

payload = {
    "model": model,
    "messages": [
        {"role": "user", "content": "Write one short Korean greeting."},
    ],
}

request = urllib.request.Request(
    f"{base_url}/chat/completions",
    data=json.dumps(payload).encode("utf-8"),
    headers={
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    },
    method="POST",
)

with urllib.request.urlopen(request, timeout=300) as response:
    data = json.load(response)

print(data["choices"][0]["message"]["content"])
