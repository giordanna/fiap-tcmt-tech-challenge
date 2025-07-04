export interface Recomendacao {
  produto: string;
  pontuacao: number;
  motivo: string;
}

export interface DadosUsuario {
  comportamento: "conservador" | "moderado" | "desconhecido";
  historico_compras: string[];
}

export interface DataLake {
  [usuarioId: string]: DadosUsuario;
}

export interface DocumentoRecomendacao {
  recomendacoes: Recomendacao[];
}

// Novas interfaces para DLQ
export interface MensagemPubSub {
  messageId: string;
  data: string;
  attributes?: { [key: string]: string };
  publishTime: string;
}
