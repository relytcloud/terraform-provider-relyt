package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"math"
	"strconv"
	"terraform-provider-relyt/internal/provider/client"
	"time"
)

func RouteRegionUri(ctx context.Context, dwsuId string, relytClient *client.RelytClient,
	diag *diag.Diagnostics) *client.OpenApiMetaInfo {
	meta, err := CommonRetry[client.OpenApiMetaInfo](ctx,
		func() (*client.OpenApiMetaInfo, error) {
			return relytClient.GetDwsuOpenApiMeta(ctx, dwsuId)
		})
	if err != nil || meta == nil {
		errMsg := "get RegionApi is nil"
		if err != nil {
			errMsg = err.Error()
		}
		diag.AddError("error get region api", "fail to get Region uri address dwsuID:"+
			""+dwsuId+" error: "+errMsg)
		return meta
	}
	return meta
}

func RetryFunction[T any](ctx context.Context, retryNum, intervalSecond int,
	backoffCoefficient float64,
	retryableFunc func() (*T, error)) (*T, error) {
	var err error
	var result *T
	for i := 0; i < retryNum; i++ {
		result, err = retryableFunc()
		if err == nil {
			return result, nil
		}
		time.Sleep(time.Duration(intervalSecond))
		intervalSecond = int(math.Ceil(backoffCoefficient * float64(intervalSecond)))
		tflog.Warn(ctx, "retry failed func! backoff second:"+strconv.Itoa(intervalSecond)+"error msg!"+err.Error())
	}

	return result, err
}

func CommonRetry[T any](ctx context.Context, retryableFunc func() (*T, error)) (*T, error) {
	return RetryFunction(ctx, 5, 1, 1.0, retryableFunc)
}

func TimeOutTask(timeoutSec int64, checkIntervalSec int32, task func() (any, error)) (any, error) {
	// 设置超时时间
	timeout := time.Duration(timeoutSec) * time.Second
	interval := time.Duration(checkIntervalSec) * time.Second

	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//// 启动任务
	//done := make(chan bool)
	//f := func() (any, error) {
	//
	//}
	//go f
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Task timed out")
			//done <- false
			return nil, fmt.Errorf("timeout")
		default:
			a, err := task()
			if err == nil {
				//done <- true
				return a, err
			}
			time.Sleep(interval)
		}
	}
}
