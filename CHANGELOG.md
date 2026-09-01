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
