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

module "cloud_run" {
  source           = "./modules/cloud_run"
  project          = var.project
  discord_channels = var.discord_channels
  discord_token    = var.discord_token
  image_tag        = var.circle_sha1
  cockroachdb_url  = var.cockroachdb_url

  depends_on = [
    module.services,
    module.storage
  ]
}

module "scheduler" {
  source                 = "./modules/scheduler"
  project                = var.project
  region                 = var.region
  uri                    = "${module.cloud_run.url}/v1/cron"
  cloud_run_service_name = module.cloud_run.name

  depends_on = [
    module.services,
    module.cloud_run,
  ]
}
