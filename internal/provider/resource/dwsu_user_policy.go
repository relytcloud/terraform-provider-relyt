package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	tfModel "terraform-provider-relyt/internal/provider/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &dwsuUserPolicy{}
	_ resource.ResourceWithConfigure = &dwsuUserPolicy{}
	//_ resource.ResourceWithImportState = &dwsuUserPolicy{}
)

// NewdwsuUserPolicy is a helper function to simplify the provider implementation.
func NewdwsuUserPolicy() resource.Resource {
	return &dwsuUserPolicy{}
}

// orderResource is the resource implementation.
type dwsuUserPolicy struct {
	RelytClientResource
	//client *client.RelytClient
}

// Metadata returns the resource type name.
func (r *dwsuUserPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_user_policy"
}

// Schema defines the schema for the resource.
func (r *dwsuUserPolicy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"dwsu_id":             schema.StringAttribute{Required: true, Description: "The ID of the service unit."},
			"mfa":                 schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("OPTIONAL"), Description: "The mfa policy of the dwsu user. Default 'OPTIONAL'"},
			"reset_init_password": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "The choice whether user need to reset their init password. Default 'false'"},
			//"mfa_protection_scopes": schema.SetAttribute{Optional: true, ElementType: types.StringType, Description: "The mfa protection scopes."},
		},
	}
}

// Create a new resource.
func (r *dwsuUserPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var securityPolicy tfModel.DwsuUserSecurityPolicy
	diags := req.Plan.Get(ctx, &securityPolicy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.patchUserSecurityPolicy(ctx, &securityPolicy, &resp.Diagnostics)
	resp.State.Set(ctx, &securityPolicy)
}

// Read resource information.
func (r *dwsuUserPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var securityPolicy tfModel.DwsuUserSecurityPolicy
	diags := req.State.Get(ctx, &securityPolicy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	meta := common.RouteRegionUri(ctx, securityPolicy.DwsuId.ValueString(), r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI

	clientPolicy, err := common.CommonRetry(ctx, func() (*client.UserSecurityPolicy, error) {
		return r.client.GetUserSecurityPolicy(ctx, regionUri, securityPolicy.DwsuId.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError("error read user security policy!", "failed to read user security policy!"+err.Error())
		return
	}
	securityPolicy.MFA = types.StringValue(clientPolicy.MFAStrategy)
	securityPolicy.ResetInitPassword = types.BoolValue(clientPolicy.RequiredChangingInitPassword)
	//proctionSet, diagnostics := types.SetValueFrom(ctx, types.StringType, clientPolicy.ExtraMfaProtectionScopes)
	//if diagnostics.HasError() {
	//	resp.Diagnostics.Append(diagnostics...)
	//	return
	//}
	//securityPolicy.MFAProtectionScopes = proctionSet
	resp.State.Set(ctx, securityPolicy)
	return
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dwsuUserPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var securityPolicy tfModel.DwsuUserSecurityPolicy
	diags := req.Plan.Get(ctx, &securityPolicy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.patchUserSecurityPolicy(ctx, &securityPolicy, &resp.Diagnostics)
	resp.State.Set(ctx, &securityPolicy)
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dwsuUserPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete Action Notice!", "This action just delete local states! Make nothing change to user policy. ")
}

func (r *dwsuUserPolicy) patchUserSecurityPolicy(ctx context.Context, securityPolicy *tfModel.DwsuUserSecurityPolicy, diag *diag.Diagnostics) {

	meta := common.RouteRegionUri(ctx, securityPolicy.DwsuId.ValueString(), r.client, diag)
	if diag.HasError() {
		return
	}
	regionUri := meta.URI
	//request
	policy := client.UserSecurityPolicy{
		MFAStrategy:                  securityPolicy.MFA.ValueString(),
		RequiredChangingInitPassword: securityPolicy.ResetInitPassword.ValueBool(),
	}

	//diags := types.Set.ElementsAs(securityPolicy.MFAProtectionScopes, ctx, &policy.ExtraMfaProtectionScopes, true)
	//diag.Append(diags...)
	//if diag.HasError() {
	//	return
	//}

	_, err := common.CommonRetry(ctx, func() (*string, error) {
		return r.client.PatchUserSecurityPolicy(ctx, regionUri, securityPolicy.DwsuId.ValueString(), policy)
	})
	if err != nil {
		diag.AddError(
			"Error patch dwsu user policy",
			"Could not patch user policy, unexpected error: "+err.Error(),
		)
		return
	}
}
