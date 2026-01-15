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

// TestMain handles the setup and teardown for the entire test suite
func TestMain(m *testing.M) {
	// Set environment for middleware
	os.Setenv("API_KEY", "test-secret-key")

	// Initialize the Test Database
	if err := handlers.InitDB(testDBFile); err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	code := m.Run()

	// Cleanup
	handlers.DB.Close()
	os.Remove(testDBFile)

	os.Exit(code)
}

// --- HELPERS ---

func clearTable() {
	handlers.DB.Exec("DELETE FROM todos")
	handlers.DB.Exec("DELETE FROM sqlite_sequence WHERE name='todos'")
}

func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", "test-secret-key")
}

// --- TIER 1: LOGIC (UNIT) TESTS ---
// Tests internal logic without needing the database connection

func TestSanitizationLogic(t *testing.T) {
	// Create a request with dirty HTML tags
	rawJSON := `{"title":"<script>alert('hack')</script>Hello","completed":false}`
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer([]byte(rawJSON)))
	rr := httptest.NewRecorder()

	todo, err := handlers.DecodeAndSanitize(rr, req)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	expected := "&lt;script&gt;alert(&#39;hack&#39;)&lt;/script&gt;Hello"
	if todo.Title != expected {
		t.Errorf("Sanitization failed. Got: %s", todo.Title)
	}
}

// --- TIER 2: INTEGRATION TESTS ---
// Tests the full flow: Handler -> Database -> JSON Response

func TestFullCreateAndFetchFlow(t *testing.T) {
	clearTable()

	// 1. Test POST (Create)
	payload := []byte(`{"title":"Integration Task","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer(payload))
	setAuthHeader(req)
	rr := httptest.NewRecorder()
	handlers.AuthMiddleware(handlers.CreateTodo).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %v", rr.Code)
	}

	// 2. Test GET (Fetch)
	reqGet, _ := http.NewRequest("GET", "/todos/v1", nil)
	rrGet := httptest.NewRecorder()
	handlers.GetTodos(rrGet, reqGet)

	var response handlers.TodoResponse
	json.Unmarshal(rrGet.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Fatalf("Expected 1 todo in list, got %d", len(response.Data))
	}

	if response.Data[0].Title != "Integration Task" {
		t.Errorf("Data mismatch. Got: %s", response.Data[0].Title)
	}

	if response.Data[0].CreatedAt.IsZero() {
		t.Error("CreatedAt was not populated by the database")
	}
}

func TestGetTodoByID(t *testing.T) {
	clearTable()
	handlers.DB.Exec("INSERT INTO todos (id, title, completed) VALUES (?, ?, ?)", 50, "Specific Item", true)

	req, _ := http.NewRequest("GET", "/todos/v1/item?id=50", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodoByID(rr, req)

	var found handlers.Todo
	json.Unmarshal(rr.Body.Bytes(), &found)

	if found.ID != 50 || found.Title != "Specific Item" {
		t.Errorf("Could not retrieve specific ID. Got: %+v", found)
	}
}

func TestPaginationAndSearch(t *testing.T) {
	clearTable()
	for i := 1; i <= 15; i++ {
		title := "Task " + strconv.Itoa(i)
		if i <= 5 {
			title = "Unique " + title
		} // 5 unique items
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
	}

	// Test Search
	req, _ := http.NewRequest("GET", "/todos/v1?search=Unique", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)
	var searchRes handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &searchRes)
	if searchRes.TotalCount != 5 {
		t.Errorf("Search failed. Expected 5 items, got %d", searchRes.TotalCount)
	}

	// Test Pagination
	req, _ = http.NewRequest("GET", "/todos/v1?page=2&limit=10", nil)
	rr = httptest.NewRecorder()
	handlers.GetTodos(rr, req)
	var pageRes handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &pageRes)
	if len(pageRes.Data) != 5 {
		t.Errorf("Pagination failed. Expected 5 items on page 2, got %d", len(pageRes.Data))
	}
}

// --- TIER 3: LOAD TESTS (BENCHMARKS) ---
// Measures performance under repeated execution

func BenchmarkGetTodos(b *testing.B) {
	clearTable()
	// Fill with 100 items for a realistic read test
	for i := 0; i < 100; i++ {
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", "Benchmark", false)
	}

	req, _ := http.NewRequest("GET", "/todos/v1?limit=10", nil)
	b.ResetTimer() // Don't count the setup time

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handlers.GetTodos(rr, req)
	}
}
