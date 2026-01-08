package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Testing42/golangtodo/handlers"
)

// TestMain allows us to set up environment variables before ANY tests run
func TestMain(m *testing.M) {
	// Set the API key in the environment so the middleware can find it
	os.Setenv("API_KEY", "test-secret-key")
	os.Exit(m.Run())
}

// Helper function updated to use the test environment key
func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "test-secret-key")
}

func TestGetTodos(t *testing.T) {
	req, _ := http.NewRequest("GET", "/todos/v1", nil) // Updated Path
	rr := httptest.NewRecorder()

	handlers.GetTodos(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestCreateTodo(t *testing.T) {
	payload := []byte(`{"title":"Learn Go Testing","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer(payload)) // Updated Path

	rr := httptest.NewRecorder()
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
	// FIX: Use a pointer to Todo (&handlers.Todo)
	handlers.Todos = []*handlers.Todo{{ID: 99, Title: "Find Me", Completed: false}}

	req, _ := http.NewRequest("GET", "/todos/v1/item?id=99", nil) // Updated Path
	rr := httptest.NewRecorder()

	handlers.GetTodoByID(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200, got %v", status)
	}
}

func TestUpdateTodo(t *testing.T) {
	// FIX: Use a pointer to Todo
	handlers.Todos = []*handlers.Todo{{ID: 1, Title: "Old Title", Completed: false}}
	payload := []byte(`{"title":"New Title","completed":true}`)

	req, _ := http.NewRequest("PUT", "/todos/v1/item?id=1", bytes.NewBuffer(payload)) // Updated Path
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	handler := handlers.AuthMiddleware(handlers.UpdateTodo)
	handler.ServeHTTP(rr, req)

	if handlers.Todos[0].Title != "New Title" {
		t.Errorf("Expected New Title, got %v", handlers.Todos[0].Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	// FIX: Use a pointer to Todo
	handlers.Todos = []*handlers.Todo{{ID: 1, Title: "Delete Me", Completed: false}}

	req, _ := http.NewRequest("DELETE", "/todos/v1/item?id=1", nil) // Updated Path
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	handler := handlers.AuthMiddleware(handlers.DeleteTodo)
	handler.ServeHTTP(rr, req)

	if len(handlers.Todos) != 0 {
		t.Errorf("Expected 0 todos, got %v", len(handlers.Todos))
	}
}
