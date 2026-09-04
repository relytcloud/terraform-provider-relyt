package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	tfModel "terraform-provider-relyt/internal/provider/model"
)

func TestLakeFormationActionFor(t *testing.T) {
	tests := []struct {
		name  string
		value types.String
		want  lakeFormationAction
	}{
		// Omitting the attribute must never delete the binding: on Alibaba Cloud it
		// carries the Unity Catalog user mapping that an admin sets out of band.
		{"omitted from config", types.StringNull(), lakeFormationSkip},
		{"not resolved yet", types.StringUnknown(), lakeFormationSkip},
		{"explicitly emptied", types.StringValue(""), lakeFormationDelete},
		{"configured value", types.StringValue("lakehouse-user@example.com"), lakeFormationSet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lakeFormationActionFor(tt.value); got != tt.want {
				t.Errorf("lakeFormationActionFor(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestPickAlias(t *testing.T) {
	v := types.StringValue
	tests := []struct {
		name      string
		preferred types.String
		legacy    types.String
		wantVal   types.String
		wantSet   bool
		wantErr   bool
	}{
		{"neither set", types.StringNull(), types.StringNull(), types.StringNull(), false, false},
		{"unknown counts as unset", types.StringUnknown(), types.StringNull(), types.StringNull(), false, false},
		{"new name only", v("acs:ram::1:role/a"), types.StringNull(), v("acs:ram::1:role/a"), true, false},
		{"legacy name only", types.StringNull(), v("arn:aws:iam::1:role/a"), v("arn:aws:iam::1:role/a"), true, false},
		{"both equal", v("x"), v("x"), v("x"), true, false},
		{"empty string is a value, not unset", v(""), types.StringNull(), v(""), true, false},
		{"both set, different", v("x"), v("y"), types.StringNull(), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, set, err := pickAlias(tt.preferred, tt.legacy)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if set != tt.wantSet || !got.Equal(tt.wantVal) {
				t.Errorf("pickAlias() = (%v, %v), want (%v, %v)", got, set, tt.wantVal, tt.wantSet)
			}
		})
	}
}

func TestApplyAsyncRoleState(t *testing.T) {
	role := types.StringValue("acs:ram::1:role/a")
	cases := []struct {
		name    string
		cfgNew  types.String
		cfgOld  types.String
		wantNew types.String
		wantOld types.String
	}{
		{"configured with new name", role, types.StringNull(), role, types.StringNull()},
		{"configured with legacy name", types.StringNull(), role, types.StringNull(), role},
		{"configured with both", role, role, role, role},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var model, cfg tfModel.DWUserModel
			cfg.AsyncQueryResultLocationRoleArn, cfg.AsyncQueryResultLocationAwsRoleArn = tt.cfgNew, tt.cfgOld
			applyAsyncRoleState(&model, &cfg, role)
			if !model.AsyncQueryResultLocationRoleArn.Equal(tt.wantNew) || !model.AsyncQueryResultLocationAwsRoleArn.Equal(tt.wantOld) {
				t.Errorf("got (%v, %v), want (%v, %v)", model.AsyncQueryResultLocationRoleArn, model.AsyncQueryResultLocationAwsRoleArn, tt.wantNew, tt.wantOld)
			}
		})
	}
}

func TestFillAsyncRoleState(t *testing.T) {
	old := types.StringValue("stale")
	cases := []struct {
		name     string
		stateNew types.String
		stateOld types.String
		wantNew  types.String
		wantOld  types.String
	}{
		{"legacy name tracked keeps using legacy name", types.StringNull(), old, types.StringNull(), types.StringValue("fresh")},
		{"new name tracked", old, types.StringNull(), types.StringValue("fresh"), types.StringNull()},
		{"both tracked", old, old, types.StringValue("fresh"), types.StringValue("fresh")},
		{"nothing tracked (import) prefers new name", types.StringNull(), types.StringNull(), types.StringValue("fresh"), types.StringNull()},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			state := tfModel.DWUserModel{AsyncQueryResultLocationRoleArn: tt.stateNew, AsyncQueryResultLocationAwsRoleArn: tt.stateOld}
			fillAsyncRoleState(&state, "fresh")
			if !state.AsyncQueryResultLocationRoleArn.Equal(tt.wantNew) || !state.AsyncQueryResultLocationAwsRoleArn.Equal(tt.wantOld) {
				t.Errorf("got (%v, %v), want (%v, %v)", state.AsyncQueryResultLocationRoleArn, state.AsyncQueryResultLocationAwsRoleArn, tt.wantNew, tt.wantOld)
			}
		})
	}
}
