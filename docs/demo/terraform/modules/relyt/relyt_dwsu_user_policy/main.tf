terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  role     = var.role
}


resource "relyt_dwsu_user_policy" "security_constraints" {
  dwsu_id             = var.dwsu_id
  mfa                 = "OPTIONAL"
  reset_init_password = true
}
