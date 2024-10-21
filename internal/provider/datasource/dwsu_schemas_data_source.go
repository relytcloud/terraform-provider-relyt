package datasource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

var (
	_ datasource.DataSource              = &DwsuSchemasDataSource{}
	_ datasource.DataSourceWithConfigure = &DwsuSchemasDataSource{}
)

func NewDwsuSchemasDataSource() datasource.DataSource {
	return &DwsuSchemasDataSource{}
}

type DwsuSchemasDataSource struct {
	RelytClientDatasource
}

func (d *DwsuSchemasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_schemas"
}

// Schema defines the schema for the data source.
func (d *DwsuSchemasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{Required: true, Description: "The name of the database."},
			"schemas": schema.ListNestedAttribute{Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Computed: true, Description: "The name of the schema."},
						"catalog":  schema.StringAttribute{Computed: true, Description: "The catalog of the schema. null is returned if the schema is not an external schema."},
						"database": schema.StringAttribute{Computed: true, Description: "The database of the schema."},
						"owner":    schema.StringAttribute{Computed: true, Description: "The owner of schema."},
						"external": schema.BoolAttribute{Computed: true, Description: "Whether the schema is an external schema. true indicates yes; false indicates no."},
					},
				}, Description: "The list of schema."},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DwsuSchemasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state model.DwsuSchemas
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	dbClient := common.ParseAccessConfig(ctx, d.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	//schemas, err := dbClient.ListSchemas(ctx, client.SchemaPageQuery{
	//	PageQuery: client.PageQuery{
	//		PageSize:   1000,
	//		PageNumber: 1,
	//	},
	//	Database: state.Database.ValueString(),
	//})
	//if err != nil {
	//	msg := "schemas read failed"
	//	if err != nil {
	//		msg = err.Error()
	//	}
	//	resp.Diagnostics.AddError("Failed list schemas", "error list schema "+msg)
	//	return
	//}

	records, _ := common.ScrollPageRecords(&resp.Diagnostics, func(pageSize, pageNum int) ([]*client.SchemaMeta, error) {
		listRecords, err := common.CommonRetry(ctx, func() (*client.CommonPage[client.SchemaMeta], error) {
			schemas, err := dbClient.ListSchemas(ctx, client.SchemaPageQuery{
				PageQuery: client.PageQuery{
					PageSize:   pageSize,
					PageNumber: pageNum,
				},
				Database: state.Database.ValueStringPointer(),
			})
			return schemas, err
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
		"name":     types.StringType,
		"catalog":  types.StringType,
		"database": types.StringType,
		"owner":    types.StringType,
		"external": types.BoolType,
	}}
	if records != nil && len(records) > 0 {
		var tfRecords []model.DwsuSchemaMeta
		for _, record := range records {
			tfRecords = append(tfRecords, model.DwsuSchemaMeta{
				Database: types.StringPointerValue(record.Database),
				Catalog:  types.StringPointerValue(record.Catalog),
				Name:     types.StringPointerValue(record.Name),
				Owner:    types.StringPointerValue(record.Owner),
				External: types.BoolPointerValue(record.External),
			})
		}
		from, diagnostics := types.ListValueFrom(ctx, elementType, tfRecords)
		if diagnostics.HasError() {
			resp.Diagnostics.Append(diagnostics...)
			return
		}
		state.Schemas = from
	}
	resp.State.Set(ctx, state)
}
