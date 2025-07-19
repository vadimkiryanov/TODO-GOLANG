# Загружаем переменные из .env и экспортируем их
ifneq (,$(wildcard db-service/.env))
    include db-service/.env
    export
endif

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

# Миграции
# Убедитесь, что у вас установлен golang-migrate
# go install -v github.com/golang-migrate/migrate/v4/cmd/migrate@latest
DB_URL="postgres://postgres:$(DB_PASSWORD)@localhost:5436/postgres?sslmode=disable"

migrate-create:
	@read -p "Введите имя миграции: " name; \
	migrate create -ext sql -dir db-service/migrations -seq $$name

migrate-up:
	migrate -path db-service/migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path db-service/migrations -database "$(DB_URL)" -verbose down

# Подключиться к psql в контейнере
db-connect:
	@read -p "Введите ID или name контейнера: " container_id; \
	docker exec -it $$container_id psql -U postgres
# "\dt" выведет список таблиц
# "Ctrl+D" или "\q" выйти из psql
