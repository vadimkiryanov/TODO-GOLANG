// db-service/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/vadimkiryanov/db-service/internal/pkg/server"
	"github.com/vadimkiryanov/db-service/model"
)

var mock_db = map[int]model.TodoItem{
	1: {Title: "Some title 1", Done: false},
	2: {Title: "Some title 2", Done: false},
	3: {Title: "Some title 3", Done: false},
}

var count_index = len(mock_db) + 1

func get(w http.ResponseWriter, _ *http.Request) {
	// Проксируем JSON обратно
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mock_db); err != nil {
		http.Error(w, "Ошибка декодирования: ", http.StatusInternalServerError)
		return
	}
	fmt.Println("Произошел запрос")
}
func create(w http.ResponseWriter, req *http.Request) {
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

	fmt.Printf("resp: %v\n", string(resp))

	if len(item.Title) == 0 {
		fmt.Print("Body пустой \n")
		return
	}

	mock_db[count_index] = model.TodoItem{Title: fmt.Sprintf("%v %v", item.Title, count_index)}
	count_index = len(mock_db) + 1

	// Проксируем JSON обратно
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mock_db); err != nil {
		http.Error(w, "Ошибка декодирования: ", http.StatusInternalServerError)
		return
	}

}

// curl -X POST http://localhost:8080/delete -d '{"id":"3"}'
func delete_(w http.ResponseWriter, req *http.Request) {
	// Читаем reqBody
	resp, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}
	defer req.Body.Close()

	// Get idDelete from req
	var item model.TodoItemId
	err = json.Unmarshal(resp, &item)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	// Convert str to int
	itemIdNum, err := strconv.Atoi(item.Id)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	// deleting
	delete(mock_db, itemIdNum)

	fmt.Printf("itemIdNum: %v\n", itemIdNum)
	fmt.Printf("mock_db: %v\n", mock_db)

}
func put(w http.ResponseWriter, req *http.Request) {
	// Читаем reqBody
	resp, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}
	defer req.Body.Close()

	var item model.TodoItemId
	err = json.Unmarshal(resp, &item)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	// Convert str to int
	itemIdNum, err := strconv.Atoi(item.Id)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	mock_db[itemIdNum] = model.TodoItem{Title: mock_db[itemIdNum].Title, Done: true}

	fmt.Printf("itemIdNum: %v\n", itemIdNum)
	fmt.Printf("mock_db: %v\n", mock_db)

}
func handleList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		get(w, r)
	case http.MethodPost:
		create(w, r)
	case http.MethodDelete:
		delete_(w, r)
	case http.MethodPut:
		put(w, r)

	default:
		http.Error(w, "Метод не доступен: ", http.StatusInternalServerError)
	}

	fmt.Println("Произошел запрос: ", r.Method)
}

func main() {
	err := initConfig()
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
	}

	if err := initDb(); err != nil {
		log.Fatal("Ошибка подключения:", err)
	}

	sm := http.NewServeMux()
	sm.HandleFunc("/api/list", handleList)

	s := server.NewServerHTTPClient("9000", sm)
	err = s.Run()

	if err != nil {
		fmt.Printf("\"Ошибка запуска сервера\": %v\n", "Ошибка запуска сервера")
	}
}

func initDb() error {
	// Строка подключения
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s", viper.GetString("db_user"),
		viper.GetString("db_password"),
		viper.GetString("db_host"),
		viper.GetString("db_port"),
		viper.GetString("sslmode"),
	)

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
		return err
	}
	defer conn.Close(ctx)

	var greeting string
	err = conn.QueryRow(ctx, "SELECT 'Hello, PostgreSQL!'").Scan(&greeting)
	if err != nil {
		log.Fatal("Ошибка выполнения запроса:", err)
		return err
	}

	fmt.Println(greeting)

	return nil
}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
