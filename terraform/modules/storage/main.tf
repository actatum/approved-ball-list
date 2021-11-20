resource "google_storage_bucket" "state_bucket" {
  name          = "${var.project}-tfstate"
  location      = "us-central1"
  force_destroy = true
  versioning {
    enabled = true
  }
}

# resource "google_storage_bucket" "function_bucket" {
#   name          = "${var.project}-function"
#   location      = "us-central1"
#   force_destroy = true
#   versioning {
#     enabled = true
#   }
# }
