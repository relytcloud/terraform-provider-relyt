package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &DwsuDatabaseResource{}
	_ resource.ResourceWithConfigure   = &DwsuDatabaseResource{}
	_ resource.ResourceWithImportState = &DwsuDatabaseResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewDwsuDatabaseResource() resource.Resource {
	return &DwsuDatabaseResource{}
}

// orderResource is the resource implementation.
type DwsuDatabaseResource struct {
	RelytClientResource
}

// Metadata returns the resource type name.
func (r *DwsuDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dwsu_database"
}

// Schema defines the schema for the resource.
func (r *DwsuDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"name":  schema.StringAttribute{Required: true, Description: "The name of the database. The database name must not exceed 127 characters."},
			"owner": schema.StringAttribute{Computed: true, Description: "The owner of the database."},
		},
	}
}

// Create a new resource.
func (r *DwsuDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	database := model.DwsuDatabaseMeta{}
	diags := req.Plan.Get(ctx, &database)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	dbClient := common.ParseAccessConfig(ctx, r.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	createDatabase, err := dbClient.CreateDatabase(ctx, client.Database{
		Name: database.Name.ValueStringPointer(),
	})
	//todo 这里幂等怎么做？创建一个已经存在的database是报失败还是报成功？
	if err != nil || createDatabase == nil {
		resp.Diagnostics.AddError("Failed to create database", " Error info: "+err.Error())
		return
	}
	database.Owner = types.StringPointerValue(createDatabase.Owner)
	resp.State.Set(ctx, database)
}

// Read resource information.
func (r *DwsuDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "try read")
	dbClient := common.ParseAccessConfig(ctx, r.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	database := model.DwsuDatabaseMeta{}
	diags := req.State.Get(ctx, &database)
	resp.Diagnostics.Append(diags...)
	getDatabase, err := common.CommonRetry(ctx, func() (*client.Database, error) {
		return dbClient.GetDatabase(ctx, database.Name.ValueString())
	})
	if err != nil || getDatabase == nil {
		msg := " database not found!"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("Failed read database", "error read database "+msg)
		return
	}
	database.Owner = types.StringPointerValue(getDatabase.Owner)
	resp.State.Set(ctx, &database)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DwsuDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Not Support！", "database not support update! please rollback your change")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DwsuDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	dbClient := common.ParseAccessConfig(ctx, r.client, req.ProviderMeta, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	database := model.DwsuDatabaseMeta{}
	diags := req.State.Get(ctx, &database)
	resp.Diagnostics.Append(diags...)

	getDatabase, err := common.CommonRetry(ctx, func() (*client.Database, error) {
		return dbClient.GetDatabase(ctx, database.Name.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed drop database", "error read database before drop! :"+err.Error())
		return
	}
	if getDatabase == nil {
		return
	}

	succ, err := common.CommonRetry(ctx, func() (*bool, error) {
		dropDatabase, err := dbClient.DropDatabase(ctx, database.Name.ValueString())
		return &dropDatabase, err
	})
	if err != nil {
		return
	}
	if err != nil || succ == nil || *succ == false {
		msg := "database drop not success"
		if err != nil {
			msg = err.Error()
		}
		resp.Diagnostics.AddError("Failed drop database", "error drop database "+msg)
		return
	}
}

func (r *DwsuDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
