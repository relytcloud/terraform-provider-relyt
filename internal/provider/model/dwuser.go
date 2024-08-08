package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type ServiceAccountModel struct {
	DwsuId types.String `tfsdk:"dwsu_id"`
	//ServiceAccountInfos ServiceAccountInfo `tfsdk:"account_infos"`
	ServiceAccountInfos types.List `tfsdk:"account_infos"`
}

type ServiceAccountInfo struct {
	Type        types.String `tfsdk:"type"`
	AccountInfo types.Map    `tfsdk:"account_info"`
}

type Boto3AccessInfoModel struct {
	DwsuId           types.String `tfsdk:"dwsu_id"`
	DwUserId         types.String `tfsdk:"account_id"`
	Boto3AccessInfos types.List   `tfsdk:"boto3_access_infos"`
}

type Boto3AccessInfo struct {
	AccessKeyId types.String `tfsdk:"access_key_id"`
	AccessKey   types.String `tfsdk:"access_key"`
	SecretKey   types.String `tfsdk:"secret_key"`
}
