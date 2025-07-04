import { HttpsError } from "firebase-functions/v2/https";
import { logger } from "firebase-functions/v2";
import util from "util";
import { DataLake, DadosUsuario, Recomendacao } from "../types";
import admin from "firebase-admin";

const definirTempoPromise = util.promisify(setTimeout);

// Simulação de Conexão com Data Lake Cloud
export async function buscarDadosDoDataLake(): Promise<DataLake> {
  /** Simula a busca de dados brutos de um data lake. */
  logger.info("Worker: Buscando dados do Data Lake...");
  await definirTempoPromise(2000); // Simula latência
  return {
    "usuario_carlos_silva": {"comportamento": "conservador", "historico_compras": ["RF", "Tesouro"]},
    "usuario_ana_paula": {"comportamento": "moderado", "historico_compras": ["Multimercado", "Ações"]},
    "usuario_novo_cliente": {"comportamento": "desconhecido", "historico_compras": []}
  };
}

// Simulação de Geração de Recomendações (Lógica do Modelo)
export function gerarLogicaRecomendacoes(dadosUsuario: DadosUsuario): Recomendacao[] {
  /** Simula a geração de recomendações baseada nos dados do usuário. */
  const recomendacoes: Recomendacao[] = [];
  if (dadosUsuario.comportamento === "conservador") {
    recomendacoes.push({
      "produto": "Fundo de Renda Fixa Premium", 
      "pontuacao": parseFloat((Math.random() * (0.99 - 0.8) + 0.8).toFixed(2)), 
      "motivo": "Perfil conservador, alta segurança"
    });
  } else if (dadosUsuario.comportamento === "moderado") {
    recomendacoes.push({
      "produto": "Fundo de Ações Europeias", 
      "pontuacao": parseFloat((Math.random() * (0.9 - 0.7) + 0.7).toFixed(2)), 
      "motivo": "Busca diversificação e crescimento"
    });
  } else { // desconhecido ou outro
    recomendacoes.push({
      "produto": "CDB de Liquidez Diária", 
      "pontuacao": parseFloat((Math.random() * (0.8 - 0.6) + 0.6).toFixed(2)), 
      "motivo": "Recomendação inicial segura"
    });
  }
  return recomendacoes;
}

// Armazenamento no Banco NoSQL para Recomendações (Firestore)
export async function armazenarRecomendacoesNoBd(usuarioId: string, recomendacoes: Recomendacao[]): Promise<void> {
  const bd = admin.firestore();
  /** Armazena as recomendações no Firestore. */
  logger.info(`Worker: Armazenando recomendações para ${usuarioId}:`, recomendacoes);
  const refDocumento = bd.collection("recomendacoes_por_usuario").doc(usuarioId);
  await refDocumento.set({ recomendacoes: recomendacoes }, { merge: true });
  logger.info(`Worker: Recomendações para ${usuarioId} atualizadas com sucesso.`);
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
    
    const dadosDoDataLake = await buscarDadosDoDataLake();

    for (const usuarioId in dadosDoDataLake) {
      if (Object.prototype.hasOwnProperty.call(dadosDoDataLake, usuarioId)) {
        const dadosUsuario = dadosDoDataLake[usuarioId];
        logger.info(`Worker: Processando usuário ${usuarioId}...`);
        const recomendacoesGeradas = gerarLogicaRecomendacoes(dadosUsuario);
        await armazenarRecomendacoesNoBd(usuarioId, recomendacoesGeradas);
        logger.info("-".repeat(30));
      }
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