

resource "relyt_dwsu_external_schema" "ex_schema" {
  provider     = "relyt.db"
  name         = "external"
  database     = var.database_name
  catalog      = "catalog"
  table_format = "DELTA"
  properties   = {
    "metastore.type"           = "glue"
    "glue.region"              = "us-east-1"
    "s3.region"                = "us-east-1"
    "glue.access-control.mode" = "lake-formation"
  }
}