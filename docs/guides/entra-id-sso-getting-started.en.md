---
page_title: "Configure Microsoft Entra ID SSO with Terraform"
subcategory: "Guides"
description: |-
  Zero-to-working setup for managing a DWSU's Microsoft Entra ID (Azure AD) single sign-on configuration via Terraform.
---

# Configure Microsoft Entra ID SSO with Terraform

This guide is for **DWSU admins / SREs / platform engineers**. The goal: get Microsoft single sign-on working on the DMS Console via Terraform in about 15 minutes.

No prior Terraform experience required — concepts are introduced as we use them.

---

## 1. What is this & why use it

**Microsoft Entra ID** (formerly Azure AD) is Microsoft's enterprise identity service. After configuring Entra ID SSO:

- Employees no longer need a separate DMS username/password.
- They click "Sign in with Microsoft" on the DMS Console → Microsoft login page → SAML/OAuth flow → automatically signed into DMS.
- When an employee is offboarded in Azure, their DMS access is cut off immediately (no separate cleanup on the DMS side).

The `relyt_dwsu_entraid_config` Terraform resource turns the "DMS-side SSO config" (i.e. the `PUT /api/entraid-config` call) into a declarative resource, giving you:

- **Reproducible**: config lives in git; one template across prod / staging / dev.
- **Auditable**: every change has a git diff and PR trail.
- **Reversible**: `terraform destroy` (or reverting a commit) turns SSO off.
- **Drift-detected**: if someone hand-edits the backend via curl, the next `terraform plan` flags it.

> ⚠️ This resource only manages the **DMS-side** registration. The Azure-side App Registration / App Roles / user assignments still need to be done in the Azure portal (or with the `azuread` provider separately). This guide walks you through both sides.

---

## 2. End-to-end flow

```
┌──────────────────────────┐           ┌──────────────────────────┐
│   Azure (manual)         │           │   Terraform (this guide) │
│                          │           │                          │
│ 1. Create App Reg.       │           │ 4. provider config       │
│ 2. Define App Roles      │  ─────▶   │    (auth_key + role)     │
│    (value = DMS user)    │           │ 5. relyt_dwsu_entraid_   │
│ 3. Assign Azure users    │           │    config resource       │
│                          │           │ 6. terraform apply       │
└──────────────────────────┘           └──────────────────────────┘
              │                                      │
              │       tenant_id / client_id          │
              └─────────────────┬────────────────────┘
                                ▼
                       ┌────────────────────┐
                       │ DMS Console        │
                       │ "Sign in w/ MS"    │ ◀── End user
                       └────────────────────┘
```

Three roles involved:

| Role | What they do |
|---|---|
| **Azure admin** | Steps 1-3: register app, define roles, assign users in Azure portal |
| **DWSU admin (you)** | Steps 4-6: the main subject of this guide |
| **End user** | Logs in with a Microsoft account |

---

## 3. Prerequisites checklist

Before running any of the commands below, confirm you have:

