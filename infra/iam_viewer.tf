resource "google_project_iam_member" "cloudrun_pubsub_viewer" {
  project = var.gcp_project_id
  role    = "roles/pubsub.viewer"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}
