// db-service/main.go
package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // драйвер для pg
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"github.com/vadimkiryanov/db-service/internal/handler"
	"github.com/vadimkiryanov/db-service/internal/pkg/server"
	"github.com/vadimkiryanov/db-service/internal/repository"
)

func init() {
	err := initConfig()
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
	}

	// Инициализация env
	if err := gotenv.Load(); err != nil {
		log.Fatalf("err %v", err.Error())
	}
}

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

	h := handler.NewHandlersService()
	s := server.NewServerHTTPClient("9000", h.InitRouters(db))

	if err := s.Run(); err != nil {
		fmt.Printf("\"Ошибка запуска сервера\": %v\n", err.Error())
		return
	}
}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
