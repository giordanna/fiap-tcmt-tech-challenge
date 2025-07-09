import { HttpsError } from "firebase-functions/v2/https";
import { logger } from "firebase-functions/v2";
import { 
  Recomendacao, 
  Cliente, 
  Produto, 
  Transacao, 
  Interacao, 
  DadosDataLake,
  DetalhesUsuario,
  DocumentoUsuario
} from "../types";
import admin from "firebase-admin";
import csv from "csv-parser";
import { Readable } from "stream";


// Função para ler CSV do Firebase Storage
async function lerCSVDoStorage(fileName: string): Promise<any[]> {
  try {
    const bucket = admin.storage().bucket();
    const file = bucket.file(`datalake/${fileName}`);
    
    const [exists] = await file.exists();
    if (!exists) {
      logger.warn(`Arquivo ${fileName} não encontrado no Storage`);
      return [];
    }

    const [buffer] = await file.download();
    const csvString = buffer.toString('utf8');
    
    const results: any[] = [];
    const stream = Readable.from(csvString).pipe(csv());
    
    try {
      for await (const data of stream) {
        results.push(data);
      }
    } catch (error) {
      logger.error(`Erro ao processar o stream do CSV ${fileName}:`, error);
      return [];
    }
    
    return results;
  } catch (error) {
    logger.error(`Erro ao ler CSV ${fileName}:`, error);
    return [];
  }
}

// Função para buscar todos os dados do datalake
export async function buscarDadosDoDataLake(): Promise<DadosDataLake> {
  logger.info("Worker: Buscando dados do Data Lake...");
  
  try {
    const [clientes, produtos, transacoes, interacoes, dados_mercado] = await Promise.all([
      lerCSVDoStorage('clientes.csv'),
      lerCSVDoStorage('produtos.csv'),
      lerCSVDoStorage('transacoes.csv'),
      lerCSVDoStorage('interacoes.csv'),
      lerCSVDoStorage('dados_mercado.csv')
    ]);

    // Converter strings para números onde necessário
    const clientesProcessados = clientes.map(cliente => ({
      ...cliente,
      idade: parseInt(cliente.idade),
      renda_mensal_estimada: parseFloat(cliente.renda_mensal_estimada),
      patrimonio_total_estimado: parseFloat(cliente.patrimonio_total_estimado)
    }));

    const produtosProcessados = produtos.map(produto => ({
      ...produto,
      rentabilidade_historica_12m: parseFloat(produto.rentabilidade_historica_12m),
      rentabilidade_historica_36m: parseFloat(produto.rentabilidade_historica_36m),
      taxa_administracao: parseFloat(produto.taxa_administracao),
      aplicacao_minima: parseFloat(produto.aplicacao_minima)
    }));

    const transacoesProcessadas = transacoes.map(transacao => ({
      ...transacao,
      valor_transacao: parseFloat(transacao.valor_transacao)
    }));

    const interacoesProcessadas = interacoes.map(interacao => ({
      ...interacao,
      duracao_interacao_segundos: interacao.duracao_interacao_segundos ? 
        parseFloat(interacao.duracao_interacao_segundos) : undefined
    }));

    const dadosMercadoProcessados = dados_mercado.map(dado => ({
      ...dado,
      valor_indice: parseFloat(dado.valor_indice),
      taxa_selic: parseFloat(dado.taxa_selic),
      cotacao_dolar: parseFloat(dado.cotacao_dolar)
    }));

    logger.info(`Dados carregados: ${clientesProcessados.length} clientes, ${produtosProcessados.length} produtos, ${transacoesProcessadas.length} transações`);
    
    return {
      clientes: clientesProcessados,
      produtos: produtosProcessados,
      transacoes: transacoesProcessadas,
      interacoes: interacoesProcessadas,
      dados_mercado: dadosMercadoProcessados
    };
  } catch (error) {
    logger.error("Erro ao buscar dados do datalake:", error);
    throw error;
  }
}

