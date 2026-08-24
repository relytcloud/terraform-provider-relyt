package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dwsuResource{}
	_ resource.ResourceWithConfigure   = &dwsuResource{}
	_ resource.ResourceWithImportState = &dwsuResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewDwsuResource() resource.Resource {
	return &dwsuResource{}
}

// orderResource is the resource implementation.
type dwsuResource struct {
	RelytClientResource
}

// Metadata returns the resource type name.
func (r *dwsuResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu"
}

// Schema defines the schema for the resource.
func (r *dwsuResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, Description: "The ID of the service unit.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			// cloud/region/domain/variant/edition are fixed at creation: the control
			// plane exposes no API to change them (the only PATCH on a DWSU is the
			// network policy). Editing one is rejected in Update rather than marked
			// RequiresReplace — a replacement plan would offer to destroy a live
			// warehouse over what is usually a typo or historic state drift, and an
			// auto-approved pipeline would carry it out.
			"cloud":   schema.StringAttribute{Required: true, Description: "The ID of the cloud provider. Cannot be changed after creation; an update is rejected."},
			"region":  schema.StringAttribute{Required: true, Description: "The ID of the region. Cannot be changed after creation; an update is rejected."},
			"domain":  schema.StringAttribute{Required: true, Description: "The domain name of the service unit. Cannot be changed after creation; an update is rejected."},
			"variant": schema.StringAttribute{Optional: true, Computed: true, Description: "The variables. Cannot be changed after creation; an update is rejected.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Default: stringdefault.StaticString("basic")},
			"edition": schema.StringAttribute{Optional: true, Computed: true, Description: "The ID of the edition. Cannot be changed after creation; an update is rejected.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Default: stringdefault.StaticString("standard")},
			"alias":   schema.StringAttribute{Optional: true, Description: "The alias of the service unit."},
			//"last_updated": schema.Int64Attribute{Computed: true},
			//"status":       schema.StringAttribute{Computed: true},
			"default_dps": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					//"dwsu_id":     schema.StringAttribute{Computed: true, Optional: true, Description: "The ID of the service unit.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
					//"id":          schema.StringAttribute{Computed: true, Optional: true, Description: "The ID of the DPS cluster.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
					"name":        schema.StringAttribute{Required: true, Description: "The name of the DPS cluster."},
					"description": schema.StringAttribute{Optional: true, Description: "The description of the DPS cluster."},
					"engine":      schema.StringAttribute{Required: true, Description: "The type of the DPS cluster. hybrid, extreme, vector"},
					"size":        schema.StringAttribute{Required: true, Description: "The name of the DPS cluster specification."},
					"status":      schema.StringAttribute{Computed: true, Description: "The status of the DPS cluster."},
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:    true,
				Description: "endpoints of dwsu",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"extensions": schema.MapAttribute{Computed: true,
							ElementType: types.StringType,
							Description: "extension info of endpoint"},
						"host":     schema.StringAttribute{Computed: true, Description: "The name of the host used by the endpoint."},
						"id":       schema.StringAttribute{Computed: true, Description: "The ID of the endpoint."},
						"open":     schema.BoolAttribute{Computed: true, Description: "Public network access"},
						"port":     schema.Int64Attribute{Computed: true, Description: "The port number used by the endpoint."},
						"protocol": schema.StringAttribute{Computed: true, Description: "The protocol used by the endpoint. enum: {HTTP, HTTPS, JDBC}"},
						"type":     schema.StringAttribute{Computed: true, Description: "The type of the endpoint. enum: {openapi, web_console, database}"},
						"uri":      schema.StringAttribute{Computed: true, Description: "The URI of the endpoint."},
					},
				},
			},
		},
	}
}

