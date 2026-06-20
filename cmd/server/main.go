package main

import (
	"log"
	"net/http"
	"time"

	"agent-gateway/internal/api"
	"agent-gateway/internal/config"
	"agent-gateway/internal/provider"
)

func main() {
	cfg := config.Load()

	router := provider.NewRouter(cfg)
	handler := api.NewHandler(cfg, router)

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("agent-gateway listening on http://%s", cfg.Addr())
	log.Fatal(server.ListenAndServe())
}
