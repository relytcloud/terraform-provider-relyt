

resource "relyt_privatelink" "privatelink" {
  dwsu_id      = "dwsu-id-from-an-duws-resource"
  service_type = "private link target service type"
  service_name = "the service name to discovery whin yor cloud provider"
  allow_principles = [
    { principle = "*" }, { principle = "arn:aws:iam::093584080162:user/*" }
  ]
}