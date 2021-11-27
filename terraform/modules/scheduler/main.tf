# resource "google_app_engine_application" "app" {
#   project     = var.project
#   location_id = var.region
# }

resource "google_cloud_scheduler_job" "job" {
  project     = var.project
  name        = "approvedBallListJob"
  description = "Run once weekly to check for new approved balls"
  schedule    = "0 12 * * *"
  time_zone   = "EST"
  region      = var.region

  pubsub_target {
    # topic.id is the topic's full resource name.
    topic_name = var.pubsub_topic
    data       = base64encode("test")
  }
}
