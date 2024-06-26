---
page_title: "hcp_group_members Resource - terraform-provider-hcp"
subcategory: "Cloud Platform"
description: |-
  The group members resource manages the members of an HCP Group.
  The user or service account that is running Terraform when creating an hcp_group_members resource must have roles/admin on the organization.
---

# hcp_group_members (Resource)

The group members resource manages the members of an HCP Group.

The user or service account that is running Terraform when creating an `hcp_group_members` resource must have `roles/admin` on the organization.

## Example Usage

```terraform
resource "hcp_group_members" "example" {
  group = hcp_group.example.resource_name
  members = [
    hcp_user_principal.example1.user_id,
    hcp_user_principal.example2.user_id,
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group` (String) The group's resource name in the format `iam/organization/<organization_id>/group/<name>`
- `members` (List of String) A list of user principal IDs to add to the group.

## Import

Import is supported using the following syntax:

```shell
# Group Members can be imported by specifying the group resource name
terraform import hcp_group_members.example "iam/organization/org_id/group/group-name"
```
