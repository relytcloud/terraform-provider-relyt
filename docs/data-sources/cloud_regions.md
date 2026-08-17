---
page_title: "relyt_cloud_regions Data Source - relyt"
subcategory: ""
description: |-
  The regions of one cloud provider.
---

# relyt_cloud_regions (Data Source)

The regions of one cloud provider. Region IDs differ per deployment; discover the cloud ID first via `relyt_clouds`.

## Example Usage

```terraform
data "relyt_clouds" "all" {}

data "relyt_cloud_regions" "regions" {
  cloud = data.relyt_clouds.all.clouds[0].id
}

output "region_ids" {
  value = [for r in data.relyt_cloud_regions.regions.regions : r.id]
}
```

## Schema

### Required

- `cloud` (String) The ID of the cloud provider.

### Read-Only

- `regions` (Attributes List) regions of the cloud (see [below for nested schema](#nestedatt--regions))

<a id="nestedatt--regions"></a>
### Nested Schema for `regions`

Read-Only:

- `id` (String) The ID of the region, used as `region` in other resources.
- `name` (String) The display name of the region.
- `area` (String) The geographic area of the region.
- `public` (Boolean) Whether the region is public.
