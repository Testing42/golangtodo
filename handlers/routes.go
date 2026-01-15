package handlers

import "net/http"

// RegisterRoutes defines all the API endpoints and connects them to handlers
func RegisterRoutes() {
	http.HandleFunc("/todos/v1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetTodos(w, r)
		case http.MethodPost:
			AuthMiddleware(CreateTodo)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/todos/v1/item", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetTodoByID(w, r)
		case http.MethodPut:
			AuthMiddleware(UpdateTodo)(w, r)
		case http.MethodDelete:
			AuthMiddleware(DeleteTodo)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Health Check endpoint moved here as well
	http.HandleFunc("/health", HealthCheck)
}
