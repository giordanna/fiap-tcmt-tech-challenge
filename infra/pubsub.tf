# Configuração do GCP Pub/Sub para Recomendações
# Este arquivo deve ser integrado ao Terraform principal

# Tópico para geração de recomendações individuais
resource "google_pubsub_topic" "gerar_recomendacao" {
  name    = "gerar-recomendacao"
  project = var.gcp_project_id

  message_retention_duration = "86400s" # 24 horas

  labels = {
    environment = var.environment
    service     = "recomendacoes"
  }
}

# Subscription para o worker de recomendações
resource "google_pubsub_subscription" "gerar_recomendacao_sub" {
  name    = "gerar-recomendacao-sub"
  topic   = google_pubsub_topic.gerar_recomendacao.name
  project = var.gcp_project_id

  # Tempo de confirmação (ACK deadline)
  ack_deadline_seconds = 60

  # Política de retry
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  # Dead Letter Queue (opcional, mas recomendado)
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dlq.id
    max_delivery_attempts = 5
  }

  # Filtro de mensagens (opcional)
  # filter = "attributes.priority = \"high\""

  # Política de expiração (remove subscription se não houver atividade)
  expiration_policy {
    ttl = "2678400s" # 31 dias
  }

  labels = {
    environment = var.environment
    service     = "recomendacoes"
  }
}

# Dead Letter Queue (DLQ) para mensagens com falha
resource "google_pubsub_topic" "dlq" {
  name    = "recomendacoes-dlq"
  project = var.gcp_project_id

  message_retention_duration = "604800s" # 7 dias

  labels = {
    environment = var.environment
    service     = "recomendacoes"
    type        = "dlq"
  }
}

# Subscription para monitorar a DLQ
resource "google_pubsub_subscription" "dlq_sub" {
  name    = "recomendacoes-dlq-sub"
  topic   = google_pubsub_topic.dlq.name
  project = var.gcp_project_id

  # Mantém mensagens por mais tempo para análise
  message_retention_duration = "604800s" # 7 dias

  labels = {
    environment = var.environment
    service     = "recomendacoes"
    type        = "dlq"
  }
}

# IAM: Permissão para Cloud Run publicar mensagens
resource "google_pubsub_topic_iam_member" "cloud_run_publisher" {
  project = var.gcp_project_id
  topic   = google_pubsub_topic.gerar_recomendacao.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

# IAM: Permissão para Cloud Run consumir mensagens
resource "google_pubsub_subscription_iam_member" "cloud_run_subscriber" {
  project      = var.gcp_project_id
  subscription = google_pubsub_subscription.gerar_recomendacao_sub.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Alerta para mensagens na DLQ
resource "google_monitoring_alert_policy" "dlq_messages" {
  display_name = "Pub/Sub - Mensagens na DLQ"
  project      = var.gcp_project_id

  conditions {
    display_name = "Mensagens não processadas na DLQ"

    condition_threshold {
      filter          = "resource.type = \"pubsub_subscription\" AND resource.labels.subscription_id = \"${google_pubsub_subscription.dlq_sub.name}\" AND metric.type = \"pubsub.googleapis.com/subscription/num_undelivered_messages\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 10

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "604800s" # 7 dias
  }
}

# Alerta para mensagens antigas não processadas
resource "google_monitoring_alert_policy" "old_messages" {
  display_name = "Pub/Sub - Mensagens antigas não processadas"
  project      = var.gcp_project_id

  conditions {
    display_name = "Mensagens com mais de 1 hora na fila"

    condition_threshold {
      filter          = "resource.type = \"pubsub_subscription\" AND resource.labels.subscription_id = \"${google_pubsub_subscription.gerar_recomendacao_sub.name}\" AND metric.type = \"pubsub.googleapis.com/subscription/oldest_unacked_message_age\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 3600 # 1 hora

      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  notification_channels = var.notification_channels

  alert_strategy {
    auto_close = "86400s" # 24 horas
  }
}

# Outputs
output "pubsub_topic_name" {
  description = "Nome do tópico Pub/Sub"
  value       = google_pubsub_topic.gerar_recomendacao.name
}

output "pubsub_subscription_name" {
  description = "Nome da subscription Pub/Sub"
  value       = google_pubsub_subscription.gerar_recomendacao_sub.name
}

output "pubsub_dlq_topic_name" {
  description = "Nome do tópico DLQ"
  value       = google_pubsub_topic.dlq.name
}
