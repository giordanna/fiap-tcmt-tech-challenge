-- habilita a extensão para gerar uuids automaticamente no banco
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- tabela de clientes (mantém id string para compatibilidade com legado/csv)
CREATE TABLE IF NOT EXISTS clientes (
    id_cliente VARCHAR(50) PRIMARY KEY,
    nome_cliente VARCHAR(255) NOT NULL,
    perfil_risco VARCHAR(50) NOT NULL, -- Conservador, Moderado, Arrojado
    patrimonio_total_estimado DECIMAL(15, 2),
    data_cadastro TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- tabela de produtos (mantém id string para compatibilidade)
CREATE TABLE IF NOT EXISTS produtos (
    id_produto VARCHAR(50) PRIMARY KEY,
    nome_produto VARCHAR(255) NOT NULL,
    risco_associado VARCHAR(20), -- Baixo, Médio, Alto
    rentabilidade_12m DECIMAL(5, 2),
    aplicacao_minima DECIMAL(15, 2),
    status_produto VARCHAR(20) DEFAULT 'Ativo'
);

-- tabela de transacoes (histórico para regra de diversificação)
CREATE TABLE IF NOT EXISTS transacoes (
    id_transacao VARCHAR(50) PRIMARY KEY,
    id_cliente VARCHAR(50) REFERENCES clientes(id_cliente),
    id_produto VARCHAR(50) REFERENCES produtos(id_produto),
    tipo_transacao VARCHAR(20), -- Aplicacao, Resgate
    valor_transacao DECIMAL(15, 2),
    data_transacao TIMESTAMP
);

-- tabela de interacoes (histórico para regra de interesse)
CREATE TABLE IF NOT EXISTS interacoes (
    id_interacao VARCHAR(50) PRIMARY KEY,
    id_cliente VARCHAR(50) REFERENCES clientes(id_cliente),
    id_produto VARCHAR(50) REFERENCES produtos(id_produto),
    tipo_interacao VARCHAR(50),
    data_interacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- tabela de recomendacoes (nova arquitetura usa uuid)
CREATE TABLE IF NOT EXISTS recomendacoes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_cliente VARCHAR(50) REFERENCES clientes(id_cliente),
    produtos_json JSONB NOT NULL, -- snapshot da recomendação gerada
    data_geracao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);