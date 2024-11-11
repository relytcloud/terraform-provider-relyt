package modifier

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

//var _ planmodifier.String = StringIgnoreCaseModifier{}

type StringIgnoreCaseModifier struct {
}

func GetStringIgnoreCaseModifier() planmodifier.String {
	return StringIgnoreCaseModifier{}
}

func (s StringIgnoreCaseModifier) Description(ctx context.Context) string {
	return "The value of this attribute in will ignore case."
}

func (s StringIgnoreCaseModifier) MarkdownDescription(ctx context.Context) string {
	return "The value of this attribute in will ignore case."
}

func (s StringIgnoreCaseModifier) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if request.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if request.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if request.ConfigValue.IsUnknown() {
		return
	}
	response.PlanValue = types.StringValue(strings.ToLower(request.PlanValue.ValueString()))
}