| # | Item | How to verify |
|---|---|---|
| 1 | Azure AD tenant admin rights | Can log into [Azure portal](https://portal.azure.com) and create App registrations |
| 2 | An existing DWSU | Visible in DMS Console, or queryable via `data "relyt_dwsus" "all" {}` |
| 3 | An **ACCOUNTADMIN** user on the DWSU | Their SecretKey will be used for all PUT/DELETE calls in this guide |
| 4 | At least one regular DMS user already exists (the one you want to enable SSO for) | DMS Console → Users. SSO will not auto-create DMS users. |
| 5 | Terraform CLI installed locally | `terraform version` reports ≥ 1.0 |
| 6 | The Relyt provider installed | See [provider installation](https://registry.terraform.io/providers/relytcloud/relyt) |

> Item 3 is the most common pitfall: **ACCOUNTADMIN ≠ SYSTEMADMIN**. These are orthogonal roles. SYSTEMADMIN is cloud-level (used to create DWSUs); ACCOUNTADMIN is DWSU-internal account admin (used to manage SSO, user security policies, etc., on that DWSU). A SYSTEMADMIN-only key will be rejected by the backend for this resource.

---

## 4. Azure-side setup (one-time)

### 4.1 Create the App Registration

1. Sign into [Azure portal](https://portal.azure.com)
2. Go to **Microsoft Entra ID** → **App registrations** → **+ New registration**
3. Fill in:
   - **Name**: `relyt-dms-<dwsu-name>` (example naming, your choice)
   - **Supported account types**: `Accounts in this organizational directory only` (corresponds to `tenant_type = "single"` in Terraform)
   - **Redirect URI**:
     - Choose **Single-page application (SPA)**
     - Set the value to the DMS frontend callback URL, e.g. `https://<dwsu-domain>/console/auth/entra-id/callback`
4. Click **Register**

### 4.2 Copy Tenant ID + Client ID

Open the newly created app → **Overview**:

- **Directory (tenant) ID** → copy it; in Terraform this is `tenant_id`
- **Application (client) ID** → copy it; in Terraform this is `client_id`

### 4.3 Define App Roles (one role per DMS user)

App registration → **App roles** → **+ Create app role**:

| Field | Value |
|---|---|
| Display name | `DMS - <username>`, e.g. `DMS - wenbo` |
| Allowed member types | `Users/Groups` |
| **Value** | `wenbo` ← **must exactly match an existing DMS username** |
| Description | `Login as DMS user wenbo` |
| Enabled | ☑ |

> ⚠️ Once issued, an App Role's `Value` is hard to change. One App Role maps to one DMS identity. To enable SSO for multiple DMS users, create multiple App Roles, each with a Value matching a DMS username.

### 4.4 Assign Azure users to the App Role

**Enterprise applications** → find the app you just created → **Users and groups** → **+ Add user/group**:

1. Pick the Azure user (e.g. `wenbo@yourcompany.com`)
2. Pick the role (e.g. `DMS - wenbo`)
3. Click **Assign**

> ⚠️ The same Azure user can be assigned **only one** DMS-related role per application. Multiple assignments will be rejected at login.

---

## 5. Relyt side: configure SSO with Terraform

### 5.1 First-time minimal config

Create a new file `main.tf`:

```hcl
terraform {
  required_providers {
    relyt = {
      source  = "relytcloud/relyt"
      version = ">= 1.5.0"   # earliest version that ships the entraid_config resource
    }
  }
}

provider "relyt" {
  # auth_key is the SecretKey of an ACCOUNTADMIN user.
  # You can also omit this line and use the RELYT_AUTH_KEY env var instead.
  auth_key = "<accountadmin-secret-key>"
  role     = "ACCOUNTADMIN"
}

# Reference an existing DWSU. If you only have one, [0] is enough; for multiple
# DWSUs, iterate .dwsu_list[*] and filter.
# If you already know the dwsu_id, you can hardcode it directly.
data "relyt_dwsus" "all" {}

resource "relyt_dwsu_entraid_config" "sso" {
  dwsu_id   = data.relyt_dwsus.all.dwsu_list[0].id
  tenant_id = "<paste-your-tenant-id-from-step-4.2>"
  client_id = "<paste-your-client-id-from-step-4.2>"

  # These two are optional (defaults shown):
  # tenant_type = "single"
  # enabled     = true
}
```

Field reference:

| Field | Required | Description |
|---|---|---|
| `dwsu_id` | ✓ | Target DWSU's ID. **Changing it triggers destroy-and-recreate** (one entraid-config binds to one DWSU) |
| `tenant_id` | ✓ | Azure tenant GUID (from step 4.2) |
| `client_id` | ✓ | Azure application GUID (from step 4.2) |
| `tenant_type` | ✗ | `"single"` allows only users in the configured tenant (recommended); `"multi"` allows users from any Azure org. Default `"single"` |
| `enabled` | ✗ | `true` activates SSO; `false` retains config but blocks SSO logins. Default `true` |

### 5.2 Run apply

```bash
# Initialize (downloads the provider)
terraform init

# See what would happen, without making changes
terraform plan
```

Expected output:

```
Terraform will perform the following actions:

  # relyt_dwsu_entraid_config.sso will be created
  + resource "relyt_dwsu_entraid_config" "sso" {
      + client_id   = "00000000-..."
      + dwsu_id     = "dwsu-abc123"
      + enabled     = true
      + tenant_id   = "11111111-..."
      + tenant_type = "single"
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

If it looks right, apply:

```bash
terraform apply
# type yes to confirm
```

### 5.3 Verify the config landed

```bash
# Curl the backend directly; should see what you just PUT
DMS=$(terraform output -raw dmsHost 2>/dev/null || echo "https://api-<your-dwsu-domain>")
curl -s $DMS/api/entraid-config | python3 -m json.tool
```

Expected response contains your `tenantId` / `clientId` / `tenantType` / `enabled: true`.

Then try the Microsoft authorization URL:

```bash
curl -s $DMS/api/auth/entra-id/authorize | python3 -m json.tool
```

Expected: `data.authorizationUri` is not null.

Finally, the **real login test**: open the DMS Console in a browser. A "Sign in with Microsoft" button should appear. Click it → Microsoft login page → sign in with the Azure user assigned in step 4.4 → bounced back into the DMS console, signed in.

---

## 6. Common operations

### 6.1 Change tenant_id / client_id (rebind to a new Azure app)

Just edit the values in `.tf` and `terraform apply`:

```hcl
resource "relyt_dwsu_entraid_config" "sso" {
  dwsu_id   = ...
  tenant_id = "new tenant id"   # ← change here
  client_id = "new client id"   # ← change here
}
```

Plan shows:

```
  ~ resource "relyt_dwsu_entraid_config" "sso" {
      ~ client_id = "old" -> "new"
      ~ tenant_id = "old" -> "new"
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

The `~ change` is an in-place update (PUT), not a destroy-and-recreate.

### 6.2 Temporarily disable SSO (keep the config)

```hcl
resource "relyt_dwsu_entraid_config" "sso" {
  # ... other fields unchanged
  enabled = false   # ← add this
}
```

After `apply`, the "Sign in with Microsoft" button is hidden on the DMS Console, but the config (tenant/client) remains on the backend. To re-enable, remove the `enabled = false` line (restoring the default `true`) and apply again.

### 6.3 Fully remove SSO

**Option A** (recommended): delete the `resource "relyt_dwsu_entraid_config" "sso" { ... }` block entirely. `terraform apply` will call DELETE.

**Option B**: tear down the whole config:

```bash
terraform destroy
```

The backend's entraid-config is cleared; the DMS Console no longer shows the Microsoft login entry.

### 6.4 Import an existing SSO config into Terraform management

Scenario: you previously configured SSO via curl (per the internal SSO setup guide), and now want to manage it through Terraform.

**Step 1**: Add a placeholder resource block in `.tf` (the placeholder values will be overwritten after import):

```hcl
resource "relyt_dwsu_entraid_config" "sso" {
  dwsu_id   = "dwsu-abc123"
  tenant_id = "placeholder"
  client_id = "placeholder"
}
```

**Step 2**:

```bash
terraform import relyt_dwsu_entraid_config.sso dwsu-abc123
#                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^
#                TF resource address                         import ID = dwsu_id
```

**Step 3**: `terraform plan`. The plan will likely show diffs between your `.tf` placeholders and the actual backend values. Update the `.tf` to match reality (or vice versa); plan again should show `no changes`.

---

## 6.5 Air-gapped / DMS-only deployments (use `dms_host` to skip the control plane)

**Scenario**: your Relyt environment only deploys DMS, with no GCS (control-plane service), or the control plane is not reachable from where Terraform runs. By default, the resource calls `control plane GetDwsu(dwsu_id) → Endpoints[openapi].URI` to find the host — if the control plane isn't reachable, this step hangs or fails.

Solution: pass `dms_host` directly, bypassing the control-plane lookup:

```hcl
resource "relyt_dwsu_entraid_config" "sso" {
  dwsu_id   = "h713"                          # any non-empty string (used only for log correlation)
  dms_host  = "http://<dms-pod-ip-or-host>"   # points directly at the DMS API host
  tenant_id = "..."
  client_id = "..."
  # other fields unchanged
}
```

Behavior:

- `dms_host` **unset** (default): existing logic — call `GET /dwsu/<id>` on the control plane to resolve the host.
- `dms_host` **set**: use the value directly; skip the control-plane call.

Use cases: single-tenant private deployments, CI dev clusters, throwaway test environments.

## 7. Advanced: provider alias for multiple roles

**Scenario**: you want to manage both DWSU lifecycle (requires SYSTEMADMIN) and SSO config (requires ACCOUNTADMIN) in the same `.tf`.

Terraform supports multiple configured instances of the same provider via **provider alias**:

```hcl
provider "relyt" {
  alias    = "system"
  auth_key = "<systemadmin-secret-key>"
  role     = "SYSTEMADMIN"
}

provider "relyt" {
  alias    = "account"
  auth_key = "<accountadmin-secret-key>"
  role     = "ACCOUNTADMIN"
}

resource "relyt_dwsu" "dw" {
  provider = relyt.system        # ← explicitly use the system alias
  cloud  = "aws"
  region = "ap-southeast-1"
  domain = "my-dw"
  default_dps = {
    name   = "hdps"
    engine = "hybrid"
    size   = "S"
  }
}

resource "relyt_dwsu_entraid_config" "sso" {
  provider = relyt.account       # ← explicitly use the account alias
  dwsu_id  = relyt_dwsu.dw.id    # TF auto-detects the dependency: DWSU first, then SSO
  tenant_id = "..."
  client_id = "..."
}
```

> Add `provider = relyt.<alias>` inside a resource block to route that resource's API calls to that specific provider config. Resources express dependencies via references like `relyt_dwsu.dw.id`, and Terraform builds the dependency graph automatically.

---

## 8. Troubleshooting cheatsheet

Look up by the error message you see:

| Error message | Cause | Fix |
|---|---|---|
| `HTTP 403 Permission denied: Account admin role required to manage SSO config` | `auth_key` is not from an ACCOUNTADMIN user | Use the ACCOUNTADMIN's SecretKey, or set up a provider alias |
| `error fetching DWSU ... DWSU not found` | `dwsu_id` is wrong, or the DWSU has been deleted | Double-check `dwsu_id`; list with `data "relyt_dwsus"` |
| `openapi endpoint not found on DWSU` | `DwsuModel.Endpoints[]` doesn't contain a `Type==openapi` entry | Likely a deployment-side config issue — contact Relyt support |
| `No 'roles' claim in token` (end-user login) | The Azure user is not assigned an App Role | Azure portal → Enterprise applications → Users and groups → assign the role |
| `Multiple app roles assigned: [a, b]` | The same Azure user has multiple DMS-related roles | Remove the extras; keep exactly one |
| `DMS user not found for role: xxx` | The App Role's Value doesn't match any DMS username | Create the user in DMS, or fix the App Role's Value |
| Browser: `AADSTS50011: ...redirect URI...` | The Azure app hasn't registered the current callback URI | App registration → Authentication → add the URI |
| Plan wants to destroy-and-recreate but you didn't change `dwsu_id` | Rare — usually state has diverged badly from `.tf` | `terraform refresh` → `terraform plan`; consider `terraform import` |

If none of the above match:

1. Run `TF_LOG=DEBUG terraform apply` to see the full HTTP requests/responses.
2. Check DMS pod logs: `kubectl -n <ns> logs deploy/hdpscluster-<id>-dms --since=10m | grep -iE "entra|verify|token|roles"`
3. Contact Relyt support with the full plan/apply output.

---

## 9. FAQ

**Q1: Can one SSO config cover multiple DWSUs?**
No. `entraid-config` is per-DWSU; each DWSU needs its own config. A common pattern is Terraform `for_each`:

```hcl
locals {
  dwsus_with_sso = {
    "prod"    = { dwsu_id = relyt_dwsu.prod.id,    tenant_id = "...", client_id = "..." }
    "staging" = { dwsu_id = relyt_dwsu.staging.id, tenant_id = "...", client_id = "..." }
  }
}

resource "relyt_dwsu_entraid_config" "sso" {
  for_each  = local.dwsus_with_sso
  dwsu_id   = each.value.dwsu_id
  tenant_id = each.value.tenant_id
  client_id = each.value.client_id
}
```

**Q2: Is `tenant_type = "multi"` safe?**
`multi` lets users from any Azure organization initiate the login flow against your app. Whether they can actually log into DMS still depends on App Role assignment — anyone without a role will pass Microsoft auth but be rejected by DMS. **However**, cross-tenant requires Azure-side multi-tenant to be enabled, or Microsoft will block first. Unless you have a real cross-org collaboration need, `single` is the safer default.

**Q3: If I change App Roles / user assignments on the Azure side, do I need to apply Terraform?**
No. This resource only manages the DMS-side record of tenant_id/client_id/enabled. The Azure-side App Roles are outside its scope. **However**, if you want to manage the Azure side via Terraform too, look at the `hashicorp/azuread` provider — separate concern.

**Q4: After destroy, are already-logged-in users kicked out?**
**No**. Already-issued access tokens stay valid until they expire. Destroy only blocks **new** Microsoft logins. To kick users immediately, you need a separate "revoke token" operation — out of scope here.

**Q5: Can I put `auth_key` directly in `.tf`?**
**Strongly discouraged**. `.tf` usually ends up in git; a leaked ACCOUNTADMIN key is equivalent to a leaked password. Recommended alternatives:

- Use an env var: `export RELYT_AUTH_KEY="..."` and omit `auth_key` in the provider block.
- Inject via Terraform Cloud / Vault / Doppler / etc.

**Q6: What if `apply` hangs or times out?**
Under the hood the resource uses `CommonRetry` (5 attempts × 1s backoff), with a per-HTTP-call timeout from the provider's `client_timeout` (default 10s). If it keeps hanging:

- Check that the DWSU is in `READY` state.
- Check that the DWSU's openapi endpoint is reachable: `curl <openapi-uri>/api/entraid-config`.
- Run `TF_LOG=DEBUG terraform apply` to see exactly where it's stuck.

---

## 10. Full main.tf reference template

```hcl
terraform {
  required_providers {
    relyt = {
      source  = "relytcloud/relyt"
      version = ">= 1.5.0"
    }
  }
}

provider "relyt" {
  # Recommended: inject via env var RELYT_AUTH_KEY, do not hardcode
  role = "ACCOUNTADMIN"
}

# Reference existing DWSUs
data "relyt_dwsus" "all" {}

locals {
  # Enable SSO on every DWSU; for a subset, replace this with an explicit map
  target_dwsus = {
    for d in data.relyt_dwsus.all.dwsu_list :
    d.alias => d.id
  }
}

resource "relyt_dwsu_entraid_config" "sso" {
  for_each = local.target_dwsus

  dwsu_id   = each.value
  tenant_id = "<your-azure-tenant-id>"
  client_id = "<your-azure-client-id>"

  tenant_type = "single"
  enabled     = true
}

output "configured_dwsus" {
  value = [for k, v in relyt_dwsu_entraid_config.sso : k]
}
```

---

## 11. Next steps

- Full resource reference: see [`relyt_dwsu_entraid_config`](../resources/dwsu_entraid_config.md)
- Detailed Azure-side setup: consult the internal Relyt SSO setup guide
- Want to manage the Azure side with Terraform too? Look at the [`hashicorp/azuread`](https://registry.terraform.io/providers/hashicorp/azuread/latest/docs) provider
