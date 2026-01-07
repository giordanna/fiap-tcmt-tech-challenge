package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"cloud.google.com/go/pubsub"
)

// GCPEventBus implementa EventBus usando Google Cloud Pub/Sub
type GCPEventBus struct {
	client   *pubsub.Client
	ctx      context.Context
	handlers map[string][]func(interface{})
	mu       sync.RWMutex
	subs     map[string]*pubsub.Subscription
	ambiente string
}

// NovoGCPEventBus cria uma nova instância do EventBus usando GCP Pub/Sub
func NovoGCPEventBus(ctx context.Context, projectID, ambiente string) (*GCPEventBus, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente Pub/Sub: %w", err)
	}

	bus := &GCPEventBus{
		client:   client,
		ctx:      ctx,
		handlers: make(map[string][]func(interface{})),
		subs:     make(map[string]*pubsub.Subscription),
		ambiente: ambiente,
	}

	slog.Info("GCP Pub/Sub inicializado", "projectID", projectID, "ambiente", ambiente)
	return bus, nil
}

// formatarTopico retorna o nome do tópico formatado com o ambiente
func (b *GCPEventBus) formatarTopico(topico string) string {
	if b.ambiente == "" {
		return topico
	}
	return fmt.Sprintf("%s-%s", topico, b.ambiente)
}

// Publicar publica um evento em um tópico do GCP Pub/Sub
func (b *GCPEventBus) Publicar(nomeTopico string, payload interface{}) {
	topico := b.formatarTopico(nomeTopico)

	// Serializa o payload para JSON
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Erro ao serializar payload", "topico", topico, "erro", err)
		return
	}

	// Obtém ou cria o tópico
	topic := b.client.Topic(topico)

	// Verifica se o tópico existe, se não, cria
	exists, err := topic.Exists(b.ctx)
	if err != nil {
		slog.Error("Erro ao verificar existência do tópico", "topico", topico, "erro", err)
		return
	}

	if !exists {
		topic, err = b.client.CreateTopic(b.ctx, topico)
		if err != nil {
			slog.Error("Erro ao criar tópico", "topico", topico, "erro", err)
			return
		}
		slog.Info("Tópico criado", "topico", topico)
	}

	// Publica a mensagem
	result := topic.Publish(b.ctx, &pubsub.Message{
		Data: data,
	})

	// Aguarda confirmação de forma assíncrona
	go func() {
		id, err := result.Get(b.ctx)
		if err != nil {
			slog.Error("Erro ao publicar mensagem", "topico", topico, "erro", err)
			return
		}
		slog.Info("Evento publicado", "topico", topico, "messageID", id)
	}()
}

// Assinar registra um handler para um tópico e inicia o consumo de mensagens
func (b *GCPEventBus) Assinar(nomeTopico string, handler func(payload interface{})) {
	topico := b.formatarTopico(nomeTopico)

	b.mu.Lock()
	b.handlers[topico] = append(b.handlers[topico], handler)

	// Se já existe uma subscription ativa para este tópico, não cria outra
	if _, exists := b.subs[topico]; exists {
		b.mu.Unlock()
		slog.Info("Handler adicionado ao tópico existente", "topico", topico)
		return
	}
	b.mu.Unlock()

	// Nome da subscription (pode ser customizado)
	subscriptionName := fmt.Sprintf("%s-sub", topico)

	// Obtém ou cria a subscription
	sub := b.client.Subscription(subscriptionName)
	exists, err := sub.Exists(b.ctx)
	if err != nil {
		slog.Error("Erro ao verificar existência da subscription", "subscription", subscriptionName, "erro", err)
		return
	}

	if !exists {
		// Garante que o tópico existe antes de criar a subscription
		topic := b.client.Topic(topico)
		topicExists, err := topic.Exists(b.ctx)
		if err != nil {
			slog.Error("Erro ao verificar existência do tópico", "topico", topico, "erro", err)
			return
		}

		if !topicExists {
			topic, err = b.client.CreateTopic(b.ctx, topico)
			if err != nil {
				slog.Error("Erro ao criar tópico", "topico", topico, "erro", err)
				return
			}
			slog.Info("Tópico criado", "topico", topico)
		}

		// Cria a subscription
		sub, err = b.client.CreateSubscription(b.ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 60, // 60 segundos para processar a mensagem
		})
		if err != nil {
			slog.Error("Erro ao criar subscription", "subscription", subscriptionName, "erro", err)
			return
		}
		slog.Info("Subscription criada", "subscription", subscriptionName, "topico", topico)
	}

	b.mu.Lock()
	b.subs[topico] = sub
	b.mu.Unlock()

	// Inicia o consumo de mensagens em uma goroutine
	go b.consumirMensagens(topico, sub)

	slog.Info("Assinante registrado", "topico", topico, "subscription", subscriptionName)
}

// consumirMensagens processa mensagens de uma subscription
func (b *GCPEventBus) consumirMensagens(topico string, sub *pubsub.Subscription) {
	err := sub.Receive(b.ctx, func(ctx context.Context, msg *pubsub.Message) {
		// Tenta deserializar como string primeiro (caso comum: clienteID)
		var payload interface{}
		var payloadStr string
		if err := json.Unmarshal(msg.Data, &payloadStr); err == nil {
			// É uma string simples
			payload = payloadStr
		} else {
			// Tenta como objeto genérico
			if err := json.Unmarshal(msg.Data, &payload); err != nil {
				slog.Error("Erro ao deserializar mensagem", "topico", topico, "erro", err)
				msg.Nack() // Não confirma a mensagem em caso de erro
				return
			}
		}

		// Executa todos os handlers registrados para este tópico
		b.mu.RLock()
		handlers, ok := b.handlers[topico]
		b.mu.RUnlock()

		if ok {
			for _, handler := range handlers {
				// Executa cada handler em sua própria goroutine
				go func(h func(interface{}), p interface{}) {
					defer func() {
						if r := recover(); r != nil {
							slog.Error("Panic no handler de evento", "erro", r, "topico", topico)
						}
					}()
					h(p)
				}(handler, payload)
			}
		}

		// Confirma o processamento da mensagem
		msg.Ack()
		slog.Debug("Mensagem processada", "topico", topico, "messageID", msg.ID)
	})

	if err != nil {
		slog.Error("Erro ao receber mensagens", "topico", topico, "erro", err)
	}
}

// Close fecha o cliente do Pub/Sub
func (b *GCPEventBus) Close() error {
	return b.client.Close()
}
