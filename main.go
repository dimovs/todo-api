package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"todo-api/handlers"
)

func main() {
	// database
	dsn := "postgres://dimovs@localhost:5432/tododb?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %w", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("DB is unreachable: %w", err)
	}

	fmt.Println("Connected to PostgreSQL")

	// http server
	mux := http.NewServeMux()

	// handlers
	mux.HandleFunc("/health", handlers.GetHealth)
	mux.HandleFunc("/todos", handlers.TodosHandler(db))
	mux.HandleFunc("/todos/", handlers.TodosByIDHandler(db))

	fmt.Println("Starting service on http://localhost:8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed: %w", err)
	}
}
