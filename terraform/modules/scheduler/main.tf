resource "google_service_account" "default" {
  account_id = "scheduler-sa"
  description = "Cloud Scheduler service account"
  display_name = "scheduler-sa"
}

resource "google_service_account_iam_binding" "cloud_run_invoker" {
  service_account_id = google_service_account.default.name
  role = "roles/run.invoker"

  members = [
    "serviceAccount:${google_service_account.default.email}"
  ]

  depends_on = [
    google_service_account.default
  ]
}

module "service_accounts" {
  source = "terraform-google-modules/service-accounts/google"
  version = "~> 3.0"
  project_id = var.project
  names = "scheduler-service-account"
  project_roles = [
    "${var.project}=>roles/run.invoker"
  ]
}

resource "google_cloud_scheduler_job" "job" {
  project     = var.project
  name        = "approvedBallListJob"
  description = "Run once hourly to check for new approved balls"
  schedule    = "0 * * * *"
  time_zone   = "EST"
  region      = var.region

  pubsub_target {
    # topic.id is the topic's full resource name.
    topic_name = var.pubsub_topic
    data       = base64encode("test")
  }
}

resource "google_cloud_scheduler_job" "http_job" {
  project = var.project
  name = "http-job"
  description = "Run once hourly to check for new approved/revoked balls"
  schedule = "0 * * * *"
  time_zone = "EST"
  region = var.region

  retry_config {
    retry_count = 1
  }

  http_target {
    http_method = "GET"
    uri = var.uri

    oidc_token {
      service_account_email = module.service_accounts.email
    }
  }

  depends_on = [
    google_service_account.default
  ]
}
