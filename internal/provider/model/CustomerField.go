package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type CustomerType struct {
	ID types.String `tfsdk:"id"`
}