// Create a new resource.
func (r *dwsuResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from dwsuModel
	var dwsuModel model.DwsuModel
	diags := req.Plan.Get(ctx, &dwsuModel)
	resp.Diagnostics.Append(diags...)
	//dwsuModel.Variant = types.StringValue("basic")
	//dwsuModel.Edition = types.StringValue("standard")
	if resp.Diagnostics.HasError() {
		return
	}
	relytDwsu := client.DwsuModel{
		DefaultDps: &client.DpsMode{
			Description: dwsuModel.DefaultDps.Description.ValueString(),
			Engine:      dwsuModel.DefaultDps.Engine.ValueString(),
			Name:        dwsuModel.DefaultDps.Name.ValueString(),
			Spec: &client.Spec{
				Name: dwsuModel.DefaultDps.Size.ValueString(),
			},
		},
		Domain:  dwsuModel.Domain.ValueString(),
		Alias:   dwsuModel.Alias.ValueString(),
		Variant: &client.Variant{ID: dwsuModel.Variant.ValueString()},
		Edition: &client.Edition{ID: dwsuModel.Edition.ValueString()},
		Region: &client.Region{
			Cloud: &client.Cloud{
				ID: dwsuModel.Cloud.ValueString(),
			},
			ID: dwsuModel.Region.ValueString(),
		},
	}

	if dwsuModel.ID.IsUnknown() {
		//可重入
		// Create dwsu
		createResult, err := r.client.CreateDwsu(ctx, relytDwsu)
		if err != nil || createResult.Code != 200 {
			resp.Diagnostics.AddError(
				"Error creating dwsu",
				"Could not create dwsu, unexpected error: "+err.Error(),
			)
			return
		}
		if createResult.Data == nil {
			resp.Diagnostics.AddError(
				"Error creating dwsu",
				"Could not get dwsu id, after create!",
			)
			return
		}
		//一旦拿到ID立刻保存
		dwsuModel.ID = types.StringValue(*createResult.Data)
		resp.State.Set(ctx, dwsuModel)
	}
	queryDwsuModel, err := WaitDwsuReady(ctx, r.client, dwsuModel.ID.ValueString())
	if err != nil || queryDwsuModel == nil {
		msg := "query dwsu failed! get null!"
		if err != nil {
			tflog.Error(ctx, "error wait dwsu ready"+err.Error())
			msg = err.Error()
		}
		resp.Diagnostics.AddError("create failed!", "error wait dwsu ready!"+msg)
		return
		//fmt.Println(fmt.Sprintf("drop dwsu%s", err.Error()))
	}
	relytQueryModel := queryDwsuModel.(*client.DwsuModel)
	r.mapRelytModelToTerraform(ctx, &resp.Diagnostics, &dwsuModel, relytQueryModel)
	tflog.Info(ctx, "bizId:"+relytQueryModel.ID)
	readDps(ctx, dwsuModel.ID.ValueString(), dwsuModel.ID.ValueString(), r.client, &resp.Diagnostics, dwsuModel.DefaultDps)
	if resp.Diagnostics.HasError() {
		return
	}
	//dwsuModel.LastUpdated = types.Int64Value(time.Now().UnixMilli())
	//dwsuModel.Status = types.StringValue(relytQueryModel.Status)
	// Set state to fully populated data
	diags = resp.State.Set(ctx, dwsuModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "create dwsu success: "+relytQueryModel.ID)
}

