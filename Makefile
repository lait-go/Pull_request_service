.PHONY: help build run test test-integration test-bash test-bash-comprehensive clean migrate-up migrate-down docker-up docker-down lint fmt vet deps install

# Переменные
BINARY_NAME=app
MAIN_PATH=./cmd/app
TEST_DB_SOURCE?=postgres://lait:123@localhost:5432/orders_db?sslmode=disable
API_URL?=http://localhost:8080
MIGRATIONS_PATH=./internal/adapters/postgres/migrations

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

help: ## Показать справку по командам
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-30s$(NC) %s\n", $$1, $$2}'

install: deps ## Установить зависимости и собрать проект
	@echo "$(GREEN)Installing...$(NC)"
	go install $(MAIN_PATH)

build: ## Собрать приложение
	@echo "$(GREEN)Building application...$(NC)"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run: ## Запустить приложение
	@echo "$(GREEN)Running application...$(NC)"
	go run $(MAIN_PATH)

migrate-up: ## Применить миграции
	@echo "$(GREEN)Running migrations...$(NC)"
	@if [ -z "$(DB_SOURCE)" ]; then \
		echo "$(YELLOW)Using default DB_SOURCE: $(TEST_DB_SOURCE)$(NC)"; \
		DB_SOURCE=$(TEST_DB_SOURCE) go run ./cmd/app/main.go migrate || true; \
	else \
		DB_SOURCE=$(DB_SOURCE) go run ./cmd/app/main.go migrate || true; \
	fi

migrate-down: ## Откатить миграции (требует migrate CLI)
	@echo "$(YELLOW)Rolling back migrations...$(NC)"
	@echo "$(YELLOW)Note: This requires migrate CLI tool$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(TEST_DB_SOURCE)" down || echo "$(YELLOW)Migrate CLI not found$(NC)"

test: ## Запустить unit тесты
	@echo "$(GREEN)Running integration tests...$(NC)"
	@echo "$(YELLOW)Using TEST_DB_SOURCE: $(TEST_DB_SOURCE)$(NC)"
	TEST_DB_SOURCE=$(TEST_DB_SOURCE) go test -v -short=false ./internal/controlers/http/... -run TestIntegration_AllHandlers


lint: ## Запустить линтер
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run ./... || echo "$(YELLOW)golangci-lint not installed$(NC)"

fmt: ## Форматировать код
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...

vet: ## Запустить go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet ./...

clean: ## Очистить скомпилированные файлы
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf bin/
	go clean ./...

setup: docker-up migrate-up ## Полная настройка (Docker + миграции)
	@echo "$(GREEN)Setup complete!$(NC)"

# Комбинированные команды
check: fmt vet test ## Проверить код (fmt + vet + test)

ci: deps fmt vet test test-integration ## CI pipeline (deps + fmt + vet + test + integration)