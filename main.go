package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var todos []Todo
var nextID = 1

func main() {
	// Instead of one generic handler, we define specific ones
	// if you want to keep the same URL for both:
	// Unique URL for GET
	http.HandleFunc("/todos/v1/get", getTodos)

	// Unique URL for POST
	http.HandleFunc("/todos/v1/post", createTodo)

	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

// Logic for GET only
func getTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// Logic for POST only
func createTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	newTodo.ID = nextID
	nextID++
	todos = append(todos, newTodo)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}
