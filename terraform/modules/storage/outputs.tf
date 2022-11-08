output "state_bucket_name" {
  value = google_storage_bucket.state_bucket.name
}

output "backups_bucket_name" {
  value = google_storage_bucket.backups.name
}
