
terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  auth_key = "9a3727e5b9c0mockaGbll2HVLVKLLY1AyjOilAqeyPOBAb74A7VlMOCKTi0bJWJd3"
  role     = "SYSTEMADMIN"
}