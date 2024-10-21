

data "relyt_dwsu_databases" "databases" {
  provider = relyt.db
}

data "relyt_dwsu_database" "database" {
  provider = "relyt.db"
  name = "qingdeng-test"
}


data "relyt_dwsu_schemas" "schemas" {
  database   = "database_name"
}
#

data "relyt_dwsu_external_schema" "schema" {
  database = "database_name"
  catalog  = "catalog_name"
  name     = "your_external_schema_name"
}

