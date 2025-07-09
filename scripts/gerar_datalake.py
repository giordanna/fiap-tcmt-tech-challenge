import pandas as pd
from faker import Faker
import random
from datetime import datetime, timedelta
import os # Importa o módulo os para interações com o sistema de arquivos

fake = Faker('pt_BR') # Para dados em português do Brasil

# --- Cria a pasta 'datalake_poc' se ela não existir ---
output_folder = 'datalake_poc'
if not os.path.exists(output_folder):
    os.makedirs(output_folder)
    print(f"Pasta '{output_folder}' criada com sucesso.")
else:
    print(f"Pasta '{output_folder}' já existe.")
# --- Fim da verificação de pasta ---

num_clientes = 1000
num_produtos = 50
num_transacoes = 5000
num_interacoes = 10000

# 1. Gerar Dados de Clientes
clientes_data = []
perfil_risco_opcoes = ["Conservador", "Moderado", "Arrojado"]
objetivo_investimento_opcoes = ["Aposentadoria", "Comprar Imóvel", "Reserva de Emergência", "Crescimento de Patrimônio", "Educação Filhos"]

for i in range(num_clientes):
    data_cadastro = fake.date_time_between(start_date='-5y', end_date='now')
    clientes_data.append({
        'id_cliente': f'CLI{i:05d}',
        'nome_cliente': fake.name(),
        'data_cadastro': data_cadastro.strftime('%Y-%m-%d %H:%M:%S'),
        'idade': random.randint(20, 70),
        'genero': random.choice(['Masculino', 'Feminino', 'Outro', 'Não Informado']),
        'renda_mensal_estimada': round(random.uniform(2000, 50000), 2),
        'patrimonio_total_estimado': round(random.uniform(5000, 5000000), 2),
        'perfil_risco': random.choice(perfil_risco_opcoes),
        'objetivo_investimento': random.choice(objetivo_investimento_opcoes),
        'ultima_interacao': fake.date_time_between(start_date=data_cadastro, end_date='now').strftime('%Y-%m-%d %H:%M:%S')
    })
df_clientes = pd.DataFrame(clientes_data)
print("Clientes gerados.")

# 2. Gerar Dados de Produtos
produtos_data = []
tipo_produto_opcoes = ["Fundo de Ações", "Fundo Multimercado", "Fundo Renda Fixa", "CDB", "LCI/LCA", "Previdência Privada"]
risco_produto_opcoes = ["Baixo", "Médio", "Alto"]
indexador_opcoes = ["CDI", "IPCA", "Ibovespa"]
setor_economia_opcoes = ["Tecnologia", "Bancos", "Energia", "Consumo", "Saúde", "Imobiliário", "Agronegócio", "Diversos"]
estrategia_investimento_opcoes = ["Long Only", "Macro", "Quantitativo", "Valor", "Crescimento"]

for i in range(num_produtos):
    tipo_prod = random.choice(tipo_produto_opcoes)
    risco_prod = random.choice(risco_produto_opcoes)
    
    rent_12m = round(random.uniform(-0.05, 0.30), 4) # -5% a 30%
    rent_36m = round(random.uniform(-0.10, 0.80), 4) # -10% a 80%

    produtos_data.append({
        'id_produto': f'PROD{i:03d}',
        'nome_produto': fake.word().capitalize() + " " + fake.word().capitalize() + " " + tipo_prod,
        'tipo_produto': tipo_prod,
        'risco_associado': risco_prod,
        'rentabilidade_historica_12m': rent_12m,
        'rentabilidade_historica_36m': rent_36m,
        'taxa_administracao': round(random.uniform(0.1, 2.5), 2),
        'liquidez': random.choice(['D+0', 'D+1', 'D+2', 'D+30', 'D+90']),
        'aplicacao_minima': random.choice([100, 500, 1000, 5000, 10000]),
        'indexador': random.choice(indexador_opcoes) if 'Fundo Renda Fixa' in tipo_prod or 'CDB' in tipo_prod else 'N/A',
        'setor_economia': random.choice(setor_economia_opcoes) if 'Fundo de Ações' in tipo_prod else 'N/A',
        'estrategia_investimento': random.choice(estrategia_investimento_opcoes) if 'Fundo' in tipo_prod else 'N/A',
        'data_lancamento': fake.date_time_between(start_date='-10y', end_date='-1y').strftime('%Y-%m-%d %H:%M:%S'),
        'status_produto': 'Ativo'
    })
df_produtos = pd.DataFrame(produtos_data)
print("Produtos gerados.")

# 3. Gerar Dados de Transações
transacoes_data = []
tipo_transacao_opcoes = ["Aplicacao", "Resgate"]
status_transacao_opcoes = ["Concluída", "Pendente"]

