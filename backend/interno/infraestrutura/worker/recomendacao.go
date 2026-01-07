package worker

import (
	"fmt"
	"log/slog"

	"backend/interno/casodeuso"
	"backend/interno/infraestrutura/pubsub"
)

const TopicoGerarRecomendacao = "gerar-recomendacao"

type WorkerRecomendacao struct {
	servico *casodeuso.ServicoRecomendacao
	bus     pubsub.EventBus
}

func NovoWorkerRecomendacao(servico *casodeuso.ServicoRecomendacao, bus pubsub.EventBus) *WorkerRecomendacao {
	worker := &WorkerRecomendacao{
		servico: servico,
		bus:     bus,
	}
	return worker
}

// Iniciar registra o worker como assinante do tópico e inicia a goroutine de consumo
func (w *WorkerRecomendacao) Iniciar() {
	slog.Info("Inicializando Worker de Recomendação", "topico_alvo", TopicoGerarRecomendacao)

	// A função Assinar inicia o processamento em background (goroutine)
	w.bus.Assinar(TopicoGerarRecomendacao, w.processarEvento)

	slog.Info("Worker de recomendação iniciado com sucesso")
}

func (w *WorkerRecomendacao) processarEvento(payload interface{}) {
	slog.Info("Mensagem recebida no worker de recomendação")

	clienteID, ok := payload.(string)
	if !ok {
		slog.Error("Payload inválido recebido no worker: esperava string (clienteID)",
			"payload_type", fmt.Sprintf("%T", payload),
			"payload_value", payload)
		return
	}

	if clienteID == "" {
		slog.Warn("Recebido clienteID vazio no worker")
		return
	}

	slog.Info("Iniciando processamento assíncrono para cliente", "cliente_id", clienteID)

	resultado, err := w.servico.Executar(clienteID)
	if err != nil {
		slog.Error("Erro ao processar recomendação no worker",
			"erro", err,
			"cliente_id", clienteID)
		return
	}

	slog.Info("Recomendação processada com sucesso via worker",
		"cliente_id", clienteID,
		"recomendacoes_geradas", len(resultado.Recomendacoes))
}
