variable "gcp_project_id" {
  description = "ID do projeto no Google Cloud"
}

variable "environment" {
  description = "Ambiente de deploy (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment deve ser dev, staging ou prod"
  }
}

variable "db_password" {
  description = "Senha do banco de dados (sensível)"
  type        = string
  sensitive   = true
}

variable "firebase_api_key" {
  description = "Firebase Web API Key para autenticação"
  type        = string
  sensitive   = true
}

variable "notification_channels" {
  description = "Canais de notificação para alertas (opcional)"
  type        = list(string)
  default     = []
}

# Configurações por ambiente
locals {
  # Sufixo do ambiente
  env_suffix = var.environment == "prod" ? "" : "-${var.environment}"

  # Configurações do Cloud SQL por ambiente
  db_configs = {
    dev = {
      tier                = "db-f1-micro" # Mais barato
      disk_size           = 10            # GB
      backup_enabled      = false
      deletion_protection = false
    }
    staging = {
      tier                = "db-g1-small" # Intermediário
      disk_size           = 20            # GB
      backup_enabled      = false
      deletion_protection = false
    }
    prod = {
      tier                = "db-custom-1-3840" # 1 vCPU, 3.75 GB RAM
      disk_size           = 50                 # GB
      backup_enabled      = true
      deletion_protection = true
    }
  }

  # Configuração selecionada
  db_config = local.db_configs[var.environment]
}

provider "google" {
  project = var.gcp_project_id
  region  = "southamerica-east1" # Região (São Paulo)
}

# 1. Secret Manager
variable "firebase_credentials_json" {
  description = "Conteúdo JSON das credenciais do Firebase (Base64 ou Raw)"
  type        = string
  sensitive   = true
}