// Função para criar detalhes completos do usuário
function criarDetalhesUsuario(
  cliente: Cliente,
  transacoes: Transacao[],
  interacoes: Interacao[]
): DetalhesUsuario {
  const transacoesCliente = transacoes.filter(t => t.id_cliente === cliente.id_cliente);
  const interacoesCliente = interacoes.filter(i => i.id_cliente === cliente.id_cliente);

  // Determinar comportamento baseado no perfil de risco
  let comportamento: "conservador" | "moderado" | "desconhecido" = "desconhecido";
  
  if (cliente.perfil_risco === "Conservador") {
    comportamento = "conservador";
  } else if (cliente.perfil_risco === "Moderado") {
    comportamento = "moderado";
  } else if (cliente.perfil_risco === "Arrojado") {
    comportamento = "moderado"; // Tratamos arrojado como moderado para simplificar
  }

  // Histórico de compras baseado nas transações
  const historico_compras = transacoesCliente
    .filter(t => t.tipo_transacao === "Aplicacao" && t.status_transacao === "Concluída")
    .map(t => t.id_produto);

  // Calcular estatísticas do usuário
  const volume_total_aplicado = transacoesCliente
    .filter(t => t.tipo_transacao === "Aplicacao" && t.status_transacao === "Concluída")
    .reduce((total, t) => total + t.valor_transacao, 0);

  return {
    id_cliente: cliente.id_cliente,
    nome_cliente: cliente.nome_cliente,
    data_cadastro: cliente.data_cadastro,
    idade: cliente.idade,
    genero: cliente.genero,
    renda_mensal_estimada: cliente.renda_mensal_estimada,
    patrimonio_total_estimado: cliente.patrimonio_total_estimado,
    perfil_risco: cliente.perfil_risco,
    objetivo_investimento: cliente.objetivo_investimento,
    ultima_interacao: cliente.ultima_interacao,
    comportamento,
    historico_compras,
    total_transacoes: transacoesCliente.length,
    volume_total_aplicado,
    total_interacoes: interacoesCliente.length,
    data_ultima_atualizacao: new Date().toISOString()
  };
}

// Função para calcular score de recomendação
function calcularScoreRecomendacao(
  cliente: Cliente,
  produto: Produto,
  transacoes: Transacao[],
  interacoes: Interacao[]
): number {
  let score = 0.5; // Score base

  // Ajuste baseado no perfil de risco
  if (cliente.perfil_risco === "Conservador" && produto.risco_associado === "Baixo") {
    score += 0.3;
  } else if (cliente.perfil_risco === "Moderado" && produto.risco_associado === "Médio") {
    score += 0.25;
  } else if (cliente.perfil_risco === "Arrojado" && produto.risco_associado === "Alto") {
    score += 0.2;
  }

  // Ajuste baseado na rentabilidade
  if (produto.rentabilidade_historica_12m > 0.1) {
    score += 0.1;
  }

  // Penalizar produtos com taxas de administração muito altas
  if (produto.taxa_administracao > 2.0) {
    score -= 0.1;
  }

  // Bonus para produtos com rentabilidade histórica consistente de 36m
  if (produto.rentabilidade_historica_36m > 0.08) {
    score += 0.05;
  }

  // Ajuste baseado no valor mínimo vs patrimônio do cliente
  const proporcaoAplicacao = produto.aplicacao_minima / cliente.patrimonio_total_estimado;
  if (proporcaoAplicacao < 0.05) { // Menos de 5% do patrimônio
    score += 0.1;
  }

  // Penalizar se cliente já tem muitas transações com este produto
  const transacoesProduto = transacoes.filter(t => 
    t.id_cliente === cliente.id_cliente && 
    t.id_produto === produto.id_produto
  ).length;
  
  if (transacoesProduto > 2) {
    score -= 0.2;
  }

  // Bonus se cliente demonstrou interesse (interações)
  const interacoesProduto = interacoes.filter(i => 
    i.id_cliente === cliente.id_cliente && 
    i.id_produto === produto.id_produto
  );
  
  if (interacoesProduto.length > 0) {
    score += 0.15;
    
    // Bonus adicional para interações longas (mais de 120 segundos)
    const interacoesLongas = interacoesProduto.filter(i => 
      i.duracao_interacao_segundos && i.duracao_interacao_segundos > 120
    );
    
    if (interacoesLongas.length > 0) {
      score += 0.05;
    }
  }

  // Ajuste baseado na liquidez e objetivo
  if (cliente.objetivo_investimento === "Reserva de Emergência" && produto.liquidez === "D+0") {
    score += 0.2;
  } else if (cliente.objetivo_investimento === "Aposentadoria" && produto.tipo_produto === "Previdência Privada") {
    score += 0.25;
  }

  return Math.min(Math.max(score, 0), 1); // Manter entre 0 e 1
}

// Função para gerar motivo da recomendação
function gerarMotivoRecomendacao(cliente: Cliente, produto: Produto): string {
  const motivos = [];

  if (cliente.perfil_risco.toLowerCase() === produto.risco_associado.toLowerCase()) {
    motivos.push(`Compatível com seu perfil ${cliente.perfil_risco.toLowerCase()}`);
  }

  if (produto.rentabilidade_historica_12m > 0.1) {
    motivos.push(`Boa rentabilidade histórica (${(produto.rentabilidade_historica_12m * 100).toFixed(1)}%)`);
  }

  if (produto.taxa_administracao <= 1.0) {
    motivos.push(`Taxa de administração baixa (${produto.taxa_administracao.toFixed(2)}%)`);
  }

  if (cliente.objetivo_investimento === "Reserva de Emergência" && produto.liquidez === "D+0") {
    motivos.push("Ideal para reserva de emergência");
  }

  if (cliente.objetivo_investimento === "Aposentadoria" && produto.tipo_produto === "Previdência Privada") {
    motivos.push("Produto específico para aposentadoria");
  }

  if (produto.aplicacao_minima <= cliente.patrimonio_total_estimado * 0.05) {
    motivos.push("Valor mínimo acessível");
  }

  return motivos.length > 0 ? motivos.join(", ") : "Produto adequado ao seu perfil";
}

