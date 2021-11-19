resource "google_cloud_scheduler_job" "job" {
  project     = var.project
  name        = "approvedBallListJob"
  description = "Run once weekly to check for new approved balls"
  schedule    = "0 7 * * TUE"

  pubsub_target {
    # topic.id is the topic's full resource name.
    topic_name = var.pubsub_topic
    data       = base64encode("test")
  }
}