// Read resource information.
func (r *dwsuResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	//这里只能改compute的值，改Required或option额值则会触发update
	var state model.DwsuModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	relytQueryModel, err := common.CommonRetry(ctx, func() (*client.DwsuModel, error) {
		return r.client.GetDwsu(ctx, state.ID.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError("error read dwsu", "msg: "+err.Error())
		//tflog.Error(ctx, "error read dwsu"+err.Error())
		return
	}
	if relytQueryModel == nil {
		// Gone from the backend: drop it from state so the next plan offers to
		// recreate it. Erroring here instead left the workspace wedged — refresh
		// failed, so plan, apply and destroy all became impossible and the only
		// way out was a manual `terraform state rm`.
		resp.State.RemoveResource(ctx)
		return
	}
	//state.Status = types.StringValue(dwsu.Status)
	// Set refreshed state
	r.mapRelytModelToTerraform(ctx, &resp.Diagnostics, &state, relytQueryModel)
	readDps(ctx, state.ID.ValueString(), state.ID.ValueString(), r.client, &resp.Diagnostics, state.DefaultDps)
	//if resp.Diagnostics.HasError() {
	//	if relytQueryModel.Status != client.DPS_STATUS_READY {
	//		//	dwsu not ready，throw warn rather error。avoid refresh block destroy
	//		resp.Diagnostics = diag.Diagnostics{}
	//		resp.Diagnostics.AddWarning("Skip Read", "DWSU not found or status not Ready! Can't refresh state. now: "+relytQueryModel.Status)
	//	}
	//	return
	//}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "read dwsu succ : "+state.ID.ValueString())
}

// Update updates the resource and sets the updated Terraform state on success.
//
// Only default_dps size and description can change in place. cloud/region/domain/
// variant/edition are RequiresReplace (see Schema) so they never reach here; alias
// and the DPS name/engine have no corresponding control-plane API, so they are
// rejected rather than silently dropped.
func (r *dwsuResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan = model.DwsuModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state = model.DwsuModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// These are fixed at creation and the control plane has no API to change
	// them. Rejecting the update here (instead of RequiresReplace on the schema)
	// is deliberate: a replacement plan would offer to destroy a live warehouse
	// over what is usually a typo or historic state drift, and an auto-approved
	// pipeline would go through with it. An error is recoverable; a destroyed
	// warehouse is not.
	immutable := []struct {
		name        string
		plan, state types.String
	}{
		{"cloud", plan.Cloud, state.Cloud},
		{"region", plan.Region, state.Region},
		{"domain", plan.Domain, state.Domain},
		{"variant", plan.Variant, state.Variant},
		{"edition", plan.Edition, state.Edition},
	}
	for _, f := range immutable {
		if !f.plan.Equal(f.state) {
			resp.Diagnostics.AddAttributeError(path.Root(f.name),
				f.name+" can't be updated",
				"This attribute is fixed at creation and the control plane has no API to change it. Revert it to "+
					strconv.Quote(f.state.ValueString())+", or destroy and recreate the resource.")
		}
	}
	if !plan.Alias.Equal(state.Alias) {
		resp.Diagnostics.AddAttributeError(path.Root("alias"),
			"alias can't be updated",
			"The control plane has no API to rename a service unit. Revert alias to "+
				strconv.Quote(state.Alias.ValueString())+", or destroy and recreate the resource.")
	}
	if plan.DefaultDps != nil && state.DefaultDps != nil {
		if !plan.DefaultDps.Name.Equal(state.DefaultDps.Name) {
			resp.Diagnostics.AddAttributeError(path.Root("default_dps").AtName("name"),
				"default_dps.name can't be updated",
				"The name of the default DPS is fixed at creation. Revert it to "+
					strconv.Quote(state.DefaultDps.Name.ValueString())+".")
		}
		if !plan.DefaultDps.Engine.Equal(state.DefaultDps.Engine) {
			resp.Diagnostics.AddAttributeError(path.Root("default_dps").AtName("engine"),
				"default_dps.engine can't be updated",
				"The engine of the default DPS is fixed at creation. Revert it to "+
					strconv.Quote(state.DefaultDps.Engine.ValueString())+".")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// updateDps mutates state.DefaultDps in place, recording the size/description
	// that were actually applied plus the resulting status.
	if !plan.DefaultDps.Size.Equal(state.DefaultDps.Size) ||
		!plan.DefaultDps.Description.Equal(state.DefaultDps.Description) {
		updateDps(ctx, r.client, state.DefaultDps, plan.DefaultDps, &resp.Diagnostics, state.ID.ValueString(), state.ID.ValueString())
		if resp.Diagnostics.HasError() {
			// Keep what actually happened rather than the requested plan.
			resp.State.Set(ctx, &state)
			return
		}
	}

	// Persist the plan, not the prior state: writing state back made every edit
	// look like a no-op, so terraform reported "inconsistent result after apply"
	// and replayed the same diff forever. Computed attributes come from state.
	plan.ID = state.ID
	plan.Endpoints = state.Endpoints
	plan.DefaultDps = state.DefaultDps
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *dwsuResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.DwsuModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.ID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"dwsu id is unknown ",
			"Can't drop dwsu with unknown id! Please check your status! ")
		return
	}
	dwsu, err := common.CommonRetry(ctx, func() (*client.DwsuModel, error) {
		return r.client.GetDwsu(ctx, state.ID.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get dwsu meta!", "Can't get dwsu info before drop it! err: "+err.Error())
		return
	}
	if dwsu == nil {
		//保证幂等，先读一次，如果读不到就认为删除成功了
		tflog.Info(ctx, "get dwsu meta is null! treated as success")
		return
	}

	// Delete existing dwsu
	_, err = common.CommonRetry(ctx, func() (*string, error) {
		err = r.client.DropDwsu(ctx, state.ID.ValueString())
		return nil, err
	})
	if err != nil {
		//要不要加error
		resp.Diagnostics.AddError(
			"Error Deleting dwsu",
			"Could not delete dwsu, unexpected error: "+err.Error(),
		)
		return
	}
	//等待删除完成
	_, err = common.TimeOutTask(r.client.CheckTimeOut, r.client.CheckInterval, func() (any, error) {
		dwsu, err2 := r.client.GetDwsu(ctx, state.ID.ValueString())
		if err2 != nil || dwsu == nil {
			//这里判断是否要充实
			return dwsu, err2
		}
		if dwsu == nil || dwsu.Status == client.DPS_STATUS_DROPPED {
			return dwsu, nil
		}
		return dwsu, fmt.Errorf("wait delete dwsu timeout! ")
	})
	if err != nil {
		tflog.Error(ctx, "error wait dwsu delete "+err.Error())
		resp.Diagnostics.AddError(
			"Error Deleting Dwsu ",
			"Could not delete dwsu, unexpected error: "+err.Error(),
		)
	}
	return
}

// Configure adds the provider configured client to the resource.
func (r *dwsuResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	relytClient, ok := req.ProviderData.(*client.RelytClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *RelytClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = relytClient
}

func (r *dwsuResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	//resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	//限制dwsu可以import的状态
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: dwsu_id Got: %q", req.ID),
		)
		return
	}
	dwsu, err := common.CommonRetry(ctx, func() (*client.DwsuModel, error) {
		return r.client.GetDwsu(ctx, req.ID)
	})
	if err != nil {
		resp.Diagnostics.AddError("error read dwsu", "msg: "+err.Error())
		//tflog.Error(ctx, "error read dwsu"+err.Error())
		return
	}
	if dwsu == nil {
		resp.Diagnostics = diag.Diagnostics{}
		resp.Diagnostics.AddError("Can't import", "DWSU not found!")
		return
	}
	if dwsu.Status != client.DPS_STATUS_READY {
		resp.Diagnostics = diag.Diagnostics{}
		resp.Diagnostics.AddError("Can't import", "DWSU status isn't ready!")
		return
	}
	//校验dps状态
	CheckDpsImport(ctx, r.client, req.ID, req.ID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *dwsuResource) mapRelytModelToTerraform(ctx context.Context, diagnostics *diag.Diagnostics, tfDwsuModel *model.DwsuModel, relytDwsuModel *client.DwsuModel) {
	endpointsTFType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"extensions": types.MapType{
				ElemType: types.StringType,
			},
			"host":     types.StringType,
			"id":       types.StringType,
			"open":     types.BoolType,
			"port":     types.Int32Type,
			"protocol": types.StringType,
			"type":     types.StringType,
			"uri":      types.StringType,
		},
	}
	if relytDwsuModel != nil && tfDwsuModel != nil {
		//tfDwsuModel.DefaultDps.DwsuId = types.StringValue(relytDwsuModel.ID)
		//tfDwsuModel.DefaultDps.ID = types.StringValue(relytDwsuModel.ID)
		var tfEndPoints []model.Endpoints
		if relytDwsuModel.Endpoints != nil && len(relytDwsuModel.Endpoints) > 0 {
			for _, endpoint := range relytDwsuModel.Endpoints {
				tfEndpoint := model.Endpoints{
					//Extensions: types.MapValue(types.StringType),
					Host:       types.StringValue(endpoint.Host),
					ID:         types.StringValue(endpoint.ID),
					Open:       types.BoolValue(endpoint.Open),
					Port:       types.Int32Value(endpoint.Port),
					Protocol:   types.StringValue(endpoint.Protocol),
					Type:       types.StringValue(endpoint.Type),
					URI:        types.StringValue(endpoint.URI),
					Extensions: types.MapNull(types.StringType),
				}
				//mapValue, diage := types.MapValueFrom(ctx, types.StringType, endpoint.Extensions)
				//diagnostics.Append(diage...)
				//tfEndpoint.Extensions = mapValue
				tfEndPoints = append(tfEndPoints, tfEndpoint)
				//tfDwsuModel.Endpoints = append(tfDwsuModel.Endpoints, tfEndpoint)
			}
			from, d := types.ListValueFrom(ctx, endpointsTFType, tfEndPoints)
			diagnostics.Append(d...)
			tfDwsuModel.Endpoints = from
		} else {
			from, d := types.ListValueFrom(ctx, endpointsTFType, tfEndPoints)
			diagnostics.Append(d...)
			tfDwsuModel.Endpoints = from
		}

		//only for import resource, fill property
		//if tfDwsuModel.Region.IsNull() || tfDwsuModel.Region.IsUnknown() {
		//}
		//set empty object. let fellow fill property
		if tfDwsuModel.DefaultDps == nil {
			tfDwsuModel.DefaultDps = &model.Dps{}
		}
		if relytDwsuModel.Region != nil {
			tfDwsuModel.Region = types.StringValue(relytDwsuModel.Region.ID)
			if relytDwsuModel.Region.Cloud != nil {
				tfDwsuModel.Cloud = types.StringValue(relytDwsuModel.Region.Cloud.ID)
			}
		}
		if tfDwsuModel.Domain.IsNull() || tfDwsuModel.Domain.IsUnknown() {
			tfDwsuModel.Domain = types.StringValue(relytDwsuModel.Domain)
		}
		//go 默认string为空字符串。。对一个Optional字段设置空字符串和 不设置是不一样的
		if relytDwsuModel.Alias != "" {
			tfDwsuModel.Alias = types.StringValue(relytDwsuModel.Alias)
		}
		if relytDwsuModel.Edition != nil {
			tfDwsuModel.Edition = types.StringValue(relytDwsuModel.Edition.ID)
		}
		if relytDwsuModel.Variant != nil {
			tfDwsuModel.Variant = types.StringValue(relytDwsuModel.Variant.ID)
		}
	}
}

func WaitDwsuReady(ctx context.Context, relytClient *client.RelytClient, dpsId string) (any, error) {
	queryDwsuModel, err := common.TimeOutTask(relytClient.CheckTimeOut, relytClient.CheckInterval, func() (any, error) {
		dwsu, err2 := relytClient.GetDwsu(ctx, dpsId)
		if err2 != nil {
			//这里判断是否要重试
			return dwsu, err2
		}
		if dwsu != nil && dwsu.Status == client.DPS_STATUS_READY {
			return dwsu, nil
		}
		if dwsu != nil && client.IsProvisionFailed(dwsu.Status) {
			return dwsu, common.Terminal(fmt.Errorf("dwsu provisioning failed, status: %s"+
				" (check the region service logs for the cause)", dwsu.Status))
		}
		return dwsu, fmt.Errorf("dwsu is not Ready")
	})
	return queryDwsuModel, err
}
