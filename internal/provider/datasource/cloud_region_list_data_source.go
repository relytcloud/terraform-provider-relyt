package datasource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &CloudRegionEndPointListDataSource{}
	_ datasource.DataSourceWithConfigure = &CloudRegionEndPointListDataSource{}
)

func NewCloudRegionListDataSource() datasource.DataSource {
	return &CloudRegionEndPointListDataSource{}
}

type CloudRegionEndPointListDataSource struct {
	RelytClientDatasource
	//client *client.RelytClient
}

func (d *CloudRegionEndPointListDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_region_endpoints"
}

// Schema defines the schema for the data source.
func (d *CloudRegionEndPointListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud":  schema.StringAttribute{Required: true, Description: "The ID of the cloud provider."},
			"region": schema.StringAttribute{Required: true, Description: "The ID of the region."},
			"endpoints": schema.ListNestedAttribute{
				Computed:    true,
				Description: "endpoints of cloud region",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"extensions": schema.MapAttribute{Computed: true,
							ElementType: types.StringType,
							Description: "extension info of endpoint"},
						"host":     schema.StringAttribute{Computed: true, Description: "The name of the host used by the endpoint."},
						"id":       schema.StringAttribute{Computed: true, Description: "The ID of the endpoint."},
						"open":     schema.BoolAttribute{Computed: true, Description: "Public network access"},
						"port":     schema.Int64Attribute{Computed: true, Description: "The port number used by the endpoint."},
						"protocol": schema.StringAttribute{Computed: true, Description: "The protocol used by the endpoint."},
						"type":     schema.StringAttribute{Computed: true, Description: "The type of the endpoint."},
						"uri":      schema.StringAttribute{Computed: true, Description: "The URI of the endpoint."},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *CloudRegionEndPointListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state model.CloudRegionEndpoints
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Cloud.ValueString() == "" {
		resp.Diagnostics.AddError("parameter error", "cloud can't be empty")
	}
	if state.Region.ValueString() == "" {
		resp.Diagnostics.AddError("parameter error", "region can't be empty")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	regionEndpoints, err := common.CommonRetry(ctx, func() (*[]client.RegionEndpoint, error) {
		return d.client.GetRegionEndpoints(ctx, state.Cloud.ValueString(), state.Region.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError("error read cloud region endpoints", "msg: "+err.Error())
		//tflog.Error(ctx, "error read dwsu"+err.Error())
		return
	}
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
	endpoints := []model.Endpoints{}
	if regionEndpoints != nil && len(*regionEndpoints) > 0 {
		for _, endpoint := range *regionEndpoints {
			tfEndpoint := model.Endpoints{
				Host:       types.StringValue(endpoint.Host),
				ID:         types.StringValue(endpoint.ID),
				Open:       types.BoolValue(endpoint.Open),
				Port:       types.Int32Value(endpoint.Port),
				Protocol:   types.StringValue(endpoint.Protocol),
				Type:       types.StringValue(endpoint.Type),
				URI:        types.StringValue(endpoint.URI),
				Extensions: types.MapNull(types.StringType),
			}
			endpoints = append(endpoints, tfEndpoint)
		}
	}
	from, diag := types.ListValueFrom(ctx, endpointsTFType, endpoints)
	resp.Diagnostics.Append(diag...)
	state.Endpoints = from
	resp.State.Set(ctx, &state)
}
