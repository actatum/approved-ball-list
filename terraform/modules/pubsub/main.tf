module "pubsub" {
  source  = "terraform-google-modules/pubsub/google"
  version = "~> 6.0"

  topic      = "abl"
  project_id = var.project
  push_subscriptions = [
    {
      name                  = "abl_sub"
      push_endpoint         = var.endpoint
      max_delivery_attempts = 5
      maximum_backoff       = "600s"
      dead_letter_topic     = module.dlq.outputs.topic
    }
  ]
}

module "dlq" {
  source     = "terraform-google-modules/pubsub/google"
  version    = "~> 6.0"
  topic      = "abl_dlq"
  project_id = var.project

  create_subscriptions = false
}
