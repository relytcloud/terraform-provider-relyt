package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/model"
)

type Abc struct {
	A *[]string `json:"a,omitempty"`
}

func TestPrivateLinkResource_parsePrinciple(t *testing.T) {
	abc := &[]string{}
	//abc := new([]string)
	//abc := make([]string, 0, 0)
	label := Abc{
		A: abc,
	}
	fmt.Println("abc" + strconv.Itoa(len(*label.A)))
	marshal, err := json.Marshal(label)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("json" + string(marshal))

}

// zones has to come back as a known list on both paths: a deployment that
// reports zones and one that leaves the field out of the response. An unknown
// or null value there fails the apply with "Provider returned invalid result
// object after apply".
func TestPrivateLinkResource_mapRelytToTFModelZones(t *testing.T) {
	ctx := context.Background()
	r := &PrivateLinkResource{}
	reported := []string{"cn-hongkong-b", "cn-hongkong-c"}
	cases := []struct {
		name  string
		zones *[]string
		want  []string
	}{
		{name: "reported by the service", zones: &reported, want: reported},
		{name: "absent from the response", zones: nil, want: []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var diags diag.Diagnostics
			linkModel := model.PrivateLinkModel{}
			linkInfo := client.PrivateLinkService{
				ServiceName: "epsrv-test",
				Status:      client.PRIVATE_LINK_READY,
				Zones:       c.zones,
			}
			r.mapRelytToTFModel(ctx, &linkInfo, &linkModel, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if linkModel.Zones.IsNull() || linkModel.Zones.IsUnknown() {
				t.Fatalf("zones must be a known list, got %v", linkModel.Zones)
			}
			var got []string
			linkModel.Zones.ElementsAs(ctx, &got, false)
			if len(got) != len(c.want) {
				t.Fatalf("zones = %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("zones = %v, want %v", got, c.want)
				}
			}
		})
	}
}
