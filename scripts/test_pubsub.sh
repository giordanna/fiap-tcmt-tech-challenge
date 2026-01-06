#!/bin/bash

# Script para testar o GCP Pub/Sub localmente
# Este script cria tópicos e subscriptions de teste e envia mensagens

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configurações
PROJECT_ID="${GCP_PROJECT_ID:-}"
TOPIC_NAME="gerar-recomendacao"
SUBSCRIPTION_NAME="gerar-recomendacao-sub"

# Verifica se o PROJECT_ID está configurado
if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Erro: GCP_PROJECT_ID não está configurado${NC}"
    echo "Execute: export GCP_PROJECT_ID=seu-projeto-gcp"
    exit 1
fi

echo -e "${GREEN}=== Teste do GCP Pub/Sub ===${NC}"
echo "Project ID: $PROJECT_ID"
echo ""

# Verifica autenticação
echo -e "${YELLOW}1. Verificando autenticação...${NC}"
if ! gcloud auth application-default print-access-token &>/dev/null; then
    echo -e "${RED}Erro: Não autenticado no GCP${NC}"
    echo "Execute: gcloud auth application-default login"
    exit 1
fi
echo -e "${GREEN}✓ Autenticado${NC}"
echo ""

# Cria tópico se não existir
echo -e "${YELLOW}2. Criando tópico '$TOPIC_NAME'...${NC}"
if gcloud pubsub topics describe "$TOPIC_NAME" --project="$PROJECT_ID" &>/dev/null; then
    echo -e "${GREEN}✓ Tópico já existe${NC}"
else
    gcloud pubsub topics create "$TOPIC_NAME" --project="$PROJECT_ID"
    echo -e "${GREEN}✓ Tópico criado${NC}"
fi
echo ""

# Cria subscription se não existir
echo -e "${YELLOW}3. Criando subscription '$SUBSCRIPTION_NAME'...${NC}"
if gcloud pubsub subscriptions describe "$SUBSCRIPTION_NAME" --project="$PROJECT_ID" &>/dev/null; then
    echo -e "${GREEN}✓ Subscription já existe${NC}"
else
    gcloud pubsub subscriptions create "$SUBSCRIPTION_NAME" \
        --topic="$TOPIC_NAME" \
        --project="$PROJECT_ID" \
        --ack-deadline=60
    echo -e "${GREEN}✓ Subscription criada${NC}"
fi
echo ""

# Publica mensagem de teste
echo -e "${YELLOW}4. Publicando mensagem de teste...${NC}"
TEST_CLIENT_ID="a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
MESSAGE_DATA=$(echo -n "\"$TEST_CLIENT_ID\"" | base64)

gcloud pubsub topics publish "$TOPIC_NAME" \
    --message="$MESSAGE_DATA" \
    --project="$PROJECT_ID"

echo -e "${GREEN}✓ Mensagem publicada (Cliente ID: $TEST_CLIENT_ID)${NC}"
echo ""

# Lista mensagens pendentes
echo -e "${YELLOW}5. Verificando mensagens pendentes...${NC}"
PENDING=$(gcloud pubsub subscriptions describe "$SUBSCRIPTION_NAME" \
    --project="$PROJECT_ID" \
    --format="value(numUndeliveredMessages)" 2>/dev/null || echo "0")

echo -e "${GREEN}✓ Mensagens pendentes: $PENDING${NC}"
echo ""

# Consome uma mensagem (para teste)
echo -e "${YELLOW}6. Consumindo mensagem de teste...${NC}"
gcloud pubsub subscriptions pull "$SUBSCRIPTION_NAME" \
    --project="$PROJECT_ID" \
    --limit=1 \
    --auto-ack || echo -e "${YELLOW}Nenhuma mensagem disponível${NC}"
echo ""

# Estatísticas
echo -e "${YELLOW}7. Estatísticas do tópico:${NC}"
gcloud pubsub topics describe "$TOPIC_NAME" --project="$PROJECT_ID"
echo ""

echo -e "${GREEN}=== Teste concluído com sucesso! ===${NC}"
echo ""
echo "Próximos passos:"
echo "1. Execute a aplicação: cd backend && go run main.go"
echo "2. Teste o endpoint: curl -X POST http://localhost:8080/api/v2/recomendacoes/$TEST_CLIENT_ID"
echo "3. Monitore os logs: gcloud pubsub subscriptions pull $SUBSCRIPTION_NAME --project=$PROJECT_ID --limit=5"
echo ""
echo "Para limpar os recursos de teste:"
echo "  gcloud pubsub subscriptions delete $SUBSCRIPTION_NAME --project=$PROJECT_ID"
echo "  gcloud pubsub topics delete $TOPIC_NAME --project=$PROJECT_ID"
