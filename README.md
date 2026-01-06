# Tech Challenge - Fase 4

## Sistema de Recomendações de Investimentos

[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-blue.svg)](https://www.postgresql.org)
[![GCP](https://img.shields.io/badge/GCP-Cloud%20Run-blue.svg)](https://cloud.google.com/run)

Sistema de recomendações de investimentos desenvolvido em **Golang** com **PostgreSQL**, evoluído da PoC original em Node.js/Firebase. Deploy automatizado no **Google Cloud Platform** usando **Cloud Run** e **Cloud SQL**.

## 📋 Índice

- [Arquitetura GCP](#arquitetura-gcp)
- [Diferenças da Versão Node.js](#diferenças-da-versão-nodejs)
- [Pré-requisitos](#pré-requisitos)
- [Desenvolvimento Local](#desenvolvimento-local)
- [Deploy no GCP](#deploy-no-gcp)
- [API Endpoints](#api-endpoints)
- [Estrutura do Projeto](#estrutura-do-projeto)

## ☁️ Arquitetura GCP

```
┌─────────────────────────────────────────────────┐
│              GitHub Actions (CI/CD)             │
│  • Build Docker Image                           │
│  • Push to GCR                                  │
│  • Deploy to Cloud Run                          │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│           Cloud Run (Serverless)                │
│  • Container: Golang API                        │
│  • Auto-scaling                                 │
│  • HTTPS automático                             │
└────────────────┬────────────────────────────────┘
                 │ Unix Socket
┌────────────────▼────────────────────────────────┐
│        Cloud SQL (PostgreSQL 15)                │
│  • Managed database                             │
│  • Backups automáticos                          │
│  • Região: southamerica-east1 (São Paulo)       │
└─────────────────────────────────────────────────┘
```

### Componentes GCP

- **Cloud Run**: Hospeda a aplicação Golang em containers serverless
- **Cloud SQL**: PostgreSQL 15 gerenciado
- **Pub/Sub**: Sistema de mensageria para processamento assíncrono
- **Secret Manager**: Armazena credenciais sensíveis (senha do banco)
- **Container Registry (GCR)**: Armazena imagens Docker
- **Terraform**: Infraestrutura como código (IaC)

## 🔄 Diferenças da Versão Node.js

| Aspecto            | Node.js (Original)           | Golang (Nova Versão)              |
| ------------------ | ---------------------------- | --------------------------------- |
| **Runtime**        | Node.js 22                   | Go 1.21                           |
| **Framework**      | Firebase Functions + Express | Gin (HTTP Router)                 |
| **Banco de Dados** | Firestore (NoSQL)            | Cloud SQL PostgreSQL              |
| **Deploy**         | Firebase CLI                 | GitHub Actions + Terraform        |
| **Processamento**  | Pub/Sub Workers              | GCP Pub/Sub (Workers assíncronos) |
| **Infraestrutura** | Serverless (Firebase)        | Serverless (Cloud Run)            |
| **IaC**            | Nenhum                       | Terraform                         |

### Vantagens da Nova Versão

✅ **Performance**: Go compilado é mais rápido que Node.js interpretado  
✅ **Concorrência**: Goroutines nativas para processamento paralelo  
✅ **SQL**: PostgreSQL com queries otimizadas e transações ACID  
✅ **Pub/Sub**: Sistema de mensageria gerenciado para processamento assíncrono escalável  
✅ **IaC**: Terraform para versionamento de infraestrutura  
✅ **CI/CD**: Deploy automatizado via GitHub Actions  
✅ **Custos**: Cloud Run cobra apenas pelo uso real (pay-per-request)

## 🛠️ Pré-requisitos

### Para Desenvolvimento Local

- **Docker** e **Docker Compose** (para banco local)
- **Go 1.21+** (para compilar o código)

### Para Deploy no GCP

- **Conta GCP** com billing ativado
- **Projeto GCP** criado
- **GitHub Repository** com secrets configurados
- **Terraform** instalado (para provisionamento de infraestrutura)

## 💻 Desenvolvimento Local

### 1. Configurar Ambiente

```bash
# Copiar template de variáveis
cp .env.example .env

# Gerar dependências Go
cd backend
go mod tidy
```

### 2. Subir Banco de Dados Local

```bash
# Subir apenas o PostgreSQL
docker-compose up -d postgres

# Verificar se está rodando
docker-compose ps
```

### 3. Gerar CSVs de mocks

```bash
python3 scripts/gerar_mocks.py
```

### 4. Importar Dados CSV

```bash
cd scripts
go mod tidy
go run importar_csv.go
```

### 5. Executar Aplicação

```bash
cd backend
go run main.go
```

### 6. Gerar documentação

```bash
cd backend
swag init -g main.go -o docs
```

A API estará disponível em `http://localhost:8080`

## 🚀 Deploy no GCP

### 1. Configurar Secrets no GitHub

No seu repositório GitHub, vá em **Settings → Secrets and variables → Actions** e adicione:

| Secret               | Descrição               | Exemplo                            |
| -------------------- | ----------------------- | ---------------------------------- |
| `GOOGLE_CREDENTIALS` | JSON da service account | `{"type": "service_account", ...}` |
| `GCP_PROJECT`        | ID do projeto GCP       | `my-project-123456`                |
| `DB_PASSWORD`        | Senha do PostgreSQL     | `SenhaSegura123!`                  |

### 2. Provisionar Infraestrutura (Terraform)

```bash
# Ir para o diretório de infraestrutura
cd infra

# Inicializar Terraform
terraform init

# Ver o plano de execução
terraform plan \
  -var="gcp_project_id=SEU_PROJECT_ID" \
  -var="db_password=SUA_SENHA"

# Aplicar (criar recursos)
terraform apply \
  -var="gcp_project_id=SEU_PROJECT_ID" \
  -var="db_password=SUA_SENHA"
```

**Ou via GitHub Actions:**

1. Vá em **Actions** no GitHub
2. Execute o workflow **"Provisionar Infraestrutura (Terraform)"**
3. Escolha `apply` quando solicitado

### 3. Deploy Automático

O deploy é **automático** ao fazer push para `main` ou `dev`:

```bash
git add .
git commit -m "feat: nova funcionalidade"
git push origin main
```

O GitHub Actions irá:

1. ✅ Build da imagem Docker
2. ✅ Push para GCR
3. ✅ Deploy no Cloud Run
4. ✅ Verificar healthcheck

### 4. Acessar Aplicação

Após o deploy, a URL será exibida nos logs do GitHub Actions:

```
Service URL: https://app-recomendacao-prod-xxxxx-rj.a.run.app
```

Ou via CLI:

```bash
gcloud run services describe app-recomendacao-prod \
  --region=southamerica-east1 \
  --format='value(status.url)'
```

## 📡 API Endpoints

### Health Check

```bash
GET /healthcheck
```

**Exemplo:**

```bash
curl https://app-recomendacao-prod-xxxxx.a.run.app/healthcheck
```

**Resposta:**

```json
{
  "status": "OK",
  "servico": "api-recomendacoes-golang"
}
```

### Gerar Recomendações

```bash
POST /recomendacoes/:clienteId
```

**Exemplo:**

```bash
curl -X POST https://app-recomendacao-prod-xxxxx.a.run.app/recomendacoes/a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
```

**Resposta:**

```json
{
  "id_recomendacao": "550e8400-e29b-41d4-a716-446655440000",
  "id_cliente": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "recomendacoes": [
    {
      "produto": {
        "id_produto": "cfb30520-2253-46c5-a337-1d1148450123",
        "nome_produto": "CDB Banco XYZ",
        "risco_associado": "Baixo",
        "rentabilidade_12m": 12.5,
        "aplicacao_minima": 1000.0
      },
      "pontuacao": 0.75,
      "motivo": "[perfil compativel] [boa rentabilidade] [acessivel]"
    }
  ]
}
```

### Buscar Recomendações

```bash
GET /recomendacoes/:clienteId
```

**Nota:** Atualmente retorna 404. Implementação futura buscará do banco.

### Documentação Swagger

A documentação interativa da API está disponível em:

```bash
http://localhost:8080/swagger/index.html
```

## 📨 Sistema de Mensageria (Pub/Sub)

O sistema utiliza **Google Cloud Pub/Sub** para processamento assíncrono de recomendações em massa.

### Arquitetura

```
┌─────────────┐    Publica     ┌──────────────┐    Consome    ┌─────────────┐
│  API POST   │───────────────▶│  GCP Pub/Sub │──────────────▶│   Worker    │
│/recomendacoes│                │    Tópico    │               │ Recomendação│
└─────────────┘                └──────────────┘               └─────────────┘
                                                                      │
                                                                      ▼
                                                               ┌─────────────┐
                                                               │  PostgreSQL │
                                                               └─────────────┘
```

### Tópicos Disponíveis

- **`gerar-recomendacoes`**: Geração de recomendações para um cliente específico
- **`gerar-recomendacoes-massiva`**: Geração de recomendações para todos os clientes

### Funcionamento

1. **Publicação**: Quando uma requisição POST é feita, um evento é publicado no Pub/Sub
2. **Assinatura**: Workers escutam os tópicos e processam as mensagens de forma assíncrona
3. **Processamento**: O worker gera as recomendações e salva no banco de dados
4. **Confirmação**: Após processar, a mensagem é confirmada (ACK)

### Configuração Local

Para desenvolvimento local, você precisa configurar as credenciais do GCP:

```bash
# Autenticar com GCP
gcloud auth application-default login

# Ou usar uma Service Account
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

## 📁 Estrutura do Projeto

```
tech-challenge-fase4/
├── .github/workflows/         # CI/CD
│   ├── deploy.yml             # Deploy automático
│   ├── infra.yml              # Terraform
│   └── security.yml           # CodeQL scan
├── backend/                   # Aplicação Golang
│   ├── main.go                # Entry point
│   ├── interno/
│   │   ├── casodeuso/         # Lógica de negócio (Use Cases)
│   │   ├── controladores/     # Handlers HTTP (Controllers)
│   │   ├── dominio/           # Entidades e Interfaces (Domain)
│   │   └── infraestrutura/    # Implementações (Db, Logger, etc)
│   │       ├── repositorio/   # Acesso a dados
│   │       ├── pubsub/        # Event Bus (GCP Pub/Sub)
│   │       ├── worker/        # Workers assíncronos
│   │       └── logger/        # Logging
│   ├── Dockerfile             # Container da API
│   └── go.mod                 # Dependências
├── infra/
│   └── main.tf                # Terraform (GCP)
├── migrations/                # SQL migrations
├── scripts/                   # Utilitários
│   └── importar_csv.go        # Importar dados
├── docker-compose.yml         # Dev local
└── .env.example               # Config dev
```

## 🔧 Configuração de Variáveis

### Desenvolvimento Local (`.env`)

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=fiap
DB_PASSWORD=fiap123
DB_NAME=tech_challenge
API_PORT=8080
API_LEGADA_BASE_URL=http://localhost:8081

GCP_PROJECT_ID=seu-projeto-gcp
```

## 🐛 Troubleshooting

### Erro: "connection refused" (Local)

**Solução:**

```bash
# Verificar se PostgreSQL está rodando
docker-compose ps

# Ver logs
docker-compose logs postgres
```

### Erro: "permission denied" (GCP)

**Solução:** Verificar se a Service Account tem as permissões:

- `roles/cloudsql.client`
- `roles/secretmanager.secretAccessor`

### Deploy falha no GitHub Actions

**Solução:**

1. Verificar se os secrets estão configurados
2. Verificar logs do workflow
3. Testar Terraform localmente

### Cloud Run não conecta ao Cloud SQL

**Solução:** Verificar se a annotation está correta no Terraform:

```hcl
"run.googleapis.com/cloudsql-instances" = "PROJECT:REGION:INSTANCE"
```

## 📊 Monitoramento

### Logs

```bash
# Logs do Cloud Run
gcloud run services logs read app-recomendacao-prod \
  --region=southamerica-east1 \
  --limit=50

# Logs do Cloud SQL
gcloud sql operations list \
  --instance=tech-challenge-db-prod-br
```

### Métricas

Acesse o **Cloud Console → Cloud Run → app-recomendacao-prod** para ver:

- Requisições por segundo
- Latência
- Uso de memória/CPU
- Erros

## 🔐 Segurança

- ✅ **CodeQL** - Scan automático de vulnerabilidades
- ✅ **Dependabot** - Atualização automática de dependências
- ✅ **Secret Manager** - Credenciais nunca em código
- ✅ **HTTPS** - Automático no Cloud Run
- ✅ **IAM** - Permissões mínimas necessárias

## 📝 Próximos Passos

- [x] Implementar endpoint GET para buscar recomendações
- [ ] Adicionar testes unitários e de integração
- [ ] Implementar autenticação JWT
- [ ] Adicionar cache com Memorystore (Redis)
- [ ] Configurar alertas no Cloud Monitoring
- [x] Adicionar documentação OpenAPI/Swagger

## 👥 Contribuindo

Este é um projeto acadêmico da FIAP - Tech Challenge Fase 4.

## 📄 Licença

Este projeto é parte do curso de Pós-Graduação da FIAP.
