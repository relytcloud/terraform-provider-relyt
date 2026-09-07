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
	return "A change that only differs in letter case from the value in state is not a change; the state value is kept."
}

func (s StringIgnoreCaseModifier) MarkdownDescription(ctx context.Context) string {
	return "A change that only differs in letter case from the value in state is not a change; the state value is kept."
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
	// Same value in a different case: keep what state holds, so neither the
	// resource nor anything derived from it shows a diff. A real change keeps
	// the configured spelling and is planned as usual. Lower-casing the planned
	// value here (the previous behaviour) only hid the diff on the resource;
	// refresh still rewrote state and outputs reported drift on every plan (#23).
	if strings.EqualFold(request.PlanValue.ValueString(), request.StateValue.ValueString()) {
		response.PlanValue = types.StringValue(request.StateValue.ValueString())
	}
}
