package client

import (
	"context"
)

func NewRelytDatabaseClient(config RelytDatabaseClientConfig) (RelytDatabaseClient, error) {
	return RelytDatabaseClient{config}, nil
}

type RelytDatabaseClient struct {
	RelytDatabaseClientConfig
}

type RelytDatabaseClientConfig struct {
	DmsHost       string `json:"apiHost"`
	AccessKey     string `json:"accessKey"`
	SecretKey     string `json:"secretKey"`
	ClientTimeout int32  `json:"clientTimeout"`
}

func (r *RelytDatabaseClient) CreateDatabase(ctx context.Context, database Database) (*Database, error) {
	resp := CommonRelytResponse[Database]{}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/database/create",
		"POST", &resp, database, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (r *RelytDatabaseClient) DropDatabase(ctx context.Context, name string) (bool, error) {
	resp := CommonRelytResponse[bool]{}

	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/database/drop",
		"POST", &resp, &Database{Name: &name}, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return false, err
	}
	return *resp.Data, nil
}

func (r *RelytDatabaseClient) ListDatabase(ctx context.Context, pageSize, pageNumber int) (*CommonPage[Database], error) {
	resp := CommonRelytResponse[CommonPage[Database]]{}
	pageQuery := PageQuery{PageSize: pageSize, PageNumber: pageNumber}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/database/list",
		"POST", &resp, &pageQuery, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (r *RelytDatabaseClient) GetDatabase(ctx context.Context, name string) (*Database, error) {
	resp := CommonRelytResponse[Database]{}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/database/detail",
		"POST", &resp, Database{Name: &name}, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

//func (r *RelytDatabaseClient) createSchema(ctx context.Context, schema SchemaMeta) (*SchemaMeta, error) {
//	resp := CommonRelytResponse[SchemaMeta]{}
//	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/schema/create",
//		"POST", &resp, &schema, nil, nil, &r.RelytDatabaseClientConfig,
//		nil)
//	if err != nil {
//		return nil, err
//	}
//	return resp.Data, nil
//}

func (r *RelytDatabaseClient) CreateExternalSchema(ctx context.Context, schema Schema) (*SchemaMeta, error) {
	resp := CommonRelytResponse[SchemaMeta]{}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/external-schema/create",
		"POST", &resp, &schema, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (r *RelytDatabaseClient) DropSchema(ctx context.Context, schema Schema) (bool, error) {
	resp := CommonRelytResponse[bool]{}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/schema/drop",
		"POST", &resp, Schema{Database: schema.Database, Catalog: schema.Catalog, Name: schema.Name}, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return false, err
	}
	return *resp.Data, nil
}

func (r *RelytDatabaseClient) ListSchemas(ctx context.Context, query SchemaPageQuery) (*CommonPage[SchemaMeta], error) {
	resp := CommonRelytResponse[CommonPage[SchemaMeta]]{}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/schema/list",
		"POST", &resp, query, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (r *RelytDatabaseClient) GetExternalSchema(ctx context.Context, schema Schema) (*SchemaMeta, error) {
	resp := CommonRelytResponse[SchemaMeta]{}
	err := signedHttpRequestWithHeader(nil, ctx, r.DmsHost, "/api/catalog/external-schema/detail",
		"POST", &resp, Schema{Database: schema.Database, Catalog: schema.Catalog, Name: schema.Name}, nil, nil, &r.RelytDatabaseClientConfig,
		nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
