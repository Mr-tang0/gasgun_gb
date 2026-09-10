package main

import (
	"context"
	"fmt"
	"gasgun_gb/backend"
	xinjiegasgun1 "gasgun_gb/services/xinjie_gasgun1"
	xinjiegasgun2 "gasgun_gb/services/xinjie_gasgun2"
	hepsgasgun1 "gasgun_gb/services/xinjie_heps_gasgun1"
)

type App struct {
	ctx             context.Context
	xinjieGasgun1   *xinjiegasgun1.XinjieGasGun1
	xinjieGasgun2   *xinjiegasgun2.XinjieGasGun2
	hepsGasgun1     *hepsgasgun1.HEPSGasGun1
	normalHopkinson *backend.NormalHopkinsonContoller

	updater *backend.UpdateService
}

func NewApp(xinjieGasgun1 *xinjiegasgun1.XinjieGasGun1,
	xinjieGasgun2 *xinjiegasgun2.XinjieGasGun2,
	HEPSGasgun1 *hepsgasgun1.HEPSGasGun1) *App {
	app := &App{
		xinjieGasgun1: xinjieGasgun1,
		xinjieGasgun2: xinjieGasgun2,
		hepsGasgun1:   HEPSGasgun1,
	}
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.xinjieGasgun1.Startup(ctx)
	a.xinjieGasgun2.Startup(ctx)
	a.hepsGasgun1.Startup(ctx)
	//创建更新服务
	a.updater = &backend.UpdateService{}
}

func (a *App) APIUpdate() backend.GitHubRelease {
	//获取更新信息
	release, err := a.updater.GetUpdateInfo()
	if err != nil {
		fmt.Printf("获取更新信息失败: %v\n", err)
		return backend.GitHubRelease{}
	}
	fmt.Printf("更新信息: %v\n", release)
	return release
}

func (a *App) GetCachedRelease() backend.GitHubRelease {
	return a.updater.GetCachedRelease()
}
