package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &DwsuExternalSchemaResource{}
	_ resource.ResourceWithConfigure   = &DwsuExternalSchemaResource{}
	_ resource.ResourceWithImportState = &DwsuExternalSchemaResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewDwsuExternalSchemaResource() resource.Resource {
	return &DwsuExternalSchemaResource{}
}

// orderResource is the resource implementation.
type DwsuExternalSchemaResource struct {
	RelytClientResource
}

// Metadata returns the resource type name.
func (r *DwsuExternalSchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_external_schema"
}

// Schema defines the schema for the resource.
func (r *DwsuExternalSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"name":         schema.StringAttribute{Required: true, Description: "The name of the external schema. The schema name must be consistent with the name of the target schema that exists in the external catalog.\nNote that the combined length of the catalog and schema values must not exceed 127 characters."},
			"catalog":      schema.StringAttribute{Required: true, Description: "The name of the catalog.\nNote that the combined length of the catalog and schema values must not exceed 127 characters."},
			"database":     schema.StringAttribute{Required: true, Description: "The name of the database."},
			"table_format": schema.StringAttribute{Required: true, Description: "table_format"},
			"properties": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    true,
				//Computed:    true,
				//Attributes: map[string]schema.Attribute{
				//	"metastore":                schema.StringAttribute{Computed: true, Optional: true, Description: "metastore", Default: stringdefault.StaticString("Glue")},
				//	"glue_access_control_mode": schema.StringAttribute{Computed: true, Optional: true, Description: "glue_access_control_mode", Default: stringdefault.StaticString("Lake Formation")},
				//	"glue_region":              schema.StringAttribute{Computed: true, Optional: true, Description: "glue_region", Default: stringdefault.StaticString("ap-east-1")},
				//	"s3_region":                schema.StringAttribute{Computed: true, Optional: true, Description: "s3_region", Default: stringdefault.StaticString("ap-east-1")},
				//},
				Description: "The properties of the schema."},
		},
	}
}

// Create a new resource.
func (r *DwsuExternalSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	externalSchema := model.DwsuExternalSchema{}
	diags := req.Plan.Get(ctx, &externalSchema)
	resp.Diagnostics.Append(diags...)
	dbClient := common.ParseAccessConfig(ctx, r.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	//properties := map[string]string{}
	dbSchema := client.Schema{
		Database:    externalSchema.Database.ValueStringPointer(),
		Catalog:     externalSchema.Catalog.ValueStringPointer(),
		Name:        externalSchema.Name.ValueStringPointer(),
		Properties:  externalSchema.Properties,
		TableFormat: externalSchema.TableFormat.ValueStringPointer(),
	}
	_, err := dbClient.CreateExternalSchema(ctx, dbSchema)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create schema", "error to create schema:"+err.Error())
		return
	}
	resp.State.Set(ctx, externalSchema)
}

// Read resource information.
func (r *DwsuExternalSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	externalSchema := model.DwsuExternalSchema{}
	diags := req.State.Get(ctx, &externalSchema)
	resp.Diagnostics.Append(diags...)
	dbClient := common.ParseAccessConfig(ctx, r.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	dbSchema := client.Schema{
		Database:    externalSchema.Database.ValueStringPointer(),
		Catalog:     externalSchema.Catalog.ValueStringPointer(),
		Name:        externalSchema.Name.ValueStringPointer(),
		Properties:  externalSchema.Properties,
		TableFormat: externalSchema.TableFormat.ValueStringPointer(),
	}
	getExternalSchema, err := common.CommonRetry(ctx, func() (*client.SchemaMeta, error) {
		return dbClient.GetExternalSchema(ctx, dbSchema)
	})
	if err != nil || getExternalSchema == nil {
		msg := " schema not found!"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("Failed to Read schema", "error to Read schema:"+msg)
		return
	}
	resp.State.Set(ctx, externalSchema)
	//todo 这里没读取？
	//if getExternalSchema.pro{
	//
	//}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DwsuExternalSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Not Support！", "schema not support update! please rollback your change")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DwsuExternalSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	externalSchema := model.DwsuExternalSchema{}
	diags := req.State.Get(ctx, &externalSchema)
	resp.Diagnostics.Append(diags...)
	dbClient := common.ParseAccessConfig(ctx, r.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	dbSchema := client.Schema{
		Database:    externalSchema.Database.ValueStringPointer(),
		Catalog:     externalSchema.Catalog.ValueStringPointer(),
		Name:        externalSchema.Name.ValueStringPointer(),
		Properties:  externalSchema.Properties,
		TableFormat: externalSchema.TableFormat.ValueStringPointer(),
	}

	getExternalSchema, err := common.CommonRetry(ctx, func() (*client.SchemaMeta, error) {
		return dbClient.GetExternalSchema(ctx, dbSchema)
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read schema", "error to read schema before drop! :"+err.Error())
		return
	}
	if getExternalSchema == nil {
		return
	}

	succ, err := common.CommonRetry(ctx, func() (*bool, error) {
		dropSchema, err := dbClient.DropSchema(ctx, dbSchema)
		return &dropSchema, err
	})
	if err != nil || succ == nil || *succ != true {
		msg := "drop schema return false"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("Failed to drop schema", "error to drop schema:"+msg)
		return
	}
}

func (r *DwsuExternalSchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
