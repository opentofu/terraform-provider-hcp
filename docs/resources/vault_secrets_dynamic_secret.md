---
page_title: "Resource hcp_vault_secrets_dynamic_secret - terraform-provider-hcp"
subcategory: "HCP Vault Secrets"
description: |-
  The Vault Secrets dynamic secret resource manages a dynamic secret configuration.
---

# hcp_vault_secrets_dynamic_secret (Resource)

The Vault Secrets dynamic secret resource manages a dynamic secret configuration.

## Example Usage

```terraform
resource "hcp_vault_secrets_dynamic_secret" "example_aws" {
  app_name         = "my-app-1"
  secret_provider  = "aws"
  name             = "my_aws_1"
  integration_name = "my-integration-1"
  default_ttl      = "900s"
  aws_assume_role = {
    iam_role_arn = "arn:aws:iam::<account_id>:role/<role_name>"
  }
}

resource "hcp_vault_secrets_dynamic_secret" "example_gcp" {
  app_name         = "my-app-1"
  secret_provider  = "gcp"
  name             = "my_gcp_1"
  integration_name = "my-integration-1"
  default_ttl      = "900s"
  gcp_impersonate_service_account = {
    service_account_email = "<name>@<project>.iam.gserviceaccount.com"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_name` (String) Vault Secrets application name that owns the secret.
- `integration_name` (String) The Vault Secrets integration name with the capability to manage the secret's lifecycle.
- `name` (String) The Vault Secrets secret name.
- `secret_provider` (String) The third party platform the dynamic credentials give access to. One of `aws` or `gcp`.

### Optional

- `aws_assume_role` (Attributes) AWS configuration to generate dynamic credentials by assuming an IAM role. Required if `secret_provider` is `aws`. (see [below for nested schema](#nestedatt--aws_assume_role))
- `default_ttl` (String) TTL the generated credentials will be valid for.
- `gcp_impersonate_service_account` (Attributes) GCP configuration to generate dynamic credentials by impersonating a service account. Required if `secret_provider` is `gcp`. (see [below for nested schema](#nestedatt--gcp_impersonate_service_account))
- `project_id` (String) HCP project ID that owns the HCP Vault Secrets integration. Inferred from the provider configuration if omitted.

### Read-Only

- `organization_id` (String) HCP organization ID that owns the HCP Vault Secrets integration.

<a id="nestedatt--aws_assume_role"></a>
### Nested Schema for `aws_assume_role`

Required:

- `iam_role_arn` (String) AWS IAM role ARN to assume when generating credentials.


<a id="nestedatt--gcp_impersonate_service_account"></a>
### Nested Schema for `gcp_impersonate_service_account`

Required:

- `service_account_email` (String) GCP service account email to impersonate.