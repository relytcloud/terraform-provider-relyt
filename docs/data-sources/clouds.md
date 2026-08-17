---
page_title: "relyt_clouds Data Source - relyt"
subcategory: ""
description: |-
  The cloud providers served by this control plane.
---

# relyt_clouds (Data Source)

The cloud providers served by this control plane. The valid values of `cloud` differ per deployment, so query them instead of copying them from another environment's examples.

## Example Usage

```terraform
# Discover the cloud providers served by this control plane.
data "relyt_clouds" "all" {}

output "cloud_ids" {
  value = [for c in data.relyt_clouds.all.clouds : c.id if c.is_available]
}
```

## Schema

### Read-Only

- `clouds` (Attributes List) clouds served by this control plane (see [below for nested schema](#nestedatt--clouds))

<a id="nestedatt--clouds"></a>
### Nested Schema for `clouds`

Read-Only:

- `id` (String) The ID of the cloud provider, used as `cloud` in other resources.
- `name` (String) The display name of the cloud provider.
- `link` (String) The homepage of the cloud provider.
- `is_public` (Boolean) Whether the cloud is public.
- `is_available` (Boolean) Whether the cloud is currently available for new service units.
