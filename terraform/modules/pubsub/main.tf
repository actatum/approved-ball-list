resource "google_pubsub_topic" "cron-topic" {
  name = var.topic_name

  message_retention_duration = "86600s"
}
