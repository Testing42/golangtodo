package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO required)
)

var DB *sql.DB

// InitDB initializes the SQLite connection and creates the table
func InitDB(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// Create table if it doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0
	);`
	_, err = DB.Exec(query)
	return err
}

// MigrateFromJSON moves data from todos.json to SQLite if the DB is empty
func MigrateFromJSON(jsonPath string) {
	// Check if we already have data in SQL
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if count > 0 {
		return // Already migrated
	}

	// Read JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return // No JSON file to migrate
	}

	var oldTodos []Todo
	json.Unmarshal(data, &oldTodos)

	for _, t := range oldTodos {
		DB.Exec("INSERT INTO todos (id, title, completed) VALUES (?, ?, ?)", t.ID, t.Title, t.Completed)
	}
	fmt.Println("Migration from JSON to SQLite successful.")
}
