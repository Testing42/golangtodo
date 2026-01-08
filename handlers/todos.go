package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
)

// DecodeAndSanitize: Handles Size Limits and Sanitization for both POST and PUT
func DecodeAndSanitize(w http.ResponseWriter, r *http.Request) (Todo, error) {
	// Limit body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var t Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return t, err
	}

	// Sanitize string to prevent XSS (converts <script> to &lt;script&gt;)
	t.Title = html.EscapeString(t.Title)
	return t, nil
}

// GetTodos: Logic for GET only
func GetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Todos)
}

// CreateTodo: Logic for POST only
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	newTodo, err := DecodeAndSanitize(w, r)
	if err != nil {
		http.Error(w, "Invalid input or payload too large", http.StatusBadRequest)
		return
	}

	newTodo.ID = NextID
	NextID++
	Todos = append(Todos, newTodo)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

// GetTodoByID: Logic for specific search
func GetTodoByID(w http.ResponseWriter, r *http.Request) {
	// 1. Get the "id" from the URL (?id=1)
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	// 2. Look for the item in our list
	for _, item := range Todos {
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

// UpdateTodo: Change an existing Todo
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	updatedData, err := DecodeAndSanitize(w, r)
	if err != nil {
		http.Error(w, "Invalid input or payload too large", http.StatusBadRequest)
		return
	}

	for i, item := range Todos {
		if fmt.Sprintf("%d", item.ID) == idStr {
			Todos[i].Title = updatedData.Title
			Todos[i].Completed = updatedData.Completed

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Todos[i])
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}

// DeleteTodo: Remove a Todo
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	for i, item := range Todos {
		if fmt.Sprintf("%d", item.ID) == idStr {
			// Remove from slice
			Todos = append(Todos[:i], Todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent) // 204 means Success, no content to show
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}
