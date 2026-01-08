package handlers

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"
)

// DecodeAndSanitize: Handles Size Limits and Sanitization
func DecodeAndSanitize(w http.ResponseWriter, r *http.Request) (*Todo, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	t := &Todo{}

	// FIX: Pass 't' directly, not '&t', because 't' is already a pointer
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		return nil, err
	}

	t.Title = html.EscapeString(t.Title)
	return t, nil
}

// GetTodos: Logic for GET only
func GetTodos(w http.ResponseWriter, r *http.Request) {
	// UPDATED: Added RLock because reading a slice while
	// another thread might be appending to it can cause a crash.
	Mu.RLock()
	defer Mu.RUnlock()

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

	// UPDATED: Added Lock to protect NextID and the Todos slice
	Mu.Lock()
	newTodo.ID = NextID
	NextID++
	Todos = append(Todos, newTodo)
	Mu.Unlock() // Manual unlock here so we don't hold it during the JSON encoding

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

// GetTodoByID: Logic for specific search
func GetTodoByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	Mu.RLock()
	defer Mu.RUnlock()

	for _, item := range Todos {
		if item.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}

// UpdateTodo: Change an existing Todo
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	updatedData, err := DecodeAndSanitize(w, r)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	Mu.Lock()
	defer Mu.Unlock()

	for i, item := range Todos {
		if item.ID == id {
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
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	Mu.Lock()
	defer Mu.Unlock()

	for i, item := range Todos {
		if item.ID == id {
			Todos = append(Todos[:i], Todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Todo not found", http.StatusNotFound)
}
