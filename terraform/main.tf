provider "google" {
  project = var.project
  region  = var.region
}

module "services" {
  source  = "./modules/services"
  project = var.project
}

module "storage" {
  source  = "./modules/storage"
  project = var.project

  depends_on = [
    module.services
  ]
}

module "pubsub" {
  source     = "./modules/pubsub"
  project    = var.project
  topic_name = "cron"

  depends_on = [
    module.services
  ]
}

module "cloud_run" {
  source           = "./modules/cloud_run"
  project          = var.project
  storage_bucket   = module.storage.backups_bucket_name
  discord_channels = var.discord_channels
  discord_token    = var.discord_token

  depends_on = [
    module.services,
    module.storage
  ]
}

module "scheduler" {
  source                 = "./modules/scheduler"
  project                = var.project
  pubsub_topic           = module.pubsub.topic
  region                 = var.region
  uri                    = "${module.cloud_run.url}/v1/cron"
  cloud_run_service_name = module.cloud_run.name

  depends_on = [
    module.services,
    module.cloud_run,
  ]
}
