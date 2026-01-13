package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Testing42/golangtodo/handlers"
)

// Define the test filename constant to avoid "magic strings"
const testDB = "test_todos.json"

// setupAndTeardown prepares the environment and ensures no stale test data exists
func setupAndTeardown() {
	// 1. Set the environment variable so getDBPath() in store.go picks it up
	os.Setenv("DB_FILE", testDB)
	os.Setenv("API_KEY", "test-secret-key")

	// 2. Reset global memory state
	handlers.Todos = []*handlers.Todo{}
	handlers.NextID = 1

	// 3. Remove any leftover test files from previous runs
	os.Remove(testDB)
	os.Remove(testDB + ".tmp")
}

func TestMain(m *testing.M) {
	setupAndTeardown()

	// Run all tests in the package
	code := m.Run()

	// Final cleanup: remove test artifacts so the directory stays clean
	os.Remove(testDB)
	os.Remove(testDB + ".tmp")

	os.Exit(code)
}

// Helper to set headers using the test secret
func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "test-secret-key")
}

func TestPersistence(t *testing.T) {
	// Setup: Add a Todo and manually trigger a save
	handlers.Todos = []*handlers.Todo{{ID: 1, Title: "Test Persistence", Completed: false}}
	err := handlers.SaveToJSON()
	if err != nil {
		t.Fatalf("Failed to save to %s: %v", testDB, err)
	}

	// Simulating a restart: Clear memory
	handlers.Todos = []*handlers.Todo{}

	// Load from the test file
	err = handlers.LoadFromJSON()
	if err != nil {
		t.Fatalf("Failed to load from %s: %v", testDB, err)
	}

	// Verify the data survived
	if len(handlers.Todos) != 1 || handlers.Todos[0].Title != "Test Persistence" {
		t.Errorf("Data did not persist correctly. Got: %v", handlers.Todos)
	}
}

func TestCreateTodoWithPersistence(t *testing.T) {
	// Ensure a clean start for this specific test
	handlers.Todos = []*handlers.Todo{}
	handlers.NextID = 1
	os.Remove(testDB)

	payload := []byte(`{"title":"Save Me To Disk","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer(payload))
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	handlers.AuthMiddleware(handlers.CreateTodo).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %v", rr.Code)
	}

	// Check for the TEST file, not the production file
	if _, err := os.Stat(testDB); os.IsNotExist(err) {
		t.Errorf("%s was not created after POST request", testDB)
	}
}
