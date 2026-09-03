terraform {
  required_providers {
    relyt = {
      source = "relytcloud/relyt"
    }
  }
}

provider "relyt" {
  # Defaults to https://api.data.cloud, the original AWS control plane. Any
  # other deployment has to set this: the wrong control plane answers, and the
  # mismatch only surfaces later as CLOUD_REGION_NOT_EXIST. Can also be set
  # via RELYT_API_HOST.
  api_host = "https://<your-domain>"
  auth_key = "9a3727e5b9c0mockaGbll2HVLVKLLY1AyjOilAqeyPOBAb74A7VlMOCKTi0bJWJd3"
  role     = "SYSTEMADMIN"
}

#provider to operate database and schema

provider "relyt" {
  alias    = "database"
  api_host = "https://<your-domain>"
  auth_key = "9a3727e5b9c0mockaGbll2HVLVKLLY1AyjOilAqeyPOBAb74A7VlMOCKTi0bJWJd3"
  role     = "SYSTEMADMIN"
  data_access_config = {
    access_key = "<access_key>"
    secret_key = "<secret_key>"
    endpoint   = "http://<dns_name>:8180"
  }
}
