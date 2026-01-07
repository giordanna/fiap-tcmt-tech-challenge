package controladores

import (
	"log/slog"
	"net/http"

	"backend/interno/casodeuso"

	"github.com/gin-gonic/gin"
)

type ControladorRecomendacoes struct {
	servico *casodeuso.ServicoRecomendacao
}

func NovoControladorRecomendacoes(servico *casodeuso.ServicoRecomendacao) *ControladorRecomendacoes {
	return &ControladorRecomendacoes{servico: servico}
}

// GerarRecomendacoes gera novas recomendações para um cliente de forma assíncrona
// @Summary      Solicita geração de recomendação
// @Description  Publica uma mensagem para gerar recomendações em background para o cliente informado
// @Tags         recomendacoes
// @Accept       json
// @Produce      json
// @Param        clienteId   path      string  true  "ID do Cliente"
// @Success      202  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v2/recomendacoes/{clienteId} [post]
func (h *ControladorRecomendacoes) GerarRecomendacoes(c *gin.Context) {
	clienteID := c.Param("clienteId")

	slog.Info("Solicitando geração de recomendações (async)", "cliente_id", clienteID)

	h.servico.SolicitarGeracao(clienteID)

	c.JSON(http.StatusAccepted, gin.H{
		"mensagem":   "Solicitação recebida com sucesso",
		"cliente_id": clienteID,
	})
}

// GerarRecomendacoesMassiva dispara geração de recomendações para todos os clientes
// @Summary      Gera recomendações em massa
// @Description  Dispara processo assíncrono para gerar recomendações para todos os clientes
// @Tags         recomendacoes
// @Accept       json
// @Produce      json
// @Success      202  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v2/recomendacoes [post]
func (h *ControladorRecomendacoes) GerarRecomendacoesMassiva(c *gin.Context) {
	slog.Info("Iniciando processo de geração de recomendações em massa")

	err := h.servico.GerarEmMassa()
	if err != nil {
		slog.Error("Erro ao iniciar geração em massa", "erro", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "Erro ao iniciar processo de recomendação em massa",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"mensagem": "Processo de geração de recomendações iniciado com sucesso. As recomendações serão geradas em background.",
	})
}

// BuscarRecomendacoes busca as recomendações mais recentes de um cliente
// @Summary      Busca recomendações recentes
// @Description  Retorna as últimas recomendações geradas para o cliente
// @Tags         recomendacoes
// @Accept       json
// @Produce      json
// @Param        clienteId   path      string  true  "ID do Cliente"
// @Success      200  {object}  dominio.ResultadoRecomendacao
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v2/recomendacoes/{clienteId} [get]
func (h *ControladorRecomendacoes) BuscarRecomendacoes(c *gin.Context) {
	clienteID := c.Param("clienteId")

	slog.Info("Buscando recomendações", "cliente_id", clienteID)

	resultado, err := h.servico.BuscarUltima(clienteID)
	if err != nil {
		slog.Error("Erro ao buscar recomendações", "erro", err, "cliente_id", clienteID)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	if resultado == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"mensagem": "Nenhuma recomendação encontrada para este usuário. Use POST para gerar novas recomendações.",
		})
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// HealthCheck verifica se o serviço está funcionando
// @Summary      Verificação de saúde
// @Description  Retorna status OK se a API estiver no ar
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v2/healthcheck [get]
func (h *ControladorRecomendacoes) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"servico": "api-recomendacoes-golang",
	})
}
