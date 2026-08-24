package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
