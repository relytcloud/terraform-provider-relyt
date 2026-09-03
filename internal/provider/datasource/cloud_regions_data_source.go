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
	_ datasource.DataSource              = &CloudRegionsDataSource{}
	_ datasource.DataSourceWithConfigure = &CloudRegionsDataSource{}
)

func NewCloudRegionsDataSource() datasource.DataSource {
	return &CloudRegionsDataSource{}
}

type CloudRegionsDataSource struct {
	RelytClientDatasource
}

func (d *CloudRegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_regions"
}

func (d *CloudRegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The regions of one cloud provider. Region IDs differ per deployment; discover the cloud ID first via `relyt_clouds`.",
		Attributes: map[string]schema.Attribute{
			"cloud": schema.StringAttribute{Required: true, Description: "The ID of the cloud provider."},
			"regions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "regions of the cloud",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true, Description: "The ID of the region, used as `region` in other resources."},
						"name":   schema.StringAttribute{Computed: true, Description: "The display name of the region."},
						"area":   schema.StringAttribute{Computed: true, Description: "The geographic area of the region."},
						"public": schema.BoolAttribute{Computed: true, Description: "Whether the region is public."},
					},
				},
			},
		},
	}
}

func (d *CloudRegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.CloudRegionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Cloud.ValueString() == "" {
		resp.Diagnostics.AddError("parameter error", "cloud can't be empty")
		return
	}
	regions, err := common.CommonRetry(ctx, func() (*[]*client.Region, error) {
		list, err := d.client.ListCloudRegions(ctx, state.Cloud.ValueString())
		return &list, err
	})
	if err != nil {
		resp.Diagnostics.AddError("error list cloud regions", "cloud: "+state.Cloud.ValueString()+" msg: "+err.Error())
		return
	}
	state.Regions = []model.CloudRegionModel{}
	if regions != nil {
		for _, r := range *regions {
			if r == nil {
				continue
			}
			state.Regions = append(state.Regions, model.CloudRegionModel{
				ID:     types.StringValue(r.ID),
				Name:   types.StringValue(r.Name),
				Area:   types.StringValue(r.Area),
				Public: types.BoolValue(r.Public),
			})
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
