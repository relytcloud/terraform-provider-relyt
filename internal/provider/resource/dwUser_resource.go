package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	tfModel "terraform-provider-relyt/internal/provider/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dwUserResource{}
	_ resource.ResourceWithConfigure   = &dwUserResource{}
	_ resource.ResourceWithImportState = &dwUserResource{}
)

// Alias handling for the cloud-neutral attribute names, see ValidateConfig and ModifyPlan.
var (
	_ resource.ResourceWithValidateConfig = &dwUserResource{}
	_ resource.ResourceWithModifyPlan     = &dwUserResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewdwUserResource() resource.Resource {
	return &dwUserResource{}
}

// orderResource is the resource implementation.
type dwUserResource struct {
	RelytClientResource
	//client *client.RelytClient
}

// Metadata returns the resource type name.
func (r *dwUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwuser"
}

// Schema defines the schema for the resource.
func (r *dwUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"dwsu_id":          schema.StringAttribute{Required: true, Description: "The ID of the service unit."},
			"id":               schema.StringAttribute{Computed: true, Description: "The ID of the DW user.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"account_name":     schema.StringAttribute{Required: true, Description: "The name of the DW user, which is unique in the instance. The name is the email address."},
			"account_password": schema.StringAttribute{Required: true, Description: "initPassword"},
			// datalake_identity and datalake_aws_lakeformation_role_arn are one value under
			// two names: the cloud-neutral name is the one to use, the AWS-named one stays
			// as a deprecated alias so existing configurations keep working. Both are
			// Computed so that omitting them from the configuration keeps whatever the
			// server holds. On Alibaba Cloud this field stores the Unity Catalog user
			// mapping (engine: set_user_role_arn), which an administrator sets out of
			// band; without Computed, every apply that omits it would delete it.
			// Set it to "" to remove the binding on purpose. Read writes the server
			// value into both names, so switching a configuration from the old name to
			// the new one is a no-op plan.
			// The async result role also has a cloud-neutral name plus a deprecated
			// AWS-named alias. Prefix and role are one setting: both must be present to
			// configure it, and omitting both clears it on the server, so they are not
			// Computed ("omitted" has to stay visible in the plan).
			"datalake_identity":                        schema.StringAttribute{Optional: true, Computed: true, Description: "The identity this DW user presents to the external data lake catalog. AWS: the ARN of the cross-account IAM role used with Lake Formation, e.g. `arn:aws:iam::123456789012:role/lake-r1`. Alibaba Cloud: the Unity Catalog user name to map this DW user to, e.g. `analyst@example.com`. Omit to leave the server value untouched; set to \"\" to remove the binding. Replaces the deprecated `datalake_aws_lakeformation_role_arn`; configure only one of the two."},
			"datalake_aws_lakeformation_role_arn":      schema.StringAttribute{Optional: true, Computed: true, DeprecationMessage: "Use `datalake_identity` instead. This AWS-named attribute is kept as an alias for existing configurations and will be removed in a future major release.", Description: "Deprecated alias of `datalake_identity`, same value and behaviour. Omit to leave the server value untouched; set to \"\" to remove the binding."},
			"async_query_result_location_prefix":       schema.StringAttribute{Optional: true, Description: "The prefix of the path to the S3 output location."},
			"async_query_result_location_role_arn":     schema.StringAttribute{Optional: true, Description: "ARN of the role the engine assumes to write asynchronous query results into the prefix. AWS: an IAM role ARN, e.g. `arn:aws:iam::123456789012:role/async-results`. Alibaba Cloud: a RAM role ARN, e.g. `acs:ram::1234567890123456:role/async-results`, whose trust policy allows the Relyt platform identity with this DWSU's external ID. Must be set together with `async_query_result_location_prefix`; omitting both clears the setting. Replaces the deprecated `async_query_result_location_aws_role_arn`; configure only one of the two."},
			"async_query_result_location_aws_role_arn": schema.StringAttribute{Optional: true, DeprecationMessage: "Use `async_query_result_location_role_arn` instead. This AWS-named attribute is kept as an alias for existing configurations and will be removed in a future major release.", Description: "Deprecated alias of `async_query_result_location_role_arn`, same value and behaviour."},
		},
	}
}

