// api-service/main.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/viper"
	"github.com/vadimkiryanov/api-service/pkg/server"
)

const (
	BASE_URL = "http://localhost:9000/api/list"
)

// curl -X GET http://localhost:8080/list -v
func handleGet(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(BASE_URL)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	defer resp.Body.Close() // Чтобы не было утечек памяти

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	fmt.Printf("list_todo: %v\n", string(body))
	// Отправляем JSON-ответ клиенту
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// ~ curl -X POST http://localhost:8080/create -d '{"title":"new title"}'
func handleCreate(w http.ResponseWriter, req *http.Request) {
	resp, err := http.Post(BASE_URL, "application/json", req.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	defer resp.Body.Close() // Чтобы не было утечек памяти

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	fmt.Printf("list_todo: %v\n", string(body))
}
func handleDelete(w http.ResponseWriter, req *http.Request) {
	deleteReq, err := http.NewRequest(http.MethodDelete, "http://localhost:9000/api/list", req.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(deleteReq)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	defer resp.Body.Close() // Чтобы не было утечек памяти
}

// curl -X PUT http://localhost:8080/done -d '{"id":"3"}'
func handleDone(w http.ResponseWriter, req *http.Request) {
	putReq, err := http.NewRequest(http.MethodPut, BASE_URL, req.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(putReq)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	defer resp.Body.Close() // Чтобы не было утечек памяти

}

func main() {
	err := initConfig()
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
	}

	// Мультиплексер
	sm := http.NewServeMux()
	sm.HandleFunc("/list", handleGet)
	sm.HandleFunc("/create", handleCreate)
	sm.HandleFunc("/delete", handleDelete)
	sm.HandleFunc("/done", handleDone)

	s := server.NewServerHTTPClient(viper.GetString("port"), sm)
	err = s.Run()
	if err != nil {
		fmt.Printf("\"Ошибка запуска сервера\": %v\n", "Ошибка запуска сервера")
	}

}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
