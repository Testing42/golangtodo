Go Todo API (SQLite Edition)
This project has the following features
Persistent storage using SQLite, Structured Logging: JSON-formatted logs using slog for production monitoring,
Security using Middleware-based authentication using API Keys, Clean Architecture with Separated concerns between routing, server lifecycle, and business logic.
Testing suite including (Unit, Integration, and Load Benchmarks).

Getting Started1. PrerequisitesGo 1.25.5
For testing you will need to run command line commands or use a tool like postman.

Installation git clone https://github.com/Testing42/golangtodo.git
cd golangtodo
go mod tidy
3. ConfigurationCreate a .env file in the root directory:Code snippetPORT=8080
DB_FILE=todos.db
API_KEY=super-secret-admin-key-123
4. Running the test run .
TestingRun All Tests go test -v ./...
Run Performance BenchmarksTo simulate high load and check SQLite performance:Bashgo test -bench=. -benchmem
