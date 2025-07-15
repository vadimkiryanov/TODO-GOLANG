# Запуск api-service
run-api:
	cd api-service && go run ./cmd

# Запуск db-service
run-db:
	cd db-service && go run ./cmd

# Запустить оба сервиса параллельно
run-all:
	@echo "Запуск всех сервисов..."
	cd db-service && go run ./cmd & \
	cd api-service && go run ./cmd
