resource "google_cloud_run_service" "abl" {
  name                       = "abl"
  location                   = "us-central1"
  project                    = var.project
  autogenerate_revision_name = true

  template {
    spec {
      containers {
        image = "us-central1-docker.pkg.dev/${var.project}/abl/abl:${var.image_tag}"
        env {
          name  = "ENV"
          value = "prod"
        }
        env {
          name  = "DISCORD_TOKEN"
          value = var.discord_token
        }
        env {
          name  = "DISCORD_CHANNELS"
          value = var.discord_channels
        }
        env {
          name  = "COCKROACHDB_URL"
          value = var.cockroachdb_url
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}
