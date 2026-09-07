## 1.6.0 (Unreleased)

FEATURES:

* Non-AWS deployments: `api_host` routes to any Relyt control plane, region
  endpoints are discovered per cloud, and Alibaba Cloud is supported end to end
  (#14, #15, re-landed in #19).

IMPROVEMENTS:

* `relyt_dwuser`: cloud-neutral names for the two attributes that carried AWS
  vocabulary but are used on every cloud. The old names keep working as
  deprecated aliases; configure one name or the other, not both (#21).

  | Deprecated name                              | Use instead                             |
  | :------------------------------------------- | :-------------------------------------- |
  | `datalake_aws_lakeformation_role_arn`        | `datalake_identity`                     |
  | `async_query_result_location_aws_role_arn`   | `async_query_result_location_role_arn`  |

  On AWS `datalake_identity` is the Lake Formation IAM role ARN and the async role
  is an IAM role ARN; on Alibaba Cloud they are the Unity Catalog user name and a
  RAM role ARN (`acs:ram::<account>:role/<name>`). `async_query_result_location_prefix`
  stays `s3://…` on both clouds.
* `relyt_dwuser`: descriptions now state what to fill in on AWS and on Alibaba Cloud.

BUG FIXES:

* `relyt_dwsu_external_schema`: `table_format` no longer reports drift on every
  plan. The service stores the value lower-cased; `Read` now keeps the spelling
  in state when the two differ only in case, and a configuration change that
  differs only in case is not planned as a change (#23).

NOTES:

* Release process: rc tags stay GitHub pre-releases on purpose and are not
  published to the Terraform Registry. The registry only ingests regular
  releases, and it treats the highest ingested version as the provider's default
  even when it is a pre-release; `1.6.0-rc4` briefly became the registry default
  on 2026-09-07 after its pre-release flag was removed by hand. Test rc builds
  from the GitHub release assets through a filesystem mirror (see RELEASING.md).
* Switching an existing configuration from a deprecated name to the new one is a
  no-op plan for `datalake_identity`. For the async role it shows a one-time
  in-place update that re-sends the same value; imported users default to the new
  name in state.

## 1.5.1 (May 21, 2026)

FEATURES:

* **New Resource:** `relyt_dwsu_entraid_config` — configures Microsoft Entra ID SSO
  for a DWSU, with an optional `dms_host` attribute for air-gapped deployments (#13)

IMPROVEMENTS:

* Fall back to the `web_console` endpoint when a deployment exposes no `openapi`
  endpoint type
* docs: Entra ID SSO getting-started guide, in Chinese and English

NOTES:

* The Alibaba Cloud migration changes from #14 and #15 were reverted in #17, so
  `main` is content-identical to the v1.5.1 tag.

## 0.0.1 (Unreleased)

FEATURES:
