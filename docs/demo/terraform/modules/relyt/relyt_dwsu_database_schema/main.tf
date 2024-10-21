
terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  alias              = "db"
  auth_key           = "xxxx"
  role               = "SYSTEMADMIN"
  data_access_config = {
    access_key = var.access_key
    secret_key = var.secret_key
    endpoint   = var.endpoint
  }
}


module "database" {
  source = "./database"
  providers = {
    relyt = relyt.db
  }
}

module "schema" {
  source = "./schema"
  providers = {
    relyt = relyt.db
  }
  database_name = module.database.database_name
}

module "datasource" {
  source = "./datasource"
  providers = {
    relyt = relyt.db
  }
}