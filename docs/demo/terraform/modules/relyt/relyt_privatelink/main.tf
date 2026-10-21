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

# "*" is an AWS value. On Alibaba Cloud a principal is the consumer's main
# account UID, digits only.
resource "relyt_dwsu_privatelink" "pl_api" {
  dwsu_id      = var.dwsu_id
  service_type = "data_api"
  allow_principals = [
    { principal = "*" }
  ]
}

resource "relyt_dwsu_privatelink" "pl_db" {
  dwsu_id      = var.dwsu_id
  service_type = "database"
  allow_principals = [{ principal = "*" }]
}
