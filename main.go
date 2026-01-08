package main

import (
	"fmt"
	"net/http"

	"github.com/Testing42/golangtodo/handlers"
)

func main() {
	// Instead of one generic handler, we define specific ones
	// if you want to keep the same URL for both:

	// Unique URL for POST
	http.HandleFunc("/todos/v1/get", handlers.GetTodos)
	// Unique URL for POST
	http.HandleFunc("/todos/v1/post", handlers.AuthMiddleware(handlers.CreateTodo))
	// New route for specific search
	http.HandleFunc("/todos/v1/get/item", handlers.GetTodoByID)
	http.HandleFunc("/todos/v1/update", handlers.AuthMiddleware(handlers.UpdateTodo))
	http.HandleFunc("/todos/v1/delete", handlers.AuthMiddleware(handlers.DeleteTodo))

	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
