FROM golang:1.26.4-bookworm AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/agent-gateway ./cmd/server

FROM node:22-bookworm-slim

ARG CLAUDE_CODE_PACKAGE=@anthropic-ai/claude-code
ARG CODEX_CLI_PACKAGE=@openai/codex

ENV NODE_ENV=production \
    AGENT_GATEWAY_HOST=0.0.0.0 \
    AGENT_GATEWAY_PORT=8765 \
    AGENT_GATEWAY_WORKDIR=/workspace \
    CLAUDE_BIN=claude \
    CODEX_BIN=codex

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && npm install -g "$CLAUDE_CODE_PACKAGE" "$CODEX_CLI_PACKAGE" \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --shell /bin/bash app \
    && mkdir -p /workspace \
    && chown -R app:app /workspace /home/app

COPY --from=build /out/agent-gateway /usr/local/bin/agent-gateway

USER app
WORKDIR /workspace

EXPOSE 8765

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD node -e "fetch('http://127.0.0.1:8765/health').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

ENTRYPOINT ["agent-gateway"]
