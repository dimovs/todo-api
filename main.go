package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	database "todo-api/db"
	"todo-api/handlers"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable not set")
	}

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
