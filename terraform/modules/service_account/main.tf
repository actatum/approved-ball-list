resource "google_service_account" "circleci" {
  account_id = "circleci"
  project    = var.project
}

resource "google_project_iam_binding" "circleci-cloudfunctions" {
  project = var.project
  role    = "roles/cloudfunctions.developer"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_project_iam_binding" "circleci-storage" {
  project = var.project
  role    = "roles/storage.admin"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_project_iam_binding" "circleci-pubsub" {
  project = var.project
  role    = "roles/pubsub.editor"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_project_iam_binding" "circleci-cloud-scheduler" {
  project = var.project
  role    = "roles/cloudscheduler.admin"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_project_iam_binding" "circleci-iam" {
  project = var.project
  role    = "roles/iam.securityAdmin"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_service_account_key" "circleci_key" {
  service_account_id = google_service_account.circleci.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}
