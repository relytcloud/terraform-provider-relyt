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
	_ datasource.DataSource              = &DwsuDatabaseDetailDataSource{}
	_ datasource.DataSourceWithConfigure = &DwsuDatabaseDetailDataSource{}
)

func NewDwsuDatabaseDetailDataSource() datasource.DataSource {
	return &DwsuDatabaseDetailDataSource{}
}

type DwsuDatabaseDetailDataSource struct {
	RelytClientDatasource
	//client *client.RelytClient
}

func (d *DwsuDatabaseDetailDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_database"
}

// Schema defines the schema for the data source.
func (d *DwsuDatabaseDetailDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":  schema.StringAttribute{Required: true, Description: "The name of the database, which uniquely identifies the database in the DW service unit."},
			"owner": schema.StringAttribute{Computed: true, Description: "The owner of the database."},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DwsuDatabaseDetailDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	dbClient := common.ParseAccessConfig(ctx, d.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tfDatabase := model.DwsuDatabaseMeta{}
	diags := req.Config.Get(ctx, &tfDatabase)
	resp.Diagnostics.Append(diags...)
	database, err := common.CommonRetry(ctx, func() (*client.Database, error) {
		return dbClient.GetDatabase(ctx, tfDatabase.Name.ValueString())
	})
	if err != nil {
		msg := "database read failed"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("Failed list database", "error get database "+msg)
		return
	}
	//elementType := types.ObjectType{AttrTypes: map[string]attr.Type{
	//	"name":  types.StringType,
	//	"owner": types.StringType,
	//}}
	if database != nil {
		tfDatabase.Owner = types.StringPointerValue(database.Owner)
	} else {
		resp.Diagnostics.AddError("Database Not Found", "please check whether it exist!")
		return
	}
	resp.State.Set(ctx, &tfDatabase)
}
