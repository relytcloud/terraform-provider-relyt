terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  role = "SYSTEMADMIN"
}

data "relyt_cloud_region_endpoints" "endpoint_list" {
  cloud  = var.cloud
  region = var.region
}