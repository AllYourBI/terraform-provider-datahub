terraform {
  required_providers {
    datahub = {
      source  = "hashicorp.com/aybi/aybi-datahub"
    }
  }
  # required_version = ">= 1.1.0"
}

provider "datahub" {
  base_url = "https://staging.api.datahub.allyourbi.nl"
  client_id = "xxx"
  client_secret     = "xxxxx"
}

# resource "aybi_datahub_job" "example" {
#   image = "tralala image"
  # environment = {
  #   "TEST_1" = "VALUE 1",
  #   "TEST_2" = "VALUE 2",
  # }
# }