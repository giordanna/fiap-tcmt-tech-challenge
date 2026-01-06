# Migração: InMemory Pub/Sub → GCP Pub/Sub

## Resumo

Este documento descreve a migração do sistema de mensageria in-memory para o Google Cloud Pub/Sub.

## Mudanças Realizadas

### 1. Dependências

Adicionada a biblioteca oficial do GCP Pub/Sub:

```bash
go get cloud.google.com/go/pubsub@latest
```

### 2. Nova Implementação

Criado o arquivo `backend/interno/infraestrutura/pubsub/gcp.go` com a implementação `GCPEventBus` que:

- ✅ Implementa a interface `EventBus` existente
- ✅ Cria tópicos automaticamente se não existirem
- ✅ Cria subscriptions automaticamente
- ✅ Serializa/deserializa payloads em JSON
- ✅ Suporta múltiplos handlers por tópico
- ✅ Processa mensagens de forma assíncrona
- ✅ Implementa ACK/NACK para controle de mensagens

### 3. Atualização do main.go

Substituída a inicialização:

**Antes:**

```go
bus := pubsub.NovoInMemoryEventBus()
```

**Depois:**

```go
gcpProjectID := getEnv("GCP_PROJECT_ID", "")
if gcpProjectID == "" {
    slog.Error("GCP_PROJECT_ID não configurado")
    os.Exit(1)
}

ctx := context.Background()
bus, err := pubsub.NovoGCPEventBus(ctx, gcpProjectID)
if err != nil {
    slog.Error("Erro ao inicializar GCP Pub/Sub", "erro", err)
    os.Exit(1)
}
defer bus.Close()
```

## Configuração

### Variáveis de Ambiente

Adicione ao seu `.env`:

```bash
GCP_PROJECT_ID=seu-projeto-gcp
```

### Autenticação Local

Para desenvolvimento local, configure as credenciais do GCP:

**Opção 1: Application Default Credentials (Recomendado)**

```bash
gcloud auth application-default login
```

**Opção 2: Service Account**

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

### Permissões Necessárias

A service account precisa das seguintes permissões:

- `roles/pubsub.publisher` - Para publicar mensagens
- `roles/pubsub.subscriber` - Para consumir mensagens
- `roles/pubsub.admin` - Para criar tópicos e subscriptions (apenas em dev)

Em produção, é recomendado criar os tópicos e subscriptions via Terraform e usar apenas as permissões de publisher/subscriber.

## Tópicos Utilizados

| Tópico                        | Descrição                                 | Payload              |
| ----------------------------- | ----------------------------------------- | -------------------- |
| `gerar-recomendacao`          | Gera recomendação para um cliente         | `string` (clienteID) |
| `gerar-recomendacoes-massiva` | Gera recomendações para todos os clientes | `string` (clienteID) |

## Compatibilidade

A implementação GCP mantém **100% de compatibilidade** com o código existente:

- ✅ Mesma interface `EventBus`
- ✅ Mesmos métodos `Publicar()` e `Assinar()`
- ✅ Mesma assinatura de handlers
- ✅ Nenhuma mudança necessária nos workers ou casos de uso

## Vantagens da Migração

### Escalabilidade

- ✅ Mensagens persistidas (não se perdem em caso de restart)
- ✅ Processamento distribuído entre múltiplas instâncias
- ✅ Auto-scaling baseado em carga

### Confiabilidade

- ✅ Garantia de entrega (at-least-once)
- ✅ Dead Letter Queue para mensagens com falha
- ✅ Retry automático

### Observabilidade

- ✅ Métricas no Cloud Monitoring
- ✅ Logs estruturados
- ✅ Rastreamento de mensagens

### Operacional

- ✅ Serviço gerenciado (sem manutenção)
- ✅ Backups automáticos
- ✅ SLA de 99.95%

## Rollback

Se necessário reverter para a versão in-memory, basta alterar o `main.go`:

```go
// Voltar para in-memory
bus := pubsub.NovoInMemoryEventBus()
```

A implementação in-memory permanece no código para compatibilidade.

## Testes

### Teste Local

1. Configure as credenciais:

```bash
gcloud auth application-default login
```

2. Configure o `.env`:

```bash
GCP_PROJECT_ID=seu-projeto-dev
```

3. Execute a aplicação:

```bash
cd backend
go run main.go
```

4. Teste o endpoint:

```bash
# Gerar recomendação para um cliente
curl -X POST http://localhost:8080/api/v2/recomendacoes/CLIENTE_ID

# Gerar recomendações em massa
curl -X POST http://localhost:8080/api/v2/recomendacoes
```

5. Verifique os tópicos criados:

```bash
gcloud pubsub topics list
gcloud pubsub subscriptions list
```

### Monitoramento

Visualize as mensagens no console do GCP:

```
https://console.cloud.google.com/cloudpubsub/topic/list
```

## Próximos Passos

- [ ] Adicionar Dead Letter Queue para mensagens com falha
- [ ] Configurar alertas no Cloud Monitoring
- [ ] Implementar retry exponencial com backoff
- [ ] Adicionar métricas customizadas
- [ ] Configurar Terraform para criar tópicos/subscriptions
- [ ] Implementar ordenação de mensagens (se necessário)

## Referências

- [GCP Pub/Sub Documentation](https://cloud.google.com/pubsub/docs)
- [Go Client Library](https://pkg.go.dev/cloud.google.com/go/pubsub)
- [Best Practices](https://cloud.google.com/pubsub/docs/publisher)
