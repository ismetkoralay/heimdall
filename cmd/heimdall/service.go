// Package main is the entry point for the Heimdall service. It initializes and starts the service, handling incoming requests and managing the lifecycle of the application.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/ismetkoralay/heimdall/internal/api"
	"github.com/ismetkoralay/heimdall/internal/config"
	"github.com/ismetkoralay/heimdall/internal/health"
	"github.com/ismetkoralay/heimdall/internal/provider"
	"github.com/ismetkoralay/heimdall/internal/provider/ollama"
)

var (
	ErrProviderBaseURLNotSet = errors.New("error provider base url is not set")
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	ollamaProvider := ollama.NewOllamaProvider(http.DefaultClient)
	providers, err := providerMap(cfg, []provider.Provider{ollamaProvider})
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	router := provider.NewRouter(cfg.ModelProviderMap, providers)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Handler)
	mux.Handle("POST /v1/chat/completions", api.ChatHandler(router, time.Now, uuid.New))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal (e.g., SIGINT, SIGTERM) to gracefully shut down the server
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("Shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	} else {
		logger.Info("Server gracefully stopped")
	}
}

func providerMap(cfg config.Config, providers []provider.Provider) (map[string]provider.Provider, error) {
	res := map[string]provider.Provider{}
	for _, p := range providers {
		providerName := p.Name()
		baseURL, ok := cfg.ProviderMap[providerName]
		if !ok {
			return nil, fmt.Errorf("%w: provider name: %s", ErrProviderBaseURLNotSet, providerName)
		}
		p.SetBaseURL(baseURL)
		res[providerName] = p
	}

	return res, nil
}
