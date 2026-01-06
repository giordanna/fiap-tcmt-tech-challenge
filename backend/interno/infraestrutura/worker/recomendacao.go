package worker

import (
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

// Iniciar registra o worker como assinante do tópico
func (w *WorkerRecomendacao) Iniciar() {
	w.bus.Assinar(TopicoGerarRecomendacao, w.processarEvento)
	slog.Info("Worker de recomendação iniciado e assinando tópico", "topico", TopicoGerarRecomendacao)
}

func (w *WorkerRecomendacao) processarEvento(payload interface{}) {
	clienteID, ok := payload.(string)
	if !ok {
		slog.Error("Payload inválido para recomendação", "payload", payload)
		return
	}

	slog.Info("Worker processando recomendação", "cliente_id", clienteID)
	_, err := w.servico.Executar(clienteID)
	if err != nil {
		slog.Error("Erro no worker ao processar recomendação", "erro", err, "cliente_id", clienteID)
	}
}
