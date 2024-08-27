---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "relyt_dwsu_privatelink Resource - relyt"
subcategory: ""
description: |-
  
---

# relyt_dwsu_privatelink (Resource)



## Example Usage

```terraform
resource "relyt_dwsu_privatelink" "privatelink" {
  dwsu_id      = "dwsu-id-from-an-duws-resource"
  service_type = "private link target service type"
  allow_principals = [
    { principal = "*" }, { principal = "arn:aws:iam::093584080162:user/*" }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `dwsu_id` (String) dwsuid
- `service_type` (String) (database | data_api | web_console)

### Optional

- `allow_principals` (Attributes List) allow principal (see [below for nested schema](#nestedatt--allow_principals))

### Read-Only

- `service_name` (String)
- `status` (String)

<a id="nestedatt--allow_principals"></a>
### Nested Schema for `allow_principals`

Optional:

- `principal` (String) principal