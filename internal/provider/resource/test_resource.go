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
	"terraform-provider-relyt/internal/provider/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &testResource{}
	_ resource.ResourceWithConfigure   = &testResource{}
	_ resource.ResourceWithImportState = &testResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewTestResource() resource.Resource {
	return &testResource{}
}

type TestList struct {
	//basetypes.ObjectType
	Name types.String `tfsdk:"name"`
	//MapValue types.Map    `tfsdk:"map_value"`
}

func (t TestList) name() {

}

type TestResource struct {
	ID types.String `tfsdk:"id"`
	//mmm types.Map    `tfsdk:"mmm"`
	Mmm      types.Map  `tfsdk:"mmm"`
	TestList types.List `tfsdk:"self"`
	//TestList []TestList `tfsdk:"self"`
}

// orderResource is the resource implementation.
type testResource struct {
	client *client.RelytClient
}

// Metadata returns the resource type name.
func (r *testResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test"
}

// Schema defines the schema for the resource.
func (r *testResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id":  schema.StringAttribute{Computed: true},
			"mmm": schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"self": schema.ListNestedAttribute{
				Computed: true,
				//ElementType: types.StringType,
				//NestedObject: schema.StringAttribute{Computed: true},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{"name": schema.StringAttribute{Computed: true}},
				},
			},
		},
	}
}

// Create a new resource.
func (r *testResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan TestResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Info(ctx, "i am here")
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	tflog.Info(ctx, "pass Diagnostics")
	//一旦拿到ID立刻保存
	if plan.ID.IsNull() {
		tflog.Info(ctx, "Plan Id  is null")
	}

	if plan.ID.IsUnknown() {
		tflog.Info(ctx, "Plan Id  is unknown")
	}
	objectType := types.ObjectType{
		map[string]attr.Type{
			"name": types.StringType,
		},
	}
	objectValues := []TestList{{Name: types.StringValue("abc")}}
	from, diagnostics := types.ListValueFrom(ctx, objectType, objectValues)
	resp.Diagnostics.Append(diagnostics...)
	plan.TestList = from
	mmm := map[string]string{"abc": "def"}

	valueFrom, d := types.MapValueFrom(ctx, types.StringType, mmm)
	plan.Mmm = valueFrom
	resp.Diagnostics.Append(d...)

	//tflog.Info(ctx, "test len :"+strconv.Itoa(len(plan.TestList)))
	//r.mapRelytModelToTerraform(ctx, &resp.Diagnostics, &plan, &model)
	plan.ID = types.StringValue("abc")
	resp.State.Set(ctx, plan)

	//old := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	//plan.ID = types.StringValue("fix- id test")
	//if time.Now().After(old) {
	//	//resp.Diagnostics.AddError("what", "hh")
	//	return
	//}
	//relytQueryModel := queryDwsuModel.(*client.DwsuModel)
	//tflog.Info(ctx, "bizId:"+relytQueryModel.ID)
	//plan.LastUpdated = types.Int64Value(time.Now().UnixMilli())
	//plan.Status = types.StringValue(relytQueryModel.Status)
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//var (
//	model = client.DwsuModel{Endpoints: []client.Endpoints{
//		{
//			Extensions: &map[string]string{"abc": "def"},
//			Host:       "abc",
//			ID:         "",
//			Open:       false,
//			Port:       0,
//			Protocol:   "",
//			Type:       "",
//			URI:        "",
//		}},
//	}
//)

// Read resource information.
func (r *testResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state TestResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//_, err := r.client.GetDwsu(ctx, state.ID.ValueString())
	//if err != nil {
	//	tflog.Error(ctx, "error read dwsu"+err.Error())
	//	return
	//}
	//r.mapRelytModelToTerraform(ctx, &resp.Diagnostics, &state, &model)
	//state.Status = types.StringValue(dwsu.Status)
	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "read dwsu succ : "+state.ID.ValueString())
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *testResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("not support", "update dwsu not supported")
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *testResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state TestResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	//resp.Diagnostics.AddError("err to delete", "error to delete")
	return
}

// Configure adds the provider configured client to the resource.
func (r *testResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *testResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 2 (Schema.Version)
		0: {
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) { /* ... */
			},
		},
		// State upgrade implementation from 1 (prior state version) to 2 (Schema.Version)
		1: {
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) { /* ... */
			},
		},
	}
}

func (r *testResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *testResource) mapRelytModelToTerraform(ctx context.Context, diagnostics *diag.Diagnostics, tfModel *TestResource, relytDwsuModel *client.DwsuModel) {
	if relytDwsuModel != nil && tfModel != nil {
		if len(relytDwsuModel.Endpoints) > 0 {
			for _, endpoint := range relytDwsuModel.Endpoints {
				//tfEndpoint := TestList{
				//Extensions: types.MapValue(types.StringType),
				//Host:     types.StringValue(endpoint.Host),
				//ID:       types.StringValue(endpoint.ID),
				//Open:     types.BoolValue(endpoint.Open),
				//Port:     types.Int64Value(int64(endpoint.Port)),
				//Protocol: types.StringValue(endpoint.Protocol),
				//Type:     types.StringValue(endpoint.Type),
				//Name: types.StringValue("abc"),
				//}
				if endpoint.Extensions != nil {
					//tfMap := make(map[string]attr.Value, len(*endpoint.Extensions))
					//for key, v := range *endpoint.Extensions {
					//	tfMap[key] = types.StringValue(v)
					//}
					//tfEndpoint.Extensions.ElementsAs(ctx, &tfMap, false)
					//mapValue, diage := types.MapValueFrom(ctx, types.StringType, endpoint.Extensions)
					//diagnostics.Append(diage...)
					//tfEndpoint.Extensions = mapValue
				}
				//types.ListValueFrom(ctx, types.ObjectType{}, tfEndpoint)
				//tfModel.TestList = append(tfModel.TestList, tfEndpoint)
				//tfModel.TestList = append(tfModel.TestList, "tfEndpoint")
			}
		}
	}
}
