output "service_account_private_key" {
  value = google_service_account_key.circleci_key.private_key
}
