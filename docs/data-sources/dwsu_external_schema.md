---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "relyt_dwsu_external_schema Data Source - relyt"
subcategory: ""
description: |-
  
---

# relyt_dwsu_external_schema (Data Source)



## Example Usage

```terraform
data "relyt_dwsu_external_schema" "schema" {
  database = "your_database_name"
  catalog  = "your_catalog_name"
  name     = "your_schema_name"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `catalog` (String) The catalog of the schema.
- `database` (String) The database of the schema.
- `name` (String) The name of the schema.

### Read-Only

- `external` (Boolean) Whether the schema is an external schema. true indicates yes; false indicates no.
- `owner` (String) The owner of schema.