package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// Discovery data sources: relyt_clouds / relyt_cloud_regions / relyt_dps_specs.
// Valid cloud, region and spec values differ per deployment and are not
// guessable, so they must be queryable rather than documented constants.

type CloudsModel struct {
	Clouds []CloudModel `tfsdk:"clouds"`
}

type CloudModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Link        types.String `tfsdk:"link"`
	IsPublic    types.Bool   `tfsdk:"is_public"`
	IsAvailable types.Bool   `tfsdk:"is_available"`
}

type CloudRegionsModel struct {
	Cloud   types.String       `tfsdk:"cloud"`
	Regions []CloudRegionModel `tfsdk:"regions"`
}

type CloudRegionModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Area   types.String `tfsdk:"area"`
	Public types.Bool   `tfsdk:"public"`
}

type DpsSpecsModel struct {
	Edition types.String   `tfsdk:"edition"`
	Engine  types.String   `tfsdk:"engine"`
	Cloud   types.String   `tfsdk:"cloud"`
	Region  types.String   `tfsdk:"region"`
	Specs   []DpsSpecModel `tfsdk:"specs"`
}

type DpsSpecModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
