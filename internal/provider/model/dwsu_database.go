package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DataAccessConfig struct {
	AccessKey     types.String `tfsdk:"access_key"`
	SecretKey     types.String `tfsdk:"secret_key"`
	Endpoint      types.String `tfsdk:"endpoint"`
	ClientTimeout types.Int32  `tfsdk:"client_timeout"`
}

//type OptionalProviderConfig struct {
//	Auth DataAccessConfig `tfsdk:"data_access_config"`
//}

type DwsuExternalSchema struct {
	Name        types.String       `tfsdk:"name"`
	Database    types.String       `tfsdk:"database"`
	Catalog     types.String       `tfsdk:"catalog"`
	TableFormat types.String       `tfsdk:"table_format"`
	Properties  map[string]*string `tfsdk:"properties"`
	//Properties   ExternalSchemaProperties `tfsdk:"properties"`
}

//type ExternalSchemaProperties struct {
//	Metastore             types.String `tfsdk:"metastore"`
//	GlueAccessControlMode types.String `tfsdk:"glue_access_control_mode"`
//	GlueRegion            types.String `tfsdk:"glue_region"`
//	S3Region              types.String `tfsdk:"s3_region"`
//}

type DwsuDatabases struct {
	Databases types.List `tfsdk:"databases"`
}

type DwsuDatabaseMeta struct {
	Name  types.String `tfsdk:"name"`
	Owner types.String `tfsdk:"owner"`
}

type DwsuSchemas struct {
	Database types.String `tfsdk:"database"`
	Schemas  types.List   `tfsdk:"schemas"`
}

type DwsuSchemaMeta struct {
	Database types.String `tfsdk:"database"`
	Catalog  types.String `tfsdk:"catalog"`
	Name     types.String `tfsdk:"name"`
	Owner    types.String `tfsdk:"owner"`
	External types.Bool   `tfsdk:"external"`
}

//var (
//	ResourceAuthSchema = resSchema.SingleNestedAttribute{
//		Required: true,
//		Attributes: map[string]resSchema.Attribute{
//			"access_key": resSchema.StringAttribute{Required: true, Description: "access key"},
//			"secret_key": resSchema.StringAttribute{Required: true, Description: "secret key"},
//		},
//		Description: "The Auth AccessKey And SecretKey.",
//	}
//	DatasourceAuthSchema = datasourceSchema.SingleNestedAttribute{
//		Required: true,
//		Attributes: map[string]datasourceSchema.Attribute{
//			"access_key": datasourceSchema.StringAttribute{Required: true, Description: "access key"},
//			"secret_key": datasourceSchema.StringAttribute{Required: true, Description: "secret key"},
//		},
//		Description: "The Auth AccessKey And SecretKey.",
//	}
//)
