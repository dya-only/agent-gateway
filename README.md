# agent-gateway

Local OpenAI-compatible gateway for authenticated Codex CLI and Claude Code sessions.

This is an MVP focused on personal local use. It exposes a small OpenAI-compatible HTTP API and forwards requests to local CLI tools.

## Features

- `GET /health`
- `GET /v1/models`
- `POST /v1/chat/completions`
- Streaming chat completions via Server-Sent Events
- Optional Bearer API key auth
- Claude Code adapter via `claude --print`
- Codex CLI adapter via `codex exec --skip-git-repo-check`
- Local-only default bind address: `127.0.0.1:8765`
- Request timeout and optional fixed working directory

Tool calling, `/v1/responses`, and session persistence are intentionally not included in the first MVP.

## Requirements

- Go 1.26+
- Authenticated `claude` CLI if using the Claude model
- Authenticated `codex` CLI if using the Codex model

## Run

```bash
cp .env.example .env
go run ./cmd/server
```

The server listens on:

```text
http://127.0.0.1:8765
```

## Docker Compose

Build and start the gateway:

```bash
cp .env.example .env
mkdir -p workspace
docker compose up --build
```

The compose service publishes the gateway on the host at:

```text
http://127.0.0.1:8765
```

Other services in the same Compose project can call it at:

```text
http://agent-gateway:8765/v1
```

Example environment for another service:

```yaml
environment:
  OPENAI_BASE_URL: http://agent-gateway:8765/v1
  OPENAI_API_KEY: ${AGENT_GATEWAY_API_KEY}
```

The default `compose.yaml` mounts:

- `./workspace` to `/workspace`
- `${HOME}/.claude` to `/home/app/.claude`
- `${HOME}/.codex` to `/home/app/.codex`

Those mounts allow the containerized CLI tools to reuse local login/session state. If your Claude or Codex CLI stores auth elsewhere, update the volume paths accordingly.

The Docker image installs these npm packages by default:

- `@anthropic-ai/claude-code`
- `@openai/codex`

You can override package names at build time if needed:

```bash
docker compose build \
  --build-arg CLAUDE_CODE_PACKAGE=@anthropic-ai/claude-code \
  --build-arg CODEX_CLI_PACKAGE=@openai/codex
```

## Configuration

Configuration is read from environment variables.

| Variable | Default | Description |
| --- | --- | --- |
| `AGENT_GATEWAY_HOST` | `127.0.0.1` | Server bind host |
| `AGENT_GATEWAY_PORT` | `8765` | Server port |
| `AGENT_GATEWAY_API_KEY` | empty | Optional Bearer auth key |
| `AGENT_GATEWAY_TIMEOUT_SECONDS` | `300` | CLI request timeout |
| `AGENT_GATEWAY_WORKDIR` | empty | Optional fixed subprocess working directory |
| `CLAUDE_BIN` | `claude` | Claude CLI binary |
| `CLAUDE_ARGS` | `--print` | Base Claude CLI args |
| `CLAUDE_MODEL` | `sonnet` | Model passed to Claude adapter |
| `CODEX_BIN` | `codex` | Codex CLI binary |
| `CODEX_ARGS` | `exec --skip-git-repo-check` | Base Codex CLI args |
| `CODEX_MODEL` | `gpt-5.5` | Model passed to Codex adapter |

## Models

The MVP exposes two model IDs:

- `claude-sonnet`
- `codex`

Check available models:

```bash
curl http://127.0.0.1:8765/v1/models \
  -H "Authorization: Bearer local-secret"
```

## Chat Completions Example

```bash
curl http://127.0.0.1:8765/v1/chat/completions \
  -H "Authorization: Bearer local-secret" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet",
    "messages": [
      {"role": "user", "content": "Explain this repository structure briefly."}
    ]
  }'
```

Codex example:

```bash
curl http://127.0.0.1:8765/v1/chat/completions \
  -H "Authorization: Bearer local-secret" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "codex",
    "messages": [
      {"role": "user", "content": "Write a small Go hello world program."}
    ]
  }'
```

## Streaming Chat Completions Example

```bash
curl -N http://127.0.0.1:8765/v1/chat/completions \
  -H "Authorization: Bearer local-secret" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet",
    "stream": true,
    "messages": [
      {"role": "user", "content": "Write a short greeting."}
    ]
  }'
```

The gateway returns OpenAI-compatible Server-Sent Events with `chat.completion.chunk` payloads and a final `data: [DONE]` event.

## Notes

- This gateway shells out to local CLI tools and returns stdout as the assistant message.
- Streaming forwards CLI stdout as it is emitted. Some CLI tools may buffer output and only emit near the end.
- `usage` values are approximate token estimates.
- Keep `AGENT_GATEWAY_HOST=127.0.0.1` unless you deliberately want network access.
- The exact Codex CLI flags can vary by version. Adjust `CODEX_ARGS` if your installed CLI expects a different non-interactive command.
