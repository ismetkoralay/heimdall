// Package main is the entry point for the Heimdall service. It initializes and starts the service, handling incoming requests and managing the lifecycle of the application.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ismetkoralay/heimdall/internal/config"
	"github.com/ismetkoralay/heimdall/internal/health"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Handler)

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
