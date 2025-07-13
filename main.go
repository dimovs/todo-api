package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	database "todo-api/db"
	"todo-api/handlers"
)

func main() {
	dsn := "postgres://dimovs@localhost:5432/tododb?sslmode=disable"

	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %w", err)
	}

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
