package model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RelytProviderModel describes the provider data model.
type RelytProviderModel struct {
	ApiHost               types.String      `tfsdk:"api_host"`
	AuthKey               types.String      `tfsdk:"auth_key"`
	Role                  types.String      `tfsdk:"role"`
	ResourceCheckTimeout  types.Int64       `tfsdk:"resource_check_timeout"`
	ResourceCheckInterval types.Int64       `tfsdk:"resource_check_interval"`
	ClientTimeout         types.Int64       `tfsdk:"client_timeout"`
	DataAccessConfig      *DataAccessConfig `tfsdk:"data_access_config"`
}
type Endpoints struct {
	Extensions types.Map    `tfsdk:"extensions"`
	Host       types.String `tfsdk:"host"`
	ID         types.String `tfsdk:"id"`
	Open       types.Bool   `tfsdk:"open"`
	Port       types.Int32  `tfsdk:"port"`
	Protocol   types.String `tfsdk:"protocol"`
	Type       types.String `tfsdk:"type"`
	URI        types.String `tfsdk:"uri"`
}

type Dps struct {
	Description types.String `tfsdk:"description"`
	Engine      types.String `tfsdk:"engine"`
	Name        types.String `tfsdk:"name"`
	Size        types.String `tfsdk:"size"`
	Status      types.String `tfsdk:"status"`
}

type PlainDwsuModel struct {
	ID      types.String `tfsdk:"id"`
	Alias   types.String `tfsdk:"alias"`
	Cloud   types.String `tfsdk:"cloud"`
	Region  types.String `tfsdk:"region"`
	Variant types.String `tfsdk:"variant"`
	Edition types.String `tfsdk:"edition"`
	Domain  types.String `tfsdk:"domain"`
	//Endpoints  []Endpoints  `tfsdk:"endpoints"`
	Endpoints types.List `tfsdk:"endpoints"`
	//LastUpdated types.Int64  `tfsdk:"last_updated"`
	//Status      types.String `tfsdk:"status"`
}

type DwsuModel struct {
	PlainDwsuModel
	DefaultDps *Dps `tfsdk:"default_dps"`
}

type DpsModel struct {
	DwsuId types.String `tfsdk:"dwsu_id"`
	ID     types.String `tfsdk:"id"`
	Dps
	//Name        types.String `tfsdk:"name"`
	//Description types.String `tfsdk:"description"`
	//Engine      types.String `tfsdk:"engine"`
	//Size        types.String `tfsdk:"size"`
	//LastUpdated types.String `tfsdk:"last_updated"`
	//Status      types.String `tfsdk:"status"`
}

type DWUserModel struct {
	DwsuId                             types.String `tfsdk:"dwsu_id"`
	ID                                 types.String `tfsdk:"id"`
	AccountName                        types.String `tfsdk:"account_name"`
	AccountPassword                    types.String `tfsdk:"account_password"`
	DatalakeAwsLakeformationRoleArn    types.String `tfsdk:"datalake_aws_lakeformation_role_arn"`
	AsyncQueryResultLocationPrefix     types.String `tfsdk:"async_query_result_location_prefix"`
	AsyncQueryResultLocationAwsRoleArn types.String `tfsdk:"async_query_result_location_aws_role_arn"`
	//LastUpdated                        types.String `tfsdk:"last_updated"`
	//Status                             types.String `tfsdk:"status"`
}

type DwsuListModel struct {
	DwsuList types.List `tfsdk:"dwsu_list"`
}

type CloudRegionEndpoints struct {
	Cloud     types.String `tfsdk:"cloud"`
	Region    types.String `tfsdk:"region"`
	Endpoints types.List   `tfsdk:"endpoints"`
}
