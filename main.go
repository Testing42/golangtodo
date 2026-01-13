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
	// 1. Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system default environment variables")
	}

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

	// 2. Load existing todos from JSON
	if err := handlers.LoadFromJSON(); err != nil {
		log.Printf("Note: Starting with fresh data (%v)", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 3. Configure the Server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // Uses DefaultServeMux
	}

	// 4. Create a channel to listen for termination signals (Ctrl+C or Kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 5. Run the server in a goroutine so it doesn't block
	go func() {
		fmt.Printf("Starting server on port %s using data file: %s\n", port, handlers.GetFileName())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// 6. Wait for the signal
	<-stop
	fmt.Println("\nShutting down gracefully...")

	// 7. Final Persistence Save
	// This ensures that any last-minute memory changes are flushed to the JSON file
	if err := handlers.SaveToJSON(); err != nil {
		fmt.Printf("Final save failed: %v\n", err)
	}

	// 8. Shutdown Context
	// We give the server 5 seconds to finish processing existing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting. Goodbye!")
}
