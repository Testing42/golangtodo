package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/Testing42/golangtodo/handlers"
)

// TestMain handles the setup and teardown for the entire test suite
func TestMain(m *testing.M) {
	//Make test logs look like production json logs
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	// 1. FORCE environment for testing (Overrides .env)
	os.Setenv("API_KEY", "test-secret-key")
	os.Setenv("DB_FILE", "test_todos.db")

	// 2. Initialize the Test Database using the env variable
	if err := handlers.InitDB(os.Getenv("DB_FILE")); err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	code := m.Run()

	// 3. Cleanup
	handlers.DB.Close()
	os.Remove(os.Getenv("DB_FILE"))

	os.Exit(code)
}

// --- HELPERS ---

func clearTable() {
	handlers.DB.Exec("DELETE FROM todos")
	handlers.DB.Exec("DELETE FROM sqlite_sequence WHERE name='todos'")
}

func setAuthHeader(req *http.Request) {
	req.Header.Set("X-API-KEY", os.Getenv("API_KEY"))
}

// --- TIER 1: LOGIC (UNIT) TESTS ---

func TestSanitizationLogic(t *testing.T) {
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

func TestFullCreateAndFetchFlow(t *testing.T) {
	clearTable()

	payload := []byte(`{"title":"Integration Task","completed":false}`)
	req, _ := http.NewRequest("POST", "/todos/v1", bytes.NewBuffer(payload))
	setAuthHeader(req)
	rr := httptest.NewRecorder()
	handlers.AuthMiddleware(handlers.CreateTodo).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %v", rr.Code)
	}

	reqGet, _ := http.NewRequest("GET", "/todos/v1", nil)
	rrGet := httptest.NewRecorder()
	handlers.GetTodos(rrGet, reqGet)

	var response handlers.TodoResponse
	json.Unmarshal(rrGet.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Fatalf("Expected 1 todo in list, got %d", len(response.Data))
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
		}
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", title, false)
	}

	req, _ := http.NewRequest("GET", "/todos/v1?search=Unique", nil)
	rr := httptest.NewRecorder()
	handlers.GetTodos(rr, req)
	var searchRes handlers.TodoResponse
	json.Unmarshal(rr.Body.Bytes(), &searchRes)
	if searchRes.TotalCount != 5 {
		t.Errorf("Search failed. Expected 5 items, got %d", searchRes.TotalCount)
	}

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

func BenchmarkGetTodos(b *testing.B) {
	clearTable()
	for i := 0; i < 100; i++ {
		handlers.DB.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", "Benchmark", false)
	}

	req, _ := http.NewRequest("GET", "/todos/v1?limit=10", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handlers.GetTodos(rr, req)
	}
}

// BenchmarkHighLoadSQLite simulates 10,000 operations
func BenchmarkHighLoadSQLite(b *testing.B) {
	clearTable()

	// Pre-fill one item to ensure we are testing "updates" or "reads"
	handlers.DB.Exec("INSERT INTO todos (id, title, completed) VALUES (1, 'Initial', 0)")

	req, _ := http.NewRequest("GET", "/todos/v1/item?id=1", nil)

	b.ResetTimer() // Start the clock AFTER setup
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handlers.GetTodoByID(rr, req)
	}
}
