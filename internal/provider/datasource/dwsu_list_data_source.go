package datasource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &DwsuListDataSource{}
	_ datasource.DataSourceWithConfigure = &DwsuListDataSource{}
)

func NewDwsuListDataSource() datasource.DataSource {
	return &DwsuListDataSource{}
}

type DwsuListDataSource struct {
	RelytClientDatasource
	//client *client.RelytClient
}

func (d *DwsuListDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsus"
}

// Schema defines the schema for the data source.
func (d *DwsuListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dwsu_list": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					//resource与database定义所引用的schema是不一样的。。。。所以此处无法复用
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true, Description: "The ID of the service unit."},
						"cloud":   schema.StringAttribute{Required: true, Description: "The ID of the cloud provider."},
						"region":  schema.StringAttribute{Required: true, Description: "The ID of the region."},
						"domain":  schema.StringAttribute{Required: true, Description: "The domain name of the service unit."},
						"variant": schema.StringAttribute{Optional: true, Computed: true, Description: "The variables."},
						"edition": schema.StringAttribute{Optional: true, Computed: true, Description: "The ID of the edition."},
						"alias":   schema.StringAttribute{Optional: true, Description: "The alias of the service unit."},
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
				},
				Description: "The dwsu list",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DwsuListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	relytDwsuList, err := common.ScrollPageRecords(&resp.Diagnostics,
		func(pageSize, pageNum int) ([]*client.DwsuModel, error) {
			return d.client.ListDwsu(ctx, pageSize, pageNum)
		})
	if err != nil {
		resp.Diagnostics.AddError("error read dwsu list", "msg: "+err.Error())
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
	dwsuModelList := []model.PlainDwsuModel{}
	if len(relytDwsuList) > 0 {
		for _, dwsuModel := range relytDwsuList {
			slice := model.PlainDwsuModel{
				ID:     types.StringValue(dwsuModel.ID),
				Alias:  types.StringValue(dwsuModel.Alias),
				Domain: types.StringValue(dwsuModel.Domain),
			}
			if dwsuModel.Region.Cloud != nil {
				slice.Cloud = types.StringValue(dwsuModel.Region.Cloud.ID)
				slice.Region = types.StringValue(dwsuModel.Region.ID)
			}
			if dwsuModel.Variant != nil {
				slice.Variant = types.StringValue(dwsuModel.Variant.ID)
			}
			if dwsuModel.Edition != nil {
				slice.Edition = types.StringValue(dwsuModel.Edition.ID)
			}

			var tfEndPoints []model.Endpoints
			for _, endpoint := range dwsuModel.Endpoints {
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
				//resp.Diagnostics.Append(diage...)
				tfEndPoints = append(tfEndPoints, tfEndpoint)
			}
			from, d := types.ListValueFrom(ctx, endpointsTFType, tfEndPoints)
			resp.Diagnostics.Append(d...)
			slice.Endpoints = from
			dwsuModelList = append(dwsuModelList, slice)
		}
	}
	DwsuModelType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":        types.StringType,
		"cloud":     types.StringType,
		"region":    types.StringType,
		"domain":    types.StringType,
		"variant":   types.StringType,
		"edition":   types.StringType,
		"alias":     types.StringType,
		"endpoints": types.ListType{ElemType: endpointsTFType},
	}}
	tflog.Info(ctx, "convert list size{}"+strconv.Itoa(len(dwsuModelList)))
	dwsuList, diagnostics := types.ListValueFrom(ctx, DwsuModelType, dwsuModelList)
	if diagnostics.HasError() {
		resp.Diagnostics.Append(diagnostics...)
		//return
	}
	tfListModel := model.DwsuListModel{DwsuList: dwsuList}
	resp.State.Set(ctx, &tfListModel)
}
