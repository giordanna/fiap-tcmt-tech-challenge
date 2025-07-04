import { onRequest } from "firebase-functions/v2/https";
import { onMessagePublished } from "firebase-functions/v2/pubsub";
import { app } from "./apis";
import { executarWorkerRecomendacoes } from "./services/worker"
import admin from "firebase-admin";
import { logger } from "firebase-functions/v2";

// Inicializar Firebase Admin (apenas uma vez)
if (!admin.apps.length) {
  admin.initializeApp();
}

const topicoDaRecomendacao = "gerar-recomendacoes";

// --- API principal ---
export const apiRecomendacoes = onRequest(
  {
    region: "southamerica-east1",
    timeoutSeconds: 60,
  }, 
  app
);

// --- Worker principal com retry ---
export const workerGerarRecomendacoes = onMessagePublished(
  {
    topic: topicoDaRecomendacao,
    region: "southamerica-east1",
    timeoutSeconds: 540,
    memory: "512MiB",
    retry: true
  },
  async (event) => {
    const mensagemId = event.data.message?.messageId || "unknown";
    const dadosMensagem = event.data.message?.data ? 
      JSON.parse(Buffer.from(event.data.message.data, 'base64').toString()) : {};
    
    logger.info(`Worker iniciado - Mensagem ID: ${mensagemId}`, { dadosMensagem });
    
    await executarWorkerRecomendacoes(mensagemId, dadosMensagem);
  }
);
