package pubsub

// EventBus define a interface para publicação e assinatura de eventos
// Implementações disponíveis:
// - GCPEventBus: Usa Google Cloud Pub/Sub (produção)
type EventBus interface {
	Publicar(topico string, payload interface{})
	Assinar(topico string, handler func(payload interface{}))
}
