// api-service
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/vadimkiryanov/api-service/internal/handlers"
	"github.com/vadimkiryanov/api-service/internal/pkg/kafka"
	"github.com/vadimkiryanov/api-service/internal/pkg/server"
)

var address = []string{"localhost:9091", "localhost:9092", "localhost:9093"}

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("ошибка при инициализации конфига: %s", err.Error())
	}
	// Мультиплексер
	producer, err := kafka.NewProducer(address)
	if err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
	}
	defer producer.Close()
	hs := handlers.NewHandlersService(producer)
	sm := hs.InitRouters()

	srv := server.NewServerHTTPClient(viper.GetString("port"), sm)

	// Запускаем сервер в отдельной горутине, чтобы он не блокировал основной поток.
	// Это позволяет нам слушать системные сигналы в основном потоке.
	go func() {
		if err != nil {
			log.Fatalf("Ошибка: %v", err.Error())
		}

		// srv.Run() - это блокирующая операция.
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске сервера: %s", err.Error())
		}

	}()

	fmt.Println("Api-service стартовал")

	// Создаем канал для получения сигналов от операционной системы.
	// Мы будем слушать SIGTERM (стандартный сигнал для завершения) и SIGINT (сигнал от Ctrl+C).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Блокируем основной поток до тех пор, пока не получим сигнал в канал quit.
	<-quit

	fmt.Println("Api-service завершает работу")

	// Создаем контекст с таймаутом для graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel() гарантирует, что ресурсы, связанные с контекстом, будут освобождены.
	defer cancel()

	// Вызываем Shutdown для "вежливой" остановки сервера.
	// Он перестает принимать новые запросы и ждет завершения старых.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке сервера: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
