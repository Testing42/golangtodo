package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Testing42/golangtodo/handlers"
)

// We use a separate database file for testing
const testDBFile = "test_todos.db"

func TestMain(m *testing.M) {
	// 1. Setup: Set API Key for middleware
	os.Setenv("API_KEY", "test-secret-key")

	// 2. Initialize the Test Database
	if err := handlers.InitDB(testDBFile); err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// 3. Run the tests
	code := m.Run()

	// 4. Teardown: Close connection and remove test file
	handlers.DB.Close()
	os.Remove(testDBFile)

	os.Exit(code)
}

// Helper to clear the table between tests to ensure a clean slate
func clearTable() {
	handlers.DB.Exec("DELETE FROM todos")
	// Reset the auto-increment counter in SQLite
	handlers.DB.Exec("DELETE FROM sqlite_sequence WHERE name='todos'")
}

// Helper to set auth headers
func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "test-secret-key")
}

func TestPersistence(t *testing.T) {
	clearTable()

	// 1. Setup data in DB
	_, err := handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", "Test Persistence", false)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// 2. Call the handler
	req, _ := http.NewRequest("GET", "/todos/v1", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	// 3. THE BEST METHOD: Unmarshal the response body into a slice
	var actualTodos []handlers.Todo
	err = json.Unmarshal(rr.Body.Bytes(), &actualTodos)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 4. Perform logical checks on the data
	if len(actualTodos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(actualTodos))
	}

	if actualTodos[0].Title != "Test Persistence" {
		t.Errorf("Expected title 'Test Persistence', got '%s'", actualTodos[0].Title)
	}

	if actualTodos[0].ID != 1 {
		t.Errorf("Expected ID 1, got %d", actualTodos[0].ID)
	}
}

func TestCreateTodoWithSQL(t *testing.T) {
	clearTable()

	payload := []byte(`{"title":"Save Me To SQLite","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer(payload))
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	// Use the actual AuthMiddleware and Handler
	handlers.AuthMiddleware(handlers.CreateTodo).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %v", rr.Code)
	}

	// 4. Verify the record actually exists in the SQL table
	var title string
	err := handlers.DB.QueryRow("SELECT title FROM todos WHERE id = 1").Scan(&title)
	if err == sql.ErrNoRows {
		t.Errorf("Record was not found in the database after POST")
	} else if title != "Save Me To SQLite" {
		t.Errorf("Expected title 'Save Me To SQLite', got '%s'", title)
	}
}

// Add this function to your main_test.go file

func TestSearchTodos(t *testing.T) {
	clearTable()

	// 1. Setup: Insert multiple items to search through
	items := []string{"Buy Milk", "Buy Bread", "Wash Car"}
	for _, title := range items {
		_, err := handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
		if err != nil {
			t.Fatalf("Failed to setup search data: %v", err)
		}
	}

	// 2. Scenario A: Search for "Buy" (should return 2 items)
	req, _ := http.NewRequest("GET", "/todos/v1?search=Buy", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var searchResults []handlers.Todo
	json.Unmarshal(rr.Body.Bytes(), &searchResults)

	if len(searchResults) != 2 {
		t.Errorf("Expected 2 search results for 'Buy', got %d", len(searchResults))
	}

	// 3. Scenario B: Search for "Milk" (should return 1 item)
	req, _ = http.NewRequest("GET", "/todos/v1?search=Milk", nil)
	rr = httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	json.Unmarshal(rr.Body.Bytes(), &searchResults)
	if len(searchResults) != 1 || searchResults[0].Title != "Buy Milk" {
		t.Errorf("Expected 'Buy Milk', got %v", searchResults)
	}

	// 4. Scenario C: Search for "Elephant" (should return 0 items)
	req, _ = http.NewRequest("GET", "/todos/v1?search=Elephant", nil)
	rr = httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	json.Unmarshal(rr.Body.Bytes(), &searchResults)
	if len(searchResults) != 0 {
		t.Errorf("Expected 0 results for 'Elephant', got %d", len(searchResults))
	}
}
