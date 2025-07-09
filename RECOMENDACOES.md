# Sistema de Recomendações baseado em Datalake

## Visão Geral

Este sistema foi desenvolvido para gerar recomendações personalizadas para usuários baseado em dados estruturados de um datalake. Ele analisa:

- **Dados de Clientes**: Perfil, renda, patrimônio, objetivos
- **Produtos Disponíveis**: CDBs, Fundos, Previdência, etc.
- **Transações Históricas**: Aplicações e resgates
- **Interações**: Pesquisas, visualizações, cliques
- **Dados de Mercado**: Ibovespa, Selic, Dólar

## Arquitetura

### 1. Armazenamento de Dados
- **Firebase Storage**: Armazena os arquivos CSV do datalake na pasta `datalake/`
- **Firestore**: Armazena as recomendações geradas

### 2. Processamento
- **Cloud Functions**: Worker que processa os dados e gera recomendações
- **Pub/Sub**: Dispara o processamento de forma assíncrona

## Estrutura dos Dados

### Clientes (clientes.csv)
```csv
id_cliente,nome_cliente,data_cadastro,idade,genero,renda_mensal_estimada,patrimonio_total_estimado,perfil_risco,objetivo_investimento,ultima_interacao
```

### Produtos (produtos.csv)
```csv
id_produto,nome_produto,tipo_produto,risco_associado,rentabilidade_historica_12m,rentabilidade_historica_36m,taxa_administracao,liquidez,aplicacao_minima,indexador,setor_economia,estrategia_investimento,data_lancamento,status_produto
```

### Transações (transacoes.csv)
```csv
id_transacao,id_cliente,id_produto,tipo_transacao,valor_transacao,data_transacao,status_transacao
```

### Interações (interacoes.csv)
```csv
id_interacao,id_cliente,tipo_interacao,id_produto,data_interacao,duracao_interacao_segundos,termo_pesquisa
```

### Dados de Mercado (dados_mercado.csv)
```csv
data,nome_indice,valor_indice,taxa_selic,cotacao_dolar
```

## Algoritmo de Recomendação

O sistema utiliza um algoritmo de scoring que considera:

### 1. Compatibilidade de Perfil de Risco
- **Conservador** → Produtos de **Baixo Risco** (+0.3 pontos)
- **Moderado** → Produtos de **Médio Risco** (+0.25 pontos)
- **Arrojado** → Produtos de **Alto Risco** (+0.2 pontos)

### 2. Rentabilidade Histórica
- Produtos com rentabilidade > 10% nos últimos 12 meses (+0.1 pontos)

### 3. Acessibilidade Financeira
- Aplicação mínima < 5% do patrimônio total (+0.1 pontos)

### 4. Diversificação
- Penaliza produtos que o cliente já possui muito (-0.2 pontos)

### 5. Interesse Demonstrado
- Bonus para produtos com interações do cliente (+0.15 pontos)

### 6. Adequação ao Objetivo
- **Reserva de Emergência** → Produtos com liquidez D+0 (+0.2 pontos)
- **Aposentadoria** → Previdência Privada (+0.25 pontos)

## Limitações Atuais

- Processamento batch (não tempo real)
- Algoritmo de scoring simples
- Não considera sazonalidade
- Não usa dados históricos de mercado na recomendação
