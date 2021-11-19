locals {
  timestamp = formatdate("YYMMDDhhmmss", timestamp())
  root_dir  = abspath("../")
}

data "archive_file" "source" {
  type        = "zip"
  source_dir  = local.root_dir
  output_path = "/tmp/function-${local.timestamp}.zip"
}

resource "google_storage_bucket_object" "zip" {
  name   = "source.zip#${data.archive_file.source.output_md5}"
  bucket = var.bucket_name
  source = data.archive_file.source.output_path
}

resource "google_project_service" "cf" {
  project = var.project
  service = "cloudfunctions.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_project_service" "cb" {
  project = var.project
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_cloudbuild_trigger" "build_trigger" {
  project = var.project
  trigger_template {
    branch_name = "main"
    repo_name   = "github.com/actatum/approved-ball-list"
  }
  filename = "cloudbuild.yaml"
}

# resource "google_cloudfunctions_function" "function" {
#   name    = var.function_name
#   runtime = "go113"

#   available_memory_mb   = 128
#   source_archive_bucket = var.bucket_name
#   source_archive_object = google_storage_bucket_object.zip.name

#   event_trigger {
#     event_type = "google.pubsub.topic.publish"
#     resource   = var.pubsub_topic
#     failure_policy {
#       retry = true
#     }
#   }
#   entry_point = var.function_entry_point
# }

# resource "google_cloudfunctions_function_iam_member" "invoker" {
#   project        = google_cloudfunctions_function.function.project
#   region         = google_cloudfunctions_function.function.region
#   cloud_function = google_cloudfunctions_function.function.name

#   role   = "roles/cloudfunctions.invoker"
#   member = "allUsers"
# }
