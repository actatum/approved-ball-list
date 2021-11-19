output "state_bucket_name" {
  value = google_storage_bucket.state_bucket.name
}

output "function_bucket_name" {
  value = google_storage_bucket.function_bucket.name
}
