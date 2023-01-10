terraform {
  required_providers {
    datahub = {
      source = "registry.terraform.io/aybi/datahub"
    }
  }
  # required_version = ">= 1.1.0"
}

provider "datahub" {
  base_url      = "https://127.0.0.1:5000"
  client_id     = "16179cd5-0483-48d2-bd7c-1d0035715728"
  client_secret = "50YZbAU@EON1#2L!pmJ5"
}

resource "datahub_job" "example" {
  name  = "hallo3"
  type  = "full"
  image = "aybcr.azurecr.io/datahub/test"
  
  environment = {
    "TEST_1" = "UPDATED1",
    "TEST_2" = "VALUE 3",
  }
  
  secrets = {
    "SECRET_1" = "secret sauce",
  }

}