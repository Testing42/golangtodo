package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTodos(t *testing.T) {
	// Setup: add a test todo
	todos = []Todo{{ID: 1, Title: "Test Todo", Completed: false}}
	nextID = 2

	// Create a test request
	req, _ := http.NewRequest("GET", "/todos", nil)
	w := httptest.NewRecorder()

	// Call the handler
	getTodos(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check response body
	var resp []Todo
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 || resp[0].Title != "Test Todo" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestCreateTodo(t *testing.T) {
	// Reset state for this test
	todos = []Todo{}
	nextID = 1

	// Create a test request body
	body := []byte(`{"title": "Test Todo", "completed": false}`)
	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the handler
	createTodo(w, req)

	// Check status
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Check response body
	var createdTodo Todo
	json.NewDecoder(w.Body).Decode(&createdTodo)
	if createdTodo.ID != 1 || createdTodo.Title != "Test Todo" {
		t.Errorf("unexpected created todo: %v", createdTodo)
	}

	// Check if it's in the todos slice
	if len(todos) != 1 || todos[0].ID != 1 {
		t.Errorf("todo not stored correctly")
	}
}
