package handlers

import (
	"database/sql"
	"encoding/json"
	"html"
	"log/slog" // NEW: Structured logging
	"net/http"
	"strconv"
	"time"
)

// sendJSONError: Helper to return consistent JSON error objects
func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  code,
	})
}

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

// GetTodos: Logic for GET all, Search by title, and Pagination with Metadata
func GetTodos(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM todos WHERE title LIKE ?"
	err := DB.QueryRow(countQuery, "%"+searchTerm+"%").Scan(&totalCount)
	if err != nil {
		slog.Error("Count query failed", "error", err)
		sendJSONError(w, "Database error counting items", http.StatusInternalServerError)
		return
	}

	var rows *sql.Rows
	if searchTerm != "" {
		query := "SELECT id, title, completed, created_at FROM todos WHERE title LIKE ? LIMIT ? OFFSET ?"
		rows, err = DB.Query(query, "%"+searchTerm+"%", limit, offset)
	} else {
		query := "SELECT id, title, completed, created_at FROM todos LIMIT ? OFFSET ?"
		rows, err = DB.Query(query, limit, offset)
	}

	if err != nil {
		slog.Error("Fetch query failed", "error", err)
		sendJSONError(w, "Database error fetching data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			slog.Warn("Failed to scan row", "error", err)
			continue
		}
		todos = append(todos, t)
	}

	response := TodoResponse{
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		Data:       todos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateTodo: Logic for POST
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	newTodo, err := DecodeAndSanitize(w, r)
	if err != nil {
		sendJSONError(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", newTodo.Title, newTodo.Completed)
	if err != nil {
		slog.Error("Insert failed", "error", err)
		sendJSONError(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()

	err = DB.QueryRow("SELECT id, title, completed, created_at FROM todos WHERE id = ?", id).
		Scan(&newTodo.ID, &newTodo.Title, &newTodo.Completed, &newTodo.CreatedAt)

	if err != nil {
		slog.Error("Retrieve after insert failed", "id", id, "error", err)
		sendJSONError(w, "Internal error retrieving record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

// GetTodoByID: Logic for GET one
func GetTodoByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendJSONError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var t Todo
	query := "SELECT id, title, completed, created_at FROM todos WHERE id = ?"
	err = DB.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt)

	if err == sql.ErrNoRows {
		sendJSONError(w, "Todo not found", http.StatusNotFound)
		return
	} else if err != nil {
		slog.Error("GetByID scan failed", "id", id, "error", err)
		sendJSONError(w, "Internal database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// UpdateTodo: Logic for PUT
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendJSONError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	updatedData, err := DecodeAndSanitize(w, r)
	if err != nil {
		sendJSONError(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", updatedData.Title, updatedData.Completed, id)
	if err != nil {
		slog.Error("Update failed", "id", id, "error", err)
		sendJSONError(w, "Database update error", http.StatusInternalServerError)
		return
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		sendJSONError(w, "Todo not found", http.StatusNotFound)
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
		sendJSONError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		slog.Error("Delete failed", "id", id, "error", err)
		sendJSONError(w, "Database delete error", http.StatusInternalServerError)
		return
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		sendJSONError(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck: A professional endpoint to verify the API and DB are alive
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if the database is actually reachable
	err := DB.Ping()
	if err != nil {
		slog.Error("Health check failed: DB unreachable", "error", err)

		// Use our professional error helper for the response
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "down",
			"database": "disconnected",
		})
		return
	}

	// If everything is fine
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "up",
		"database": "connected",
		"time":     time.Now().Format(time.RFC3339),
	})
}
