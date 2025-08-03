package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type ConfigRedis struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db"`           // идентификатор базы данных
	MaxRetries  int           `yaml:"max_retries"`  // максимальное количество попыток подключения
	DialTimeout time.Duration `yaml:"dial_timeout"` // таймаут для установления новых соединений
	Timeout     time.Duration `yaml:"timeout"`      // таймаут для записи и чтения
}

func NewRedisDB(ctx context.Context, cfg ConfigRedis) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		logrus.Fatalf("ошибка подлкючения к redis-db: %s\n", err.Error())
		return nil, err
	}
	logrus.Info("успешное подключение к redis-db")

	return db, nil
}

// моковые данные
// docker run --name redis-db -p 6379:6379 -d redis:7.0-alpine redis-server --requirepass test1234 --user testuser on '~*' +@all
