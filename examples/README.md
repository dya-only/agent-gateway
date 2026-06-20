# agent-gateway Examples

These examples call the local OpenAI-compatible gateway.

Default connection values:

```text
Base URL: http://127.0.0.1:8765/v1
API key: local-secret
Model: claude-sonnet
```

Start the gateway first:

```bash
cp .env.example .env
go run ./cmd/server
```

## 1. curl

```bash
./examples/curl-chat.sh
```

Streaming:

```bash
./examples/curl-stream.sh
```

Or run directly:

```bash
curl http://127.0.0.1:8765/v1/chat/completions \
  -H "Authorization: Bearer local-secret" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet",
    "messages": [
      {"role": "user", "content": "Write one short Korean greeting."}
    ]
  }'
```

## 2. Python

Uses only the Python standard library.

```bash
python3 examples/python_chat.py
```

Streaming:

```bash
python3 examples/python_stream.py
```

## 3. Node.js

Uses built-in `fetch`, available in Node 18+.

```bash
node examples/node-chat.mjs
```

## 4. OpenAI Python SDK

Install the SDK:

```bash
pip install openai
```

Run:

```bash
python3 examples/openai_sdk_python.py
```

## 5. OpenAI Node SDK

Install the SDK:

```bash
npm install openai
```

Run:

```bash
node examples/openai-sdk-node.mjs
```

## Environment Overrides

All examples support these environment variables:

```bash
export AGENT_GATEWAY_BASE_URL=http://127.0.0.1:8765/v1
export AGENT_GATEWAY_API_KEY=local-secret
export AGENT_GATEWAY_MODEL=claude-sonnet
```

Use Codex instead:

```bash
AGENT_GATEWAY_MODEL=codex python3 examples/python_chat.py
```
