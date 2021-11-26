# terraform {
#   backend "local" {

#   }
# }

terraform {
  backend "gcs" {
    bucket = "approved-ball-list-4-tfstate"
    prefix = "terraform/state"
  }
}
