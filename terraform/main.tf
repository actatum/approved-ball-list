provider "google" {
  project = var.project
  region  = var.region
}

module "storage" {
  source  = "./modules/storage"
  project = var.project
}

module "service_account" {
  source  = "./modules/service_account"
  project = var.project
}

module "pubsub" {
  source     = "./modules/pubsub"
  project    = var.project
  topic_name = "cron"
}

module "scheduler" {
  source       = "./modules/scheduler"
  project      = var.project
  pubsub_topic = module.pubsub.topic
}

module "function" {
  source               = "./modules/function"
  project              = var.project
  function_name        = "approvedBallList"
  function_entry_point = "ApprovedBallList"
  bucket_name          = module.storage.function_bucket_name
  pubsub_topic         = module.pubsub.topic
}
