package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-relyt/internal/provider/client"
	"terraform-provider-relyt/internal/provider/common"
	"terraform-provider-relyt/internal/provider/model"
)

func updateDps(ctx context.Context, relytClient *client.RelytClient, state, plan *model.DpsModel, diagnostics diag.Diagnostics, dwsuId, dpsId string) bool {
	meta := common.RouteRegionUri(ctx, dwsuId, relytClient, &diagnostics)
	if diagnostics.HasError() {
		return true
	}
	regionUri := meta.URI
	patchDps := client.DpsMode{
		Description: plan.Description.ValueString(),
		Engine:      plan.Engine.ValueString(),
		Name:        plan.Name.ValueString(),
		Spec:        &client.Spec{Name: plan.Size.ValueString()},
	}
	_, err := relytClient.PatchDps(ctx, regionUri, dwsuId, dpsId, patchDps)
	if err != nil {
		tflog.Error(ctx, "error update dps"+err.Error())
		diagnostics.AddError("update dps failed!", "error update dps!"+err.Error())
		return true
	}
	_, err = WaitDpsReady(ctx, relytClient, regionUri, dwsuId, dpsId, diagnostics)
	if err != nil {
		tflog.Error(ctx, "error wait dps ready after update"+err.Error())
		diagnostics.AddError("update dps failed!", "error wait dps ready after update!"+err.Error())
		return true
	}
	state.Size = plan.Size
	return false
}

func WaitDpsReady(ctx context.Context, relytClient *client.RelytClient, regionUri string, dwsuId, dpsId string, diagnostics diag.Diagnostics) (any, error) {
	queryDpsMode, err := common.TimeOutTask(relytClient.CheckTimeOut, relytClient.CheckInterval, func() (any, error) {
		dps, err2 := relytClient.GetDps(ctx, regionUri, dwsuId, dpsId)
		if err2 != nil {
			//这里判断是否要充实
			return dps, err2
		}
		if dps != nil && dps.Status == client.DPS_STATUS_READY {
			return dps, nil
		}
		return dps, fmt.Errorf("dps is not Ready")
	})
	if err != nil {
		tflog.Error(ctx, "error wait dps ready"+err.Error())
		diagnostics.AddError(
			"create dps failed!", "error wait dps ready! "+err.Error(),
		)
		return nil, err
		//fmt.Println(fmt.Sprintf("drop dwsu%s", err.Error()))
	}
	return queryDpsMode, err
}