// ValidateConfig rejects a configuration that sets an attribute and its
// deprecated alias to different values; the two names are one setting.
func (r *dwUserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg tfModel.DWUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, _, err := pickAlias(cfg.DatalakeIdentity, cfg.DatalakeAwsLakeformationRoleArn); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("datalake_identity"), "Conflicting attribute values",
			"datalake_identity and its deprecated alias datalake_aws_lakeformation_role_arn are the same setting, but "+err.Error()+". Configure only datalake_identity.")
	}
	if _, _, err := pickAlias(cfg.AsyncQueryResultLocationRoleArn, cfg.AsyncQueryResultLocationAwsRoleArn); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("async_query_result_location_role_arn"), "Conflicting attribute values",
			"async_query_result_location_role_arn and its deprecated alias async_query_result_location_aws_role_arn are the same setting, but "+err.Error()+". Configure only async_query_result_location_role_arn.")
	}
}

// ModifyPlan keeps the two Computed names of the data lake identity in step while
// planning. Terraform marks a Computed attribute that is absent from the
// configuration as "(known after apply)" whenever anything else on the resource
// changes; for a name that only mirrors its sibling that is noise. The configured
// name drives and the other one takes the same planned value. When neither is
// configured the prior state is kept, matching handleAccountConfig, which then
// leaves the server value alone.
func (r *dwUserResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return // destroy
	}
	var cfg, plan tfModel.DWUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfgNew, cfgOld := !cfg.DatalakeIdentity.IsNull(), !cfg.DatalakeAwsLakeformationRoleArn.IsNull()
	switch {
	case cfgNew && cfgOld:
		return // both written; ValidateConfig guarantees they agree
	case cfgNew:
		plan.DatalakeAwsLakeformationRoleArn = plan.DatalakeIdentity
	case cfgOld:
		plan.DatalakeIdentity = plan.DatalakeAwsLakeformationRoleArn
	default:
		if req.State.Raw.IsNull() {
			return // create with nothing configured: stays unknown, apply writes null
		}
		var state tfModel.DWUserModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.DatalakeIdentity = state.DatalakeIdentity
		plan.DatalakeAwsLakeformationRoleArn = state.DatalakeAwsLakeformationRoleArn
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// Create a new resource.
func (r *dwUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from dwUserModel
	var dwUserModel tfModel.DWUserModel
	diags := req.Plan.Get(ctx, &dwUserModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The configuration, not the plan, tells which of the aliased names the user
	// wrote; the plan carries Computed placeholders for the other one.
	var cfg tfModel.DWUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	meta := common.RouteRegionUri(ctx, dwUserModel.DwsuId.ValueString(), r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	relytAccount := client.Account{
		InitPassword: dwUserModel.AccountPassword.ValueString(),
		Name:         dwUserModel.AccountName.ValueString(),
	}
	// Create new order
	createResult, err := r.client.CreateAccount(ctx, regionUri, dwUserModel.DwsuId.ValueString(), relytAccount)
	if err != nil || createResult.Code != 200 {
		resp.Diagnostics.AddError(
			"Error creating dwuser",
			"Could not create dwuser, unexpected error: "+err.Error(),
		)
		return
	}
	dwUserModel.ID = types.StringValue(relytAccount.Name)
	r.handleAccountConfig(ctx, &dwUserModel, &cfg, regionUri, &resp.Diagnostics)
	// State goes in after handleAccountConfig so the normalized (never unknown)
	// lakeformation value lands in it. Written even when the config calls failed:
	// the account exists on the server, so dropping it from state orphans it.
	// No rollback on failure: the account exists on the server, the user decides.
	resp.Diagnostics.Append(resp.State.Set(ctx, &dwUserModel)...)
}

// Read resource information.
func (r *dwUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state tfModel.DWUserModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.ID.IsNull() {
		resp.Diagnostics.AddError("can't read dwuser", "dwuser id is nil")
		return
	}
	//state.ID = state.AccountName
	meta := common.RouteRegionUri(ctx, state.DwsuId.ValueString(), r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := common.CommonRetry(ctx, func() (*client.AsyncResult, error) {
		return r.client.GetAsyncAccountConfig(ctx, meta.URI, state.DwsuId.ValueString(), state.ID.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error get dwuser asyncAccountConfig",
			"Could not get config asyncAccountConfig, unexpected error: "+err.Error(),
		)
		return
	}
	if config != nil {
		if config.S3LocationPrefix != "" {
			state.AsyncQueryResultLocationPrefix = types.StringValue(config.S3LocationPrefix)
		}
		if config.AwsIamArn != "" {
			fillAsyncRoleState(&state, config.AwsIamArn)
		}

	}

	lakeInfo, err := common.CommonRetry(ctx, func() (*client.LakeFormation, error) {
		return r.client.GetLakeFormationConfig(ctx, meta.URI, state.DwsuId.ValueString(), state.ID.ValueString())
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error get dwuser LakeFormationConfig",
			"Could not get config LakeFormationConfig, unexpected error: "+err.Error(),
		)
		return
	}
	if lakeInfo != nil && lakeInfo.IAMRole != "" {
		// Both names are Computed, so writing the server value into both keeps the
		// plan clean whichever one the configuration uses.
		state.DatalakeIdentity = types.StringValue(lakeInfo.IAMRole)
		state.DatalakeAwsLakeformationRoleArn = types.StringValue(lakeInfo.IAMRole)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	return
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dwUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//resp.Diagnostics.AddError("not support", "update account not supported")
	var plan tfModel.DWUserModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var cfg tfModel.DWUserModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// read old status
	var stat tfModel.DWUserModel
	req.State.Get(ctx, &stat)
	if resp.Diagnostics.HasError() {
		return
	}
	if stat.AccountName.ValueString() != plan.AccountName.ValueString() {
		resp.Diagnostics.AddError("not support", "can't update account name!")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	//plan.ID = plan.AccountName
	meta := common.RouteRegionUri(ctx, plan.DwsuId.ValueString(), r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI

	if stat.AccountPassword.ValueString() != plan.AccountPassword.ValueString() {
		//resp.Diagnostics.AddError("not support", "can't update init password!")
		_, err := common.CommonRetry(ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.PatchAccount(ctx, regionUri, plan.DwsuId.ValueString(), plan.ID.ValueString(), plan.AccountPassword.ValueString())
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed update password", " patch password failed with:"+err.Error())
			return
		}
	}

	tflog.Info(ctx, "accountId:"+plan.ID.ValueString())
	r.handleAccountConfig(ctx, &plan, &cfg, regionUri, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dwUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state tfModel.DWUserModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta := common.RouteRegionUri(ctx, state.DwsuId.ValueString(), r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI

	// Delete existing account
	_, err := common.CommonRetry(ctx, func() (*any, error) {
		err := r.client.DropAccount(ctx, regionUri, state.DwsuId.ValueString(), state.ID.ValueString())
		return nil, err
	})
	//err := r.client.DropAccount(ctx, regionUri, state.DwsuId.ValueString(), state.ID.ValueString())
	if err != nil {
		//要不要加error
		resp.Diagnostics.AddError(
			"Error Deleting dwuser",
			"Could not delete dwuser, unexpected error: "+err.Error(),
		)
	}
}

func (r *dwUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute

	// Retrieve import ID and save to id attribute
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: dwsu_id,account_name. Got: %q", req.ID),
		)
		return
	}
	dwsuId := idParts[0]
	accountName := idParts[1]

	meta := common.RouteRegionUri(ctx, dwsuId, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	account, err := common.CommonRetry(ctx, func() (*client.Account, error) {
		account, err := r.client.GetAccount(ctx, regionUri, dwsuId, accountName)
		if err != nil {
			return nil, err
		}
		return account.Data, nil
	})
	if err != nil || account == nil {
		msg := "account not found!"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("import account failed!", "error read account, "+msg)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("dwsu_id"), dwsuId)
	resp.State.SetAttribute(ctx, path.Root("id"), accountName)
	resp.State.SetAttribute(ctx, path.Root("account_name"), accountName)
	//password，not show
	resp.State.SetAttribute(ctx, path.Root("account_password"), types.StringValue(""))

}

// handleAccountConfig pushes the async result location and the data lake identity
// to the server and normalizes the model that becomes state. dwUserModel is the
// plan (what state must end up as); cfg is the raw configuration, consulted to
// learn which of the aliased attribute names the user actually wrote.
func (r *dwUserResource) handleAccountConfig(ctx context.Context, dwUserModel *tfModel.DWUserModel, cfg *tfModel.DWUserModel, regionUri string, diagnostics *diag.Diagnostics) {
	//dwUserModel.ID = dwUserModel.AccountName
	tflog.Info(ctx, fmt.Sprintf("=======uknown %t nil %t", dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsUnknown(), dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsNull()))
	//if dwUserModel.AsyncQueryResultLocationPrefix.IsUnknown() {
	//	dwUserModel.AsyncQueryResultLocationPrefix = types.StringNull()
	//}
	//if dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsUnknown() {
	//	dwUserModel.AsyncQueryResultLocationAwsRoleArn = types.StringNull()
	//}
	// --- async query result location: prefix + role, one setting under two role names ---
	roleVal, roleSet, err := pickAlias(cfg.AsyncQueryResultLocationRoleArn, cfg.AsyncQueryResultLocationAwsRoleArn)
	if err != nil {
		// ValidateConfig catches this earlier; kept as a guard for unknown values that
		// only resolve during apply.
		diagnostics.AddError("Error config dwuser", "async_query_result_location_role_arn / async_query_result_location_aws_role_arn: "+err.Error())
		return
	}
	prefix := cfg.AsyncQueryResultLocationPrefix
	prefixSet := !prefix.IsNull() && !prefix.IsUnknown()
	switch {
	case roleSet && prefixSet:
		asyncResult := client.AsyncResult{AwsIamArn: roleVal.ValueString(), S3LocationPrefix: prefix.ValueString()}
		_, err := common.CommonRetry[client.CommonRelytResponse[string]](ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.AsyncAccountConfig(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString(), asyncResult)
		})
		if err != nil {
			diagnostics.AddError(
				"Error config dwuser",
				"Could not config dwuser async, unexpected error: "+err.Error(),
			)
			//return
		}
		dwUserModel.AsyncQueryResultLocationPrefix = prefix
		applyAsyncRoleState(dwUserModel, cfg, roleVal)
	case !roleSet && !prefixSet:
		// Both omitted: clear the setting on the server, as before.
		_, err := common.CommonRetry[client.CommonRelytResponse[string]](ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.DeleteAsyncAccountConfig(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString())
		})
		if err != nil {
			diagnostics.AddError(
				"Error config dwuser",
				"Could not drop dwuser async config, unexpected error: "+err.Error(),
			)
			//return
		}
		dwUserModel.AsyncQueryResultLocationPrefix = types.StringNull()
		dwUserModel.AsyncQueryResultLocationRoleArn = types.StringNull()
		dwUserModel.AsyncQueryResultLocationAwsRoleArn = types.StringNull()
	default:
		//只有一个属性的时候报给用户异常
		diagnostics.AddError(
			"Error config dwuser",
			"Could not config dwuser async, arn and prefix should be set together",
		)
	}

	// --- data lake identity: one value under two Computed names ---
	lfVal, lfSet, err := pickAlias(cfg.DatalakeIdentity, cfg.DatalakeAwsLakeformationRoleArn)
	if err != nil {
		diagnostics.AddError("Error config dwuser", "datalake_identity / datalake_aws_lakeformation_role_arn: "+err.Error())
		return
	}
	if !lfSet {
		// Nothing configured: leave the server value alone (Read fills state from it),
		// but the Computed attributes have to be known once apply finishes.
		if dwUserModel.DatalakeIdentity.IsUnknown() {
			dwUserModel.DatalakeIdentity = types.StringNull()
		}
		if dwUserModel.DatalakeAwsLakeformationRoleArn.IsUnknown() {
			dwUserModel.DatalakeAwsLakeformationRoleArn = types.StringNull()
		}
		return
	}
	// Whichever name was configured, both carry the effective value in state.
	dwUserModel.DatalakeIdentity = lfVal
	dwUserModel.DatalakeAwsLakeformationRoleArn = lfVal
	lakeFormation := client.LakeFormation{IAMRole: lfVal.ValueString()}
	switch lakeFormationActionFor(lfVal) {
	case lakeFormationSkip:
		// leave the server value alone; Read fills state from it
	case lakeFormationDelete:
		_, err := common.CommonRetry[client.CommonRelytResponse[string]](ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.DeleteLakeFormationConfig(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString())
		})
		if err != nil {
			diagnostics.AddError(
				"Error config dwuser",
				"Could not delete dwuser lakeformation, unexpected error: "+err.Error(),
			)
		}
	case lakeFormationSet:
		_, err := common.CommonRetry[client.CommonRelytResponse[string]](ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.LakeFormationConfig(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString(), lakeFormation)
		})
		if err != nil {
			diagnostics.AddError(
				"Error config dwuser",
				"Could not config dwuser lakeformation, unexpected error: "+err.Error(),
			)
		}
	}
}

// pickAlias resolves one setting exposed under two attribute names: the
// cloud-neutral name (preferred) and its deprecated cloud-flavoured alias
// (legacy). It returns the configured value and whether either name was set.
// Unknown values count as not set — they only resolve during apply and cannot be
// compared. Setting both names to different values is a configuration error.
func pickAlias(preferred, legacy types.String) (types.String, bool, error) {
	pSet := !preferred.IsNull() && !preferred.IsUnknown()
	lSet := !legacy.IsNull() && !legacy.IsUnknown()
	switch {
	case pSet && lSet:
		if preferred.ValueString() != legacy.ValueString() {
			return types.StringNull(), false, fmt.Errorf("both are set with different values (%q vs %q)", preferred.ValueString(), legacy.ValueString())
		}
		return preferred, true, nil
	case pSet:
		return preferred, true, nil
	case lSet:
		return legacy, true, nil
	default:
		return types.StringNull(), false, nil
	}
}

// applyAsyncRoleState writes the effective async result role into the model
// under the name(s) the configuration used. The role attributes are not Computed,
// so state must mirror the configuration exactly: the name that was configured
// gets the value, the other one stays null.
func applyAsyncRoleState(model, cfg *tfModel.DWUserModel, role types.String) {
	if !cfg.AsyncQueryResultLocationRoleArn.IsNull() && !cfg.AsyncQueryResultLocationRoleArn.IsUnknown() {
		model.AsyncQueryResultLocationRoleArn = role
	} else {
		model.AsyncQueryResultLocationRoleArn = types.StringNull()
	}
	if !cfg.AsyncQueryResultLocationAwsRoleArn.IsNull() && !cfg.AsyncQueryResultLocationAwsRoleArn.IsUnknown() {
		model.AsyncQueryResultLocationAwsRoleArn = role
	} else {
		model.AsyncQueryResultLocationAwsRoleArn = types.StringNull()
	}
}

// fillAsyncRoleState writes the role read back from the server into the name(s)
// the state already tracks, so a configuration that keeps using the deprecated
// name sees no diff. When neither name is tracked yet (fresh import), the
// cloud-neutral name is used.
func fillAsyncRoleState(state *tfModel.DWUserModel, role string) {
	v := types.StringValue(role)
	newTracked := !state.AsyncQueryResultLocationRoleArn.IsNull()
	oldTracked := !state.AsyncQueryResultLocationAwsRoleArn.IsNull()
	switch {
	case newTracked && oldTracked:
		state.AsyncQueryResultLocationRoleArn = v
		state.AsyncQueryResultLocationAwsRoleArn = v
	case oldTracked:
		state.AsyncQueryResultLocationAwsRoleArn = v
	default:
		state.AsyncQueryResultLocationRoleArn = v
	}
}

type lakeFormationAction int

const (
	lakeFormationSkip lakeFormationAction = iota
	lakeFormationDelete
	lakeFormationSet
)

// lakeFormationActionFor decides what to do with the lakeformation binding for a
// given configured value. An absent value means "don't touch the server", not
// "delete": on Alibaba Cloud this field holds the Unity Catalog user mapping that
// an administrator sets out of band (engine: set_user_role_arn), so deleting on
// absence silently broke external schemas for anyone who left this AWS-named
// field out of their configuration. Removing the binding takes an explicit "".
func lakeFormationActionFor(roleArn types.String) lakeFormationAction {
	switch {
	case roleArn.IsNull() || roleArn.IsUnknown():
		return lakeFormationSkip
	case roleArn.ValueString() == "":
		return lakeFormationDelete
	default:
		return lakeFormationSet
	}
}
