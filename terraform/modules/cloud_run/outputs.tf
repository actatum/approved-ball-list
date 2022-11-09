output "url" {
  value = google_cloud_run_service.abl.status[0].url
}

output "name" {
  value = google_cloud_run_service.abl.name
}