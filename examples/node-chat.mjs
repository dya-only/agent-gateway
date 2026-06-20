const baseUrl = process.env.AGENT_GATEWAY_BASE_URL ?? "http://127.0.0.1:8765/v1";
const apiKey = process.env.AGENT_GATEWAY_API_KEY ?? "local-secret";
const model = process.env.AGENT_GATEWAY_MODEL ?? "claude-sonnet";

const response = await fetch(`${baseUrl}/chat/completions`, {
  method: "POST",
  headers: {
    Authorization: `Bearer ${apiKey}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    model,
    messages: [
      { role: "user", content: "Write one short Korean greeting." },
    ],
  }),
});

if (!response.ok) {
  throw new Error(`Gateway request failed: ${response.status} ${await response.text()}`);
}

const data = await response.json();
console.log(data.choices[0].message.content);
