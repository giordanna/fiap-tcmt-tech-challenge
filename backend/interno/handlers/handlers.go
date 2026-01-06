package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	casodeuso "github.com/giordanna/fiap-tcmt-tech-challenge/backend/interno/casos_de_uso"
)

type HandlerRecomendacoes struct {
	servico *casodeuso.ServicoRecomendacao
}

func NovoHandlerRecomendacoes(servico *casodeuso.ServicoRecomendacao) *HandlerRecomendacoes {
	return &HandlerRecomendacoes{servico: servico}
}

// GerarRecomendacoes gera novas recomendações para um cliente
// POST /recomendacoes/:clienteId
func (h *HandlerRecomendacoes) GerarRecomendacoes(c *gin.Context) {
	clienteID := c.Param("clienteId")

	slog.Info("Gerando recomendações", "cliente_id", clienteID)

	resultado, err := h.servico.Executar(clienteID)
	if err != nil {
		slog.Error("Erro ao gerar recomendações", "erro", err, "cliente_id", clienteID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro":     "Erro ao gerar recomendações",
			"detalhes": err.Error(),
		})
		return
	}

	slog.Info("Recomendações geradas com sucesso",
		"cliente_id", clienteID,
		"recomendacao_id", resultado.ID,
		"total_recomendacoes", len(resultado.Recomendacoes))

	c.JSON(http.StatusOK, resultado)
}

// BuscarRecomendacoes busca as recomendações mais recentes de um cliente
// GET /recomendacoes/:clienteId
func (h *HandlerRecomendacoes) BuscarRecomendacoes(c *gin.Context) {
	clienteID := c.Param("clienteId")

	slog.Info("Buscando recomendações", "cliente_id", clienteID)

	// Por enquanto, retorna 404 - implementação futura buscará do banco
	// TODO: Implementar repositório.BuscarUltimaRecomendacao(clienteID)
	c.JSON(http.StatusNotFound, gin.H{
		"mensagem": "Nenhuma recomendação encontrada para este usuário. Use POST para gerar novas recomendações.",
	})
}

// HealthCheck verifica se o serviço está funcionando
// GET /healthcheck
func (h *HandlerRecomendacoes) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"servico": "api-recomendacoes-golang",
	})
}
