# Tech Challenge - Fase 4: Sistema de Recomendações de Investimentos

[![Go](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-blue.svg)](https://www.postgresql.org)
[![GCP](https://img.shields.io/badge/GCP-Cloud%20Run-blue.svg)](https://cloud.google.com/run)

Sistema de recomendações de investimentos desenvolvido em **Golang** com **PostgreSQL**. A aplicação utiliza arquitetura clean, processamento assíncrono via **GCP Pub/Sub** e deploy automatizado no **Google Cloud Platform (Cloud Run)**.

## 🚀 Como Iniciar (Desenvolvimento Local)

### Pré-requisitos

- **Go 1.21+**
- **Docker** e **Docker Compose**
- **Python 3** (para scripts de mock)

### Passo a Passo

1. **Configurar Variáveis de Ambiente**

   ```bash
   cp .env.example .env
   # Ajuste as variáveis no arquivo .env conforme necessário (DB, GCP Project, etc)
   ```

2. **Subir Banco de Dados**

   ```bash
   docker-compose up -d postgres
   ```

3. **Popular Banco de Dados (Mocks)**

   ```bash
   # Gerar arquivos CSV de exemplo
   python3 scripts/gerar_mocks.py

   # Importar dados para o banco
   cd scripts
   go mod tidy
   go run importar_csv.go
   cd ..
   ```

4. **Executar a Aplicação**

   ```bash
   cd backend
   go mod tidy
   go run main.go
   ```

   A API estará disponível em: `http://localhost:8080`

### 📚 Documentação da API (Swagger)

Após iniciar a aplicação, acesse a documentação interativa:

- `http://localhost:8080/api/v2/swagger/index.html`

## ☁️ Infraestrutura e Deploy

O projeto utiliza **Terraform** para IaC e **GitHub Actions** para CI/CD.

- **Infraestrutura**: Diretório `infra/`. Use `terraform apply` para provisionar recursos no GCP (Cloud SQL, Pub/Sub, Cloud Run).
- **Deploy Automático**: Push na branch `main` dispara o pipeline de deploy via GitHub Actions.

## 📁 Estrutura do Projeto

- `backend/`: Código fonte da aplicação API (Clean Architecture).
- `infra/`: Scripts Terraform para provisionamento GCP.
- `scripts/`: Scripts auxiliares para carga de dados e testes.
- `.github/workflows/`: Pipelines de CI/CD.

---

_Projeto acadêmico da FIAP - Pós Tech._
