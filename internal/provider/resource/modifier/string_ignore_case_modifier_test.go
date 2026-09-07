package modifier

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringIgnoreCaseModifier(t *testing.T) {
	cases := []struct {
		name          string
		state, config types.String
		plan          types.String
		want          types.String
	}{
		{"no state (create) keeps the configured value", types.StringNull(), types.StringValue("DELTA"), types.StringValue("DELTA"), types.StringValue("DELTA")},
		{"same value in a different case is not a change", types.StringValue("delta"), types.StringValue("DELTA"), types.StringValue("DELTA"), types.StringValue("delta")},
		{"state keeps the configured spelling after apply", types.StringValue("DELTA"), types.StringValue("delta"), types.StringValue("delta"), types.StringValue("DELTA")},
		{"identical value is untouched", types.StringValue("DELTA"), types.StringValue("DELTA"), types.StringValue("DELTA"), types.StringValue("DELTA")},
		{"a real change keeps the configured spelling", types.StringValue("delta"), types.StringValue("ICEBERG"), types.StringValue("ICEBERG"), types.StringValue("ICEBERG")},
		{"unknown plan is left alone", types.StringValue("delta"), types.StringUnknown(), types.StringUnknown(), types.StringUnknown()},
		{"unknown config is left alone", types.StringValue("delta"), types.StringUnknown(), types.StringValue("DELTA"), types.StringValue("DELTA")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &planmodifier.StringResponse{PlanValue: c.plan}
			GetStringIgnoreCaseModifier().PlanModifyString(context.Background(), planmodifier.StringRequest{
				StateValue:  c.state,
				ConfigValue: c.config,
				PlanValue:   c.plan,
			}, resp)
			if !resp.PlanValue.Equal(c.want) {
				t.Fatalf("plan value = %v, want %v", resp.PlanValue, c.want)
			}
		})
	}
}
