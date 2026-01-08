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

func TestGetTodoByID(t *testing.T) {
	// Manually add a todo to the list so we have something to find
	todos = append(todos, Todo{ID: 99, Title: "Find Me", Completed: false})

	// Create request with query parameter ?id=99
	req, _ := http.NewRequest("GET", "/todos/v1/get/item?id=99", nil)
	rr := httptest.NewRecorder()

	getTodoByID(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200, got %v", status)
	}

	var found Todo
	json.Unmarshal(rr.Body.Bytes(), &found)
	if found.Title != "Find Me" {
		t.Errorf("Expected 'Find Me', got %v", found.Title)
	}
}

func TestUpdateTodo(t *testing.T) {
	todos = []Todo{{ID: 1, Title: "Old Title", Completed: false}}

	payload := []byte(`{"title":"New Title","completed":true}`)
	req, _ := http.NewRequest("PUT", "/todos/v1/update?id=1", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()

	updateTodo(rr, req)

	if todos[0].Title != "New Title" {
		t.Errorf("Expected New Title, got %v", todos[0].Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	todos = []Todo{{ID: 1, Title: "Delete Me", Completed: false}}

	req, _ := http.NewRequest("DELETE", "/todos/v1/delete?id=1", nil)
	rr := httptest.NewRecorder()

	deleteTodo(rr, req)

	if len(todos) != 0 {
		t.Errorf("Expected 0 todos, got %v", len(todos))
	}
}
