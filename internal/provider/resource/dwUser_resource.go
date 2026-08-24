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
			"dwsu_id":                             schema.StringAttribute{Required: true, Description: "The ID of the service unit."},
			"id":                                  schema.StringAttribute{Computed: true, Description: "The ID of the DW user.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"account_name":                        schema.StringAttribute{Required: true, Description: "The name of the DW user, which is unique in the instance. The name is the email address."},
			"account_password":                    schema.StringAttribute{Required: true, Description: "initPassword"},
			"datalake_aws_lakeformation_role_arn": schema.StringAttribute{Optional: true, Description: "The ARN of the cross-account IAM role, optional."},
			"async_query_result_location_prefix":  schema.StringAttribute{Optional: true, Description: "The prefix of the path to the S3 output location."},
			"async_query_result_location_aws_role_arn": schema.StringAttribute{Optional: true, Description: "The ARN of the role to access the output location, optional."},
		},
	}
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
	diags = resp.State.Set(ctx, &dwUserModel)
	r.handleAccountConfig(ctx, &dwUserModel, regionUri, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//这里注释掉主动回滚，应该由用户回滚
	//err := r.client.DropAccount(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString())
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"Error rollback create dwuser",
	//		"Could not rollback dwuser! please clear it with destroy or manual! userId: "+dwUserModel.ID.ValueString()+""+err.Error(),
	//	)
	//}
	//}
	if resp.Diagnostics.HasError() {
		//如果有异常，dwuser不要写状态
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
			state.AsyncQueryResultLocationAwsRoleArn = types.StringValue(config.AwsIamArn)
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
	r.handleAccountConfig(ctx, &plan, regionUri, &resp.Diagnostics)
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

func (r *dwUserResource) handleAccountConfig(ctx context.Context, dwUserModel *tfModel.DWUserModel, regionUri string, diagnostics *diag.Diagnostics) {
	//dwUserModel.ID = dwUserModel.AccountName
	asyncResult := client.AsyncResult{
		AwsIamArn:        dwUserModel.AsyncQueryResultLocationAwsRoleArn.ValueString(),
		S3LocationPrefix: dwUserModel.AsyncQueryResultLocationPrefix.ValueString(),
	}
	lakeFormation := client.LakeFormation{
		IAMRole: dwUserModel.DatalakeAwsLakeformationRoleArn.ValueString(),
	}
	tflog.Info(ctx, fmt.Sprintf("=======uknown %t nil %t", dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsUnknown(), dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsNull()))
	//if dwUserModel.AsyncQueryResultLocationPrefix.IsUnknown() {
	//	dwUserModel.AsyncQueryResultLocationPrefix = types.StringNull()
	//}
	//if dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsUnknown() {
	//	dwUserModel.AsyncQueryResultLocationAwsRoleArn = types.StringNull()
	//}
	if !dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsNull() && !dwUserModel.AsyncQueryResultLocationPrefix.IsNull() {
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
	} else if dwUserModel.AsyncQueryResultLocationPrefix.IsNull() && dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsNull() {
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
	} else if dwUserModel.AsyncQueryResultLocationPrefix.IsNull() || dwUserModel.AsyncQueryResultLocationAwsRoleArn.IsNull() {
		//只有一个属性的时候报给用户异常
		diagnostics.AddError(
			"Error config dwuser",
			"Could not config dwuser async, arn and prefix should be set together",
		)
	}
	if dwUserModel.DatalakeAwsLakeformationRoleArn.IsUnknown() {
		dwUserModel.DatalakeAwsLakeformationRoleArn = types.StringNull()
	}
	if !dwUserModel.DatalakeAwsLakeformationRoleArn.IsNull() {
		_, err := common.CommonRetry[client.CommonRelytResponse[string]](ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.LakeFormationConfig(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString(), lakeFormation)
		})
		if err != nil {
			diagnostics.AddError(
				"Error config dwuser",
				"Could not config dwuser lakeformation, unexpected error: "+err.Error(),
			)
			//return
		}
	} else if dwUserModel.DatalakeAwsLakeformationRoleArn.IsNull() {
		_, err := common.CommonRetry[client.CommonRelytResponse[string]](ctx, func() (*client.CommonRelytResponse[string], error) {
			return r.client.DeleteLakeFormationConfig(ctx, regionUri, dwUserModel.DwsuId.ValueString(), dwUserModel.ID.ValueString())
		})
		if err != nil {
			diagnostics.AddError(
				"Error config dwuser",
				"Could not delete dwuser lakeformation, unexpected error: "+err.Error(),
			)
			//return
		}
	}
}
