include .env

.PHONY: help setup dev dev-up dev-down build test clean docker docker-down install-deps install-frontend dev-frontend build-frontend test-frontend run

BINARY_NAME=stats-central-api
DOCKER_COMPOSE_FILE=docker-compose.yml
DOCKER_COMPOSE_DEV_FILE=docker-compose.dev.yml

MIGRATE_CMD=migrate
MIGRATE_PATH=migrations

ifeq ($(DB_DRIVER),sqlite)
DB_URL=sqlite3://$(DB_NAME)
else
DB_URL=mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?multiStatements=true
endif

# Instalar dependências
install-deps: ## Instala as dependências do Go
	@echo "📦 Instalando dependências..."
	@go mod download
	@go mod tidy

install-frontend: ## Instala as dependências do frontend (Node.js)
	@echo "📦 Instalando dependências do frontend..."
	@npm install

dev-frontend: ## Inicia o servidor de desenvolvimento do frontend (Vite)
	@echo "🔥 Iniciando frontend com hot reload..."
	@echo "📍 Frontend rodará em: http://localhost:5173 (proxy /api -> localhost:3003)"
	@npm run dev

build-frontend: ## Compila o frontend para internal/web/dist (embutido no binário Go)
	@echo "🔨 Compilando frontend..."
	@npm run build
	@echo "✅ Frontend compilado em internal/web/dist"

test-frontend: ## Executa os testes do frontend (Vitest)
	@echo "🧪 Executando testes do frontend..."
	@npm test

# Instalar ferramentas de desenvolvimento
install-tools:
	@echo "🛠️  Instalando ferramentas de desenvolvimento..."
	@go install github.com/cespare/reflex@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Ferramentas instaladas"

# Desenvolvimento
dev-up: ## Sobe os serviços de desenvolvimento (PostgreSQL e Redis)
	@echo "🐳 Subindo serviços de desenvolvimento..."
	@docker compose up postgres keycloak -d
	@echo "✅ Serviços rodando:"

dev-down: ## Para os serviços de desenvolvimento
	@echo "🛑 Parando serviços de desenvolvimento..."
	@docker compose down postgres keycloak

# Build
build: ## Compila a aplicação
	@echo "🔨 Compilando aplicação..."
	@go build -ldflags='-s -w' -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "✅ Binário criado: bin/$(BINARY_NAME)"

# Executar aplicação compilada
run: build ## Executa a aplicação compilada
	@echo "🚀 Executando aplicação..."
	@./bin/$(BINARY_NAME)

run-dev: ## Inicia o servidor com hot reload
	@echo "🔥 Iniciando servidor com hot reload..."
	@echo "📍 Servidor rodará em: http://localhost:$(SERVER_PORT)"
	@echo "🔄 Arquivos monitorados para reload automático"
	@echo "Press Ctrl+C to stop"
	@reflex -c reflex.conf

# Testes
test: ## Executa os testes
	@echo "🧪 Executando testes..."
	@go test ./...

test-e2e: ## Executa os testes de integração (E2E)
	@echo "🧪 Executando testes E2E..."
	@go test -tags=integration -v ./test/e2e/...

test-v: ## Executa os testes
	@echo "🧪 Executando testes..."
	@go test -v ./...

test-cover: ## Executa testes com cobertura
	@echo "🧪 Executando testes com cobertura..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Relatório de cobertura gerado: coverage.html"

# Linting
lint: ## Executa o linter
	@echo "🔍 Executando linter..."
	@golangci-lint run

# Limpeza
clean: ## Remove arquivos temporários e binários
	@echo "🧹 Limpando arquivos temporários..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@docker system prune -f
	@echo "✅ Limpeza concluída"

# Database
migrate-create:
	@echo "🔨 Criando migração..."
	@$(MIGRATE_CMD) create -dir=$(MIGRATE_PATH) -ext=sql -seq $(NAME)

migrate:
	@echo "📊 Executando migrações..."
	@$(MIGRATE_CMD) -path $(MIGRATE_PATH) -database "$(DB_URL)" up
	@echo "✅ Migração concluída"

rollback:
	@echo "⏪ Executando rollback das migrações..."
	@$(MIGRATE_CMD) -path $(MIGRATE_PATH) -database "$(DB_URL)" down 1
	@echo "✅ Rollback concluído"

seed-stock:
	@echo "🌱 Executando seed de stocks..."
	@go run cmd/seed/main.go $(ARGS)
	@echo "✅ Seed de stocks concluído"