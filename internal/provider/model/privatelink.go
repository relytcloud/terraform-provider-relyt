package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type PrivateLinkModel struct {
	DwsuId          types.String `tfsdk:"dwsu_id"`
	ServiceType     types.String `tfsdk:"service_type"`
	ServiceName     types.String `tfsdk:"service_name"`
	Status          types.String `tfsdk:"status"`
	Zones           types.List   `tfsdk:"zones"`
	AllowPrincipals types.List   `tfsdk:"allow_principals"`
}

type AllowPrinciple struct {
	Principal types.String `tfsdk:"principal"`
}
