// db-service/main.go
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

	_ "github.com/lib/pq" // драйвер для pg
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"github.com/vadimkiryanov/db-service/internal/handler"
	"github.com/vadimkiryanov/db-service/internal/pkg/server"
	"github.com/vadimkiryanov/db-service/internal/repository"
)

func init() {
	// Инициализация env
	if err := gotenv.Load(); err != nil {
		log.Fatalf("ошибка при инициализации env: %s", err.Error())
	}

	if err := initConfig(); err != nil {
		log.Fatalf("ошибка при инициализации конфига: %s", err.Error())
	}
}

var address = []string{"localhost:9091", "localhost:9092", "localhost:9093"}
var TOPIC = "my-topic-handlers"

func main() {
	db, err := repository.NewPostgresDB(&repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.user"),
		DBName:   viper.GetString("db.name"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
	}
	defer db.Close()

	h := handler.NewHandlersService()
	srv := server.NewServerHTTPClient("9000", h.InitRouters(db))

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ошибка при запуске сервера: %s", err.Error())
		}
	}()

	fmt.Println("DB-service стартовал")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	fmt.Println("DB-service завершает работу")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("ошибка при остановке сервера: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
