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
		DmsHost:   "http://127.0.0.1:8180",
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
	database, err := databaseClient.CreateDatabase(context.TODO(), Database{Name: &name, Comments: &name})
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
	database := "ex_db"
	catalog := "test"
	name := "tpch"
	//format := "DELTA"
	//glue := "glue"
	//region := "us-east-1"
	//lake := "lake-formation"
	schema := Schema{
		Database: &database,
		Catalog:  &catalog,
		Name:     &name,
		//Properties:  map[string]*string{"metastore.type": &glue, "glue.region": &region, "s3.region": &region, "glue.access-control.mode": &lake},
		//TableFormat: &format,
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
	database := "abc"
	catalog := "qingdeng"
	name := "schema"
	format := "DELTA"
	glue := "glue"
	region := "us-east-1"
	lake := "lake-formation"
	schema := Schema{
		Database:    &database,
		Catalog:     &catalog,
		Name:        &name,
		Properties:  map[string]*string{"metastore.type": &glue, "glue.region": &region, "s3.region": &region, "glue.access-control.mode": &lake},
		TableFormat: &format,
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
		Database: &database,
	})
	if err != nil {
		fmt.Println("err  " + err.Error())
		return
	}
	marshal, _ = json.Marshal(list)
	fmt.Println("drop: " + string(marshal))
}
