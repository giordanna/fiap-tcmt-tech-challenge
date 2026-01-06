.PHONY: help setup up down logs migrate seed test clean

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Configura o ambiente (cria .env e gera go.sum)
	@echo "Configurando ambiente..."
	@if [ ! -f .env ]; then cp .env.example .env; echo ".env criado"; fi
	@cd backend && go mod tidy && echo "Dependências Go atualizadas"

up: ## Sobe os containers (docker-compose up)
	@echo "Subindo containers..."
	docker-compose up -d
	@echo "Aguardando banco de dados ficar pronto..."
	@sleep 5
	@echo "Containers iniciados!"

down: ## Para os containers (docker-compose down)
	@echo "Parando containers..."
	docker-compose down

logs: ## Mostra logs dos containers
	docker-compose logs -f

migrate: ## Executa as migrações do banco de dados
	@echo "Executando migrações..."
	docker-compose exec postgres psql -U fiap -d tech_challenge -f /docker-entrypoint-initdb.d/001_schema_inicial.up.sql
	@echo "Migrações executadas!"

seed: ## Importa dados CSV para o banco (requer script de import)
	@echo "Importando dados CSV..."
	@if [ -f scripts/import_csv.go ]; then \
		cd scripts && go run import_csv.go; \
	else \
		echo "Script de importação não encontrado. Execute manualmente."; \
	fi

test: ## Executa testes
	@echo "Executando testes..."
	cd backend && go test ./...

clean: ## Remove containers, volumes e imagens
	@echo "Limpando ambiente..."
	docker-compose down -v
	@echo "Ambiente limpo!"

build: ## Compila a aplicação Go
	@echo "Compilando aplicação..."
	cd backend && go build -o ../bin/api-gateway ./cmd/main.go
	@echo "Compilação concluída! Binário em bin/api-gateway"
