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

// getEnv handles environment variable retrieval with a default fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// runServer manages the startup and graceful shutdown of the HTTP server
func runServer(srv *http.Server, dbFile, port string) {
	// Setup channel to listen for interrupt signals (Ctrl+C, etc.)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run server in a goroutine so it doesn't block the main thread
	go func() {
		slog.Info("Starting server", "port", port, "db", dbFile)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait here until we receive a signal
	<-stop
	slog.Info("Shutting down gracefully...")

	// Create a 5-second window to finish active requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting. Goodbye!")
}
