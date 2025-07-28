package handler

import (
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

const logDirPermissions = 0755
const logFilePermissions = 0666

type Handler struct {
	kafkaLogger *logrus.Logger
}

func NewHandler() *Handler {
	logger := logrus.New()

	// Создаем директорию logs, если она не существует
	if err := os.MkdirAll("logs", logDirPermissions); err != nil {
		logrus.Fatalf("Не удалось создать директорию для логов: %v", err)
	}

	// Открываем лог-файл для записи сообщений в файл
	logFile, err := os.OpenFile("logs/kafka_events.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePermissions)
	if err != nil {
		logrus.Fatalf("Не удалось открыть лог-файл для Kafka: %v", err)
	}
	logger.SetOutput(logFile)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &Handler{
		kafkaLogger: logger,
	}
}

func (h *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, cn int) error {
	h.kafkaLogger.WithFields(logrus.Fields{
		"consumer":  cn,
		"offset":    topic.Offset,
		"partition": topic.Partition,
	}).Infof("Новое сообщение: %s", string(message))
	return nil
}
