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
  }
}