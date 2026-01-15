package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Testing42/golangtodo/handlers"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// 1. Setup Structured Logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 2. Configuration
	dbFile := getEnv("DB_FILE", "./todos.db")
	port := getEnv("PORT", "8080")

	// 3. Database Initialization
	if err := handlers.InitDB(dbFile); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer handlers.DB.Close()

	// 4. Migrations & Routes
	handlers.MigrateFromJSON("todos.json")
	handlers.RegisterRoutes()

	// 5. Execution
	runServer(&http.Server{Addr: ":" + port}, dbFile, port)
}
