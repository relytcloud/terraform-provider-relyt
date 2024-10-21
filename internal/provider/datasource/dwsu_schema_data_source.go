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
	_ datasource.DataSource              = &DwsuSchemaDetailDataSource{}
	_ datasource.DataSourceWithConfigure = &DwsuSchemaDetailDataSource{}
)

func NewDwsuSchemaDetailDataSource() datasource.DataSource {
	return &DwsuSchemaDetailDataSource{}
}

type DwsuSchemaDetailDataSource struct {
	RelytClientDatasource
	//client *client.RelytClient
}

func (d *DwsuSchemaDetailDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_external_schema"
}

// Schema defines the schema for the data source.
func (d *DwsuSchemaDetailDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{Required: true, Description: "The database of the schema."},
			"catalog":  schema.StringAttribute{Required: true, Description: "The catalog of the schema."},
			"name":     schema.StringAttribute{Required: true, Description: "The name of the schema."},
			"owner":    schema.StringAttribute{Computed: true, Description: "The owner of schema."},
			"external": schema.BoolAttribute{Computed: true, Description: "Whether the schema is an external schema. true indicates yes; false indicates no."},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DwsuSchemaDetailDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	dbClient := common.ParseAccessConfig(ctx, d.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tfSchema := model.DwsuSchemaMeta{}
	diags := req.Config.Get(ctx, &tfSchema)
	resp.Diagnostics.Append(diags...)

	schemaMeta, err := common.CommonRetry(ctx, func() (*client.SchemaMeta, error) {
		return dbClient.GetExternalSchema(ctx, client.Schema{
			Database: tfSchema.Database.ValueStringPointer(),
			Catalog:  tfSchema.Catalog.ValueStringPointer(),
			Name:     tfSchema.Name.ValueStringPointer(),
		})
	})
	if err != nil {
		msg := "schema read failed"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("Failed get schemas", "error get schema "+msg)
		return
	}
	//elementType := types.ObjectType{AttrTypes: map[string]attr.Type{
	//	"name":  types.StringType,
	//	"owner": types.StringType,
	//}}
	if schemaMeta != nil {
		tfSchema.Owner = types.StringPointerValue(schemaMeta.Owner)
		tfSchema.External = types.BoolPointerValue(schemaMeta.External)
	} else {
		resp.Diagnostics.AddError("Schema Not Found", "please check whether it exist!")
		return
	}
	resp.State.Set(ctx, tfSchema)
}
