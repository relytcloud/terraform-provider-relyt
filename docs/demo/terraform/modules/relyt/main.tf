terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  role     = "SYSTEMADMIN"
}

module "dwsu" {
  source = "./dwsu"
}

module "edps" {
  source = "./edps"
  dwsu_id = module.dwsu.dwsu_id
}

module "dw_user" {
  source = "./dw_user"
  dwsu_id = module.dwsu.dwsu_id
}

module "integration_info" {
  source = "./relyt_dwsu_integration_info"
  dwsu_id = module.dwsu.dwsu_id
  external_id = "20240821"
}

module "privatelink" {
  source = "./relyt_privatelink"
  dwsu_id = module.dwsu.dwsu_id
}

module "database_schema" {
  source = "./relyt_dwsu_database_schema"
  endpoint = "your privatelink endpoint"
  access_key = data.relyt_dwsu_boto3_access_info.boto3.boto3_access_infos[0].access_key
  secret_key = data.relyt_dwsu_boto3_access_info.boto3.boto3_access_infos[0].secret_key
}