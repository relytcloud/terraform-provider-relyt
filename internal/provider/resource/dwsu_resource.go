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
			"id":      schema.StringAttribute{Computed: true, Description: "The ID of the service unit.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"cloud":   schema.StringAttribute{Required: true, Description: "The ID of the cloud provider."},
			"region":  schema.StringAttribute{Required: true, Description: "The ID of the region."},
			"domain":  schema.StringAttribute{Required: true, Description: "The domain name of the service unit."},
			"variant": schema.StringAttribute{Optional: true, Computed: true, Description: "The variables.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Default: stringdefault.StaticString("basic")},
			"edition": schema.StringAttribute{Optional: true, Computed: true, Description: "The ID of the edition.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Default: stringdefault.StaticString("standard")},
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
		//	dwsu not found，throw error will cause refresh failed! block destroy. but warning will make import sucess
		//
		resp.Diagnostics = diag.Diagnostics{}
		resp.Diagnostics.AddError("Skip Read", "DWSU not found!")
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
func (r *dwsuResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//resp.Diagnostics.AddError("not support", "update dwsu not supported! please rollback your change!")
	var plan = model.DwsuModel{}
	req.Plan.Get(ctx, &plan)
	var state = model.DwsuModel{}
	req.State.Get(ctx, &state)
	if plan.DefaultDps.Size != state.DefaultDps.Size {
		updateDps(ctx, r.client, state.DefaultDps, plan.DefaultDps, &resp.Diagnostics, state.ID.ValueString(), state.ID.ValueString())
		//反馈给用户，当前dps状态
		resp.State.Set(ctx, &state)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.State.Set(ctx, &state)
	return
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
		return dwsu, fmt.Errorf("dwsu is not Ready")
	})
	return queryDwsuModel, err
}
