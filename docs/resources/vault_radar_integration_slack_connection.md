---
page_title: "hcp_vault_radar_integration_slack_connection Resource - terraform-provider-hcp"
subcategory: "HCP Vault Radar"
description: |-
  This terraform resource manages an Integration Slack Connection in Vault Radar.
---

# hcp_vault_radar_integration_slack_connection (Resource)

-> **Note:** This feature is currently in private beta.

This terraform resource manages an Integration Slack Connection in Vault Radar.

## Example Usage

```terraform
variable "slack_token" {
  type      = string
  sensitive = true
}

resource "hcp_vault_radar_integration_slack_connection" "slack_connection" {
  name  = "example connection to slack"
  token = var.slack_token
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of connection. Name must be unique.
- `token` (String, Sensitive) Slack bot user OAuth token. Example: Bot token strings begin with 'xoxb'.

### Optional

- `project_id` (String) The ID of the HCP project where Vault Radar is located. If not specified, the project specified in the HCP Provider config block will be used, if configured.

### Read-Only

- `id` (String) The ID of this resource.