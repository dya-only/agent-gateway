import OpenAI from "openai";

const client = new OpenAI({
  baseURL: process.env.AGENT_GATEWAY_BASE_URL ?? "http://127.0.0.1:8765/v1",
  apiKey: process.env.AGENT_GATEWAY_API_KEY ?? "local-secret",
});

const response = await client.chat.completions.create({
  model: process.env.AGENT_GATEWAY_MODEL ?? "claude-sonnet",
  messages: [
    { role: "user", content: "Write one short Korean greeting." },
  ],
});

console.log(response.choices[0].message.content);
