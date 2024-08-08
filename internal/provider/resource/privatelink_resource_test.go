package resource

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
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
