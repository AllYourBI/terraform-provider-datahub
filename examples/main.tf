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
  image = "aybcr.azurecr.io/aybi/dh-test-image"
  
  environment = {
    "TEST_1" = "UPDATED1",
    "TEST_2" = "VALUE 3",
  }
  
  secrets = {
    "SECRET_1" = "secret sauce",
  }

  # oauth= {
  #   application = "exact_online"
  #   flow = "authorization_code"
  #   token_url = "https://x.nl/token"
  #   authorization_url = "https://x.nl/auth"
  #   scope = "tralal"
  #   config_prefix = "EXACT_ONLINE_"
  # }

}

data "datahub_oauth_url" "test" {
  job_id = datahub_job.example.job_id
}

output "URL" {
  value = data.datahub_oauth_url.test.redirect
}

# resource "datahub_init_run" "init-test" {
#   name  = "initrun"
#   image = "aybcr.azurecr.io/aybi/dh-test-image:non-existing-tag"
  
#   environment = {
#     "TEST_1" = "UPDATED1",
#     "TEST_2" = "VALUE 3",
#   }
  
#   secrets = {
#     "SECRET_1" = "secret sauce",
#   }
# }

resource "datahub_client" "client-test" {
  customer_code = "ayby"
  customer_name = "test-client1"
}

output "client_secret" {
  value = datahub_client.client-test.client_secret
  sensitive = true
}

resource "datahub_schedule" "test-sched"{
  schedule = "* * * * * *"

  bindings = [
    {
      job_id = datahub_job.example.job_id
      environment = {
        "TEST_1" = "UPDATED1",
        "TEST_2" = "VALUE 3",
      }
    }
  ]

}