#!/usr/bin/env sh
set -eu

BASE_URL="${AGENT_GATEWAY_BASE_URL:-http://127.0.0.1:8765/v1}"
API_KEY="${AGENT_GATEWAY_API_KEY:-local-secret}"
MODEL="${AGENT_GATEWAY_MODEL:-claude-sonnet}"

curl -sS "$BASE_URL/chat/completions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"$MODEL\",\"messages\":[{\"role\":\"user\",\"content\":\"Write one short Korean greeting.\"}]}"
