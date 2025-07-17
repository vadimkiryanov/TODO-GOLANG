// api-service/main.go
package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/vadimkiryanov/api-service/internal/handlers"
	"github.com/vadimkiryanov/api-service/pkg/server"
)

func main() {
	err := initConfig()
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
	}

	// Мультиплексер
	hs := handlers.NewHandlersService()
	sm := hs.InitRouters()

	s := server.NewServerHTTPClient(viper.GetString("port"), sm)
	err = s.Run()
	if err != nil {
		fmt.Printf("\"Ошибка запуска сервера\": %v\n", err.Error())
	}

}

func initConfig() error {
	viper.AddConfigPath("internal/configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
