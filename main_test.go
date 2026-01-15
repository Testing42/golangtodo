package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/Testing42/golangtodo/handlers"
)

const testDBFile = "test_todos.db"

func TestMain(m *testing.M) {
	os.Setenv("API_KEY", "test-secret-key")

	// Initialize the Test Database
	if err := handlers.InitDB(testDBFile); err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	code := m.Run()

	handlers.DB.Close()
	os.Remove(testDBFile)

	os.Exit(code)
}

func clearTable() {
	handlers.DB.Exec("DELETE FROM todos")
	handlers.DB.Exec("DELETE FROM sqlite_sequence WHERE name='todos'")
}

func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "test-secret-key")
}

// TestPersistence: Verifies list fetching and the CreatedAt timestamp
func TestPersistence(t *testing.T) {
	clearTable()

	_, err := handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", "Test Persistence", false)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	req, _ := http.NewRequest("GET", "/todos/v1", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var response handlers.TodoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	actualTodos := response.Data

	if len(actualTodos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(actualTodos))
	}

	// Check if the date was actually set (should not be year 0001)
	if actualTodos[0].CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be populated, but it was the zero time")
	}
}

// TestCreateTodoWithSQL: Verifies that POSTing a todo generates a valid record with a date
func TestCreateTodoWithSQL(t *testing.T) {
	clearTable()

	payload := []byte(`{"title":"Save Me To SQLite","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer(payload))
	setAuthHeader(req)
	rr := httptest.NewRecorder()

	handlers.AuthMiddleware(handlers.CreateTodo).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %v", rr.Code)
	}

	// Unmarshal the response to check the date returned by the API
	var created handlers.Todo
	json.Unmarshal(rr.Body.Bytes(), &created)

	if created.CreatedAt.IsZero() {
		t.Error("The created todo returned a zero timestamp")
	}

	// Verify the database record directly
	var title string
	var createdAt string // Scan as string for raw DB check
	err := handlers.DB.QueryRow("SELECT title, created_at FROM todos WHERE id = 1").Scan(&title, &createdAt)
	if err != nil {
		t.Errorf("Record was not found in the database: %v", err)
	}
}

func TestGetTodoByID(t *testing.T) {
	clearTable()

	// Insert a specific item
	handlers.DB.Exec("INSERT INTO todos (id, title, completed) VALUES (?, ?, ?)", 99, "Find Me", false)

	req, _ := http.NewRequest("GET", "/todos/v1/item?id=99", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodoByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %v", rr.Code)
	}

	var found handlers.Todo
	json.Unmarshal(rr.Body.Bytes(), &found)

	if found.Title != "Find Me" || found.ID != 99 {
		t.Errorf("Expected 'Find Me' with ID 99, got ID %d Title '%s'", found.ID, found.Title)
	}
}

func TestSearchTodos(t *testing.T) {
	clearTable()

	items := []string{"Buy Milk", "Buy Bread", "Wash Car"}
	for _, title := range items {
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
	}

	// Search for "Buy"
	req, _ := http.NewRequest("GET", "/todos/v1?search=Buy", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var response handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 search results for 'Buy', got %d", len(response.Data))
	}
}

func TestPagination(t *testing.T) {
	clearTable()

	for i := 1; i <= 15; i++ {
		title := "Task " + strconv.Itoa(i)
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
	}

	// Scenario: Page 1, Limit 10
	req, _ := http.NewRequest("GET", "/todos/v1?page=1&limit=10", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var responseP1 handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &responseP1)
	if len(responseP1.Data) != 10 {
		t.Errorf("Expected 10 items on page 1, got %d", len(responseP1.Data))
	}
	if responseP1.TotalCount != 15 {
		t.Errorf("Expected total_count 15, got %d", responseP1.TotalCount)
	}

	// Scenario: Page 2, Limit 10
	req, _ = http.NewRequest("GET", "/todos/v1?page=2&limit=10", nil)
	rr = httptest.NewRecorder()
	handlers.GetTodos(rr, req)

	var responseP2 handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &responseP2)
	if len(responseP2.Data) != 5 {
		t.Errorf("Expected 5 items on page 2, got %d", len(responseP2.Data))
	}
}
