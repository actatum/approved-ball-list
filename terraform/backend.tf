# terraform {
#   backend "local" {

#   }
# }

terraform {
  backend "gcs" {
    bucket = "approved-ball-list-13-tfstate"
    prefix = "terraform/state"
  }
}
