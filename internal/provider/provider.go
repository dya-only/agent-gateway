package provider

import (
	"context"
	"errors"

	"agent-gateway/internal/config"
	"agent-gateway/internal/openai"
)

var ErrUnknownModel = errors.New("unknown model")

type Provider interface {
	Complete(ctx context.Context, model config.Model, messages []openai.ChatMessage) (string, error)
	StreamComplete(ctx context.Context, model config.Model, messages []openai.ChatMessage) (<-chan string, <-chan error, error)
}

type Router struct {
	models    map[string]config.Model
	providers map[string]Provider
}

func NewRouter(cfg config.Config) *Router {
	models := make(map[string]config.Model, len(cfg.Models))
	for _, model := range cfg.Models {
		models[model.Name] = model
	}

	return &Router{
		models: models,
		providers: map[string]Provider{
			"claude": NewCLIProvider(cfg.ClaudeBin, cfg.ClaudeArgs, cfg.Workdir, cliStyleClaude),
			"codex":  NewCLIProvider(cfg.CodexBin, cfg.CodexArgs, cfg.Workdir, cliStyleCodex),
		},
	}
}

func (r *Router) Models() []config.Model {
	models := make([]config.Model, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, model)
	}
	return models
}

func (r *Router) Complete(ctx context.Context, modelName string, messages []openai.ChatMessage) (string, error) {
	model, ok := r.models[modelName]
	if !ok {
		return "", ErrUnknownModel
	}

	client, ok := r.providers[model.Provider]
	if !ok {
		return "", ErrUnknownModel
	}

	return client.Complete(ctx, model, messages)
}

func (r *Router) StreamComplete(ctx context.Context, modelName string, messages []openai.ChatMessage) (<-chan string, <-chan error, error) {
	model, ok := r.models[modelName]
	if !ok {
		return nil, nil, ErrUnknownModel
	}

	client, ok := r.providers[model.Provider]
	if !ok {
		return nil, nil, ErrUnknownModel
	}

	return client.StreamComplete(ctx, model, messages)
}
