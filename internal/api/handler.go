package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"agent-gateway/internal/config"
	"agent-gateway/internal/openai"
	"agent-gateway/internal/provider"
)

type Handler struct {
	cfg    config.Config
	router *provider.Router
	sem    chan struct{}
}

func NewHandler(cfg config.Config, router *provider.Router) *Handler {
	return &Handler{
		cfg:    cfg,
		router: router,
		sem:    make(chan struct{}, cfg.MaxWorkers),
	}
}

func (h *Handler) acquire(ctx context.Context) bool {
	select {
	case h.sem <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (h *Handler) release() {
	<-h.sem
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.health)
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /v1/models", h.withAuth(h.models))
	mux.HandleFunc("POST /v1/chat/completions", h.withAuth(h.chatCompletions))
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) models(w http.ResponseWriter, r *http.Request) {
	items := make([]openai.ModelInfo, 0)
	for _, model := range h.router.Models() {
		items = append(items, openai.ModelInfo{ID: model.Name, Object: "model", OwnedBy: model.Provider})
	}
	writeJSON(w, http.StatusOK, openai.ModelList{Object: "list", Data: items})
}

func (h *Handler) chatCompletions(w http.ResponseWriter, r *http.Request) {
	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON body")
		return
	}
	if req.Stream {
		h.streamChatCompletions(w, r, req)
		return
	}
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.Timeout)
	defer cancel()

	if !h.acquire(ctx) {
		writeError(w, http.StatusServiceUnavailable, "capacity_error", "server at capacity")
		return
	}
	defer h.release()

	content, err := h.router.Complete(ctx, req.Model, req.Messages)
	if err != nil {
		if errors.Is(err, provider.ErrUnknownModel) {
			writeError(w, http.StatusBadRequest, "invalid_request_error", "unknown model")
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, "timeout_error", "CLI request timed out")
			return
		}
		writeError(w, http.StatusBadGateway, "provider_error", err.Error())
		return
	}

	promptTokens := estimateTokensFromMessages(req.Messages)
	completionTokens := estimateTokens(content)
	resp := openai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []openai.ChatChoice{{
			Index:        0,
			Message:      openai.ChatMessage{Role: "assistant", Content: content},
			FinishReason: "stop",
		}},
		Usage: openai.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) streamChatCompletions(w http.ResponseWriter, r *http.Request, req openai.ChatCompletionRequest) {
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages is required")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "server_error", "streaming is not supported by this server")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.Timeout)
	defer cancel()

	if !h.acquire(ctx) {
		writeError(w, http.StatusServiceUnavailable, "capacity_error", "server at capacity")
		return
	}
	defer h.release()

	chunks, errs, err := h.router.StreamComplete(ctx, req.Model, req.Messages)
	if err != nil {
		if errors.Is(err, provider.ErrUnknownModel) {
			writeError(w, http.StatusBadRequest, "invalid_request_error", "unknown model")
			return
		}
		writeError(w, http.StatusBadGateway, "provider_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	id := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	for chunk := range chunks {
		writeSSE(w, openai.ChatCompletionChunk{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   req.Model,
			Choices: []openai.ChunkChoice{{
				Index: 0,
				Delta: openai.ChunkDelta{Role: "assistant", Content: chunk},
			}},
		})
		flusher.Flush()
	}

	if err := <-errs; err != nil {
		writeSSE(w, openai.ErrorResponse{Error: openai.ErrorBody{Message: err.Error(), Type: "provider_error"}})
		flusher.Flush()
		return
	}

	finishReason := "stop"
	writeSSE(w, openai.ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []openai.ChunkChoice{{
			Index:        0,
			Delta:        openai.ChunkDelta{},
			FinishReason: &finishReason,
		}},
	})
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.cfg.APIKey == "" {
			next(w, r)
			return
		}

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(token), []byte(h.cfg.APIKey)) != 1 {
			writeError(w, http.StatusUnauthorized, "authentication_error", "invalid API key")
			return
		}

		next(w, r)
	}
}

func estimateTokensFromMessages(messages []openai.ChatMessage) int {
	total := 0
	for _, message := range messages {
		total += estimateTokens(message.Content)
	}
	return total
}

func estimateTokens(value string) int {
	if value == "" {
		return 0
	}
	return len([]rune(value))/4 + 1
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, typ string, message string) {
	writeJSON(w, status, openai.ErrorResponse{Error: openai.ErrorBody{Message: message, Type: typ}})
}

func writeSSE(w http.ResponseWriter, value any) {
	fmt.Fprint(w, "data: ")
	_ = json.NewEncoder(w).Encode(value)
	fmt.Fprint(w, "\n")
}
