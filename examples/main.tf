terraform {
  required_providers {
    datahub = {
      source = "registry.terraform.io/aybi/datahub"
    }
  }
  # required_version = ">= 1.1.0"
}

provider "datahub" {
  base_url      = "https://staging.api.datahub.allyourbi.nl"
  # base_url      = "https://localhost:5000"
  
}

resource "datahub_job" "example" {
  name  = "test-scheduling-job"
  type  = "full"
  image = "aybcr.azurecr.io/aybi/dh-echo:latest"
  
  environment = {
    "TEST_1" = "UPDATED1",
    "TEST_2" = "VALUE 3",
  }
  
  secrets = {
    "SECRET_1" = "secret sauce",
  }

  command = [
    "echo",
    "-command",
    "echo", 
  ]

  # oauth= {
  #   application = "exact_online"
  #   flow = "authorization_code"
  #   token_url = "https://x.nl/token"
  #   authorization_url = "https://x.nl/auth"
  #   scope = "tralal"
  #   config_prefix = "EXACT_ONLINE_"
  # }

}

# data "datahub_oauth_url" "test" {
#   job_id = datahub_job.example.job_id
# }

# output "URL" {
#   value = data.datahub_oauth_url.test.redirect
# }

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

# resource "datahub_client" "client-test" {
#   customer_code = "ayby"
#   customer_name = "test-client1"
# }

# output "client_secret" {
#   value = datahub_client.client-test.client_secret
#   sensitive = true
# }

locals {
  numbers = range(0)
}

resource "datahub_schedule" "test-sched"{
  for_each = toset([for i in local.numbers : tostring(i)])

  schedule = "0 0 */${5 + each.key % 10} * * *"

  bindings = [
    {
      job_id = datahub_job.example.job_id
      # environment = {
      #   "TEST_1" = "UPDATED1",
      #   "TEST_2" = "VALUE 3",
      # }
    }
  ]

}



# --- Runner Resources ---

resource "datahub_runner" "example" {
  name         = "my-custom-runner"
  display_name = "My Custom Runner"
}

output "runner_id" {
  value = datahub_runner.example.id
}

output "runner_bootstrap_token" {
  value     = datahub_runner.example.bootstrap_token
  sensitive = true
}

data "datahub_runners" "all" {}

output "available_runners" {
  value = [for r in data.datahub_runners.all.runners : r.name]
}
