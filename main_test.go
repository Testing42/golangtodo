package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

	// 3. THE BEST METHOD: Unmarshal the response body into the TodoResponse wrapper
	var response handlers.TodoResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Now, you must look inside 'response.Data' to find your todos
	actualTodos := response.Data

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

	items := []string{"Buy Milk", "Buy Bread", "Wash Car"}
	for _, title := range items {
		_, err := handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
		if err != nil {
			t.Fatalf("Failed to setup search data: %v", err)
		}
	}

	// 2. Scenario A: Search for "Buy"
	req, _ := http.NewRequest("GET", "/todos/v1?search=Buy", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	// FIX: Use the Wrapper
	var response handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 search results for 'Buy', got %d", len(response.Data))
	}

	// 3. Scenario B: Search for "Milk"
	req, _ = http.NewRequest("GET", "/todos/v1?search=Milk", nil)
	rr = httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var responseB handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &responseB)
	if len(responseB.Data) != 1 || responseB.Data[0].Title != "Buy Milk" {
		t.Errorf("Expected 'Buy Milk', got %v", responseB.Data)
	}
}

func TestPagination(t *testing.T) {
	clearTable()

	// 1. Setup: Insert 15 items
	for i := 1; i <= 15; i++ {
		title := "Task " + strconv.Itoa(i)
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
	}

	// 2. Scenario: Page 1, Limit 10
	req, _ := http.NewRequest("GET", "/todos/v1?page=1&limit=10", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	// FIX: Use the Wrapper
	var responseP1 handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &responseP1)
	if len(responseP1.Data) != 10 {
		t.Errorf("Expected 10 items on page 1, got %d", len(responseP1.Data))
	}

	// 3. Scenario: Page 2, Limit 10
	req, _ = http.NewRequest("GET", "/todos/v1?page=2&limit=10", nil)
	rr = httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var responseP2 handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &responseP2)
	if len(responseP2.Data) != 5 {
		t.Errorf("Expected 5 items on page 2, got %d", len(responseP2.Data))
	}
}
