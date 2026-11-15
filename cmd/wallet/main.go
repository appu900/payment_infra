package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("health check", "remote", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errChan := make(chan error, 1)

	go func() {
		logger.Info("wallet service listening", "addr", ":8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received, stopping server")

	case er := <-errChan:
		logger.Error("server error", "stopping", "err", er)
		os.Exit(1)

	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "err", err)
		os.Exit(1)
	}

}
