

variable "endpoint" {
  type        = string
  description = "privatelink endpoint"
  default = "http://pl-4679805844736-api-d3410a34c78f4386.elb.us-east-1.amazonaws.com:8180"
}

variable "access_key" {
  type        = string
  description = "access_key"


  default = "AK8DoEFMRPWBGG0eY1JyNBVj7OnrTO3B6t3uJFyibDcGwz56HrAlg8uKtxf9hQeoHphJzOw"
}

variable "secret_key" {
  type        = string
  description = "secret_key"

  default = "HHJU4NBSLKZVGKTGRM41FCLGZVH4VPWS"
}