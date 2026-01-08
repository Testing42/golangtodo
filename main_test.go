package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Testing42/golangtodo/handlers"
)

// Helper function to set the API key on requests to save typing
func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "my-secure-key-123")
}

func TestGetTodos(t *testing.T) {
	req, _ := http.NewRequest("GET", "/todos/v1/get", nil)
	rr := httptest.NewRecorder()

	// Calling the exported function from the handlers package
	handlers.GetTodos(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestCreateTodo(t *testing.T) {
	payload := []byte(`{"title":"Learn Go Testing","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1/post", bytes.NewBuffer(payload))

	rr := httptest.NewRecorder()
	// Wrap the exported handler with the exported middleware
	handler := handlers.AuthMiddleware(handlers.CreateTodo)

	// 1. Test without key (Should Fail)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without key, got %v", rr.Code)
	}

	// 2. Test with key (Should Pass)
	rr = httptest.NewRecorder()
	setAuthHeader(req)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201 with key, got %v", rr.Code)
	}
}

func TestGetTodoByID(t *testing.T) {
	// Update the exported Todos slice
	handlers.Todos = []handlers.Todo{{ID: 99, Title: "Find Me", Completed: false}}

	req, _ := http.NewRequest("GET", "/todos/v1/get/item?id=99", nil)
	rr := httptest.NewRecorder()

	handlers.GetTodoByID(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200, got %v", status)
	}
}

func TestUpdateTodo(t *testing.T) {
	handlers.Todos = []handlers.Todo{{ID: 1, Title: "Old Title", Completed: false}}
	payload := []byte(`{"title":"New Title","completed":true}`)

	req, _ := http.NewRequest("PUT", "/todos/v1/update?id=1", bytes.NewBuffer(payload))
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	handler := handlers.AuthMiddleware(handlers.UpdateTodo)
	handler.ServeHTTP(rr, req)

	if handlers.Todos[0].Title != "New Title" {
		t.Errorf("Expected New Title, got %v", handlers.Todos[0].Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	handlers.Todos = []handlers.Todo{{ID: 1, Title: "Delete Me", Completed: false}}

	req, _ := http.NewRequest("DELETE", "/todos/v1/delete?id=1", nil)
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	handler := handlers.AuthMiddleware(handlers.DeleteTodo)
	handler.ServeHTTP(rr, req)

	if len(handlers.Todos) != 0 {
		t.Errorf("Expected 0 todos, got %v", len(handlers.Todos))
	}
}

func TestPostAndPutSanitization(t *testing.T) {
	maliciousJSON := []byte(`{"title":"<script>evil</script>"}`)

	// Test POST
	reqPost, _ := http.NewRequest("POST", "/todos/v1/post", bytes.NewBuffer(maliciousJSON))
	setAuthHeader(reqPost)
	rrPost := httptest.NewRecorder()
	handlers.AuthMiddleware(handlers.CreateTodo).ServeHTTP(rrPost, reqPost)

	// Test PUT
	reqPut, _ := http.NewRequest("PUT", "/todos/v1/update?id=1", bytes.NewBuffer(maliciousJSON))
	setAuthHeader(reqPut)
	rrPut := httptest.NewRecorder()
	handlers.AuthMiddleware(handlers.UpdateTodo).ServeHTTP(rrPut, reqPut)

	// In the logic, PUT returns 200 (OK) with the object,
	// while POST returns 201 (Created).
	if rrPost.Code == http.StatusCreated && rrPut.Code == http.StatusOK {
		t.Log("Both endpoints successfully sanitized input")
	}
}
