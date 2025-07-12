package main

import (
	"database/sql"
	"encoding/json"
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

	// handlers
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/todos", getTodosHandler(db))
	mux.HandleFunc("/todos/create", createTodoHandler(db))

	fmt.Println("Starting service on http://localhost:8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed: %w", err)
	}
}

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	CreatedAt string `json:"created_at"`
}

func getTodosHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("select id, title, completed, created_at from todos order by id")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		todos := []Todo{}
		for rows.Next() {
			var todo Todo
			err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			todos = append(todos, todo)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(todos)
	}
}

func createTodoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var input struct {
			Title string `json:"title"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		var todo Todo

		err := db.QueryRow(`
			insert into todos (title) values ($1) returning id, title, completed, created_at`,
			input.Title,
		).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)

		if err != nil {
			http.Error(w, "Failed to create a todo", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(todo)
	}
}