resource "google_secret_manager_secret" "firebase_credentials" {
  secret_id = "firebase-credentials"
  replication {
    user_managed {
      replicas {
        location = "southamerica-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "firebase_credentials_version" {
  secret      = google_secret_manager_secret.firebase_credentials.id
  secret_data = var.firebase_credentials_json
}

resource "google_secret_manager_secret" "firebase_api_key" {
  secret_id = "firebase-api-key"
  replication {
    user_managed {
      replicas {
        location = "southamerica-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "firebase_api_key_version" {
  secret      = google_secret_manager_secret.firebase_api_key.id
  secret_data = var.firebase_api_key
}

# Senha do banco É versionada por ambiente (única diferença)
resource "google_secret_manager_secret" "db_pass_secret" {
  secret_id = "db-password${local.env_suffix}"
  replication {
    user_managed {
      replicas {
        location = "southamerica-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "db_pass_version" {
  secret      = google_secret_manager_secret.db_pass_secret.id
  secret_data = var.db_password
}

# 2. Service Accounts (compartilhadas entre ambientes)
resource "google_service_account" "cloudrun_sa" {
  account_id   = "cloudrun-sa"
  display_name = "Cloud Run Service Account (todos os ambientes)"
}



resource "google_secret_manager_secret_iam_member" "secret_access" {
  secret_id = google_secret_manager_secret.db_pass_secret.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

resource "google_secret_manager_secret_iam_member" "firebase_secret_access" {
  secret_id = google_secret_manager_secret.firebase_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

resource "google_secret_manager_secret_iam_member" "firebase_api_key_secret_access" {
  secret_id = google_secret_manager_secret.firebase_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# 3. Cloud SQL (versionado por ambiente - PRECISA SER DIFERENTE)
resource "google_sql_database_instance" "postgres" {
  name             = "tech-challenge-db${local.env_suffix}"
  database_version = "POSTGRES_15"
  region           = "southamerica-east1"

  settings {
    tier = local.db_config.tier

    disk_size = local.db_config.disk_size
    disk_type = "PD_SSD"

    backup_configuration {
      enabled                        = local.db_config.backup_enabled
      start_time                     = "03:00"
      point_in_time_recovery_enabled = local.db_config.backup_enabled
      transaction_log_retention_days = local.db_config.backup_enabled ? 7 : null
      backup_retention_settings {
        retained_backups = local.db_config.backup_enabled ? 7 : 1
      }
    }

    location_preference {
      zone = "southamerica-east1-a"
    }

    ip_configuration {
      ipv4_enabled    = true # Habilita IP público (necessário se não houver VPC configurada)
      private_network = null
    }
  }

  deletion_protection = local.db_config.deletion_protection
}

# 4. Artifact Registry
resource "google_artifact_registry_repository" "app_repo" {
  location      = "southamerica-east1"
  repository_id = "app-repo"
  description   = "Docker repository for Application"
  format        = "DOCKER"
}

# 5. Cloud Run
resource "google_cloud_run_service" "backend" {
  name     = "app-recomendacao${local.env_suffix}"
  location = "southamerica-east1"

  template {
    metadata {
      annotations = {
        "run.googleapis.com/cloudsql-instances" = "${var.gcp_project_id}:southamerica-east1:${google_sql_database_instance.postgres.name}"
      }
    }

    spec {
      service_account_name = google_service_account.cloudrun_sa.email

      containers {
        image = "${google_artifact_registry_repository.app_repo.location}-docker.pkg.dev/${var.gcp_project_id}/${google_artifact_registry_repository.app_repo.repository_id}/app:latest"

        env {
          name  = "DB_HOST"
          value = "/cloudsql/${var.gcp_project_id}:southamerica-east1:${google_sql_database_instance.postgres.name}"
        }
        env {
          name  = "DB_USER"
          value = "postgres"
        }
        env {
          name  = "DB_NAME"
          value = "postgres"
        }
        env {
          name = "DB_PASSWORD"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret.db_pass_secret.secret_id
              key  = "latest"
            }
          }
        }
        env {
          name  = "API_PORT"
          value = "8080"
        }
        env {
          name  = "GCP_PROJECT_ID"
          value = var.gcp_project_id
        }
        env {
          name  = "APP_ENV"
          value = var.environment
        }
        env {
          name = "FIREBASE_API_KEY"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret.firebase_api_key.secret_id
              key  = "latest"
            }
          }
        }
        # Firebase usa Application Default Credentials (ADC) no Cloud Run
        # Não precisa de FIREBASE_CREDENTIALS_PATH pois a service account já tem permissão
        
        # Limites de recursos
        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }
      
      # Concorrência e Timeout
      container_concurrency = 30
      timeout_seconds       = 300
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

# Permite acesso público (unauthenticated) à API
resource "google_cloud_run_service_iam_member" "public_access" {
  location = google_cloud_run_service.backend.location
  project  = google_cloud_run_service.backend.project
  service  = google_cloud_run_service.backend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "url_api" {
  value = google_cloud_run_service.backend.status[0].url
}

# Permissões do IAM para a Service Account do Cloud Run

# 1. Permissão para conectar ao Cloud SQL (CRÍTICO para startup)
resource "google_project_iam_member" "cloudrun_sql_client" {
  project = var.gcp_project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# 2. Permissões para Pub/Sub (Publisher e Subscriber) - Nível de projeto para garantir acesso
resource "google_project_iam_member" "cloudrun_pubsub_publisher" {
  project = var.gcp_project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

resource "google_project_iam_member" "cloudrun_pubsub_subscriber" {
  project = var.gcp_project_id
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# 3. Permissões para Firebase Auth (para GetUserByEmail e verificação)
resource "google_project_iam_member" "cloudrun_firebase_admin" {
  project = var.gcp_project_id
  role    = "roles/firebaseauth.admin"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# 4. Permissão para assinar tokens (Custom Tokens)
resource "google_project_iam_member" "cloudrun_token_creator" {
  project = var.gcp_project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# Service Account para GitHub Actions (Deploy)
resource "google_service_account" "github_actions_sa" {
  account_id   = "github-actions-deploy"
  display_name = "GitHub Actions Deploy Service Account"
}

# Permissões do GitHub Actions

# 1. Permissão para subir imagens no Artifact Registry
resource "google_artifact_registry_repository_iam_member" "github_actions_ar_writer" {
  project    = var.gcp_project_id
  location   = google_artifact_registry_repository.app_repo.location
  repository = google_artifact_registry_repository.app_repo.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.github_actions_sa.email}"
}

# 2. Permissão para gerenciar o Cloud Run (fazer o deploy)
resource "google_project_iam_member" "github_actions_run_admin" {
  project = var.gcp_project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${google_service_account.github_actions_sa.email}"
}

# 3. Permissão para "agir como" a service account de runtime (cloudrun-sa)
# Isso permite que o GitHub Actions faça deploy de um serviço que RODA como cloudrun-sa
resource "google_service_account_iam_member" "github_actions_act_as_runtime" {
  service_account_id = google_service_account.cloudrun_sa.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.github_actions_sa.email}"
}

output "github_actions_sa_email" {
  description = "Email da Service Account para configurar no GitHub Secrets"
  value       = google_service_account.github_actions_sa.email
}