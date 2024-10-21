package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func doHttpRequest[T any](p *RelytClient, ctx context.Context, host, path, method string,
	respMode *CommonRelytResponse[T],
	request any,
	parameter map[string]string,
	codeHandler func(response *CommonRelytResponse[T], respDumpByte []byte) (*CommonRelytResponse[T], error)) (err error) {
	return doHttpRequestWithHeader(p, ctx, host, path, method, respMode, request, parameter, nil, codeHandler)
}

func doHttpRequestWithHeader[T any](p *RelytClient, ctx context.Context, host, path, method string,
	respMode *CommonRelytResponse[T],
	request any,
	parameter map[string]string,
	header map[string]string,
	codeHandler func(response *CommonRelytResponse[T], respDumpByte []byte) (*CommonRelytResponse[T], error)) (err error) {
	return signedHttpRequestWithHeader(p, ctx, host, path, method,
		respMode, request, parameter, header, nil, codeHandler)
	//if host == "" {
	//	host = p.ApiHost
	//}
	//var jsonData = []byte("")
	//if request != nil && "" != request {
	//	requestJson, err := json.Marshal(request)
	//	if err != nil {
	//		tflog.Error(ctx, "fmt request json error:"+err.Error())
	//	}
	//	tflog.Info(ctx, "request data :"+string(requestJson))
	//	jsonData = requestJson // POST请求发送的数据
	//}
	//hostApi := host + path
	//parsedHostApi, err := url.Parse(hostApi)
	//if err != nil {
	//	return err
	//}
	//queryParams := url.Values{}
	//if parameter != nil {
	//	for k, v := range parameter {
	//		queryParams.Add(k, v)
	//	}
	//}
	//parsedHostApi.RawQuery = queryParams.Encode()
	//
	//req, err := http.NewRequest(method, parsedHostApi.String(), bytes.NewBuffer(jsonData))
	//if err != nil {
	//	tflog.Error(ctx, "Error creating request:"+err.Error())
	//	return err
	//}
	//req.Header.Set("x-maxone-api-key", p.AuthKey)
	//req.Header.Set("x-maxone-role-id", p.Role)
	//req.Header.Set("Content-Type", "application/json")
	//if header != nil {
	//	for k, v := range header {
	//		req.Header.Set(k, v)
	//	}
	//}
	//requestString, _ := httputil.DumpRequestOut(req, true)
	//tflog.Info(ctx, "== request: "+string(requestString))
	//client := &http.Client{Timeout: 10 * time.Second}
	//resp, err := client.Do(req)
	//if err != nil {
	//	tflog.Error(ctx, "Error sending request:"+err.Error())
	//	return err
	//}
	//defer resp.Body.Close()
	//responseString, _ := httputil.DumpResponse(resp, true)
	//tflog.Info(ctx, "== response: "+string(responseString))
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	tflog.Error(ctx, "Error reading responseString body:"+err.Error())
	//	return err
	//}
	//if resp.StatusCode != CODE_SUCCESS {
	//	tflog.Error(ctx, "Error status http code not 200! "+resp.Status)
	//	//printResp(ctx, resp)
	//	return fmt.Errorf("Error http code not 200! respCode: %s!\n%s ", resp.Status, string(body))
	//}
	//
	//err = json.Unmarshal(body, respMode)
	//if err != nil {
	//	tflog.Error(ctx, "read json respFail:"+err.Error())
	//	return err
	//}
	//if respMode.Code != CODE_SUCCESS {
	//	tflog.Warn(ctx, "error call api! resp code not 200: "+string(body))
	//}
	//if codeHandler != nil {
	//	tflog.Trace(ctx, "use code handle func!")
	//	handler, err := codeHandler(respMode, body)
	//	if handler != nil {
	//		respMode.Code = handler.Code
	//		respMode.Data = handler.Data
	//		respMode.Msg = handler.Msg
	//	}
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	if respMode.Code != CODE_SUCCESS {
	//		tflog.Error(ctx, "error call api! resp code not 200: "+string(body))
	//		return fmt.Errorf(string(body))
	//	}
	//}
	//return nil
}

func signedHttpRequestWithHeader[T any](p *RelytClient, ctx context.Context, host, path,
	method string,
	respMode *CommonRelytResponse[T],
	request any,
	parameter map[string]string,
	header map[string]string,
	databaseClientConfig *RelytDatabaseClientConfig,
	codeHandler func(response *CommonRelytResponse[T], respDumpByte []byte) (*CommonRelytResponse[T], error)) (err error) {
	if host == "" {
		host = p.ApiHost
	}
	jsonBody := false
	var jsonData = []byte("")
	if request != nil && "" != request {
		jsonBody = true
		requestJson, err := json.Marshal(request)
		if err != nil {
			tflog.Error(ctx, "fmt request json error:"+err.Error())
		}
		tflog.Info(ctx, "request data :"+string(requestJson))
		jsonData = requestJson // POST请求发送的数据
	}
	hostApi := host + path
	parsedHostApi, err := url.Parse(hostApi)
	if err != nil {
		return err
	}
	queryParams := url.Values{}
	if parameter != nil {
		for k, v := range parameter {
			queryParams.Add(k, v)
		}
	}
	parsedHostApi.RawQuery = queryParams.Encode()
	//parsedHostApi.Opaque = host
	//parsedHostApi.RawQuery

	req, err := http.NewRequest(method, parsedHostApi.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		tflog.Error(ctx, "Error creating request:"+err.Error())
		return err
	}
	if p != nil {
		req.Header.Set("x-maxone-api-key", p.AuthKey)
		req.Header.Set("x-maxone-role-id", p.Role)
	}
	if jsonBody {
		req.Header.Set("Content-Type", "application/json")
	}
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	clientTimeout := 10 * time.Second
	if databaseClientConfig != nil {
		err = AwsSignHttp(databaseClientConfig, req, jsonData)
		if err != nil {
			tflog.Error(ctx, "error sign request"+err.Error())
			return err
		}
		if databaseClientConfig.ClientTimeout > 0 {
			clientTimeout = time.Duration(databaseClientConfig.ClientTimeout) * time.Second
		}
	}
	if p != nil {
		clientTimeout = time.Duration(p.ClientTimeout) * time.Second
	}

	requestId := ""
	requestUUID, uuidErr := uuid.NewUUID()
	if uuidErr == nil {
		requestId = requestUUID.String()
	}
	requestString, _ := httputil.DumpRequestOut(req, true)
	tflog.Info(ctx, "== apiId : "+requestId+" request: "+string(requestString))
	client := &http.Client{Timeout: clientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		tflog.Error(ctx, "Error sending request:"+err.Error())
		return err
	}
	defer resp.Body.Close()
	responseString, _ := httputil.DumpResponse(resp, true)
	tflog.Info(ctx, "== apiId : "+requestId+" response: "+string(responseString))
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tflog.Error(ctx, "Error reading responseString body:"+err.Error())
		return err
	}
	if resp.StatusCode != CODE_SUCCESS {
		tflog.Error(ctx, "Error status http code not 200! "+resp.Status)
		//printResp(ctx, resp)
		return fmt.Errorf("Error http code not 200! respCode: %s!\n%s ", resp.Status, string(body))
	}

	err = json.Unmarshal(body, respMode)
	if err != nil {
		tflog.Error(ctx, "read json respFail:"+err.Error())
		return err
	}
	if respMode.Code != CODE_SUCCESS {
		tflog.Warn(ctx, "error call api! resp code not 200: "+string(body))
	}
	if codeHandler != nil {
		tflog.Trace(ctx, "use code handle func!")
		handler, err := codeHandler(respMode, body)
		if handler != nil {
			respMode.Code = handler.Code
			respMode.Data = handler.Data
			respMode.Msg = handler.Msg
		}
		if err != nil {
			return err
		}
	} else {
		if respMode.Code != CODE_SUCCESS {
			tflog.Error(ctx, "error call api! resp code not 200: "+string(body))
			return fmt.Errorf(string(body))
		}
	}
	return nil
}

func AwsSignHttp(aksk *RelytDatabaseClientConfig, req *http.Request, body []byte) error {
	credentials := aws.Credentials{
		AccessKeyID:     aksk.AccessKey,
		SecretAccessKey: aksk.SecretKey,
		SessionToken:    "",
		Source:          "",
		CanExpire:       false,
		Expires:         time.Time{},
		AccountID:       "",
	}
	signer := v4.NewSigner()
	hash := sha256.Sum256(body)
	payloadHash := hex.EncodeToString(hash[:])
	err := signer.SignHTTP(context.TODO(), credentials, req, payloadHash, "relyt", "default", time.Now())
	if err != nil {
		return err
	}
	return nil
}
