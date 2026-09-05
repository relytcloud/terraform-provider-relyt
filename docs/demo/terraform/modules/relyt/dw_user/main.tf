terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  role = var.role
}

resource "relyt_dwuser" "dw_user1" {
  dwsu_id                              = var.dwsu_id
  account_name                         = var.account_name
  account_password                     = var.account_password
  datalake_identity                    = var.datalake_identity
  async_query_result_location_prefix   = var.async_query_result_location_prefix
  async_query_result_location_role_arn = var.async_query_result_location_role_arn
}

data "relyt_dwsu_boto3_access_info" "boto3" {
  dwsu_id    = relyt_dwuser.dw_user1.dwsu_id
  account_id = relyt_dwuser.dw_user1.account_name
}
