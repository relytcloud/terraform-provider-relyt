package datasource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &serviceAccountDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceAccountDataSource{}
)

// coffeesModel maps coffees schema data.
//type SpecQueryModel struct {
//	//standard
//	Edition types.String `tfsdk:"edition"`
//	//hybrid extreme
//	Type     types.String `tfsdk:"type"`
//	Cloud    types.String `tfsdk:"cloud"`
//	Region   types.String `tfsdk:"region"`
//	SpecName types.String `tfsdk:"spec_name"`
//}
//
//type SpecModel struct {
//	ID   types.Int64  `tfsdk:"id"`
//	Name types.String `tfsdk:"name"`
//}

func NewServiceAccountDataSource() datasource.DataSource {
	return &serviceAccountDataSource{}
}

type serviceAccountDataSource struct {
	client *client.RelytClient
}

func (d *serviceAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_service_account"
}

// Schema defines the schema for the data source.
func (d *serviceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dwsu_id": schema.StringAttribute{Required: true},
			"account_infos": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":         schema.StringAttribute{Computed: true},
						"account_info": schema.MapAttribute{ElementType: types.StringType, Computed: true},
					},
				},
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *serviceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.ServiceAccountModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.DwsuId.IsNull() {
		resp.Diagnostics.AddError("dwsu id is nil", "can't query service account with nil dwsu id")
		return
	}
	meta := common.RouteRegionUri(ctx, state.DwsuId.ValueString(), d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := d.client.GetDwsuServiceAccount(ctx, meta.URI, state.DwsuId.ValueString())
	if err != nil {
		tflog.Error(ctx, "error read service account:"+err.Error())
		resp.Diagnostics.AddError("read failed!", "error read service account:"+err.Error())
		return
	}
	if len(account) > 0 {
		var saList []model.ServiceAccountInfo
		for _, serviceAccount := range account {
			mapAttr, diagnostics := types.MapValueFrom(ctx, types.StringType, serviceAccount.AccountInfo)
			resp.Diagnostics.Append(diagnostics...)
			saList = append(saList, model.ServiceAccountInfo{
				Type:        types.StringValue(serviceAccount.Type),
				AccountInfo: mapAttr,
			})
		}
		saListType := types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":         types.StringType,
			"account_info": types.MapType{ElemType: types.StringType},
		}}
		from, diagnostics := types.ListValueFrom(ctx, saListType, saList)
		resp.Diagnostics.Append(diagnostics...)
		state.ServiceAccountInfos = from
	}
	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *serviceAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = relytClient
}
