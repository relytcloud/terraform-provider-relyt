package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &PrivateLinkResource{}
	_ resource.ResourceWithConfigure   = &PrivateLinkResource{}
	_ resource.ResourceWithImportState = &PrivateLinkResource{}
)

type PrivateLinkResource struct {
	RelytClientResource
}

func NewPrivateLinkResource() resource.Resource {
	return &PrivateLinkResource{}
}

// Metadata returns the resource type name.
func (r *PrivateLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_privatelink"
}

// Schema defines the schema for the resource.
func (r *PrivateLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"dwsu_id":      schema.StringAttribute{Required: true, Description: "dwsuid"},
			"service_type": schema.StringAttribute{Required: true, Description: "(database | data_api | web_console)"},
			"service_name": schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"allow_principals": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"principal": schema.StringAttribute{Required: true, Description: "principal"},
					},
				},
				Required: true, Description: "allow principal"},
		},
	}
}

// Create a new resource.
func (r *PrivateLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	//var plan tfModel.PrivateLinkModel
	var plan model.PrivateLinkModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dwsuId := plan.DwsuId.ValueString()
	meta := common.RouteRegionUri(ctx, dwsuId, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	service := client.PrivateLinkService{
		ServiceType:       plan.ServiceType.ValueString(),
		AllowedPrincipals: new([]string),
	}
	r.parsePrinciple(ctx, plan.AllowPrincipals, &service)
	if plan.Status.IsUnknown() {
		_, err := r.client.CreatePrivateLinkService(ctx, regionUri, dwsuId, service)
		if err != nil {
			tflog.Error(ctx, "error create private link"+err.Error())
			resp.Diagnostics.AddError("create failed!", "failed to create private link!"+err.Error())
			return
		}
	}
	//先写一下Status，下次读一下。如果有Status属性则创建一半
	plan.Status = types.StringValue(client.PRIVATE_LINK_UNKNOWN)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	privateLinkInfo, err := common.TimeOutTask(r.client.CheckTimeOut, r.client.CheckInterval, func() (any, error) {
		linkService, errGet := r.client.GetPrivateLinkService(ctx, regionUri, dwsuId, plan.ServiceType.ValueString())
		if errGet != nil {
			return nil, errGet
		}
		if linkService != nil && linkService.Status == client.PRIVATE_LINK_READY {
			return linkService, nil
		}
		return linkService, fmt.Errorf("status not ready")
	})
	if err != nil {
		resp.Diagnostics.AddError("wait ready failed!", "failed to wait privateLink! "+err.Error())
		return
	}
	pl, ok := privateLinkInfo.(*client.PrivateLinkService)
	if !ok {
		resp.Diagnostics.AddError("return type not privatelink", "type convert error")
		return
	}
	r.mapRelytToTFModel(nil, pl, &plan, &resp.Diagnostics)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *PrivateLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state model.PrivateLinkModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dwsuId := state.DwsuId.ValueString()
	meta := common.RouteRegionUri(ctx, dwsuId, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	retry, err := common.CommonRetry(ctx, func() (*client.PrivateLinkService, error) {
		return r.client.GetPrivateLinkService(ctx, regionUri, dwsuId, state.ServiceType.ValueString())
	})
	if err != nil || retry == nil {
		resp.Diagnostics.AddError("error get private link", "get private link failed!"+err.Error())
		return
	}
	r.mapRelytToTFModel(ctx, retry, &state, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PrivateLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state model.PrivateLinkModel
	var plan model.PrivateLinkModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.DwsuId.Equal(state.DwsuId) || !plan.ServiceType.Equal(state.ServiceType) {
		resp.Diagnostics.AddError("update contain not support field!", "can't update dwsuId or serviceType")
		return
	}
	dwsuId := state.DwsuId.ValueString()
	meta := common.RouteRegionUri(ctx, dwsuId, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	service := client.PrivateLinkService{AllowedPrincipals: new([]string)}
	r.parsePrinciple(ctx, plan.AllowPrincipals, &service)
	_, err := common.CommonRetry(ctx, func() (*client.CommonRelytResponse[client.PrivateLinkService], error) {
		return r.client.PatchPrivateLinkService(ctx, regionUri, dwsuId, state.ServiceType.ValueString(), service)
	})
	if err != nil {
		resp.Diagnostics.AddError("error update private link", "update private link failed!"+err.Error())
		return
	}
	state.AllowPrincipals = plan.AllowPrincipals
	resp.State.Set(ctx, &state)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *PrivateLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state model.PrivateLinkModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dwsuId := state.DwsuId.ValueString()
	meta := common.RouteRegionUri(ctx, dwsuId, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	_, err := common.CommonRetry(ctx, func() (*client.CommonRelytResponse[string], error) {
		return r.client.DeletePrivateLinkService(ctx, regionUri, dwsuId, state.ServiceType.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError("error delete private link", "delete private link failed!"+err.Error())
		return
	}
	_, err = common.TimeOutTask(r.client.CheckTimeOut, r.client.CheckInterval, func() (any, error) {
		linkService, errGet := r.client.GetPrivateLinkService(ctx, regionUri, dwsuId, state.ServiceType.ValueString())
		if errGet != nil {
			return nil, errGet
		}
		if linkService != nil {
			return linkService, fmt.Errorf("still wait dropping")
		}
		return linkService, nil
	})
	if err != nil {
		msg := "get privatelink still exist"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("failed to wait delete", "failed to wait Delete"+msg)
		return
	}
}

func (r *PrivateLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: dwsu_id,service_type. Got: %q", req.ID),
		)
		return
	}
	dwsuId := idParts[0]
	serviceType := idParts[1]
	//是否可以import的校验
	meta := common.RouteRegionUri(ctx, dwsuId, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	regionUri := meta.URI
	privatelink, err := common.CommonRetry(ctx, func() (*client.PrivateLinkService, error) {
		return r.client.GetPrivateLinkService(ctx, regionUri, dwsuId, serviceType)
	})
	if err != nil || privatelink == nil {
		msg := "can't find private link!"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("error get private link", "get private link failed! "+msg)
		return
	}
	if privatelink.Status != client.PRIVATE_LINK_READY {
		//4679805844736
		resp.Diagnostics.AddError("error get private link", "private link is not ready! please check and wait")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dwsu_id"), dwsuId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_type"), serviceType)...)
	//resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PrivateLinkResource) mapRelytToTFModel(ctx context.Context, linkInfo *client.PrivateLinkService, linkModel *model.PrivateLinkModel, diagnostic *diag.Diagnostics) {
	if linkInfo != nil && linkModel != nil {
		linkModel.Status = types.StringValue(linkInfo.Status)
		linkModel.ServiceName = types.StringValue(linkInfo.ServiceName)
		objectType := types.ObjectType{AttrTypes: map[string]attr.Type{
			"principal": types.StringType,
		}}
		if linkInfo.AllowedPrincipals == nil {
			linkInfo.AllowedPrincipals = new([]string)
		}
		readPrincipal := true
		if len(*linkInfo.AllowedPrincipals) == len(linkModel.AllowPrincipals.Elements()) {
			//为了保持客户端顺序，这里从服务端读取后判断一下是否长度一样，内容一样。如果不是则发生过变动需要更新
			set := map[string]bool{}
			for _, p := range *linkInfo.AllowedPrincipals {
				set[p] = true
			}
			tmpService := client.PrivateLinkService{AllowedPrincipals: &[]string{}}
			r.parsePrinciple(ctx, linkModel.AllowPrincipals, &tmpService)
			for _, n := range *tmpService.AllowedPrincipals {
				if _, exist := set[n]; !exist {
					break
				}
			}
			readPrincipal = false
			return
		}
		if readPrincipal {
			principleList := make([]model.AllowPrinciple, 0, len(*linkInfo.AllowedPrincipals))
			if len(*linkInfo.AllowedPrincipals) > 0 {
				for _, allowPrinciple := range *linkInfo.AllowedPrincipals {
					principleList = append(principleList, model.AllowPrinciple{Principal: types.StringValue(allowPrinciple)})
				}
			}
			from, diagnostics := types.ListValueFrom(ctx, objectType, principleList)
			diagnostics.Append(diagnostics...)
			linkModel.AllowPrincipals = from
		}
	}
}

func (r *PrivateLinkResource) parsePrinciple(ctx context.Context, allowPrincipals types.List, service *client.PrivateLinkService) {
	principle := make([]model.AllowPrinciple, 0, len(allowPrincipals.Elements()))
	if !allowPrincipals.IsNull() && !allowPrincipals.IsUnknown() && len(allowPrincipals.Elements()) > 0 {
		allowPrincipals.ElementsAs(ctx, &principle, false)
		for _, principle := range principle {
			principals := append(*service.AllowedPrincipals, principle.Principal.ValueString())
			service.AllowedPrincipals = &principals
		}
	} else {
		service.AllowedPrincipals = &[]string{}
	}
}
