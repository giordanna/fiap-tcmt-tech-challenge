-- habilita a extensão para gerar uuids automaticamente no banco
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- tabela de clientes (LGPD: Apenas dados essenciais para o MVP de recomendação)
CREATE TABLE IF NOT EXISTS clientes (
    id_cliente UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    perfil_risco VARCHAR(50) NOT NULL, -- Conservador, Moderado, Arrojado
    patrimonio_total_estimado DECIMAL(15, 2),
    objetivo_investimento VARCHAR(100)
);

-- tabela de produtos
CREATE TABLE IF NOT EXISTS produtos (
    id_produto UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome_produto VARCHAR(255) NOT NULL,
    tipo_produto VARCHAR(50),
    risco_associado VARCHAR(20), -- Baixo, Médio, Alto
    rentabilidade_historica_12m DECIMAL(5, 2),
    rentabilidade_historica_36m DECIMAL(5, 2),
    taxa_administracao DECIMAL(5, 2),
    aplicacao_minima DECIMAL(15, 2),
    liquidez VARCHAR(50),
    status_produto VARCHAR(20) DEFAULT 'Ativo'
);

-- tabela de transacoes
CREATE TABLE IF NOT EXISTS transacoes (
    id_transacao UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_cliente UUID REFERENCES clientes(id_cliente),
    id_produto UUID REFERENCES produtos(id_produto),
    tipo_transacao VARCHAR(20), -- Aplicacao, Resgate
    valor_transacao DECIMAL(15, 2),
    data_transacao TIMESTAMP,
    status_transacao VARCHAR(50)
);

-- tabela de interacoes
CREATE TABLE IF NOT EXISTS interacoes (
    id_interacao UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_cliente UUID REFERENCES clientes(id_cliente),
    id_produto UUID REFERENCES produtos(id_produto),
    tipo_interacao VARCHAR(50),
    data_interacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duracao_interacao_segundos INT
);

-- tabela de recomendacoes
CREATE TABLE IF NOT EXISTS recomendacoes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_cliente UUID REFERENCES clientes(id_cliente),
    produtos_json JSONB NOT NULL, -- snapshot da recomendação gerada
    data_geracao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);