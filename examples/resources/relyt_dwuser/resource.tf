
resource "relyt_dwuser" "user1" {
  dwsu_id                              = "dwsu-id-from-an-dwsu-resource"
  account_name                         = "UniqueAccountName"
  account_password                     = "daf#$dgdfe&Abce%64"
  datalake_identity                    = "arn:aws:iam::123456789012:role/lake-r1" # AWS: Lake Formation IAM role ARN. Alibaba Cloud: Unity Catalog user name. Replaces the deprecated datalake_aws_lakeformation_role_arn.
  async_query_result_location_prefix   = "s3=//bucket-name/prefix/..."
  async_query_result_location_role_arn = "arn:aws:iam::123456789012:role/async-results" # AWS: IAM role ARN. Alibaba Cloud: RAM role ARN (acs:ram::...). Set together with the prefix. Replaces the deprecated async_query_result_location_aws_role_arn.
}