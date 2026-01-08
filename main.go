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
	// Use one handler to manage different methods for the same path
	http.HandleFunc("/todos", todoHandler)

	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTodos(w, r)
	case http.MethodPost:
		createTodo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo

	// 1. Decode the JSON body from the request into newTodo
	if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 2. Assign an ID and save it
	newTodo.ID = nextID
	nextID++
	todos = append(todos, newTodo)

	// 3. Send back the created object
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}
