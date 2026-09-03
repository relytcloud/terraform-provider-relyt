package datasource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

var (
	_ datasource.DataSource              = &DpsSpecsDataSource{}
	_ datasource.DataSourceWithConfigure = &DpsSpecsDataSource{}
)

func NewDpsSpecsDataSource() datasource.DataSource {
	return &DpsSpecsDataSource{}
}

type DpsSpecsDataSource struct {
	RelytClientDatasource
}

func (d *DpsSpecsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dps_specs"
}

func (d *DpsSpecsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The DPS specs available in one cloud/region, already filtered by regional availability. The spec list is not the same everywhere — a size copied from another environment's example may not exist here and would only fail at create time.",
		Attributes: map[string]schema.Attribute{
			"edition": schema.StringAttribute{Optional: true, Description: "The ID of the edition. Defaults to 'standard'."},
			"engine":  schema.StringAttribute{Required: true, Description: "The type of the DPS cluster. enum: {hybrid, extreme}"},
			"cloud":   schema.StringAttribute{Required: true, Description: "The ID of the cloud provider."},
			"region":  schema.StringAttribute{Required: true, Description: "The ID of the region."},
			"specs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "specs available for this engine in this cloud/region",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.Int64Attribute{Computed: true, Description: "The ID of the spec."},
						"name": schema.StringAttribute{Computed: true, Description: "The name of the spec, used as `size` in relyt_dwsu.default_dps and relyt_dps."},
					},
				},
			},
		},
	}
}

func (d *DpsSpecsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.DpsSpecsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Cloud.ValueString() == "" {
		resp.Diagnostics.AddError("parameter error", "cloud can't be empty")
	}
	if state.Region.ValueString() == "" {
		resp.Diagnostics.AddError("parameter error", "region can't be empty")
	}
	if state.Engine.ValueString() == "" {
		resp.Diagnostics.AddError("parameter error", "engine can't be empty")
	}
	if resp.Diagnostics.HasError() {
		return
	}
	edition := state.Edition.ValueString()
	if edition == "" {
		edition = "standard"
	}
	specs, err := common.CommonRetry(ctx, func() (*[]client.Spec, error) {
		list, err := d.client.ListSpec(ctx, edition, state.Engine.ValueString(),
			state.Cloud.ValueString(), state.Region.ValueString())
		return &list, err
	})
	if err != nil {
		resp.Diagnostics.AddError("error list dps specs",
			"edition: "+edition+" engine: "+state.Engine.ValueString()+
				" cloud: "+state.Cloud.ValueString()+" region: "+state.Region.ValueString()+
				" msg: "+err.Error())
		return
	}
	state.Specs = []model.DpsSpecModel{}
	if specs != nil {
		for _, s := range *specs {
			state.Specs = append(state.Specs, model.DpsSpecModel{
				ID:   types.Int64Value(s.ID),
				Name: types.StringValue(s.Name),
			})
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
