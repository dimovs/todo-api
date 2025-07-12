package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
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

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Println("Starting service on http://localhost:8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed: %w", err)
	}
}