// Geração de recomendações baseada nos dados reais
export function gerarRecomendacoes(
  cliente: Cliente,
  produtos: Produto[],
  transacoes: Transacao[],
  interacoes: Interacao[]
): Recomendacao[] {
  // Filtrar apenas produtos ativos
  const produtosAtivos = produtos.filter(p => p.status_produto === "Ativo");

  // Calcular score para cada produto
  const produtosComScore = produtosAtivos.map(produto => {
    const score = calcularScoreRecomendacao(cliente, produto, transacoes, interacoes);
    const motivo = gerarMotivoRecomendacao(cliente, produto);
    
    return {
      produto: produto.nome_produto,
      pontuacao: parseFloat(score.toFixed(2)),
      motivo
    };
  });

  // Ordenar por score e pegar os top 5
  produtosComScore.sort((a, b) => b.pontuacao - a.pontuacao);
  
  return produtosComScore.slice(0, 5);
}

// Armazenamento no Banco NoSQL completo (Firestore)
export async function armazenarDadosUsuarioCompleto(
  usuarioId: string, 
  detalhesUsuario: DetalhesUsuario, 
  recomendacoes: Recomendacao[]
): Promise<void> {
  const bd = admin.firestore();
  /** Armazena os dados completos do usuário e recomendações no Firestore. */
  logger.info(`Worker: Armazenando dados completos para usuário ${usuarioId}`);
  
  const documentoUsuario: DocumentoUsuario = {
    ...detalhesUsuario,
    recomendacoes: recomendacoes
  };
  
  const refDocumento = bd.collection("recomendacoes_por_usuario").doc(usuarioId);
  await refDocumento.set(documentoUsuario, { merge: true });
  
  logger.info(`Worker: Dados completos para usuário ${usuarioId} salvos com sucesso.`);
}

// Lógica principal do worker com tratamento de erro melhorado
export async function executarWorkerRecomendacoes(mensagemId?: string, dadosMensagem?: any): Promise<void> {
  /**
   * Função principal do Worker que processa dados do Data Lake 
   * e atualiza recomendações no Firestore.
   */
  logger.info("Worker de Recomendações iniciado via Pub/Sub...", { 
    mensagemId,
    dadosMensagem 
  });

  try {
    const bd = admin.firestore();

    // Verificar se o Firestore está acessível
    await bd.collection("health_check").doc("test").get();
    
    // Buscar dados do datalake
    const dadosDataLake = await buscarDadosDoDataLake();

    // Processar recomendações para cada cliente
    for (const cliente of dadosDataLake.clientes) {
      logger.info(`Worker: Processando cliente ${cliente.id_cliente}...`);
      
      try {
        // Gerar recomendações
        const recomendacoesGeradas = gerarRecomendacoes(
          cliente,
          dadosDataLake.produtos,
          dadosDataLake.transacoes,
          dadosDataLake.interacoes
        );
        
        // Criar detalhes completos do usuário
        const detalhesUsuario = criarDetalhesUsuario(
          cliente,
          dadosDataLake.transacoes,
          dadosDataLake.interacoes
        );
        
        // Armazenar dados completos do usuário
        await armazenarDadosUsuarioCompleto(cliente.id_cliente, detalhesUsuario, recomendacoesGeradas);
        
        logger.info(`Worker: Dados completos salvos para ${cliente.id_cliente}:`);
        logger.info(`  - ${recomendacoesGeradas.length} recomendações`);
        logger.info(`  - ${detalhesUsuario.total_transacoes} transações`);
        logger.info(`  - ${detalhesUsuario.total_interacoes} interações`);
        logger.info(`  - Volume total aplicado: R$ ${detalhesUsuario.volume_total_aplicado.toFixed(2)}`);
      } catch (error) {
        logger.error(`Erro ao processar cliente ${cliente.id_cliente}:`, error);
        // Continua com o próximo cliente
      }
      
      logger.info("-".repeat(30));
    }
    
    logger.info("Worker de Recomendações finalizado com sucesso.", { mensagemId });
  } catch (erro) {
    logger.error("Erro no Worker de Recomendações:", {
      erro: erro instanceof Error ? erro.message : erro,
      stack: erro instanceof Error ? erro.stack : undefined,
      mensagemId,
      dadosMensagem
    });
    
    // Re-throw o erro para que o Pub/Sub saiba que falhou e possa fazer retry
    throw new HttpsError("internal", "Erro ao executar o worker de recomendações.", erro);
  }
}