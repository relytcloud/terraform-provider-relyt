variable "role" {
  type        = string
  default     = "SYSTEMADMIN"
  description = "Role"
}

variable "dwsu_id" {
  type        = string
  description = "DWSU Id"
}

variable "account_name" {
  type        = string
  default     = "user1@zbyte-inc.com"
  description = "Account name"
}

variable "account_password" {
  type        = string
  default     = "User1123."
  description = "Account password"
}

variable "datalake_identity" {
  type        = string
  default     = "arn:aws:iam::905418298243:role/lake-r1"
  description = "Identity used against the external data lake catalog: Lake Formation IAM role ARN on AWS, Unity Catalog user name on Alibaba Cloud"
}

variable "async_query_result_location_prefix" {
  type        = string
  default     = "s3://relytqaresult-us-east-1/user1/result1/"
  description = "s3:// prefix the engine writes async query results to (also s3:// on Alibaba Cloud / OSS)"
}

variable "async_query_result_location_role_arn" {
  type        = string
  default     = "arn:aws:iam::905418298243:role/lake-r1"
  description = "Role the engine assumes to write async query results: IAM role ARN on AWS, RAM role ARN (acs:ram::...) on Alibaba Cloud"
}
