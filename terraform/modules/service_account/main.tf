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
  role    = "roles/storage.objectAdmin"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_service_account_key" "circleci_key" {
  service_account_id = google_service_account.circleci.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}
