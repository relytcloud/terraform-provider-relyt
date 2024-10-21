// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	relytDS "terraform-provider-relyt/internal/provider/datasource"
	relytRS "terraform-provider-relyt/internal/provider/resource"
)

// Ensure RelytDatabaseProvider satisfies various provider interfaces.
var _ provider.Provider = &RelytDatabaseProvider{}
var _ provider.ProviderWithFunctions = &RelytDatabaseProvider{}

// RelytDatabaseProvider defines the provider implementation.
type RelytDatabaseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type RelytDatabaseProviderEnv struct {
	EnvKey         string
	PropertyName   string
	SummarySuggest string
	detailSuggest  string
}

func (p *RelytDatabaseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "relyt_database"
	resp.Version = p.version
}

// 定义provider能接受的参数，类型，是否可选等
func (p *RelytDatabaseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			//"endpoint": schema.StringAttribute{
			//	MarkdownDescription: "Example provider attribute",
			//	Optional:            true,
			//},
			"access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "access_key",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "secret_key",
			},
			"api_host": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "api_host",
			},
		},
	}
}

// 读取配置文件
func (p *RelytDatabaseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *RelytDatabaseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		relytRS.NewDwsuDatabaseResource,
		relytRS.NewDwsuExternalSchemaResource,
	}
}

func (p *RelytDatabaseProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		relytDS.NewDwsuDatabasesDataSource,
		relytDS.NewDwsuDatabaseDetailDataSource,
		relytDS.NewDwsuSchemasDataSource,
		relytDS.NewDwsuSchemaDetailDataSource,
	}
}

func (p *RelytDatabaseProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		//NewExampleFunction,
	}
}

func NewDataBaseProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RelytDatabaseProvider{
			version: version,
		}
	}
}
