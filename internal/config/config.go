package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host       string
	Port       string
	APIKey     string
	Timeout    time.Duration
	MaxWorkers int
	Models     []Model
	ClaudeBin  string
	ClaudeArgs []string
	CodexBin   string
	CodexArgs  []string
	Workdir    string
}

type Model struct {
	Name     string `json:"id"`
	Provider string `json:"-"`
	Model    string `json:"-"`
}

func Load() Config {
	cfg := Config{
		Host:       env("AGENT_GATEWAY_HOST", "127.0.0.1"),
		Port:       env("AGENT_GATEWAY_PORT", "8765"),
		APIKey:     os.Getenv("AGENT_GATEWAY_API_KEY"),
		Timeout:    time.Duration(envInt("AGENT_GATEWAY_TIMEOUT_SECONDS", 300)) * time.Second,
		MaxWorkers: envInt("AGENT_GATEWAY_MAX_WORKERS", 8),
		ClaudeBin:  env("CLAUDE_BIN", "claude"),
		ClaudeArgs: splitArgs(env("CLAUDE_ARGS", "--print")),
		CodexBin:   env("CODEX_BIN", "codex"),
		CodexArgs:  splitArgs(env("CODEX_ARGS", "exec --skip-git-repo-check")),
		Workdir:    os.Getenv("AGENT_GATEWAY_WORKDIR"),
	}

	cfg.Models = []Model{
		{Name: "claude-sonnet", Provider: "claude", Model: env("CLAUDE_MODEL", "sonnet")},
	}
	for _, m := range splitModels(env("CODEX_MODELS", env("CODEX_MODEL", "gpt-5.5"))) {
		cfg.Models = append(cfg.Models, Model{Name: m, Provider: "codex", Model: m})
	}

	return cfg
}

func (c Config) Addr() string {
	return c.Host + ":" + c.Port
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitArgs(value string) []string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func splitModels(value string) []string {
	var out []string
	for _, m := range strings.Split(value, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}
