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
	_ datasource.DataSource              = &CloudsDataSource{}
	_ datasource.DataSourceWithConfigure = &CloudsDataSource{}
)

func NewCloudsDataSource() datasource.DataSource {
	return &CloudsDataSource{}
}

type CloudsDataSource struct {
	RelytClientDatasource
}

func (d *CloudsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clouds"
}

func (d *CloudsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The cloud providers served by this control plane. The valid values of `cloud` differ per deployment, so query them instead of copying them from another environment's examples.",
		Attributes: map[string]schema.Attribute{
			"clouds": schema.ListNestedAttribute{
				Computed:    true,
				Description: "clouds served by this control plane",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "The ID of the cloud provider, used as `cloud` in other resources."},
						"name":         schema.StringAttribute{Computed: true, Description: "The display name of the cloud provider."},
						"link":         schema.StringAttribute{Computed: true, Description: "The homepage of the cloud provider."},
						"is_public":    schema.BoolAttribute{Computed: true, Description: "Whether the cloud is public."},
						"is_available": schema.BoolAttribute{Computed: true, Description: "Whether the cloud is currently available for new service units."},
					},
				},
			},
		},
	}
}

func (d *CloudsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.CloudsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clouds, err := common.CommonRetry(ctx, func() (*[]*client.Cloud, error) {
		list, err := d.client.ListClouds(ctx)
		return &list, err
	})
	if err != nil {
		resp.Diagnostics.AddError("error list clouds", "msg: "+err.Error())
		return
	}
	state.Clouds = []model.CloudModel{}
	if clouds != nil {
		for _, c := range *clouds {
			if c == nil {
				continue
			}
			state.Clouds = append(state.Clouds, model.CloudModel{
				ID:          types.StringValue(c.ID),
				Name:        types.StringValue(c.Name),
				Link:        types.StringValue(c.Link),
				IsPublic:    types.BoolValue(c.IsPublic),
				IsAvailable: types.BoolValue(c.IsAvailable),
			})
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
