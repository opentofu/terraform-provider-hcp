---
page_title: "hcp_packer_bucket Resource - terraform-provider-hcp"
subcategory: "HCP Packer"
description: |-
  The Packer Bucket resource allows you to manage a bucket within an active HCP Packer Registry.
---

# hcp_packer_bucket (Resource)

The Packer Bucket resource allows you to manage a bucket within an active HCP Packer Registry.

## Example Usage

```terraform
resource "hcp_packer_bucket" "staging" {
  name = "alpine"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The bucket's name.

### Optional

- `project_id` (String) The ID of the project to create the bucket under. If unspecified, the bucket will be created in the project the provider is configured with.

### Read-Only

- `created_at` (String) The creation time of this bucket
- `organization_id` (String) The ID of the HCP organization where this bucket is located.
- `resource_name` (String) The buckets's HCP resource name in the format `packer/project/<project_id>/packer/<name>`.

## Import

Import is supported using the following syntax:

```shell
# Using a HCP Packer Bucket Resource Name
# packer/project/{project_id}/bucket/{bucket_name}
terraform import hcp_packer_bucket.alpine packer/project/f709ec73-55d4-46d8-897d-816ebba28778/bucket/alpine
```
