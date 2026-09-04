
resource "relyt_dwuser" "user1" {
  dwsu_id          = "dwsu-id-from-an-dwsu-resource"
  account_name     = "user1@example.com"
  account_password = "daf#$dgdfe&Abce%64"

  # Identity used against the external data lake catalog.
  #   AWS:           Lake Formation IAM role ARN, e.g. "arn:aws:iam::123456789012:role/lake-r1"
  #   Alibaba Cloud: Unity Catalog user name,     e.g. "analyst@example.com"
  # (replaces the deprecated datalake_aws_lakeformation_role_arn)
  datalake_identity = "arn:aws:iam::123456789012:role/lake-r1"

  # Asynchronous query results: s3:// prefix (also on Alibaba Cloud / OSS) plus the
  # role the engine assumes to write there. Set both or neither.
  #   AWS:           "arn:aws:iam::123456789012:role/async-results"
  #   Alibaba Cloud: "acs:ram::1234567890123456:role/async-results"
  # (replaces the deprecated async_query_result_location_aws_role_arn)
  async_query_result_location_prefix   = "s3://bucket-name/prefix/"
  async_query_result_location_role_arn = "arn:aws:iam::123456789012:role/async-results"
}