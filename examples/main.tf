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


# eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiaHR0cHM6Ly9kZXYubG9jYWw6NTAwMCJdLCJjbGllbnRfaWQiOiIxNjE3OWNkNS0wNDgzLTQ4ZDItYmQ3Yy0xZDAwMzU3MTU3MjgiLCJleHAiOjE2NzM3NTE2ODMsImlhdCI6MTY3MzczMDA4MywiaXNzIjoiRGF0YWh1YiBFbmdpbmUiLCJzY29wZXMiOiJjbGllbnRfc2NvcGUiLCJzdWIiOiIxNjE3OWNkNS0wNDgzLTQ4ZDItYmQ3Yy0xZDAwMzU3MTU3MjgifQ.F5F6Wa1OLnaUQxPhS9iiRs1KJG9MFsLLIY4xq8geRTY

# curl --location --request POST 'https://dev.local:5000/api/v1/run' \
# --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiaHR0cHM6Ly9kZXYubG9jYWw6NTAwMCJdLCJjbGllbnRfaWQiOiIxNjE3OWNkNS0wNDgzLTQ4ZDItYmQ3Yy0xZDAwMzU3MTU3MjgiLCJleHAiOjE2NzM3NTE2ODMsImlhdCI6MTY3MzczMDA4MywiaXNzIjoiRGF0YWh1YiBFbmdpbmUiLCJzY29wZXMiOiJjbGllbnRfc2NvcGUiLCJzdWIiOiIxNjE3OWNkNS0wNDgzLTQ4ZDItYmQ3Yy0xZDAwMzU3MTU3MjgifQ.F5F6Wa1OLnaUQxPhS9iiRs1KJG9MFsLLIY4xq8geRTY' \
# --header 'Content-Type: application/json' \
# --data-raw '{
#     "job_id": "58a74c1a-2436-4547-8af4-0c050eef84d8"
# }'


resource "datahub_init_run" "init-test" {
  name  = "initrun"
  image = "aybcr.azurecr.io/aybi/dh-test-image:non-existing-tag"
  
  environment = {
    "TEST_1" = "UPDATED1",
    "TEST_2" = "VALUE 3",
  }
  
  secrets = {
    "SECRET_1" = "secret sauce",
  }
}