package handlers

import (
	"io"
	"log"
	"net/http"
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

func (s *HandlersService) InitRouters() *http.ServeMux {
	s.handler.HandleFunc("/list", handleGet)
	s.handler.HandleFunc("/create", handleCreate)
	s.handler.HandleFunc("/delete", handleDelete)
	s.handler.HandleFunc("/done", handleDone)
	return s.handler
}

// routers
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("err: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

}
