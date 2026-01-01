.PHONY: help build run-scraper run-server test clean install deps lint format

# Переменные
BINARY_NAME=etfscraper
CMD_PATH=./cmd/etfscraper
BIN_DIR=./bin
DB_PATH=etf_data.db
PORT=8080

# Помощь
help:
	@echo "ETF Scraper - Makefile команды"
	@echo ""
	@echo "Использование:"
	@echo "  make <команда>"
	@echo ""
	@echo "Команды:"
	@echo "  help          Показать эту справку"
	@echo "  deps          Установить зависимости"
	@echo "  build         Собрать бинарный файл"
	@echo "  run-scraper   Запустить скрейпинг"
	@echo "  run-server    Запустить веб-сервер"
	@echo "  test          Запустить тесты"
	@echo "  lint          Проверить код линтером"
	@echo "  format        Форматировать код"
	@echo "  clean         Удалить собранные файлы"
	@echo "  install       Установить бинарник в \$$GOPATH/bin"
	@echo ""
	@echo "Примеры:"
	@echo "  make deps             # Установить зависимости"
	@echo "  make build            # Собрать проект"
	@echo "  make run-server       # Запустить сервер"
	@echo "  DB_PATH=test.db make run-scraper  # Использовать другую БД"

# Установка зависимостей
deps:
	@echo "📦 Устанавливаю зависимости..."
	go mod download
	go mod tidy
	@echo "✓ Зависимости установлены"

# Сборка проекта
build: deps
	@echo "🔨 Собираю проект..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)/main.go
	@echo "✓ Бинарник создан: $(BIN_DIR)/$(BINARY_NAME)"

# Запуск скрейпера
run-scraper:
	@echo "🕷️  Запуск скрейпера..."
	go run $(CMD_PATH)/main.go scrape

# Запуск сервера
run-server:
	@echo "🚀 Запуск веб-сервера на порту $(PORT)..."
	@echo "📊 Откройте http://localhost:$(PORT) в браузере"
	SERVER_PORT=$(PORT) go run $(CMD_PATH)/main.go serve

# Запуск тестов
test:
	@echo "🧪 Запуск тестов..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "📊 Покрытие тестами:"
	go tool cover -func=coverage.out

# Запуск тестов с покрытием в HTML
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Отчет сохранен в coverage.html"

# Проверка кода линтером
lint:
	@echo "🔍 Проверка кода..."
	@which golangci-lint > /dev/null || (echo "❌ golangci-lint не установлен. Установите: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "✓ Проверка завершена"

# Форматирование кода
format:
	@echo "💅 Форматирование кода..."
	go fmt ./...
	gofmt -s -w .
	@echo "✓ Код отформатирован"

# Очистка
clean:
	@echo "🧹 Очистка..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	@echo "✓ Очистка завершена"

# Установка в систему
install: build
	@echo "📥 Устанавливаю в систему..."
	go install $(CMD_PATH)/main.go
	@echo "✓ Установлено в \$$GOPATH/bin/main"

# Быстрый запуск (скрейпинг + сервер)
quick: run-scraper run-server

# Показать версию Go
version:
	@echo "Go версия:"
	@go version

# Показать структуру проекта
tree:
	@echo "📁 Структура проекта:"
	@tree -I 'bin|.git|.idea|*.db' -L 3

# Docker команды (если будет добавлен Docker)
docker-build:
	@echo "🐳 Сборка Docker образа..."
	docker build -t etf-scraper:latest .

docker-run:
	@echo "🐳 Запуск в Docker..."
	docker run -p $(PORT):$(PORT) -v $$(pwd)/$(DB_PATH):/app/$(DB_PATH) etf-scraper:latest