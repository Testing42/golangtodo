package main

import (
	"fmt"
	"net/http"
	"os" // 1. Import the os package

	"github.com/Testing42/golangtodo/handlers"
)

func main() {
	// 2. Read the PORT from the environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default fallback if no PORT is set
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

	// 3. Use the dynamic port variable
	fmt.Printf("Server running on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
