import csv
import uuid
import random
from datetime import datetime, timedelta

def gerar_uuid():
    return str(uuid.uuid4())

def obter_data_aleatoria():
    data_inicio = datetime.now() - timedelta(days=365)
    dias_aleatorios = random.randint(0, 365)
    return (data_inicio + timedelta(days=dias_aleatorios)).strftime("%Y-%m-%d %H:%M:%S")

# Estruturas de dados para integridade referencial
ids_clientes = []
ids_produtos = []

# 1. Gerar Clientes
dados_clientes = []
objetivos_possiveis = [
    'Aposentadoria', 
    'Reserva de Emergência', 
    'Compra de Imóvel', 
    'Viagem', 
    'Educação', 
    'Crescimento de Capital',
    'Independência Financeira'
]

for i in range(50): # Aumentado para 50 para ter mais variedade
    id_c = gerar_uuid()
    ids_clientes.append(id_c)
    dados_clientes.append([
        id_c,
        random.uniform(10000, 2000000), # patrimonio
        random.choice(['Conservador', 'Moderado', 'Arrojado']), # perfil_risco
        random.choice(objetivos_possiveis) # objetivo variado
    ])

with open('scripts/datalake/clientes.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerow(['id_cliente', 'patrimonio_total_estimado', 'perfil_risco', 'objetivo_investimento'])
    writer.writerows(dados_clientes)

# 2. Gerar Produtos
dados_produtos = []
tipos_produtos = ['Fundo de Renda Fixa', 'Fundo Multimercado', 'Fundo de Ações', 'CDB', 'LCI/LCA', 'Previdência']
liquidez_opcoes = ['D+0', 'D+1', 'D+30', 'D+60']

for i in range(20): # Aumentado para 20 produtos
    id_p = gerar_uuid()
    ids_produtos.append(id_p)
    tipo = random.choice(tipos_produtos)
    
    # Lógica simples para coerência de risco
    if tipo in ['Fundo de Ações', 'Fundo Multimercado']:
        risco = random.choice(['Médio', 'Alto'])
        rentab_12m = random.uniform(10, 30)
    elif tipo in ['CDB', 'LCI/LCA', 'Fundo de Renda Fixa']:
        risco = 'Baixo'
        rentab_12m = random.uniform(8, 13)
    else:
        risco = random.choice(['Baixo', 'Médio'])
        rentab_12m = random.uniform(9, 15)

    dados_produtos.append([
        id_p,
        f"{tipo} Top {i+1}",
        tipo,
        risco,
        rentab_12m, # rentabilidade 12m
        rentab_12m * random.uniform(2.5, 3.2), # rentabilidade 36m (aprox)
        random.uniform(0.5, 2.0), # taxa adm
        random.choice([100, 500, 1000, 5000, 10000]), # aplicacao minima
        random.choice(liquidez_opcoes), # liquidez
        "Ativo" # status
    ])

with open('scripts/datalake/produtos.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerow(['id_produto', 'nome_produto', 'tipo_produto', 'risco_associado', 'rentabilidade_historica_12m', 'rentabilidade_historica_36m', 'taxa_administracao', 'aplicacao_minima', 'liquidez', 'status_produto'])
    writer.writerows(dados_produtos)

# 3. Gerar Transações
dados_transacoes = []
for i in range(100): # Mais transações
    id_t = gerar_uuid()
    dados_transacoes.append([
        id_t,
        random.choice(ids_clientes),
        random.choice(ids_produtos),
        random.choice(['Aplicacao', 'Resgate']),
        random.uniform(100, 50000),
        obter_data_aleatoria(),
        "Concluida"
    ])

with open('scripts/datalake/transacoes.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerow(['id_transacao', 'id_cliente', 'id_produto', 'tipo_transacao', 'valor_transacao', 'data_transacao', 'status_transacao'])
    writer.writerows(dados_transacoes)

# 4. Gerar Interações
dados_interacoes = []
for i in range(150): # Mais interações
    id_i = gerar_uuid()
    dados_interacoes.append([
        id_i,
        random.choice(ids_clientes),
        random.choice(ids_produtos),
        random.choice(['Visualizacao', 'Clique', 'Favorito', 'Simulacao']),
        obter_data_aleatoria(),
        random.randint(5, 600) # duração em segundos
    ])

with open('scripts/datalake/interacoes.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerow(['id_interacao', 'id_cliente', 'id_produto', 'tipo_interacao', 'data_interacao', 'duracao_interacao_segundos'])
    writer.writerows(dados_interacoes)

print("CSVs gerados com sucesso em scripts/datalake!")
