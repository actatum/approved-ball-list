output "topic" {
  value = google_eventarc_trigger.default.transport[0].pubsub[0].topic
}
