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
	_ datasource.DataSource              = &Boto3DataSource{}
	_ datasource.DataSourceWithConfigure = &Boto3DataSource{}
)

func NewBoto3DataSource() datasource.DataSource {
	return &Boto3DataSource{}
}

type Boto3DataSource struct {
	client *client.RelytClient
}

func (d *Boto3DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_boto3_access_info"
}

// Schema defines the schema for the data source.
func (d *Boto3DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dwsu_id":    schema.StringAttribute{Required: true, Description: "The ID of the service unit."},
			"account_id": schema.StringAttribute{Required: true, Description: "The ID of the account"},
			"boto3_access_infos": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"access_key_id": schema.StringAttribute{Computed: true, Description: "The ID of the key"},
						"access_key":    schema.StringAttribute{Computed: true, Description: "AccessKey"},
						"secret_key":    schema.StringAttribute{Computed: true, Description: "SecretKey"},
					},
				},
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *Boto3DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.Boto3AccessInfoModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	meta := common.RouteRegionUri(ctx, state.DwsuId.ValueString(), d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	boto3AccessInfo, err := d.client.GetBoto3AccessInfo(ctx, meta.URI, state.DwsuId.ValueString(), state.DwUserId.ValueString())
	if err != nil {
		tflog.Error(ctx, "error read boto3 access info:"+err.Error())
		resp.Diagnostics.AddError("read failed!", "error read boto3:"+err.Error())
		return
	}
	if len(boto3AccessInfo) > 0 {
		var saList []model.Boto3AccessInfo
		for _, boto3 := range boto3AccessInfo {
			saList = append(saList, model.Boto3AccessInfo{
				AccessKeyId: types.StringValue(boto3.AccessKeyId),
				AccessKey:   types.StringValue(boto3.AccessKey),
				SecretKey:   types.StringValue(boto3.SecretKey),
			})
		}
		saListType := types.ObjectType{AttrTypes: map[string]attr.Type{
			"access_key_id": types.StringType,
			"access_key":    types.StringType,
			"secret_key":    types.StringType,
		}}
		from, diagnostics := types.ListValueFrom(ctx, saListType, saList)
		resp.Diagnostics.Append(diagnostics...)
		state.Boto3AccessInfos = from
	}
	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *Boto3DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
