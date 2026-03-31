---
page_title: "datahub_runner Resource - terraform-provider-datahub"
subcategory: ""
description: |-
  Manages a Datahub runner.
---

# datahub_runner (Resource)

Manages a client-scoped runner in the Datahub platform. Runners execute jobs and runs on behalf of clients.

~> **Note:** The `bootstrap_token` is only available at creation time and cannot be retrieved afterward. It is stored in Terraform state. If you lose the token, use the `datahub_runner_bootstrap_token` resource to regenerate it.

~> **Note:** Since the API does not support updating runners, any change to `name` or `display_name` will force the resource to be destroyed and recreated.

## Example Usage

```hcl
resource "datahub_runner" "example" {
  name         = "my-runner"
  display_name = "My Custom Runner"
}

output "runner_bootstrap_token" {
  value     = datahub_runner.example.bootstrap_token
  sensitive = true
}
```

## Argument Reference

- `name` - (Required, Forces new resource) Unique runner identifier. Must match pattern `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`.
- `display_name` - (Optional, Forces new resource) Human-friendly display name.

## Attribute Reference

- `id` - The UUID of the runner.
- `is_global` - Whether the runner is a global runner.
- `client_id` - The client ID that owns the runner, if client-scoped.
- `status` - Current health status: `unknown`, `healthy`, `degraded`, or `offline`.
- `last_health_ping` - Timestamp of the last health ping.
- `registered_at` - Timestamp when the runner was registered.
- `certificate_expiry` - Certificate expiry timestamp.
- `fingerprint` - The runner's certificate fingerprint.
- `bootstrap_token` - (Sensitive) The one-time bootstrap token, only available at creation.

## Import

Runners can be imported using the runner UUID:

```shell
terraform import datahub_runner.example <runner-uuid>
```

~> **Note:** The `bootstrap_token` will be empty after import since it cannot be retrieved from the API.
