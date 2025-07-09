export interface Recomendacao {
  produto: string;
  pontuacao: number;
  motivo: string;
}

export interface DocumentoRecomendacao {
  recomendacoes: Recomendacao[];
}

export interface DetalhesUsuario {
  id_cliente: string;
  nome_cliente: string;
  data_cadastro: string;
  idade: number;
  genero: string;
  renda_mensal_estimada: number;
  patrimonio_total_estimado: number;
  perfil_risco: 'Conservador' | 'Moderado' | 'Arrojado';
  objetivo_investimento: string;
  ultima_interacao: string;
  comportamento: "conservador" | "moderado" | "desconhecido";
  historico_compras: string[];
  total_transacoes: number;
  volume_total_aplicado: number;
  total_interacoes: number;
  data_ultima_atualizacao: string;
}

export interface DocumentoUsuario extends DetalhesUsuario {
  recomendacoes: Recomendacao[];
}

// Novas interfaces para DLQ
export interface MensagemPubSub {
  messageId: string;
  data: string;
  attributes?: { [key: string]: string };
  publishTime: string;
}

// Interfaces para dados do CSV
export interface Cliente {
  id_cliente: string;
  nome_cliente: string;
  data_cadastro: string;
  idade: number;
  genero: string;
  renda_mensal_estimada: number;
  patrimonio_total_estimado: number;
  perfil_risco: 'Conservador' | 'Moderado' | 'Arrojado';
  objetivo_investimento: string;
  ultima_interacao: string;
}

export interface Produto {
  id_produto: string;
  nome_produto: string;
  tipo_produto: string;
  risco_associado: 'Baixo' | 'Médio' | 'Alto';
  rentabilidade_historica_12m: number;
  rentabilidade_historica_36m: number;
  taxa_administracao: number;
  liquidez: string;
  aplicacao_minima: number;
  indexador: string;
  setor_economia: string;
  estrategia_investimento: string;
  data_lancamento: string;
  status_produto: string;
}

export interface Transacao {
  id_transacao: string;
  id_cliente: string;
  id_produto: string;
  tipo_transacao: 'Aplicacao' | 'Resgate';
  valor_transacao: number;
  data_transacao: string;
  status_transacao: 'Concluída' | 'Pendente';
}

export interface Interacao {
  id_interacao: string;
  id_cliente: string;
  tipo_interacao: string;
  id_produto?: string;
  data_interacao: string;
  duracao_interacao_segundos?: number;
  termo_pesquisa?: string;
}

export interface DadoMercado {
  data: string;
  nome_indice: string;
  valor_indice: number;
  taxa_selic: number;
  cotacao_dolar: number;
}

export interface DadosDataLake {
  clientes: Cliente[];
  produtos: Produto[];
  transacoes: Transacao[];
  interacoes: Interacao[];
  dados_mercado: DadoMercado[];
}
