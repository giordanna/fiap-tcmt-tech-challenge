variable "gcp_project_id" {
  description = "ID do projeto no Google Cloud"
}

variable "db_password" {
  description = "Senha do banco de dados (sensível)"
  type        = string
  sensitive   = true
}

provider "google" {
  project = var.gcp_project_id
  region  = "southamerica-east1" # Região alterada (São Paulo)
}

# 1. Secret Manager (Global/Regional)
resource "google_secret_manager_secret" "db_pass_secret" {
  secret_id = "db-password-prod"
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

# 2. Service Account
resource "google_service_account" "cloudrun_sa" {
  account_id   = "cloudrun-sa"
  display_name = "Cloud Run Service Account"
}

resource "google_secret_manager_secret_iam_member" "secret_access" {
  secret_id = google_secret_manager_secret.db_pass_secret.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# 3. Cloud SQL (Postgres)
resource "google_sql_database_instance" "postgres" {
  name             = "tech-challenge-db-prod-br"
  database_version = "POSTGRES_15"
  region           = "southamerica-east1"
  
  settings {
    tier = "db-custom-1-3840"
    location_preference {
      zone = "southamerica-east1-a"
    }
  }
  deletion_protection = false
}

# 4. Cloud Run
resource "google_cloud_run_service" "backend" {
  name     = "app-recomendacao-prod"
  location = "southamerica-east1"

  template {
    spec {
      service_account_name = google_service_account.cloudrun_sa.email
      
      containers {
        image = "gcr.io/${var.gcp_project_id}/app:latest"
        
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
      }
      
      # Adiciona Cloud SQL Proxy sidecar
      metadata {
        annotations = {
          "run.googleapis.com/cloudsql-instances" = "${var.gcp_project_id}:southamerica-east1:${google_sql_database_instance.postgres.name}"
        }
      }
    }
  }
  
  traffic {
    percent         = 100
    latest_revision = true
  }
}

output "url_api" {
  value = google_cloud_run_service.backend.status[0].url
}