package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type PrivateLinkModel struct {
	DwsuId          types.String `tfsdk:"dwsu_id"`
	ServiceType     types.String `tfsdk:"service_type"`
	ServiceName     types.String `tfsdk:"service_name"`
	Status          types.String `tfsdk:"status"`
	AllowPrinciples types.List   `tfsdk:"allow_principles"`
}

type AllowPrinciple struct {
	Principle types.String `tfsdk:"principle"`
}
