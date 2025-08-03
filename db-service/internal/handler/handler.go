package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/vadimkiryanov/db-service/internal/pkg/kafka"
	"github.com/vadimkiryanov/db-service/model"
)

type HandlersService struct {
	handler  *http.ServeMux
	consumer *kafka.Consumer
}

const TOPIC = "my-topic-handlers"

func NewHandlersService() *HandlersService {
	// Создаем новый сервис
	return &HandlersService{
		handler: http.NewServeMux(),
	}
}

func (s *HandlersService) InitRouters(db *sqlx.DB, rdb *redis.Client, ctx *context.Context) *http.ServeMux {
	s.handler.HandleFunc("/api/list", handleList(db, rdb, *ctx))

	return s.handler
}

func handleGet(db *sqlx.DB, rdb *redis.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		todos := []model.TodoItem{}

		// Проверяем в редисе наличие данных
		key := "todo_items"
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			logrus.Errorf("Ошибка получения данных из redis-БД: %s", err.Error())
		}

		logrus.Infof("val: %s", val)
		if val != "" {
			logrus.Infof("Данные найдены в redis-БД")
			if err := json.Unmarshal([]byte(val), &todos); err != nil {
				logrus.Errorf("Ошибка декодирования JSON из redis: %s", err.Error())
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(todos); err != nil {
				logrus.Errorf("Ошибка кодирования JSON: %s", err.Error())
			}
			return
		}

		// Получение данных из постгреса
		time.Sleep(time.Second * 5)
		err = db.Select(&todos, "SELECT * FROM todo_items")
		if err != nil {
			logrus.Errorf("Ошибка получения данных из БД: %s", err.Error())
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		// Сохраняем данные в редисе
		data, err := json.Marshal(todos)
		if err != nil {
			logrus.Errorf("Ошибка кодирования JSON для redis: %s", err.Error())
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		rdb.Set(ctx, key, data, 30*time.Second)
		logrus.Infof("Данные сохранены в redis-БД")

		// Отправляем ответ клинту в JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(todos); err != nil {
			log.Printf("Ошибка кодирования JSON: %v", err)
		}
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

		// Снова получаем обновленный список
		todos := []model.TodoItem{}
		err = db.Select(&todos, "SELECT * FROM todo_items")
		if err != nil {
			log.Fatalf("Ошибка: %v", err.Error())
		}

		// Отправляем ответ клинту в JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(todos); err != nil {
			log.Printf("Ошибка кодирования JSON: %v", err)
		}
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

		// Отправляем ответ клинту в JSON
		w.WriteHeader(http.StatusOK)
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

		todos := []model.TodoItem{}
		err = db.Select(&todos, "SELECT * FROM todo_items WHERE id = $1", item.Id)
		if err != nil {
			log.Printf("Ошибка получения данных из БД: %v", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		// Отправляем ответ клинту в JSON
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(todos); err != nil {
			log.Printf("Ошибка кодирования JSON: %v", err)
		}
	}

}

func handleList(db *sqlx.DB, rdb *redis.Client, ctx context.Context) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(db, rdb, ctx)(w, r)
		case http.MethodPost:
			handleCreate(db)(w, r)
		case http.MethodDelete:
			handleDelete(db)(w, r)
		case http.MethodPut:
			handleDone(db)(w, r)

		default:
			http.Error(w, "Метод не доступен: ", http.StatusInternalServerError)
		}

		fmt.Println("\nПроизошел запрос: ", r.Method)
	}
}
