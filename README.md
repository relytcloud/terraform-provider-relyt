# Terraform Provider Relyt 

```
    terraform {
      required_providers {
        relyt = {
          source  = "relytcloud/relyt"
          version = "0.0.3"
        }  
      }
    }
    
    provider "relyt" {
      api_host = "https://<your-domain>" # Defaults to https://api.data.cloud (the original AWS control plane). Set it to your deployment's control plane address — ask your operator.
      auth_key = "<api_key>" # Copy the API key you obtained from the previous step.
      role     = "SYSTEMADMIN" # The system role for the Relyt cloud account to create, fixed to SYSTEMADMIN.
    }
    
    
    locals {
      cloud_id = {
        id = "aws"
      }
      region_id = {
        id = "<region_id>" # Set the ID the region in which the DW service unit will be created.
      }
      BASIC = {
        id = "basic"
      }
    }
    
    # Create a DW service unit and its default Hybrid DPS cluster.
      resource "relyt_dwsu" "dwsu_example" {
      cloud     = local.cloud_id.id
      region    = local.region_id.id
      domain    = "dwsu-example-tf" # The subdomain, customizable.
      alias     = "dwsu-example-test" # The alias of the DW service unit.
      default_dps = {
        name        = "hdps-test" # The name for the Hybrid DPS cluster, customizable.
        description = "The Hybrid DPS cluster" # A short description for the Hybrid DPS cluster, customizable and optional.
        engine      = "hybrid" # The type of the DPS cluster, fixed to hybrid.
        size        = "S" # The size of the Hybrid DPS cluster. Set it based on your needs.
    }
    }
    
    # Create an Extreme DPS cluster.
    resource "relyt_dps" "edps_example" {
      dwsu_id     = relyt_dwsu.dwsu_example.id
      name        = "edps1" # The name for the Extreme DPS cluster, customizable.
      description = "An Extreme DPS cluster" # A short description for the Extreme DPS cluster, customizable.
      engine      = "extreme" # The type of the DPS cluster, fixed to extreme.
      size        = "XS" # The size of the Extreme DPS cluster. Set it based on your needs.
    }
    
    # Create a DW user. You can repeat this code block to create multiple DW users.
    resource "relyt_dwuser" "user1" {
      dwsu_id          = relyt_dwsu.dwsu_example.id
      account_name     = "user1" # Name the DW user to create.
      account_password = "Qwer123!" # The password for the DW user, which must be 8 to 32 characters in length and contain at least one uppercase letter, one lowercase letter, one digit, and one special character.
    
    
    # Other optional parameters
      datalake_identity                   = "anotherRole2" # The data lake identity: IAM role ARN on AWS, Unity Catalog user name on Alibaba Cloud (replaces datalake_aws_lakeformation_role_arn).
      async_query_result_location_prefix  = "simple"       # The prefix of the path to the S3 output location.
      async_query_result_location_role_arn = "anotherSimple" # The ARN of the role to access the output location: IAM role on AWS, RAM role (acs:ram::...) on Alibaba Cloud (replaces async_query_result_location_aws_role_arn).
    }
    
    
    data "relyt_dwsu_service_account" "tes" {
      dwsu_id = relyt_dwsu.dwsu_example.id
    }
```