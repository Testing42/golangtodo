package handlers

import (
	"database/sql"
	"encoding/json"
	"html"
	"net/http"
	"strconv"
)

// DecodeAndSanitize: Handles Size Limits and Sanitization
func DecodeAndSanitize(w http.ResponseWriter, r *http.Request) (*Todo, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	t := &Todo{}
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		return nil, err
	}
	t.Title = html.EscapeString(t.Title)
	return t, nil
}

// GetTodos: Logic for GET all, Search by title, and Pagination
func GetTodos(w http.ResponseWriter, r *http.Request) {
	// NEW: Check for a search query parameter (e.g., /todos/v1?search=milk)
	searchTerm := r.URL.Query().Get("search")

	// NEW: Get Pagination parameters from URL (e.g., /todos/v1?page=2&limit=10)
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// NEW: Set Defaults for Pagination
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	// Default to 10 items, max 100 for safety
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// NEW: Calculate Offset (How many items to skip)
	offset := (page - 1) * limit

	var rows *sql.Rows
	var err error

	if searchTerm != "" {
		// UPDATED: Use LIKE for partial matches with LIMIT and OFFSET
		// The % wildcards allow matching the term anywhere in the title.
		query := "SELECT id, title, completed FROM todos WHERE title LIKE ? LIMIT ? OFFSET ?"
		rows, err = DB.Query(query, "%"+searchTerm+"%", limit, offset)
	} else {
		// UPDATED: Get all items with LIMIT and OFFSET
		query := "SELECT id, title, completed FROM todos LIMIT ? OFFSET ?"
		rows, err = DB.Query(query, limit, offset)
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Initialize as empty slice so it returns [] instead of null in JSON
	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			continue
		}
		todos = append(todos, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// CreateTodo: Logic for POST
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	newTodo, err := DecodeAndSanitize(w, r)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", newTodo.Title, newTodo.Completed)
	if err != nil {
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	newTodo.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

// GetTodoByID: Logic for GET one
func GetTodoByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var t Todo
	err = DB.QueryRow("SELECT id, title, completed FROM todos WHERE id = ?", id).Scan(&t.ID, &t.Title, &t.Completed)
	if err == sql.ErrNoRows {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// UpdateTodo: Logic for PUT
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	updatedData, err := DecodeAndSanitize(w, r)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", updatedData.Title, updatedData.Completed, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	updatedData.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedData)
}

// DeleteTodo: Logic for DELETE
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec("DELETE FROM todos WHERE id = ?", id)
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
