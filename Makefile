run-api:
	go run ./cmd/todo_service/main.go

run-db:
	go run ./cmd/db_service/main.go

run-all:
	@echo "🚀 Запуск DB сервиса..."
	go run ./cmd/db_service/main.go &
	@sleep 2
	@echo "🚀 Запуск API сервиса..."
	go run ./cmd/todo_service/main.go
