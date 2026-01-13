package main

import (
	"fmt"
	"log"
	"net/http"
	"os" // 1. Import the os package

	"github.com/Testing42/golangtodo/handlers"
	"github.com/joho/godotenv"
)

func main() {
	// 2. Load the .env file
	// If no filename is provided, it looks for ".env" in the current directory
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system default environment variables")
	}

	// GET (all) or POST (create)
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

	// GET (one), PUT (update), or DELETE (remove)
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

	// NEW: Load existing todos from the JSON file
	if err := handlers.LoadFromJSON(); err != nil {
		fmt.Printf("Warning: Could not load data: %v\n", err)
	}

	// 3. Now os.Getenv will successfully find your API_KEY and PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Load existing data from the file determined by GetFileName()
	fmt.Printf("Starting server on port %s using data file: %s\n", port, handlers.GetFileName())

	if err := handlers.LoadFromJSON(); err != nil {
		log.Printf("Note: Starting with fresh data (%v)", err)
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
