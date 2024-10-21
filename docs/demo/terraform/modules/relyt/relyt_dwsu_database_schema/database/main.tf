
resource "relyt_dwsu_database" "db" {
  provider = "relyt.db"
  name     = "databse_name"
}

resource "relyt_dwsu_external_schema" "ex_schema" {
  provider     = "relyt.db"
  name         = "external"
  database     = relyt_dwsu_database.db.name
  catalog      = "catalog"
  table_format = "DELTA"
  properties   = {
    "metastore.type"           = "glue"
    "glue.region"              = "us-east-1"
    "s3.region"                = "us-east-1"
    "glue.access-control.mode" = "lake-formation"
  }
}




data "relyt_dwsu_databases" "databases" {
  provider = relyt.db
}

data "relyt_dwsu_database" "database" {
  provider = "relyt.db"
  name = "qingdeng-test"
}


data "relyt_dwsu_schemas" "schemas" {
  database   = relyt_dwsu_database.db.name
  depends_on = [relyt_dwsu_external_schema.ex_schema]
}
#

data "relyt_dwsu_external_schema" "schema" {
  database = relyt_dwsu_database.db.name
  catalog  = relyt_dwsu_external_schema.ex_schema.catalog
  name     = relyt_dwsu_external_schema.ex_schema.name
}

