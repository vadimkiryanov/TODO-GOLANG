package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vadimkiryanov/db-service/model"
)

type HandlersService struct {
	handler *http.ServeMux
}

func NewHandlersService() *HandlersService {
	// Создаем новый сервис
	return &HandlersService{
		handler: http.NewServeMux(),
	}
}

func (s *HandlersService) InitRouters(db *sqlx.DB) *http.ServeMux {
	s.handler.HandleFunc("/api/list", handleList(db))

	return s.handler
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		todos := []model.TodoItem{}
		err := db.Select(&todos, "SELECT * FROM todo_items")
		if err != nil {
			log.Fatalf("Ошибка: %v", err.Error())
		}

		fmt.Printf("todos: %v, length: %v", todos, len(todos))
		w.Header().Set("Content-Type", "application/json")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		resp, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("err: %v\n", err)
			return
		}
		defer req.Body.Close()

		var item model.TodoItem
		err = json.Unmarshal(resp, &item)
		if err != nil {
			fmt.Println("Ошибка:", err)
			return
		}

		if len(item.Title) == 0 {
			fmt.Print("Body пустой \n")
			return
		}

		tx := db.MustBegin()
		tx.MustExec("INSERT INTO todo_items (id, title, done) VALUES ($1, $2, $3)", uuid.New().String(), item.Title, false)
		tx.Commit()

		// Query the database, storing results in a []Person (wrapped in []interface{})
		todos := []model.TodoItem{}
		err = db.Select(&todos, "SELECT * FROM todo_items")
		if err != nil {
			log.Fatalf("Ошибка: %v", err.Error())
		}

		fmt.Printf("todos: %v", todos)

		// Проксируем JSON обратно
		w.Header().Set("Content-Type", "application/json")
	}
}

// curl -X POST http://localhost:8080/delete -d '{"id":"3"}'
func handleDelete(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Читаем reqBody
		resp, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("err: %v\n", err)
			return
		}
		defer req.Body.Close()

		// Get idDelete from req
		var item model.TodoItem
		err = json.Unmarshal(resp, &item)
		if err != nil {
			log.Fatalf("err: %v\n", err)
			return
		}

		_, err = uuid.Parse(item.Id)
		if err != nil {
			http.Error(w, "Неверный формат ID", http.StatusBadRequest)
			return
		}

		tx := db.MustBegin()
		result, err := tx.Exec("DELETE FROM todo_items WHERE id = $1", item.Id)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
			return
		}
		tx.Commit()

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения количества удаленных строк: %v", err), http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			http.Error(w, "Запись с таким ID не найдена", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}

}

// curl -X PUT http://localhost:8080/done -d '{"id":"3"}'
func handleDone(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Читаем reqBody
		resp, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("err: %v\n", err)
			return
		}
		defer req.Body.Close()

		var item model.TodoItem
		err = json.Unmarshal(resp, &item)
		if err != nil {
			log.Fatalf("err: %v\n", err)
			return
		}

		// Convert str to int
		_, err = uuid.Parse(item.Id)
		if err != nil {
			log.Fatalf("err: %v\n", err)
			return
		}

		tx := db.MustBegin()
		result, err := tx.Exec("UPDATE todo_items SET done = true WHERE id = $1", item.Id)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Ошибка обновления: %v", err), http.StatusInternalServerError)
			return
		}
		tx.Commit()

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка получения количества обновленных строк: %v", err), http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			http.Error(w, "Запись с таким ID не найдена", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}

}

func handleList(db *sqlx.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(db)(w, r)
		case http.MethodPost:
			handleCreate(db)(w, r)
		case http.MethodDelete:
			handleDelete(db)(w, r)
		case http.MethodPut:
			handleDone(db)(w, r)

		default:
			http.Error(w, "Метод не доступен: ", http.StatusInternalServerError)
		}

		fmt.Println("Произошел запрос: ", r.Method)
	}
}
