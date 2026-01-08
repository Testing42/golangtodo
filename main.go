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

	// New route for specific search
	http.HandleFunc("/todos/v1/get/item", getTodoByID)

	http.HandleFunc("/todos/v1/update", updateTodo)
	http.HandleFunc("/todos/v1/delete", deleteTodo)

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

func getTodoByID(w http.ResponseWriter, r *http.Request) {
	// 1. Get the "id" from the URL (?id=1)
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	// 2. Look for the item in our list
	for _, item := range todos {
		// We convert the ID to a string to compare easily
		if fmt.Sprintf("%d", item.ID) == idStr {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)
			return
		}
	}

	// 3. If not found
	http.Error(w, "Todo not found", http.StatusNotFound)
}

// UPDATE: Change an existing Todo
func updateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	var updatedData Todo
	json.NewDecoder(r.Body).Decode(&updatedData)

	for i, item := range todos {
		if fmt.Sprintf("%d", item.ID) == idStr {
			todos[i].Title = updatedData.Title
			todos[i].Completed = updatedData.Completed
			json.NewEncoder(w).Encode(todos[i])
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}

// DELETE: Remove a Todo
func deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	for i, item := range todos {
		if fmt.Sprintf("%d", item.ID) == idStr {
			// Remove from slice
			todos = append(todos[:i], todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent) // 204 means Success, no content to show
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}
