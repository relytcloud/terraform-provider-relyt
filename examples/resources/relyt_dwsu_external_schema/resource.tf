

resource "relyt_dwsu_external_schema" "ex_schema" {
  name          = "external"
  database_name = "your_database_name"
  catalog_name  = "your_catalog_name"
  table_format  = "DELTA"
  properties = {
    "metastore.type"           = "glue"
    "glue.region"              = "us-east-1"
    "s3.region"                = "us-east-1"
    "glue.access-control.mode" = "lake-formation"
  }
}