# terraform {
#   backend "local" {

#   }
# }

terraform {
  backend "gcs" {
    bucket = "approved-ball-list-4543-tfstate"
    prefix = "terraform/state"
  }
}
