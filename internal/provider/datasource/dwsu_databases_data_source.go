package datasource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

var (
	_ datasource.DataSource              = &DwsuDatabasesDataSource{}
	_ datasource.DataSourceWithConfigure = &DwsuDatabasesDataSource{}
)

func NewDwsuDatabasesDataSource() datasource.DataSource {
	return &DwsuDatabasesDataSource{}
}

type DwsuDatabasesDataSource struct {
	RelytClientDatasource
}

func (d *DwsuDatabasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_databases"
}

// Schema defines the schema for the data source.
func (d *DwsuDatabasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"databases": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Computed: true, Description: "The name of database"},
						"owner": schema.StringAttribute{Computed: true, Description: "The owner of database"},
					},
				}, Description: "The list of database."},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DwsuDatabasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	dbClient := common.ParseAccessConfig(ctx, d.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tfDatabases := model.DwsuDatabases{}
	diags := req.Config.Get(ctx, &tfDatabases)
	resp.Diagnostics.Append(diags...)
	records, _ := common.ScrollPageRecords(&resp.Diagnostics, func(pageSize, pageNum int) ([]*client.Database, error) {
		listRecords, err := common.CommonRetry(ctx, func() (*client.CommonPage[client.Database], error) {
			database, err := dbClient.ListDatabase(ctx, pageSize, pageNum)
			return database, err
		})
		if err != nil {
			return nil, err
		}
		if listRecords == nil {
			return nil, fmt.Errorf(" shouldn't get nil CommonPage resp")
		}
		return listRecords.Records, nil
	})
	if resp.Diagnostics.HasError() {
		return
	}

	elementType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":  types.StringType,
		"owner": types.StringType,
	}}
	if records != nil && len(records) > 0 {
		var tfRecords []model.DwsuDatabaseMeta
		for _, innerRecord := range records {
			tfRecords = append(tfRecords, model.DwsuDatabaseMeta{
				Name:  types.StringPointerValue(innerRecord.Name),
				Owner: types.StringPointerValue(innerRecord.Owner),
			})
		}
		from, diagnostics := types.ListValueFrom(ctx, elementType, tfRecords)
		if diagnostics.HasError() {
			tflog.Info(ctx, "read has error")
			resp.Diagnostics.Append(diagnostics...)
			return
		}
		tfDatabases.Databases = from
	}
	resp.State.Set(ctx, &tfDatabases)
}
