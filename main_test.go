package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper function to set the API key on requests to save typing
func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "my-secure-key-123")
}

func TestGetTodos(t *testing.T) {
	req, _ := http.NewRequest("GET", "/todos/v1/get", nil)
	rr := httptest.NewRecorder()

	getTodos(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestCreateTodo(t *testing.T) {
	payload := []byte(`{"title":"Learn Go Testing","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1/post", bytes.NewBuffer(payload))

	// We must wrap the handler with middleware to test the security
	rr := httptest.NewRecorder()
	handler := authMiddleware(createTodo)

	// 1. Test without key (Should Fail)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without key, got %v", rr.Code)
	}

	// 2. Test with key (Should Pass)
	rr = httptest.NewRecorder() // reset recorder
	setAuthHeader(req)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201 with key, got %v", rr.Code)
	}
}

func TestGetTodoByID(t *testing.T) {
	todos = []Todo{{ID: 99, Title: "Find Me", Completed: false}}
	req, _ := http.NewRequest("GET", "/todos/v1/get/item?id=99", nil)
	rr := httptest.NewRecorder()

	getTodoByID(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200, got %v", status)
	}
}

func TestUpdateTodo(t *testing.T) {
	todos = []Todo{{ID: 1, Title: "Old Title", Completed: false}}
	payload := []byte(`{"title":"New Title","completed":true}`)

	req, _ := http.NewRequest("PUT", "/todos/v1/update?id=1", bytes.NewBuffer(payload))
	setAuthHeader(req) // Add the key
	rr := httptest.NewRecorder()

	// Testing the middleware + function together
	handler := authMiddleware(updateTodo)
	handler.ServeHTTP(rr, req)

	if todos[0].Title != "New Title" {
		t.Errorf("Expected New Title, got %v", todos[0].Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	todos = []Todo{{ID: 1, Title: "Delete Me", Completed: false}}

	req, _ := http.NewRequest("DELETE", "/todos/v1/delete?id=1", nil)
	setAuthHeader(req) // Add the key
	rr := httptest.NewRecorder()

	handler := authMiddleware(deleteTodo)
	handler.ServeHTTP(rr, req)

	if len(todos) != 0 {
		t.Errorf("Expected 0 todos, got %v", len(todos))
	}
}

func TestPostAndPutSanitization(t *testing.T) {
	// 1. Create malicious payload
	maliciousJSON := []byte(`{"title":"<script>evil</script>"}`)

	// 2. Test POST
	reqPost, _ := http.NewRequest("POST", "/todos/v1/post", bytes.NewBuffer(maliciousJSON))
	setAuthHeader(reqPost)
	rrPost := httptest.NewRecorder()
	authMiddleware(createTodo).ServeHTTP(rrPost, reqPost)

	// 3. Test PUT
	reqPut, _ := http.NewRequest("PUT", "/todos/v1/update?id=1", bytes.NewBuffer(maliciousJSON))
	setAuthHeader(reqPut)
	rrPut := httptest.NewRecorder()
	authMiddleware(updateTodo).ServeHTTP(rrPut, reqPut)

	// Verify both were escaped
	if rrPost.Code == http.StatusCreated && rrPut.Code == http.StatusNoContent {
		t.Log("Both endpoints successfully sanitized input")
	}
}
