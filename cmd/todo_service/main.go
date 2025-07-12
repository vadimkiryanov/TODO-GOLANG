// api-service/main.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/vadimkiryanov/todo-golang/pkg/server"
)

func handlerList_Get(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://localhost:9000/api/list")
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

// ~ curl -X POST http://localhost:8080/create -d '{"title":"new title"}'
func handlerList_Create(w http.ResponseWriter, req *http.Request) {
	resp, err := http.Post("http://localhost:9000/api/list", "application/json", req.Body)
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
func handlerList_Delete(w http.ResponseWriter, req *http.Request) {
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
func handlerList_Done(w http.ResponseWriter, req *http.Request) {
	putReq, err := http.NewRequest(http.MethodPut, "http://localhost:9000/api/list", req.Body)
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
	// Мультиплексер
	sm := http.NewServeMux()
	sm.HandleFunc("/list", handlerList_Get)
	sm.HandleFunc("/create", handlerList_Create)
	sm.HandleFunc("/delete", handlerList_Delete)
	sm.HandleFunc("/done", handlerList_Done)

	s := new(server.ServerHTTP)

	err := s.Run("8080", sm)
	if err != nil {
		fmt.Printf("\"Ошибка запуска сервера\": %v\n", "Ошибка запуска сервера")
	}

}
