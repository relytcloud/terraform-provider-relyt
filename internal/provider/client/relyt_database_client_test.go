package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
)

func init() {
	relytDatabaseClientConfig := RelytDatabaseClientConfig{
		DmsHost:   "http://pl-4679805844736-api-d3410a34c78f4386.elb.us-east-1.amazonaws.com:8180",
		AccessKey: "AK8DoEFMRPWBGG0eY1JyNBVj7OnrTO3B6t3uJFyibDcGwz56HrAlg8uKtxf9hQeoHphJzOw",
		SecretKey: "HHJU4NBSLKZVGKTGRM41FCLGZVH4VPWS",
	}
	databaseClient, _ = NewRelytDatabaseClient(relytDatabaseClientConfig)
}

var (
	databaseClient RelytDatabaseClient
)

func TestRelytDatabaseClient_testSign(t *testing.T) {
	//databaseClient.testSign()
	//hostApi := "https://baidu.com/abc"
	hostApi := "https://baidu.com/abc%2Fabc"
	parsedHostApi, err := url.Parse(hostApi)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("host", parsedHostApi.Opaque, "path ", parsedHostApi.Path)
	fmt.Println("raw path", parsedHostApi.RawPath)
}

func TestRelytDatabaseClient_createDatabase(t *testing.T) {
	name := "abc"
	database, err := databaseClient.CreateDatabase(context.TODO(), Database{Name: name, Comments: "abc"})
	if err != nil {
		fmt.Println("error create" + err.Error())
		return
	}
	marshal, _ := json.Marshal(database)
	fmt.Println("resp:" + string(marshal))
}

func TestRelytDatabaseClient_dropDatabase(t *testing.T) {
	name := "exp"
	//	 dropdatabase
	dropDatabase, err := databaseClient.DropDatabase(ctx, name)
	if err != nil {
		return
	}
	marshal, _ := json.Marshal(dropDatabase)
	fmt.Println("drop: " + string(marshal))
}

func TestRelytDatabaseClient_listAllDatabase(t *testing.T) {
	listDatabase, err := databaseClient.ListDatabase(ctx, 10, 1)
	if err != nil {
		fmt.Println("err  " + err.Error())
		return
	}
	marshal, _ := json.Marshal(listDatabase)
	fmt.Println("drop: " + string(marshal))

	db, err := databaseClient.GetDatabase(ctx, "test1")
	marshal, _ = json.Marshal(db)
	fmt.Println("get: " + string(marshal))

}

func TestRelytDatabaseClient_schma(t *testing.T) {
	schema := Schema{
		Database:    "abc",
		Catalog:     "qingdeng",
		Name:        "schema",
		Properties:  map[string]string{"metastore.type": "glue", "glue.region": "us-east-1", "s3.region": "us-east-1", "glue.access-control.mode": "lake-formation"},
		TableFormat: "DELTA",
	}
	//resp, err := databaseClient.createExternalSchema(ctx, schema)
	//if err != nil {
	//	fmt.Println("err  " + err.Error())
	//	return
	//}
	//marshal, _ := json.Marshal(resp)
	//fmt.Println("drop: " + string(marshal))
	//get
	resp, err := databaseClient.GetExternalSchema(ctx, schema)
	if err != nil {
		fmt.Println("err  " + err.Error())
		return
	}
	marshal, _ := json.Marshal(resp)
	fmt.Println("drop: " + string(marshal))

}

func TestListSchema(t *testing.T) {
	schema := Schema{
		Database:    "exp",
		Catalog:     "catalog",
		Name:        "external",
		Properties:  map[string]string{"metastore.type": "glue", "glue.region": "us-east-1", "s3.region": "us-east-1", "glue.access-control.mode": "lake-formation"},
		TableFormat: "DELTA",
	}
	//	 drop
	drop, err := databaseClient.DropSchema(ctx, schema)
	if err != nil {
		fmt.Println("err  " + err.Error())
		return
	}
	marshal, _ := json.Marshal(drop)
	fmt.Println("drop: " + string(marshal))
	//	list
	list, err := databaseClient.ListSchemas(ctx, SchemaPageQuery{
		PageQuery: PageQuery{
			PageSize:   100,
			PageNumber: 1,
		},
		Database: "abc",
	})
	if err != nil {
		fmt.Println("err  " + err.Error())
		return
	}
	marshal, _ = json.Marshal(list)
	fmt.Println("drop: " + string(marshal))
}

func TestRelytDatabaseClient_testNil(t *testing.T) {

	schemaMeta := SchemaMeta{
		Name:     "",
		Owner:    "",
		Comments: "",
		Type:     "",
		Oid:      nil,
		Database: "",
		Catalog:  "",
		UID:      "",
	}
	marshal, _ := json.Marshal(schemaMeta)
	fmt.Println(string(marshal))
	json.Unmarshal([]byte("{name:\"\"}"), &schemaMeta)
	fmt.Println("name====" + schemaMeta.Name)

	//json := "{}"
	type Inner struct {
		boo bool `json:"boo,omitempty"`
	}
	abc := Inner{}
	json.Unmarshal([]byte("{}"), &abc)
	var bool1 *bool
	bool1 = &abc.boo
	if *bool1 == true {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	var ddd []int
	ddd = nil
	fmt.Println(len(ddd))
}
