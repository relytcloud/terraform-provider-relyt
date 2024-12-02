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

data "relyt_dwsus" "dwsus" {}
