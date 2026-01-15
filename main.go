package main

import (
	"context"
	"fmt"
	"log"
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

	// 1. Initialize SQLite Database
	if err := handlers.InitDB("./todos.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
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
		fmt.Printf("Starting server on port %s using SQLite database.\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting. Goodbye!")
}
