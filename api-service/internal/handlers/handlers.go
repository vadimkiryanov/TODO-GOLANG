package handlers

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vadimkiryanov/api-service/internal/pkg/kafka"
)

type HandlersService struct {
	handler  *http.ServeMux
	producer *kafka.Producer
}

func NewHandlersService(producer *kafka.Producer) *HandlersService {
	// Создаем новый сервис
	return &HandlersService{
		handler:  http.NewServeMux(),
		producer: producer,
	}
}

func (serv *HandlersService) InitRouters() *http.ServeMux {
	serv.handler.HandleFunc("/list", serv.handleGet)
	serv.handler.HandleFunc("/create", serv.handleCreate)
	serv.handler.HandleFunc("/delete", serv.handleDelete)
	serv.handler.HandleFunc("/done", serv.handleDone)
	return serv.handler
}

// routers
const (
	BASE_URL = "http://localhost:9000/api/list"
	TOPIC    = "my-topic-handlers"
)

// curl -X GET http://localhost:8080/list -v
func (serv *HandlersService) handleGet(w http.ResponseWriter, req *http.Request) {
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

	if err := serv.producer.Produce(http.MethodGet, TOPIC, "key", time.Now()); err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
	}

	// Отправляем JSON-ответ клиенту
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// ~ curl -X POST http://localhost:8080/create -d '{"title":"new title"}'
func (serv *HandlersService) handleCreate(w http.ResponseWriter, req *http.Request) {
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

	if err := serv.producer.Produce(http.MethodPost, TOPIC, "key", time.Now()); err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (serv *HandlersService) handleDelete(w http.ResponseWriter, req *http.Request) {
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

	if err := serv.producer.Produce(http.MethodDelete, TOPIC, "key", time.Now()); err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
}

// curl -X PUT http://localhost:8080/done -d '{"id":"3"}'
func (serv *HandlersService) handleDone(w http.ResponseWriter, req *http.Request) {
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

	if err := serv.producer.Produce(http.MethodPut, TOPIC, "key", time.Now()); err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

}
