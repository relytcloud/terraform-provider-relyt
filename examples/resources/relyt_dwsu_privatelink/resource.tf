# On AWS a principal is an IAM principal ARN, or "*" for every account.
resource "relyt_dwsu_privatelink" "privatelink" {
  dwsu_id      = "dwsu-id-from-an-duws-resource"
  service_type = "database"
  allow_principals = [
    { principal = "arn:aws:iam::123456789012:user/relyt-consumer" }
  ]
}

# On Alibaba Cloud the endpoint service keeps an account whitelist: a principal
# is the consumer's main account UID, digits only.
resource "relyt_dwsu_privatelink" "privatelink_alicloud" {
  dwsu_id      = "dwsu-id-from-an-duws-resource"
  service_type = "database"
  allow_principals = [
    { principal = "1234567890123456" }
  ]
}

# The consumer creates the endpoint on their side, and Alibaba Cloud only
# allows it in a zone the endpoint service has a backend in - which is what
# zones reports.
resource "alicloud_privatelink_vpc_endpoint" "endpoint" {
  service_name = relyt_dwsu_privatelink.privatelink_alicloud.service_name
  vpc_id       = var.vpc_id

  zone {
    zone_id    = relyt_dwsu_privatelink.privatelink_alicloud.zones[0]
    vswitch_id = var.vswitch_id
  }
}
