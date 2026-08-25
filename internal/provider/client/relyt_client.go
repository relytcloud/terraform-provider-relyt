package client

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/url"
	"strconv"
)

func NewRelytClient(config RelytClientConfig) (RelytClient, error) {
	return RelytClient{config}, nil
}

type RelytClientConfig struct {
	ApiHost                   string                     `json:"apiHost"`
	AuthKey                   string                     `json:"authKey"`
	Role                      string                     `json:"role"`
	RegionApi                 string                     `json:"regionApi"`
	CheckTimeOut              int64                      `json:"checkTimeOut"`
	CheckInterval             int32                      `json:"checkInterval"`
	ClientTimeout             int32                      `json:"clientTimeout"`
	RelytDatabaseClientConfig *RelytDatabaseClientConfig `json:"relytDatabaseClientConfig"`
}

type RelytClient struct {
	RelytClientConfig
}

func (p *RelytClient) ListDwsu(ctx context.Context, pageSize, pageNumber int) ([]*DwsuModel, error) {
	resp := CommonRelytResponse[CommonPage[DwsuModel]]{}
	pageQuery := map[string]string{
		"pageSize":   strconv.Itoa(pageSize),
		"pageNumber": strconv.Itoa(pageNumber),
	}
	err := doHttpRequest(p, ctx, "", "/dwsu", "GET", &resp, nil, pageQuery, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data.Records, nil
}

func (p *RelytClient) CreateDwsu(ctx context.Context, request DwsuModel) (*CommonRelytResponse[string], error) {
	url := "/dwsu"
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, "", url, "POST", &resp, request, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetDwsu(ctx context.Context, dwServiceUnitId string) (*DwsuModel, error) {
	path := fmt.Sprintf("/dwsu/%s", dwServiceUnitId)
	resp := CommonRelytResponse[DwsuModel]{}
	handler := func(response *CommonRelytResponse[DwsuModel], respString []byte) (*CommonRelytResponse[DwsuModel], error) {
		if response.Code != CODE_SUCCESS && resp.Code != CODE_DWSU_NOT_FOUND {
			body := string(respString)
			tflog.Error(ctx, "error call api! resp code not success! body: "+body)
			return response, fmt.Errorf(body)
		}
		return response, nil
	}
	err := doHttpRequest(p, ctx, "", path, "GET", &resp, nil, nil, handler)
	if err != nil {
		tflog.Error(ctx, "Error get dwsu:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) DropDwsu(ctx context.Context, dwServiceUnitId string) error {
	path := fmt.Sprintf("/dwsu/%s", dwServiceUnitId)
	resp := CommonRelytResponse[string]{}
	handler := func(response *CommonRelytResponse[string], respString []byte) (*CommonRelytResponse[string], error) {
		if response.Code != CODE_SUCCESS && resp.Code != CODE_DWSU_NOT_FOUND {
			body := string(respString)
			tflog.Error(ctx, "error call api! resp code not success! body: "+body)
			return response, fmt.Errorf(body)
		}
		return response, nil
	}
	err := doHttpRequest(p, ctx, "", path, "DELETE", &resp, nil, nil, handler)
	if err != nil {
		tflog.Info(ctx, "delete dwsu err:"+err.Error())
		return err
	}
	return nil
}

func (p *RelytClient) ListDps(ctx context.Context, pageSize, pageNumber int, dwServiceUnitId string) ([]*DpsMode, error) {
	resp := CommonRelytResponse[CommonPage[DpsMode]]{}
	pageQuery := map[string]string{
		"pageSize":   strconv.Itoa(pageSize),
		"pageNumber": strconv.Itoa(pageNumber),
	}
	path := fmt.Sprintf("/dwsu/%s/dps", dwServiceUnitId)
	err := doHttpRequest(p, ctx, "", path, "GET", &resp, nil, pageQuery, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data.Records, nil
}

func (p *RelytClient) CreateDps(ctx context.Context, regionUri string, dwServiceUnitId string, mode DpsMode) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/dps", dwServiceUnitId)
	resp := CommonRelytResponse[string]{}
	if err := doHttpRequest(p, ctx, regionUri, path, "POST", &resp, mode, nil, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) PatchDps(ctx context.Context, regionUri string, dwServiceUnitId, dpsId string, mode DpsMode) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/dps/%s", dwServiceUnitId, dpsId)
	resp := CommonRelytResponse[string]{}
	if err := doHttpRequest(p, ctx, regionUri, path, "PATCH", &resp, mode, nil, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetDps(ctx context.Context, regionUri, dwServiceUnitId, dpsBizId string) (*DpsMode, error) {
	path := fmt.Sprintf("/dwsu/%s/dps/%s", dwServiceUnitId, dpsBizId)
	resp := CommonRelytResponse[DpsMode]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get dps:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) DropDps(ctx context.Context, regionUri, dwServiceUnitId, dpsBizId string) error {
	path := fmt.Sprintf("/dwsu/%s/dps/%s", dwServiceUnitId, dpsBizId)
	resp := CommonRelytResponse[string]{}
	handler := func(response *CommonRelytResponse[string], respString []byte) (*CommonRelytResponse[string], error) {
		if response.Code != CODE_SUCCESS && resp.Code != CODE_DPS_NOT_FOUND {
			body := string(respString)
			tflog.Error(ctx, "error call api! resp code not success! body: "+body)
			return response, fmt.Errorf(body)
		}
		return nil, nil
	}
	err := doHttpRequest(p, ctx, regionUri, path, "DELETE", &resp, nil, nil, handler)
	if err != nil {
		tflog.Info(ctx, "delete dps err:"+err.Error())
		return err
	}
	return nil
}

// ListClouds returns the cloud providers this control plane serves
// (GET /infra). Useful as the first discovery call: the valid values of
// `cloud` differ per deployment and are not guessable.
func (p *RelytClient) ListClouds(ctx context.Context) ([]*Cloud, error) {
	resp := CommonRelytResponse[[]*Cloud]{}
	err := doHttpRequest(p, ctx, "", "/infra", "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return *resp.Data, nil
}

// ListCloudRegions returns the regions of one cloud (GET /infra/{cloud}).
func (p *RelytClient) ListCloudRegions(ctx context.Context, cloud string) ([]*Region, error) {
	path := fmt.Sprintf("/infra/%s", url.PathEscape(cloud))
	resp := CommonRelytResponse[[]*Region]{}
	err := doHttpRequest(p, ctx, "", path, "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return *resp.Data, nil
}

func (p *RelytClient) ListSpec(ctx context.Context, edition, dpsType, cloud, region string) ([]Spec, error) {
	path := fmt.Sprintf("/dwsu/edition/%s/dps/%s/specs", edition, dpsType)
	specList := CommonRelytResponse[[]Spec]{}
	parameter := map[string]string{"cloud": cloud, "region": region}
	err := doHttpRequest(p, ctx, "", path, "GET", &specList, nil, parameter, nil)
	if err != nil {
		return nil, err
	}
	if specList.Data == nil {
		return nil, nil
	}
	return *specList.Data, nil
}

func (p *RelytClient) CreateAccount(ctx context.Context, regionUri string, dwsuId string, account Account) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/account", dwsuId)
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "POST", &resp, account, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) PatchAccount(ctx context.Context, regionUri string, dwsuId string, userId, passwd string) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s", dwsuId, url.PathEscape(userId))
	patchPwd := map[string]any{"initPassword": passwd, "resetMfa": true}
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "PATCH", &resp, patchPwd, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetAccount(ctx context.Context, regionUri string, dwsuId string, userId string) (*CommonRelytResponse[Account], error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[Account]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) DropAccount(ctx context.Context, regionUri string, dwsuId string, userId string) error {
	path := fmt.Sprintf("/dwsu/%s/user/%s", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[string]{}
	handler := func(response *CommonRelytResponse[string], respString []byte) (*CommonRelytResponse[string], error) {
		if response.Code != CODE_SUCCESS && resp.Code != CODE_USER_NOT_FOUND {
			body := string(respString)
			tflog.Error(ctx, "error call api! resp code not success! body: "+body)
			return response, fmt.Errorf(body)
		}
		return nil, nil
	}
	err := doHttpRequest(p, ctx, regionUri, path, "DELETE", &resp, nil, nil, handler)
	return err
}

func (p *RelytClient) AsyncAccountConfig(ctx context.Context, regionUri, dwsuId, userId string, asyncResult AsyncResult) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/asyncresult", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "PUT", &resp, asyncResult, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) DeleteAsyncAccountConfig(ctx context.Context, regionUri, dwsuId, userId string) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/asyncresult", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "DELETE", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetAsyncAccountConfig(ctx context.Context, regionUri, dwsuId, userId string) (*AsyncResult, error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/asyncresult", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[AsyncResult]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) LakeFormationConfig(ctx context.Context, regionUri, dwsuId, userId string, formation LakeFormation) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/lakeformation", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "PUT", &resp, formation, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetLakeFormationConfig(ctx context.Context, regionUri, dwsuId, userId string) (*LakeFormation, error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/lakeformation", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[LakeFormation]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) DeleteLakeFormationConfig(ctx context.Context, regionUri, dwsuId, userId string) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/lakeformation", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "DELETE", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetBoto3AccessInfo(ctx context.Context, regionUri, dwsuId, userId string) ([]*Boto3AccessInfo, error) {
	path := fmt.Sprintf("/dwsu/%s/user/%s/ak", dwsuId, url.PathEscape(userId))
	resp := CommonRelytResponse[[]*Boto3AccessInfo]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return *resp.Data, nil
}

func (p *RelytClient) GetOpenApiMeta(ctx context.Context, cloud, region string) (*OpenApiMetaInfo, error) {
	path := fmt.Sprintf("/infra/%s/%s/endpoint", url.PathEscape(cloud), url.PathEscape(region))
	resp := CommonRelytResponse[[]*OpenApiMetaInfo]{}
	err := doHttpRequest(p, ctx, "", path, "GET", &resp, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return PickRegionOpenApiMeta(cloud, region, resp.Data)
}

// PickRegionOpenApiMeta selects the entry to use as the regional API host from
// what /infra/{cloud}/{region}/endpoint returned.
//
// The endpoint is not filtered server-side: it returns every endpoint registered
// for the region. Requiring exactly one entry therefore broke as soon as an
// operator registered a second type, and the resulting "length of api 2" said
// nothing about the cause. Select by type instead, preferring "openapi" and
// falling back to "web_console" to match common.PickOpenApiURIFromEndpoints.
// A single entry of any other type is still returned as-is — the pre-migration
// behavior — so a deployment whose lone endpoint carries an unexpected type
// keeps working.
//
// Pure function (no network), unit-testable.
func PickRegionOpenApiMeta(cloud, region string, metas *[]*OpenApiMetaInfo) (*OpenApiMetaInfo, error) {
	if metas == nil || len(*metas) == 0 {
		return nil, fmt.Errorf("no endpoint is registered for %s/%s. "+
			"ask your operator to register an 'openapi' endpoint for this region", cloud, region)
	}
	var fallback, single *OpenApiMetaInfo
	nonNil := 0
	types := make([]string, 0, len(*metas))
	for _, m := range *metas {
		if m == nil {
			continue
		}
		nonNil++
		single = m
		types = append(types, m.Type)
		if m.Type == "openapi" {
			return m, nil
		}
		if m.Type == "web_console" && fallback == nil {
			fallback = m
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	// Exactly one endpoint of an unrecognized type: return it. This is what the
	// pre-migration code did unconditionally, and deployments in the field (AWS)
	// may register their single endpoint under a type this code has never seen.
	if nonNil == 1 {
		return single, nil
	}
	return nil, fmt.Errorf("no 'openapi' or 'web_console' endpoint registered for %s/%s, got types: %v",
		cloud, region, types)
}

func (p *RelytClient) GetDwsuOpenApiMeta(ctx context.Context, dwsuId string) (*OpenApiMetaInfo, error) {
	dwsu, err := p.GetDwsu(ctx, dwsuId)
	if err != nil {
		return nil, err
	}
	if dwsu == nil {
		return nil, fmt.Errorf("can't find dwsu meta! %s", dwsuId)
	}
	meta, err := p.GetOpenApiMeta(ctx, dwsu.Region.Cloud.ID, dwsu.Region.ID)
	return meta, err
}

func (p *RelytClient) GetDwsuServiceAccount(ctx context.Context, regionUri, dwServiceUnitId string) ([]*ServiceAccount, error) {
	path := fmt.Sprintf("/dwsu/%s/service-accounts", dwServiceUnitId)
	resp := CommonRelytResponse[[]*ServiceAccount]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get dwsu:"+err.Error())
		return nil, err
	}
	return *resp.Data, nil
}

func (p *RelytClient) CreatePrivateLinkService(ctx context.Context, regionUri, dwServiceUnitId string, pl PrivateLinkService) (*PrivateLinkService, error) {
	path := fmt.Sprintf("/dwsu/%s/private-link-services", dwServiceUnitId)
	resp := CommonRelytResponse[PrivateLinkService]{}
	pl.ServiceName = ""
	pl.Status = ""
	header := map[string]string{"x-maxone-idempotent": "false"}
	err := doHttpRequestWithHeader(p, ctx, regionUri, path, "PUT", &resp, pl, nil, header, nil)
	if err != nil {
		tflog.Error(ctx, "Error create private-link:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) GetPrivateLinkService(ctx context.Context, regionUri, dwServiceUnitId, serviceType string) (*PrivateLinkService, error) {
	path := fmt.Sprintf("/dwsu/%s/private-link-services/%s", dwServiceUnitId, serviceType)
	resp := CommonRelytResponse[PrivateLinkService]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get private-link:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) PatchPrivateLinkService(ctx context.Context, regionUri, dwServiceUnitId, serviceType string, pl PrivateLinkService) (*CommonRelytResponse[PrivateLinkService], error) {
	path := fmt.Sprintf("/dwsu/%s/private-link-services/%s", dwServiceUnitId, serviceType)
	pl.ServiceType = ""
	pl.ServiceName = ""
	pl.Status = ""
	resp := CommonRelytResponse[PrivateLinkService]{}
	err := doHttpRequest(p, ctx, regionUri, path, "PATCH", &resp, pl, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error patch private-link:"+err.Error())
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) DeletePrivateLinkService(ctx context.Context, regionUri, dwServiceUnitId, serviceType string) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/private-link-services/%s", dwServiceUnitId, serviceType)
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "DELETE", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error delete private-link:"+err.Error())
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetIntegration(ctx context.Context, regionUri, dwServiceUnitId string) (*IntegrationInfo, error) {
	path := fmt.Sprintf("/dwsu/%s/integration", dwServiceUnitId)
	resp := CommonRelytResponse[IntegrationInfo]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get dwsu integration:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) PatchIntegration(ctx context.Context, regionUri, dwServiceUnitId string, info IntegrationInfo) (*CommonRelytResponse[string], error) {
	path := fmt.Sprintf("/dwsu/%s/integration", dwServiceUnitId)
	//这两个字段暂不支持更新
	info.RelytVpc = ""
	info.RelytPrincipal = ""
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "PATCH", &resp, info, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error patch dwsu integration:"+err.Error())
		return nil, err
	}
	return &resp, nil
}

func (p *RelytClient) GetRegionEndpoints(ctx context.Context, cloud, region string) (*[]RegionEndpoint, error) {
	path := fmt.Sprintf("/infra/%s/%s/endpoint", url.PathEscape(cloud), url.PathEscape(region))
	resp := CommonRelytResponse[[]RegionEndpoint]{}
	err := doHttpRequest(p, ctx, "", path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error list region endpoints:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) GetUserSecurityPolicy(ctx context.Context, regionUri, dwsuId string) (*UserSecurityPolicy, error) {
	path := fmt.Sprintf("/dwsu/%s/user-security-policy", url.PathEscape(dwsuId))
	resp := CommonRelytResponse[UserSecurityPolicy]{}
	err := doHttpRequest(p, ctx, regionUri, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get user-security-policy:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) PatchUserSecurityPolicy(ctx context.Context, regionUri, dwsuId string, userSecPolicy UserSecurityPolicy) (*string, error) {
	path := fmt.Sprintf("/dwsu/%s/user-security-policy", url.PathEscape(dwsuId))
	resp := CommonRelytResponse[string]{}
	err := doHttpRequest(p, ctx, regionUri, path, "PATCH", &resp, userSecPolicy, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get user-security-policy:"+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) GetEntraIdConfig(ctx context.Context, dmsHost, dwsuId string) (*EntraIdConfig, error) {
	path := "/api/entraid-config"
	resp := CommonRelytResponse[EntraIdConfig]{}
	// Backend returns code:200, data:null when the config is absent — no special
	// not-found code needed. resp.Data == nil signals "not configured" to callers.
	err := doHttpRequest(p, ctx, dmsHost, path, "GET", &resp, nil, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error get entraid-config: "+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) PutEntraIdConfig(ctx context.Context, dmsHost, dwsuId string, cfg EntraIdConfig) (*EntraIdConfig, error) {
	path := "/api/entraid-config"
	resp := CommonRelytResponse[EntraIdConfig]{}
	err := doHttpRequest(p, ctx, dmsHost, path, "PUT", &resp, cfg, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Error put entraid-config: "+err.Error())
		return nil, err
	}
	return resp.Data, nil
}

func (p *RelytClient) DeleteEntraIdConfig(ctx context.Context, dmsHost, dwsuId string) error {
	path := "/api/entraid-config"
	resp := CommonRelytResponse[string]{}
	// Backend returns code:200 success even when the config doesn't exist — DELETE
	// is already idempotent on the server side, no special handler needed.
	err := doHttpRequest(p, ctx, dmsHost, path, "DELETE", &resp, nil, nil, nil)
	if err != nil {
		tflog.Info(ctx, "delete entraid-config err: "+err.Error())
	}
	return err
}
