resource "google_eventarc_trigger" "default" {
  name     = var.name
  location = var.region
  project  = var.project

  matching_criteria {
    attribute = "type"
    value     = "google.cloud.pubsub.topic.v1.messagePublished"
  }

  destination {
    cloud_run_service {
      service = var.cloud_run_service
      region  = var.region
      path    = var.path
    }
  }
}
