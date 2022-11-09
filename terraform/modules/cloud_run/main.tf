resource "google_cloud_run_service" "abl" {
  name = "abl"
  location = "us-central1"
  project = var.project
  autogenerate_revision_name = true

  template {
    spec {
      containers {
        image = "us-central1-docker.pkg.dev/${var.project}/abl/abl:latest"
        env {
          name = "ENV"
          value = "prod" 
        }
        env {
          name = "STORAGE_BUCKET"
          value = var.storage_bucket
        }
        env {
          name = "DISCORD_TOKEN"
          value = var.discord_token
        }
        env {
          name = "DISCORD_CHANNELS"
          value = var.discord_channels
        }
      }
    }
  }

  traffic {
    percent = 100
    latest_revision = true
  }
}