---
page_title: "datahub_runners Data Source - terraform-provider-datahub"
subcategory: ""
description: |-
  Lists all accessible runners.
---

# datahub_runners (Data Source)

Lists all runners accessible to the authenticated client, including global runners and runners owned by the client.

## Example Usage

```hcl
data "datahub_runners" "all" {}

output "runner_names" {
  value = [for r in data.datahub_runners.all.runners : r.name]
}
```

## Attribute Reference

- `runners` - A list of runner objects. Each runner has the following attributes:
  - `id` - The UUID of the runner.
  - `name` - The unique identifier of the runner.
  - `display_name` - Human-friendly display name.
  - `is_global` - Whether the runner is a global runner.
  - `client_id` - The client ID that owns the runner, if client-scoped.
  - `status` - Current health status: `unknown`, `healthy`, `degraded`, or `offline`.
  - `last_health_ping` - Timestamp of the last health ping.
  - `registered_at` - Timestamp when the runner was registered.
  - `certificate_expiry` - Certificate expiry timestamp.
