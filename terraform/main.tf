provider "google" {
  project = var.project
  region  = var.region
}

module "storage" {
  source  = "./modules/storage"
  project = var.project
}

module "services" {
  source  = "./modules/services"
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
  region       = var.region
}
