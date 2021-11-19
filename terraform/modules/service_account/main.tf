resource "google_service_account" "circleci" {
  account_id = "circleci"
  project    = var.project
}

resource "google_project_iam_binding" "circleci" {
  project = var.project
  role    = "roles/cloudfunctions.developer"

  members = [
    "serviceAccount:${google_service_account.circleci.email}",
  ]
}

resource "google_service_account_key" "circleci_key" {
  service_account_id = google_service_account.circleci.name
  public_key_type    = "TYPE_X509_PEM_FILE"
}
