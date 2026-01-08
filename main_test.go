package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTodos(t *testing.T) {
	// Create a mock request
	req, _ := http.NewRequest("GET", "/todos", nil)

	// ResponseRecorder acts like a mini-browser to capture the result
	rr := httptest.NewRecorder()

	// We call the function directly
	getTodos(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if the response is valid JSON (even if empty list)
	var response []Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Response was not valid JSON: %v", err)
	}
}

func TestCreateTodo(t *testing.T) {
	// Create a new Todo to send
	payload := []byte(`{"title":"Learn Go Testing","completed":false}`)

	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()

	// Call the separate create function
	createTodo(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Verify the ID was incremented
	var created Todo
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created.ID == 0 {
		t.Error("Expected a generated ID, but got 0")
	}
}
