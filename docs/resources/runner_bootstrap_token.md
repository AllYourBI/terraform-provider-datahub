---
page_title: "datahub_runner_bootstrap_token Resource - terraform-provider-datahub"
subcategory: ""
description: |-
  Resets a runner's bootstrap token.
---

# datahub_runner_bootstrap_token (Resource)

Resets and regenerates the bootstrap token for a Datahub runner. This invalidates the existing token and certificate, requiring the runner to re-bootstrap.

~> **Warning:** Resetting the bootstrap token will invalidate the runner's existing certificate. The runner will need to re-bootstrap using the new token.

## Example Usage

```hcl
resource "datahub_runner" "example" {
  name = "my-runner"
}

resource "datahub_runner_bootstrap_token" "example" {
  runner_id = datahub_runner.example.id
}

# Force token rotation by changing a trigger value
resource "datahub_runner_bootstrap_token" "rotating" {
  runner_id = datahub_runner.example.id

  triggers = {
    rotation = "2024-01-15"
  }
}

output "new_bootstrap_token" {
  value     = datahub_runner_bootstrap_token.example.bootstrap_token
  sensitive = true
}
```

## Argument Reference

- `runner_id` - (Required, Forces new resource) The UUID or name of the runner to reset the token for.
- `triggers` - (Optional, Forces new resource) A map of arbitrary strings that, when changed, will force the resource to be replaced, generating a new bootstrap token.

## Attribute Reference

- `id` - The UUID of the runner.
- `name` - The name of the runner.
- `bootstrap_token` - (Sensitive) The newly generated bootstrap token.
