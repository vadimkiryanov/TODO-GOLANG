package repository

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg *Config) (*sqlx.DB, error) {
	// Инициализация бд
	db, err := sqlx.Connect("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.DBName,
		cfg.Password,
		cfg.SSLMode),
	)

	if err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
		return nil, err
	}
	err = db.Ping()

	if err != nil {
		log.Fatalf("Ошибка: %v", err.Error())
		return nil, err
	}

	fmt.Printf("БД подключена на %s:%s\n", viper.GetString("db.host"), viper.GetString("db.port"))

	return db, nil
}
