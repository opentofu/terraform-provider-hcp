---
page_title: "hcp_waypoint_application Resource - terraform-provider-hcp"
subcategory: "HCP Waypoint"
description: |-
  The Waypoint Application resource managed the lifecycle of an Application that's based off of a Template.
---

# hcp_waypoint_application `Resource`

-> **Note:** HCP Waypoint is currently in public beta.

The Waypoint Application resource managed the lifecycle of an Application that's based off of a Template.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the Application.
- `template_id` (String) ID of the Template this Application is based on.

### Optional

- `application_input_variables` (Attributes Set) Input variables set for the application. (see [below for nested schema](#nestedatt--application_input_variables))
- `project_id` (String) The ID of the HCP project where the Waypoint Application is located.
- `readme_markdown` (String) Instructions for using the Application (markdown format supported). Note: this is a base64 encoded string, and can only be set in configuration after initial creation. The initial version of the README is generated from the README Template from source Template.

### Read-Only

- `id` (String) The ID of the Application.
- `namespace_id` (String) Internal Namespace ID.
- `organization_id` (String) The ID of the HCP organization where the Waypoint Application is located.
- `template_input_variables` (Attributes Set) Input variables set for the application. (see [below for nested schema](#nestedatt--template_input_variables))
- `template_name` (String) Name of the Template this Application is based on.

<a id="nestedatt--application_input_variables"></a>
### Nested Schema for `application_input_variables`

Required:

- `name` (String) Variable name
- `value` (String) Variable value
- `variable_type` (String) Variable type


<a id="nestedatt--template_input_variables"></a>
### Nested Schema for `template_input_variables`

Required:

- `name` (String) Variable name
- `value` (String) Variable value

Optional:

- `variable_type` (String) Variable type
