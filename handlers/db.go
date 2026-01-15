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

	// FIX 1: Use 'path' instead of 'dataSourceName'
	// FIX 2: Use '=' instead of ':=' to avoid shadowing the global DB variable
	DB, err = sql.Open("sqlite", path+"?_loc=auto&_parseTime=true")
	if err != nil {
		return err
	}

	// Ping the database to ensure the connection is actually valid
	if err = DB.Ping(); err != nil {
		return err
	}

	// Create table with the created_at column using local time
	query := `
    CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        completed BOOLEAN NOT NULL DEFAULT 0,
        created_at DATETIME DEFAULT (datetime('now','localtime'))
    );`
	_, err = DB.Exec(query)
	return err
}

// MigrateFromJSON remains the same, but ensure it uses the global DB
func MigrateFromJSON(jsonPath string) {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if count > 0 {
		return
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return
	}

	var oldTodos []Todo
	json.Unmarshal(data, &oldTodos)

	for _, t := range oldTodos {
		// Updated to handle the created_at if necessary, or let SQL default it
		DB.Exec("INSERT INTO todos (id, title, completed) VALUES (?, ?, ?)", t.ID, t.Title, t.Completed)
	}
	fmt.Println("Migration from JSON to SQLite successful.")
}