for i in range(num_transacoes):
    cliente_id = random.choice(df_clientes['id_cliente'])
    produto_id = random.choice(df_produtos['id_produto'])
    tipo_trans = random.choice(tipo_transacao_opcoes)
    
    # Gerar data da transação após o cadastro do cliente e lançamento do produto
    data_transacao_str_cliente = df_clientes[df_clientes['id_cliente'] == cliente_id]['data_cadastro'].iloc[0]
    data_transacao_dt_cliente = datetime.strptime(data_transacao_str_cliente, '%Y-%m-%d %H:%M:%S')

    data_lancamento_str_produto = df_produtos[df_produtos['id_produto'] == produto_id]['data_lancamento'].iloc[0]
    data_lancamento_dt_produto = datetime.strptime(data_lancamento_str_produto, '%Y-%m-%d %H:%M:%S')

    start_date_transacao = max(data_transacao_dt_cliente, data_lancamento_dt_produto)
    
    data_transacao = fake.date_time_between(start_date=start_date_transacao, end_date='now')

    valor_trans = round(random.uniform(100, 100000), 2)
    
    transacoes_data.append({
        'id_transacao': f'TRA{i:06d}',
        'id_cliente': cliente_id,
        'id_produto': produto_id,
        'tipo_transacao': tipo_trans,
        'valor_transacao': valor_trans,
        'data_transacao': data_transacao.strftime('%Y-%m-%d %H:%M:%S'),
        'status_transacao': random.choice(status_transacao_opcoes)
    })
df_transacoes = pd.DataFrame(transacoes_data)
print("Transações geradas.")

# 4. Gerar Dados de Interações
interacoes_data = []
tipo_interacao_opcoes = ["Visualizacao_Produto", "Clique_CTA_Investir", "Pesquisa_Produto", "Download_Material", "Contato_Suporte", "Acesso_Area_Logada"]

for i in range(num_interacoes):
    cliente_id = random.choice(df_clientes['id_cliente'])
    tipo_interacao = random.choice(tipo_interacao_opcoes)
    
    produto_id = None
    if tipo_interacao in ["Visualizacao_Produto", "Clique_CTA_Investir", "Pesquisa_Produto"]:
        produto_id = random.choice(df_produtos['id_produto'])
    
    termo_pesquisa = fake.word() if tipo_interacao == "Pesquisa_Produto" else None

    # Gerar data da interação após o cadastro do cliente
    data_cadastro_str_cliente = df_clientes[df_clientes['id_cliente'] == cliente_id]['data_cadastro'].iloc[0]
    data_cadastro_dt_cliente = datetime.strptime(data_cadastro_str_cliente, '%Y-%m-%d %H:%M:%S')

    data_interacao = fake.date_time_between(start_date=data_cadastro_dt_cliente, end_date='now')

    interacoes_data.append({
        'id_interacao': f'INT{i:07d}',
        'id_cliente': cliente_id,
        'tipo_interacao': tipo_interacao,
        'id_produto': produto_id,
        'data_interacao': data_interacao.strftime('%Y-%m-%d %H:%M:%S'),
        'duracao_interacao_segundos': random.randint(10, 600) if 'Visualizacao' in tipo_interacao else None,
        'termo_pesquisa': termo_pesquisa
    })
df_interacoes = pd.DataFrame(interacoes_data)
print("Interações geradas.")

# 5. Gerar Dados de Mercado
dados_mercado_data = []
# Ajustei a data de início para um período mais compatível com "agora",
# dado que a data atual é 09/07/2025.
start_date_mercado = datetime(2020, 1, 1) # Últimos ~5 anos a partir de 2025
end_date_mercado = datetime.now() 

current_date = start_date_mercado
while current_date <= end_date_mercado:
    dados_mercado_data.append({
        'data': current_date.strftime('%Y-%m-%d'),
        'nome_indice': 'Ibovespa',
        'valor_indice': round(random.uniform(80000, 150000), 2),
        'taxa_selic': round(random.uniform(2.0, 15.0), 2), # Variação da Selic
        'cotacao_dolar': round(random.uniform(4.5, 6.0), 2)
    })
    current_date += timedelta(days=1)
df_dados_mercado = pd.DataFrame(dados_mercado_data)
print("Dados de Mercado gerados.")

# Salvar todos os dataframes em arquivos CSV para o datalake
df_clientes.to_csv(os.path.join(output_folder, 'clientes.csv'), index=False)
df_produtos.to_csv(os.path.join(output_folder, 'produtos.csv'), index=False)
df_transacoes.to_csv(os.path.join(output_folder, 'transacoes.csv'), index=False)
df_interacoes.to_csv(os.path.join(output_folder, 'interacoes.csv'), index=False)
df_dados_mercado.to_csv(os.path.join(output_folder, 'dados_mercado.csv'), index=False)

print(f"\nTodos os arquivos CSV foram gerados e salvos na pasta '{output_folder}/'.")