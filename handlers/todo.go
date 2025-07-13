package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"todo-api/models"
)

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func getTodos(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query("select id, title, completed, created_at from tododb.public.todos order by id")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
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

func createTodo(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	var todo models.Todo

	err := db.QueryRow(`
			insert into tododb.public.todos (title) values ($1) returning id, title, completed, created_at`,
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

func TodosHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTodos(w, db)
		case http.MethodPost:
			createTodo(w, r, db)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func TodosByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/todos/")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getTodoByID(w, db, id)
		case http.MethodDelete:
			deleteTodoByID(w, db, id)
		case http.MethodPut:
			updateTodoByID(w, r, db, id)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func getTodoByID(w http.ResponseWriter, db *sql.DB, id int) {
	var todo models.Todo
	err := db.QueryRow(
		`select id, title, completed, created_at from tododb.public.todos where id = $1`,
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to query todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func deleteTodoByID(w http.ResponseWriter, db *sql.DB, id int) {
	res, err := db.Exec(`delete from tododb.public.todos where id = $1`, id)

	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	affected, err := res.RowsAffected()

	if err != nil {
		http.Error(w, "Failed to confirm deletion", http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateTodoByID(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
	var input struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Title == nil && input.Completed == nil {
		http.Error(w, "Nothing to update", http.StatusBadRequest)
		return
	}

	query := `update tododb.public.todos set `
	params := []any{}
	i := 1

	if input.Title != nil {
		query += fmt.Sprintf("title = $%d,", i)
		params = append(params, *input.Title)
		i++
	}

	if input.Completed != nil {
		query += fmt.Sprintf("completed = $%d,", i)
		params = append(params, *input.Completed)
		i++
	}

	query = strings.TrimRight(query, ",") + fmt.Sprintf(" WHERE id = $%d RETURNING id, title, completed, created_at", i)
	params = append(params, id)

	var todo models.Todo
	err = db.QueryRow(query, params...).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}
