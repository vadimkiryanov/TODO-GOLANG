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
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"github.com/vadimkiryanov/db-service/internal/handler"
	"github.com/vadimkiryanov/db-service/internal/pkg/kafka"
	hKafka "github.com/vadimkiryanov/db-service/internal/pkg/kafka/handler"
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
var CONSUMER_GROUP = "my-consumer-group"

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

	h := handler.NewHandlersService() // Хэндлер для HTTP
	hk := hKafka.NewHandler()         // Хэндлер для Kafka
	sm := h.InitRouters(db)           // Мультиплексер
	srv := server.NewServerHTTPClient("9000", sm)

	c1, err := kafka.NewConsumer(hk, address, TOPIC, CONSUMER_GROUP, 1)
	if err != nil {
		logrus.Fatal(err)
	}

	c2, err := kafka.NewConsumer(hk, address, TOPIC, CONSUMER_GROUP, 2)
	if err != nil {
		logrus.Fatal(err)
	}

	c3, err := kafka.NewConsumer(hk, address, TOPIC, CONSUMER_GROUP, 3)
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		c1.Start()
	}()
	go func() {
		c2.Start()
	}()
	go func() {
		c3.Start()
	}()

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ошибка при запуске сервера: %s", err.Error())
		}
	}()

	fmt.Println("DB-service стартовал")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logrus.Fatal(c1.Stop(), c2.Stop(), c3.Stop()) // Так делать нельзя в продакшн

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
