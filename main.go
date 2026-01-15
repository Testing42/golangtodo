package main

import (
	"context"
	"log/slog" // NEW: Structured logging
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Testing42/golangtodo/handlers"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// NEW: Initialize Structured JSON Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Initialize SQLite Database
	if err := handlers.InitDB("./todos.db"); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer handlers.DB.Close()

	// 2. Optional: Migrate existing JSON data to SQL
	handlers.MigrateFromJSON("todos.json")

	// Route Handlers
	http.HandleFunc("/todos/v1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTodos(w, r)
		case http.MethodPost:
			handlers.AuthMiddleware(handlers.CreateTodo)(w, r)
		default:
			// Using http.Error here is okay for "Method Not Allowed"
			// but you could use the JSON helper here too!
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/todos/v1/item", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTodoByID(w, r)
		case http.MethodPut:
			handlers.AuthMiddleware(handlers.UpdateTodo)(w, r)
		case http.MethodDelete:
			handlers.AuthMiddleware(handlers.DeleteTodo)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting server", "port", port, "db", "sqlite")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting. Goodbye!")
}
