import express from "express";
import { DocumentoRecomendacao } from "../types";
import { logger } from "firebase-functions/v2";
import admin from "firebase-admin";

// --- Microsserviço de Recomendações (Função HTTP) ---
const app = express();
app.use(express.json());

app.get("/recomendacoes/:usuarioId", async (req, res) => {
  /**
   * Endpoint para obter recomendações de investimento para um usuário específico.
   * Busca as recomendações do Firestore.
   */
  const usuarioId: string = req.params.usuarioId;
  logger.info(`Requisição recebida para o usuário: ${usuarioId}`);

  try {
    const bd = admin.firestore();

    const refDocumento = bd.collection("recomendacoes_por_usuario").doc(usuarioId);
    const documento = await refDocumento.get();

    if (documento.exists && documento.data() && (documento.data() as DocumentoRecomendacao).recomendacoes) {
      const dados = documento.data() as DocumentoRecomendacao;
      return res.json({ id_usuario: usuarioId, recomendacoes: dados.recomendacoes });
    } else {
      // Se não encontrar para o usuário específico, tenta o padrão
      const documentoPadrao = await bd.collection("recomendacoes_por_usuario").doc("usuario_padrao").get();
      if (documentoPadrao.exists && documentoPadrao.data() && (documentoPadrao.data() as DocumentoRecomendacao).recomendacoes) {
        const dados = documentoPadrao.data() as DocumentoRecomendacao;
        return res.json({ id_usuario: usuarioId, recomendacoes: dados.recomendacoes });
      } else {
        return res.status(404).json({ mensagem: "Nenhuma recomendação encontrada para este usuário ou padrão." });
      }
    }
  } catch (erro) {
    logger.error("Erro ao buscar recomendações:", erro);
    return res.status(500).json({ mensagem: "Erro interno ao buscar recomendações." });
  }
});

app.get("/healthcheck", (_, res) => {
  /**
   * Endpoint de verificação de saúde para verificar se o serviço está ativo.
   */
  res.json({ status: "OK", servico: "api-recomendacoes" });
});

export { app };